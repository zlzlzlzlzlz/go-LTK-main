package localclient

import (
	"goltk/app"
	"goltk/data"
	"goltk/front"
	"goltk/sound"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type cardReceive struct {
	id    data.PID
	cards []data.CID
}

type cardMoveRec struct {
	src, dst data.PID
	cards    []data.CID
}

type dropAbleRec struct {
	cards   []data.CID
	dropNum uint8
}

type gameStateInf struct {
	state     data.GameState
	t         time.Duration
	curPlayer data.PID
}

type usecardInf struct {
	user    data.PID
	card    data.CID
	targets []data.PID
}

type useSkillInf struct {
	user   data.PID
	skill  data.SID
	target []data.PID
	args   []byte
}

type useTmpCardInf struct {
	user    data.PID
	cname   data.CardName
	dec     data.Decor
	num     data.CNum
	tmpType data.TmpCardType
	targets []data.PID
}

type setHpInf struct {
	pid    data.PID
	hp     data.HP
	hptype data.SetHpType
}

type gsidCardInf struct {
	id    data.GSID
	cards []data.CID
}

var skillTitleImg = map[string]*ebiten.Image{}

func getSkillTitleImg(name string) *ebiten.Image {
	if img, ok := skillTitleImg[name]; ok {
		return img
	}
	img := loadImg("assets/game/skillTitle/" + name + ".png")
	skillTitleImg[name] = img
	return img
}

var skillNameImgMap = map[string]*ebiten.Image{}

func getSkillNameImg(name string) *ebiten.Image {
	if img, ok := skillNameImgMap[name]; ok {
		return img
	}
	img := loadImg("assets/player/skillName/" + name + ".png")
	skillNameImgMap[name] = img
	return img
}

type localStateImg struct {
	x, y       float64
	visibility float32
	img        *ebiten.Image
	moveInf    struct {
		counter int
		dx, dy  float64
	}
	visibilityInf struct {
		counter int
		dt      float32
	}
}

func (l *localStateImg) Draw(screen *ebiten.Image) {
	if l.visibility <= 0 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(l.x, l.y)
	op.ColorScale.ScaleAlpha(l.visibility)
	screen.DrawImage(l.img, op)
}

func (l *localStateImg) anime() {
	if l.moveInf.counter > 0 {
		l.setPos(l.x+l.moveInf.dx, l.y+l.moveInf.dy)
		l.moveInf.counter--
	}
	if l.visibilityInf.counter > 0 {
		l.visibility -= l.visibilityInf.dt
		l.visibilityInf.counter--
	} else if l.visibilityInf.counter == 0 {
		l.visibility = 0
		l.visibilityInf.counter = -1
	}
}

func (l *localStateImg) showUp() {
	l.setPos(-50, 500)
	l.visibility = 1
	l.visibilityInf.counter = -1
	l.move2Pos(10, 500)
}

func (l *localStateImg) vanish() {
	l.move2Pos(200, 500)
	l.setFateOut(60)
}

func (l *localStateImg) setPos(x, y float64) {
	l.x, l.y = x, y
}

func (l *localStateImg) move2Pos(x, y float64) {
	const speed = 2 //每tick移动像素
	dx, dy := x-l.x, y-l.y
	distence := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2)) //起点到终点的距离
	if distence == 0 {
		return
	}
	l.moveInf.counter = int(distence / speed)
	l.moveInf.dx = dx * speed / distence
	l.moveInf.dy = dy * speed / distence
	remain := int(distence) % speed
	l.setPos(l.x+dx*float64(remain)/distence, l.y+dy*float64(remain)/distence)
}

func (l *localStateImg) setFateOut(t int) {
	l.visibilityInf.counter = t
	l.visibilityInf.dt = 1.0 / float32(t)
}

var gameImg = struct {
	selbg             *ebiten.Image
	bg                *ebiten.Image //背景
	processBar        *ebiten.Image // 进度条图像
	processBg         *ebiten.Image // 进度条底图像
	processBarOther   *ebiten.Image
	processBarotherBg *ebiten.Image
	skillBg           *ebiten.Image
	skillBgBig        *ebiten.Image
	skillBgLage       *ebiten.Image
	skillCardBg       *ebiten.Image
	skillNameBg       *ebiten.Image
	roleDetailBg      *ebiten.Image
	winImg            *ebiten.Image
	loseImg           *ebiten.Image
}{
	selbg:             loadImg("assets/game/select/selbg.png"),
	bg:                loadImg("assets/bg.jpg"),
	processBar:        loadImg("assets/item/timer.png"),
	processBg:         loadImg("assets/item/timerbg.png"),
	skillBg:           loadImg("assets/game/gameSkillBg.png"),
	skillBgBig:        loadImg("assets/game/gameSkillBgBig.png"),
	skillBgLage:       loadImg("assets/game/gameSkillLarge.png"),
	skillCardBg:       loadImg("assets/game/gameSkillCardBg.png"),
	skillNameBg:       loadImg("assets/game/gameSkillNameBg.png"),
	processBarOther:   loadImg("assets/item/timerOther.png"),
	processBarotherBg: loadImg("assets/item/timerotherBg.png"),
	roleDetailBg:      loadImg("assets/menu/selectBg.png"),
	winImg:            loadImg("assets/game/win.png"),
	loseImg:           loadImg("assets/game/lose.png"),
}

type Games struct {
	state              data.GameState
	app                *app.App
	pid                data.PID //该客户端主视角玩家
	curPlayer          data.PID //当前玩家
	turnOwner          data.PID //回合拥有者
	hasSkip            bool     //是否在当前阶段按下过结束按钮
	cards              []cardI
	handleZone         []cardI
	useCardInf         chan data.UseCardInf
	dropCardInf        chan data.DropCardInf
	useSkillInf        chan data.UseSkillInf
	cardReceiver       chan cardReceive             //接收发牌信息的chan
	removeReceiver     chan cardReceive             //接收弃牌信息的chan
	moveCardRec        chan cardMoveRec             //接收移动卡牌信息的chan
	useReceiver        chan usecardInf              //接受用牌信息的chan
	useTmpRec          chan useTmpCardInf           //接收使用临时卡信息的chan
	availableTargetrec chan data.AvailableTargetInf //接收可用目标信息的chan
	gsidCardReceiver   chan gsidCardInf             //游戏技能的卡牌接收器
	pidReceiver        chan data.PID                //接收主视角pid的chan
	playerInfRec       chan []data.PlayerInf        //接受玩家列表信息的chan
	useAbleReceiver    chan []data.CID              //可用的卡接收器
	useAbleSkillRec    chan []data.SID              //可用的主动技接收器
	useSkillRspRec     chan data.UseSkillRsp        //用主动技能的回应的接收器
	dropAbleRec        chan dropAbleRec             //可丢弃的卡的接收器
	turnOwnerRec       chan data.PID                //当前回合拥有者接收器
	skillSelectRec     chan data.SID                //问玩家要不要用技能的chan
	useSkillRec        chan useSkillInf             //接收使用技能信息的chan
	targetQuest        chan data.CID                //向服务端询问卡片可用目标
	availableRoleRec   chan []data.Role             //接收可选角色列表
	roleInf            chan data.Role               //向服务端发送选择的角色
	setHpReceiver      chan setHpInf
	gameStateRec       chan gameStateInf
	closeSignal        chan struct{}
	quitBtn            buttonI
	playList           []*player
	roleList           []data.Role
	selRole            *data.Role
	btnList            []buttonI
	dropAbleInf        dropAbleRec
	promptText         *front.TextItem2 //提示文本
	promptTextOther    *front.TextItem2
	processInf         struct {
		rate   float64 //进度条比例
		total  float64 //进度条共持续多少gt
		remain float64 //当前剩余多少gt
	}
	bgm              *audio.Player
	useAbleCards     []data.CID
	wgfdCards        []cardI
	ssqyCards        []cardI
	ghcqCards        []cardI
	qlgCards         []cardI
	judgeCard        cardI   //要判定的卡
	judgeResult      cardI   //判定结果
	otherCards       []cardI //别人的卡
	selNum           int     //要丢弃的卡牌数量
	selcard          []data.CID
	skillTitle       *ebiten.Image
	selectSkill      data.SID //要选择的技能
	skillTargetInf   data.AvailableTargetInf
	skilltargetList  []data.PID //已选的技能目标列表
	judgeSkill       data.SID
	skillJudgeResult cardI
	heapTop          []cardI
	heapButtom       []cardI
	localStateImgs   [6]localStateImg
	littleHeaptext   *front.TextItem2
	littleHeapNum    uint8
	roledetail       *front.TextItem2
	showRoleDetail   bool
	roleDetailNum    data.PID
	settleAnime      struct {
		visibility float32
		img        *ebiten.Image
	}
	wxkjInf struct {
		count  uint8 //偶数为锦囊生效
		cName  data.CardName
		target data.PID
		user   data.PID
	}
}

func NewGames(app *app.App) *Games {
	g := Games{
		app:                app,
		pid:                -1,
		useCardInf:         make(chan data.UseCardInf, 1),
		dropCardInf:        make(chan data.DropCardInf, 1),
		useSkillInf:        make(chan data.UseSkillInf, 1),
		cardReceiver:       make(chan cardReceive, 1),
		removeReceiver:     make(chan cardReceive, 1),
		moveCardRec:        make(chan cardMoveRec, 1),
		useReceiver:        make(chan usecardInf, 1),
		useTmpRec:          make(chan useTmpCardInf, 1),
		availableTargetrec: make(chan data.AvailableTargetInf, 1),
		pidReceiver:        make(chan data.PID, 1),
		playerInfRec:       make(chan []data.PlayerInf, 1),
		useAbleReceiver:    make(chan []data.CID, 1),
		useAbleSkillRec:    make(chan []data.SID, 1),
		dropAbleRec:        make(chan dropAbleRec, 1),
		turnOwnerRec:       make(chan data.PID, 1),
		skillSelectRec:     make(chan data.SID, 1),
		useSkillRec:        make(chan useSkillInf, 4),
		useSkillRspRec:     make(chan data.UseSkillRsp, 1),
		gsidCardReceiver:   make(chan gsidCardInf, 1),
		targetQuest:        make(chan data.CID, 1),
		availableRoleRec:   make(chan []data.Role, 1),
		roleInf:            make(chan data.Role, 1),
		setHpReceiver:      make(chan setHpInf, 1),
		gameStateRec:       make(chan gameStateInf, 1),
		closeSignal:        make(chan struct{}),
		promptText:         front.NewTextItem2("", 0, 0, 24, 2.5, 24),
		promptTextOther:    front.NewTextItem2("", 0, 0, 18, 0, 18),
		quitBtn:            newQuitGameBtn(),
		littleHeaptext:     front.NewTextItem2("160", 1205, 35, 30, 4, 0),
		littleHeapNum:      160,
		roledetail:         front.NewTextItem2("", 530, 160, 26, 2, 24),
	}
	for i := 0; i < len(g.localStateImgs); i++ {
		g.localStateImgs[i].img = getImg("assets/player/localState/" + strconv.Itoa(i) + ".png")
	}
	return &g
}

func (g *Games) GetUseCardInf() <-chan data.UseCardInf {
	return g.useCardInf
}

func (g *Games) GetDropCardInf() <-chan data.DropCardInf {
	return g.dropCardInf
}

func (g *Games) GetUseSkillInf() <-chan data.UseSkillInf {
	return g.useSkillInf
}

func (g *Games) SetPid(id data.PID) {
	g.pidReceiver <- id
}

func (g *Games) SendAvailableRole(roles ...data.Role) {
	g.availableRoleRec <- roles
}

func (g *Games) GetRole() <-chan data.Role {
	return g.roleInf
}

func (g *Games) SendCard(id data.PID, cards ...data.CID) {
	g.cardReceiver <- cardReceive{id: id, cards: cards}
}

func (g *Games) RemoveCard(id data.PID, cards ...data.CID) {
	g.removeReceiver <- cardReceive{id: id, cards: cards}
}

func (g *Games) MoveCard(src, dst data.PID, cards ...data.CID) {
	g.moveCardRec <- cardMoveRec{src: src, dst: dst, cards: cards}
}

func (g *Games) UseCard(user data.PID, c data.CID, targets ...data.PID) {
	g.useReceiver <- usecardInf{user: user, card: c, targets: targets}
}

func (g *Games) UseTmpCard(user data.PID, name data.CardName, dec data.Decor, num data.CNum,
	tmpType data.TmpCardType, target ...data.PID) {
	g.useTmpRec <- useTmpCardInf{user: user, cname: name, dec: dec, num: num, tmpType: tmpType, targets: target}
}

func (g *Games) SendUseAbleCards(cards []data.CID) {
	g.useAbleReceiver <- cards
}

func (g *Games) SendUseAbleSkill(skills []data.SID) {
	g.useAbleSkillRec <- skills
}

func (g *Games) SendDropAbleCard(cards []data.CID, dropNum uint8) {
	g.dropAbleRec <- dropAbleRec{cards: cards, dropNum: dropNum}
}

func (g *Games) SendSkillSelect(sid data.SID) {
	g.skillSelectRec <- sid
}

func (g *Games) SendUseSkillRsp(rsp data.UseSkillRsp) {
	g.useSkillRspRec <- rsp
}

func (g *Games) UseSkill(user data.PID, skill data.SID, target []data.PID, args ...byte) {
	g.useSkillRec <- useSkillInf{user: user, skill: skill, target: target, args: args}
}

func (g *Games) SendAvailableTarget(inf data.AvailableTargetInf) {
	g.availableTargetrec <- inf
}

func (g *Games) SendPlayerInf(inf []data.PlayerInf) {
	g.playerInfRec <- inf
}

func (g *Games) SendGSCards(id data.GSID, cards ...data.CID) {
	g.gsidCardReceiver <- gsidCardInf{id: id, cards: cards}
}

func (g *Games) SetHP(pid data.PID, hp data.HP, dmgtype data.SetHpType) {
	g.setHpReceiver <- setHpInf{pid: pid, hp: hp, hptype: dmgtype}
}

func (g *Games) SetGameState(state data.GameState, t time.Duration, curPlayer data.PID) {
	g.gameStateRec <- gameStateInf{state: state, t: t, curPlayer: curPlayer}
}

func (g *Games) SetTurnOwner(pid data.PID) {
	g.turnOwnerRec <- pid
}

func (g *Games) GetTargetQuest() <-chan data.CID {
	return g.targetQuest
}

func (g *Games) GetClientType() data.PlayerType {
	return data.LocalPlayer
}

func (g *Games) Close() {
	close(g.closeSignal)
}

func (g *Games) Draw(screen *ebiten.Image) {
	//绘制背景
	screen.DrawImage(gameImg.bg, nil)
	// 绘制小牌堆
	g.drawLittleHeap(screen)
	//绘制退出按钮
	g.quitBtn.Draw(screen)
	//画自己阶段小图像
	for i := 0; i < len(g.localStateImgs); i++ {
		g.localStateImgs[i].Draw(screen)
	}
	//绘制处理区
	for i := 0; i < len(g.handleZone); i++ {
		g.handleZone[i].Draw(screen)
	}
	//绘制玩家
	for i := len(g.playList) - 1; i >= 0; i-- {
		g.playList[i].Draw(screen)
	}
	//画阶段提示
	g.promptText.Draw(screen)
	g.promptTextOther.Draw(screen)
	//如果五谷丰登卡槽不为空则绘制
	if len(g.wgfdCards) != 0 {
		g.drawWGFD(screen)
	}
	if g.curPlayer == g.pid {
		switch g.state {
		case data.SSQYState:
			g.drawSSQY(screen)
		case data.GHCQState:
			g.drawGHCQ(screen)
		case data.DropOtherCardState:
			g.drawDropOtherCard(screen)
		case data.QLGState:
			g.drawQLG(screen)
		case data.PoJunState:
			g.drawDropOtherCard(screen)
		case data.FanKuiState:
			g.drawDropOtherCard(screen)
		case data.GuanXingState:
			g.drawGuanXing(screen)
		}
		if len(g.btnList) != 0 {
			for i := 0; i < len(g.btnList); i++ {
				g.btnList[i].Draw(screen)
			}
		}
	}
	switch g.state {
	case data.InitState:
		g.drawInit(screen)
	case data.WinState:
		g.drawSettle(screen)
	case data.LoseState:
		g.drawSettle(screen)
	}
	//判定
	if g.judgeCard != nil {
		g.drawMakeJudge(screen)
	}
	//如果为技能判定则绘制
	if g.state == data.SkillJudgeState {
		g.drawSkillJudge(screen)
	}
	if g.state == data.InitState {
		for i := 0; i < len(g.btnList); i++ {
			g.btnList[i].Draw(screen)
		}
	}
	if g.showRoleDetail {
		p := g.playList[g.roleDetailNum]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1.5, 1.2)
		op.GeoM.Translate(50, 120)
		op.ColorScale.ScaleAlpha(0.8)
		screen.DrawImage(gameImg.roleDetailBg, op)
		op.GeoM.Reset()
		op.GeoM.Scale(1.4, 1.4)
		op.GeoM.Translate(160, 150)
		screen.DrawImage(getImg("assets/role/faction/"+strconv.Itoa(int(p.side))+".png"), op)
		op.GeoM.Reset()
		op.GeoM.Scale(1.4, 1.4)
		img := p.img
		op.GeoM.Translate(204, 369-float64(img.Bounds().Dy()))
		screen.DrawImage(img, op)
		g.roledetail.Draw(screen)
		front.DrawText(screen, p.dspName, 188, 194, 28, true)
	}
	//debug 绘制当前阶段
	//front.DrawText(screen, g.state.String(), 20, 500, 32, false)
}

func (g *Games) drawLittleHeap(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(1200, 10)
	screen.DrawImage(getImg("assets/item/littleheap.png"), op)
	g.littleHeaptext.Draw(screen)
}

func (g *Games) drawInit(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(52, 16)
	screen.DrawImage(getImg("assets/menu/dragon.png"), op)
	op.GeoM.Reset()
	op.GeoM.Translate(360, 80)
	screen.DrawImage(gameImg.roleDetailBg, op)
	op.GeoM.Reset()
	op.GeoM.Translate(72, 14)
	screen.DrawImage(getImg("assets/menu/roleBg.png"), op)
	op.GeoM.Reset()
	op.GeoM.Translate(106, 568)
	screen.DrawImage(getImg("assets/menu/bottomBg.png"), op)
}

func (g *Games) drawSettle(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(300, 200)
	op.ColorScale.ScaleAlpha(g.settleAnime.visibility)
	screen.DrawImage(g.settleAnime.img, op)
}

func (g *Games) drawWGFD(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 120)
	screen.DrawImage(gameImg.skillBg, op)
	title := getSkillTitleImg("wgfd")
	op.GeoM.Translate(float64(667-title.Bounds().Dx()/2), 10)
	screen.DrawImage(title, op)
	for i := 0; i < len(g.wgfdCards); i++ {
		g.wgfdCards[i].Draw(screen)
	}
}

func (g *Games) drawSSQY(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 120)
	screen.DrawImage(gameImg.skillBgBig, op)
	title := getSkillTitleImg("ssqy")
	op.GeoM.Translate(float64(667-title.Bounds().Dx()/2), 10)
	screen.DrawImage(title, op)
	for i := 0; i < len(g.ssqyCards); i++ {
		g.ssqyCards[i].Draw(screen)
	}
}

func (g *Games) drawGHCQ(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 120)
	screen.DrawImage(gameImg.skillBgBig, op)
	title := getSkillTitleImg("ghcq")
	op.GeoM.Translate(float64(667-title.Bounds().Dx()/2), 10)
	screen.DrawImage(title, op)
	for i := 0; i < len(g.ghcqCards); i++ {
		g.ghcqCards[i].Draw(screen)
	}
}

func (g *Games) drawDropOtherCard(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 120)
	screen.DrawImage(gameImg.skillBgBig, op)
	if g.skillTitle != nil {
		op.GeoM.Translate(float64(667-g.skillTitle.Bounds().Dx()/2), 10)
		screen.DrawImage(g.skillTitle, op)
	}
	for i := 0; i < len(g.otherCards); i++ {
		g.otherCards[i].Draw(screen)
	}
}

func (g *Games) drawQLG(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 120)
	screen.DrawImage(gameImg.skillBg, op)
	title := getSkillTitleImg("qlg")
	op.GeoM.Translate(float64(667-title.Bounds().Dx()/2), 10)
	screen.DrawImage(title, op)
	//绘制卡片背景
	op.GeoM.Reset()
	op.GeoM.Translate(546, 202)
	screen.DrawImage(gameImg.skillCardBg, op)
	op.GeoM.Reset()
	op.GeoM.Translate(688, 202)
	screen.DrawImage(gameImg.skillCardBg, op)
	//绘制名字
	op.GeoM.Reset()
	op.GeoM.Translate(497, 212)
	screen.DrawImage(gameImg.skillNameBg, op)
	front.DrawText(screen, "装备牌", 520, 240, 24, true)
	for i := 0; i < len(g.qlgCards); i++ {
		if g.qlgCards[i].getID() != 0 {
			g.qlgCards[i].Draw(screen)
		}
	}
}

func (g *Games) drawMakeJudge(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 70)
	screen.DrawImage(gameImg.skillBg, op)
	var title *ebiten.Image
	switch g.judgeCard.getCardName() {
	case data.LBSS:
		title = getSkillTitleImg("lbss")
	case data.BLCD:
		title = getSkillTitleImg("blcd")
	case data.Lightning:
		title = getSkillTitleImg("lightning")
	}
	op.GeoM.Translate(float64(667-title.Bounds().Dx()/2), 10)
	screen.DrawImage(title, op)
	//绘制卡片背景
	op.GeoM.Reset()
	op.GeoM.Translate(462, 134)
	screen.DrawImage(gameImg.skillCardBg, op)
	op.GeoM.Translate(293, 0)
	screen.DrawImage(gameImg.skillCardBg, op)
	//绘制名字
	op.GeoM.Reset()
	op.GeoM.Translate(688, 144)
	screen.DrawImage(gameImg.skillNameBg, op)
	front.DrawText(screen, "判定结果", 711, 174, 24, true)
	g.judgeCard.Draw(screen)
	if g.judgeResult != nil {
		g.judgeResult.Draw(screen)
	}
}

func (g *Games) drawSkillJudge(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 70)
	screen.DrawImage(gameImg.skillBg, op)
	if g.judgeSkill != 0 {
		title := getSkillTitleImg(g.judgeSkill.String())
		op.GeoM.Translate(float64(667-title.Bounds().Dx()/2), 10)
		screen.DrawImage(title, op)
	}
	//绘制卡片背景
	op.GeoM.Reset()
	op.GeoM.Translate(636, 134)
	screen.DrawImage(gameImg.skillCardBg, op)
	//绘制名字
	op.GeoM.Reset()
	op.GeoM.Translate(569, 144)
	screen.DrawImage(gameImg.skillNameBg, op)
	front.DrawText(screen, "判定结果", 592, 174, 24, true)
	if g.skillJudgeResult != nil {
		g.skillJudgeResult.Draw(screen)
	}
}

func (g *Games) drawGuanXing(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 100)
	screen.DrawImage(gameImg.skillBgLage, op)
	op.GeoM.Translate(324, 68)
	for i := 0; i < 5; i++ {
		screen.DrawImage(gameImg.skillCardBg, op)
		op.GeoM.Translate(0, 193)
		screen.DrawImage(gameImg.skillCardBg, op)
		op.GeoM.Translate(141, -193)
	}
	op.GeoM.Reset()
	op.GeoM.Translate(257, 178)
	screen.DrawImage(gameImg.skillNameBg, op)
	op.GeoM.Translate(0, 193)
	screen.DrawImage(gameImg.skillNameBg, op)
	front.DrawText(screen, "牌堆顶", 290, 238, 24, true)
	front.DrawText(screen, "牌堆底", 290, 431, 24, true)
	op.GeoM.Reset()
	op.GeoM.Translate(617, 110)
	screen.DrawImage(g.skillTitle, op)
	for _, c := range g.heapTop {
		c.Draw(screen)
	}
	for _, c := range g.heapButtom {
		c.Draw(screen)
	}
	for _, b := range g.btnList {
		b.Draw(screen)
	}
}

// 完成游戏的动画计算
func (g *Games) anime() {
	for i := 0; i < len(g.playList); i++ {
		g.playList[i].anime()
	}
	for i := 0; i < len(g.handleZone); i++ {
		g.handleZone[i].anime()
	}
	for i := 0; i < len(g.localStateImgs); i++ {
		g.localStateImgs[i].anime()
	}
	if g.state == data.WinState || g.state == data.LoseState {
		g.animeSettle()
	}
}

func (g *Games) animeSettle() {
	if g.settleAnime.visibility >= 1 {
		return
	}
	g.settleAnime.visibility += 0.01
}

func (g *Games) switchState(inf gameStateInf) {
	//取消选择被选中的牌
	p := g.getPlayer(g.pid)
	for i := len(p.selCard) - 1; i >= 0; i-- {
		p.findCard(p.selCard[i]).deSelect(p)
	}
	//关闭所有主动技
	for _, s := range p.skills {
		s.deselect(g, p)
		s.setActive(false)
	}
	//重置hasSkip
	g.hasSkip = false
	//清空玩家按钮槽
	g.getPlayer(g.pid).btnGrup = nil
	//清空游戏按钮槽
	g.btnList = nil
	//当下一阶段不为五谷丰登或无懈可击时，清空五谷丰登卡槽
	//高风险
	if inf.state == data.UseCardState {
		g.wgfdCards = nil
	}
	//当某一阶段结束时要做的操作
	switch g.state {
	case data.PrepareState:
		//准备阶段结束清空处理区
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
		if g.curPlayer == g.pid {
			g.localStateImgs[0].vanish()
		}
	case data.JudgedState:
		if g.curPlayer == g.pid {
			g.localStateImgs[1].vanish()
		}
	case data.SendCardState:
		if g.curPlayer == g.pid {
			g.localStateImgs[2].vanish()
		}
	case data.UseCardState:
		//出牌阶段结束时清空处理区中不为锦囊牌也不为杀的牌
		iterators(g.handleZone, func(c cardI) {
			if c.getCardType() != data.TipsCardType && c.getCardName() != data.Attack {
				g.moveCard2Drop(c.getID())
			}
		})
		//清理铁索连环（重铸）:
		if inf.state != data.WXKJState {
			iterators(g.handleZone, func(c cardI) {
				if c.getCardName() == data.TSLH {
					g.moveCard2Drop(c.getID())
				}
			})
		}
		if g.curPlayer == g.pid {
			g.localStateImgs[3].vanish()
		}
	case data.DropCardState:
		if g.curPlayer == g.pid {
			g.localStateImgs[4].vanish()
		}
	case data.DodgeState:
		//出闪阶段结束时清空处理区中的闪和杀
		iterators(g.handleZone, func(c cardI) {
			if isItemInList([]data.CardName{data.Dodge, data.Attack, data.LightnAttack, data.FireAttack}, c.getCardName()) {
				g.moveCard2Drop(c.getID())
			}
		})
	case data.WXKJState:
		//当无懈可击阶段结束且下一个阶段不为无懈可击时，清空处理区中不在排除列表的锦囊牌
		//黑名单匹配表，表中的卡牌只有下一个阶段不是指定阶段时才被清除
		blackTable := map[data.CardName]data.GameState{data.Duel: data.DuelState,
			data.SSQY: data.SSQYState, data.GHCQ: data.GHCQState}
		//白名单匹配表，表中的卡牌只有下一个阶段是指定阶段时才被清除
		whiteTable := map[data.CardName]data.GameState{data.NMRQ: data.UseCardState, data.WJQF: data.UseCardState,
			data.WGFD: data.UseCardState, data.Burn: data.UseCardState}
		if inf.state != data.WXKJState {
			iterators(g.handleZone, func(c cardI) {
				if c.getCardType() == data.TipsCardType {
					if s, ok := blackTable[c.getCardName()]; ok {
						if s == inf.state {
							return
						}
					}
					if s, ok := whiteTable[c.getCardName()]; ok {
						if s != inf.state {
							return
						}
					}
					g.moveCard2Drop(c.getID())
				}
			})
		}
		//当无懈可击结束，且下一阶段不为做判定或无懈时，清空判定区
		if inf.state != data.MakeJudgeState && inf.state != data.WXKJState {
			g.judgeCard = nil
		}
	case data.DuelState:
		//当决斗阶段结束时清空杀
		iterators(g.handleZone, func(c cardI) {
			if isItemInList([]data.CardName{data.Attack, data.LightnAttack, data.FireAttack}, c.getCardName()) {
				g.moveCard2Drop(c.getID())
			}
		})
		//当下一个阶段不为决斗时清空决斗
		if inf.state != data.DuelState {
			iterators(g.handleZone, func(c cardI) {
				if c.getCardName() == data.Duel {
					g.moveCard2Drop(c.getID())
				}
			})
		}
	case data.NMRQState:
		//当南蛮入侵阶段结束时清空杀
		iterators(g.handleZone, func(c cardI) {
			if isItemInList([]data.CardName{data.Attack, data.LightnAttack, data.FireAttack}, c.getCardName()) {
				g.moveCard2Drop(c.getID())
			}
		})
	case data.WJQFState:
		//当万箭齐发阶段结束时清空闪
		iterators(g.handleZone, func(c cardI) {
			if c.getCardName() == data.Dodge {
				g.moveCard2Drop(c.getID())
			}
		})
	case data.WGFDState:
		//当五谷丰登阶段结束，且下一阶段不为五谷丰登或无懈可击时，清空五谷丰登卡牌列表与处理区
		if inf.state != data.WGFDState && inf.state != data.WXKJState {
			iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
			g.wgfdCards = nil
		}
	case data.BurnDropState:
		//当火攻弃置阶段结束时清空处理区
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
	case data.SSQYState:
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
	case data.GHCQState:
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
	case data.QLGState:
		g.qlgCards = nil
	case data.MakeJudgeState:
		//判定阶段结束将判定结果送至处理区随后清空判定区与处理区
		g.moveCard2Handle(g.judgeResult)
		g.judgeCard = nil
		g.judgeResult = nil
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
	case data.DyingState:
		//濒死阶段结束时清空处理区里的桃和酒
		iterators(g.handleZone, func(c cardI) {
			if c.getCardName() == data.Peach || c.getCardName() == data.Drunk {
				g.moveCard2Drop(c.getID())
			}
		})
	case data.DieState:
		//死亡阶段结束时清空处理区
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
	case data.EndState:
		if g.curPlayer == g.pid {
			g.localStateImgs[5].vanish()
		}
		//结束阶段结束时还原玩家回合内属性
		g.getPlayer(g.curPlayer).isDrunk = false
		//清空所有玩家所有不在排除列表里的技能框
		denyList := []data.SID{data.WuKuSkill, data.QuanJiSkill, data.ZhenGuSkill,
			data.XionHuoSkill, data.JiQiaoSkill, data.LueYingSkill, data.ShouXiSkill}
		for _, p := range g.playList {
			for sid := range p.skillBoxs {
				if !isItemInList(denyList, sid) {
					delete(p.skillBoxs, sid)
				}
			}
		}
	case data.SkillSelectState:
		//技能选择阶段结束时清空目标列表并将所有玩家设为可选
		g.skilltargetList = nil
		for i := 0; i < len(g.playList); i++ {
			g.playList[i].unSelAble = false
			g.playList[i].isSelected = false
		}
	case data.SkillJudgeState:
		g.moveCard2Handle(g.skillJudgeResult)
		g.skillJudgeResult = nil
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
	case data.DropSelfAllCards:
		if g.curPlayer == g.pid {
			p := g.getPlayer(g.pid)
			for _, c := range p.equipSlot {
				if c == nil || c.getID() == 0 {
					continue
				}
				c.setPos(p.x+10, p.y+20)
				c.setVisibility(0)
				c.setDrawEquipTip(false)
			}
			p.calculatePos()
		}
	case data.GSFState:
		if g.curPlayer == g.pid {
			p := g.getPlayer(g.pid)
			for _, c := range p.equipSlot {
				if c == nil || c.getID() == 0 || c.getCardType() == data.WeaponCardType {
					continue
				}
				c.setPos(p.x+10, p.y+20)
				c.setVisibility(0)
				c.setDrawEquipTip(false)
			}
			p.calculatePos()
		}
	}
	//设置gameState
	g.state = inf.state
	//设置进度条
	if inf.t.Seconds() < 1 {
		g.processInf.total = 0
		g.processInf.remain = g.processInf.total
		g.processInf.rate = 0
	} else {
		g.processInf.total = inf.t.Seconds() * 60
		g.processInf.remain = g.processInf.total
		g.processInf.rate = 1
	}
	//将本地玩家所有牌设为不可用
	localCards := g.getPlayer(g.pid).cards
	for i := 0; i < len(localCards); i++ {
		localCards[i].setSelectedAble(false)
	}
	//切换玩家
	g.curPlayer = inf.curPlayer
	//当自己切换到usecard,wxkj,dying,dodge.Duel,nmrq,BurnDrop...时,接收可用牌信息
	if g.state == data.WXKJState || (g.curPlayer == g.pid &&
		isItemInList([]data.GameState{data.UseCardState, data.DyingState, data.DodgeState, data.DuelState,
			data.NMRQState, data.WJQFState, data.BurnDropState, data.JDSRState, data.QlYYDState, data.WenJiState,
			data.EnYuanState, data.LuanWuState}, g.state)) {
		g.useAbleCards = <-g.useAbleReceiver
		//检查玩家是否已经拥有所有可用牌
	checkpoint:
		pCards := []data.CID{}
		for i := 0; i < len(p.cards); i++ {
			pCards = append(pCards, p.cards[i].getID())
		}
		for _, c := range g.useAbleCards {
			if isItemInList(pCards, c) {
				continue
			}
			g.addCard(<-g.cardReceiver)
			goto checkpoint
		}
		for _, card := range p.cards {
			card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
		}
	}
	//当自己切换到dropCard,dropSelfHand...时，接受可丢弃牌信息
	if g.curPlayer == g.pid &&
		isItemInList([]data.GameState{data.DropCardState, data.DropSelfHandCard, data.DropSelfAllCards,
			data.GSFState}, g.state) {
		g.dropAbleInf = <-g.dropAbleRec
		g.selNum = int(g.dropAbleInf.dropNum)
		localCards := g.getPlayer(g.pid).cards
		for i := 0; i < len(localCards); i++ {
			localCards[i].setSelectedAble(isItemInList(g.dropAbleInf.cards, localCards[i].getID()))
		}
	}
	//当自己切换到useCardState...时，接受可用技能信息
	if g.curPlayer == g.pid &&
		isItemInList([]data.GameState{data.UseCardState, data.DuelState, data.NMRQState,
			data.DodgeState, data.WJQFState, data.QlYYDState, data.DyingState,
			data.JDSRState, data.LuanWuState}, g.state) {
		useAbleSkill := <-g.useAbleSkillRec
		for _, s := range g.getPlayer(g.pid).skills {
			s.setActive(isItemInList(useAbleSkill, s.getID()))
		}
	}
	//进入无懈可击接受技能
	if g.state == data.WXKJState {
		useAbleSkill := <-g.useAbleSkillRec
		for _, s := range g.getPlayer(g.pid).skills {
			s.setActive(isItemInList(useAbleSkill, s.getID()))
		}
	} else {
		g.wxkjInf.count = 0
		g.wxkjInf.cName = data.NoName
	}
	//当某一阶段开始时要做的操作
	switch inf.state {
	case data.PrepareState:
		if g.curPlayer == g.pid {
			g.localStateImgs[0].showUp()
		}
	case data.JudgedState:
		if g.curPlayer == g.pid {
			g.localStateImgs[1].showUp()
		}
	case data.SendCardState:
		if g.curPlayer == g.pid {
			g.localStateImgs[2].showUp()
		}
	case data.UseCardState:
		//出牌阶段开始时清空所有牌
		iterators(g.handleZone, func(c cardI) { g.moveCard2Drop(c.getID()) })
		if g.curPlayer == g.pid {
			g.localStateImgs[3].showUp()
		}
	case data.DropCardState:
		if g.curPlayer == g.pid {
			g.localStateImgs[4].showUp()
		}
	case data.EndState:
		if g.curPlayer == g.pid {
			g.localStateImgs[5].showUp()
		}
	case data.DropSelfAllCards:
		g.selNum = int(g.dropAbleInf.dropNum)
		//如果为自己的回合
		if g.curPlayer == g.pid {
			p := g.getPlayer(g.pid)
			//将武器之外的装备牌指针添加到手牌堆然后再删掉
			previousLen := len(p.cards)
			for _, c := range p.equipSlot {
				if c == nil || c.getID() == 0 {
					continue
				}
				c.setVisibility(1)
				c.setDrawEquipTip(true)
				p.cards = append(p.cards, c)
			}
			p.calculatePos()
			//将本地玩家所有牌设为可用
			for i := 0; i < len(p.cards); i++ {
				p.cards[i].setSelectedAble(true)
			}
			p.cards = p.cards[:previousLen]
		}
	case data.GSFState:
		g.selNum = int(g.dropAbleInf.dropNum)
		//如果为自己的回合
		if g.curPlayer == g.pid {
			p := g.getPlayer(g.pid)
			//将武器之外的装备牌指针添加到手牌堆然后再删掉
			previousLen := len(p.cards)
			for _, c := range p.equipSlot {
				if c == nil || c.getID() == 0 || c.getCardType() == data.WeaponCardType {
					continue
				}
				c.setVisibility(1)
				c.setDrawEquipTip(true)
				p.cards = append(p.cards, c)
			}
			p.calculatePos()
			//将本地玩家所有牌设为可用
			for i := 0; i < len(p.cards); i++ {
				p.cards[i].setSelectedAble(true)
			}
			p.cards = p.cards[:previousLen]
		}
	case data.WenJiState:
		if g.curPlayer == g.pid {
			p := g.getPlayer(g.pid)
			//将武器之外的装备牌指针添加到手牌堆然后再删掉
			previousLen := len(p.cards)
			for _, c := range p.equipSlot {
				if c == nil || c.getID() == 0 {
					continue
				}
				c.setVisibility(1)
				c.setDrawEquipTip(true)
				p.cards = append(p.cards, c)
			}
			p.calculatePos()
			//将本地玩家所有牌设为可用
			for i := 0; i < len(p.cards); i++ {
				p.cards[i].setSelectedAble(true)
			}
			p.cards = p.cards[:previousLen]
		}
	case data.BurnShowState:
		//如果为自己的回合
		if g.curPlayer == g.pid {
			//将本地玩家所有牌设为可用
			localCards := g.getPlayer(g.pid).cards
			for i := 0; i < len(localCards); i++ {
				localCards[i].setSelectedAble(true)
			}
		}
	case data.DieState:
		sound.PlayAudio("assets/audio/dead.mp3")
		p := g.getPlayer(g.curPlayer)
		p.death = true
		p.dying = false
		sound.PlayDeathSound(p.name)
		p.skillBoxs = make(map[data.SID]skillBoxI)
	case data.SetHpState:
		//检查设置血量指令
		g.setHp(<-g.setHpReceiver)
	case data.SkillSelectState:
		//进入技能选择阶段时读取要选择的技能
		g.selectSkill = <-g.skillSelectRec
		if g.curPlayer == g.pid { //如果是自己则接受目标列表
			g.skillTargetInf = <-g.availableTargetrec
			if g.skillTargetInf.TargetNum != 0 {
				for i := 0; i < len(g.playList); i++ {
					if !isItemInList(g.skillTargetInf.TargetList, g.playList[i].pid) {
						g.playList[i].unSelAble = true
					}
				}
			}
		}
	case data.PoJunState:
		if g.curPlayer == g.pid {
		checkPoint:
			rsp := <-g.useSkillRspRec
			if rsp.ID != data.PoJunSkill || len(rsp.Args) != 1 {
				goto checkPoint
			}
			g.selNum = int(rsp.Args[0])
			g.otherCards = nil
			g.skillTitle = getSkillTitleImg("pojun")
			//装备牌
			for i, cid := range rsp.Cards[:4] {
				if cid == 0 {
					continue
				}
				c := copyCardByID(cid, g)
				c.setPos(float64(i)*140+146, 384)
				c.setSelectedAble(true)
				g.otherCards = append(g.otherCards, c)
			}
			//绘制手牌堆
			dx := 136.
			if len(rsp.Cards)-7 > 8 {
				dx = 1088. / float64(len(rsp.Cards)-4)
			}
			for i, cid := range rsp.Cards[4:] {
				c := copyCardByID(cid, g)
				c.setPos(float64(i)*dx+148, 192)
				c.setSelectedAble(true)
				c.setShowBack(true)
				g.otherCards = append(g.otherCards, c)
			}
		}
	case data.FanKuiState:
		if g.curPlayer == g.pid {
			rsp := <-g.useSkillRspRec
			g.otherCards = nil
			g.skillTitle = getSkillTitleImg("fankui")
			//装备牌
			for i, cid := range rsp.Cards[:4] {
				if cid == 0 {
					continue
				}
				c := copyCardByID(cid, g)
				c.setPos(float64(i)*140+146, 384)
				c.setSelectedAble(true)
				g.otherCards = append(g.otherCards, c)
			}
			//绘制手牌堆
			dx := 136.
			if len(rsp.Cards)-7 > 8 {
				dx = 1088. / float64(len(rsp.Cards)-4)
			}
			for i, cid := range rsp.Cards[4:] {
				c := copyCardByID(cid, g)
				c.setPos(float64(i)*dx+148, 192)
				c.setSelectedAble(true)
				c.setShowBack(true)
				g.otherCards = append(g.otherCards, c)
			}
		}
	case data.LiyuState:
		if g.curPlayer == g.pid {
			rsp := <-g.useSkillRspRec
			g.selNum = int(rsp.Args[0])
			for _, c := range g.getPlayer(g.pid).cards {
				c.setSelectedAble(true)
			}
		}
	case data.GuanXingState:
		if g.curPlayer == g.pid {
			inf := <-g.useSkillRspRec
			g.skillTitle = getSkillTitleImg("guanxing")
			g.heapTop = nil
			g.heapButtom = nil
			for _, c := range inf.Cards {
				card := copyCardByID(c, g)
				card.setVisibility(1)
				card.setSelectedAble(true)
				g.heapTop = append(g.heapTop, card)
			}
			calculateCardPos(321, 172, 141, 705, true, g.heapTop...)
		}
	case data.WinState:
		g.bgm.Close()
		sound.PlayAudio("assets/audio/win.mp3")
		g.settleAnime.img = gameImg.winImg
		g.settleAnime.visibility = 0
	case data.LoseState:
		g.bgm.Close()
		sound.PlayAudio("assets/audio/fail.mp3")
		g.settleAnime.img = gameImg.loseImg
		g.settleAnime.visibility = 0
	}
	g.updatePromptText()
}

func (g *Games) setRoleDetail(sid []data.SID) {
	text := ""
	for _, skill := range sid {
		const maxLen = 24 //每行的最大长度
		str := []rune(skill.Name() + "：" + skill.Text())
		for i := maxLen; i < len(str); i += maxLen {
			str = append(str[:i], append([]rune("\n"), str[i:]...)...)
		}
		text += "[yellow]" + skill.Name() + "：[white]" + string(str[len([]rune(skill.Name()))+1:])
		if len(str) != 0 && str[len(str)-1] != rune('\n') {
			text += "\n\n"
		}
	}
	g.roledetail.SetText(text)
}

func (g *Games) getPlayer(pid data.PID) *player {
	index := pid - g.pid
	if index < 0 {
		return g.playList[len(g.playList)+int(index)]
	}
	return g.playList[index]
}

// 将牌移动到处理区
func (g *Games) moveCard2Handle(c cardI) {
	c.resetFadeOut()
	c.setVisibility(1)
	c.setShowBack(false)
	c.setDrawEquipTip(false)
	c.setSelectedAble(true)
	c.setInHandleZoon(true)
	g.handleZone = append(g.handleZone, c)
	g.calculateHandleZonePos()
}

// 计算处理区牌的位置
func (g *Games) calculateHandleZonePos() {
	const maxLength = 600.
	if len(g.handleZone) == 1 {
		g.handleZone[0].move2Pos(cardHeapPos.x, cardHeapPos.y, nil)
	} else {
		dx := 130.
		if float64(len(g.handleZone))*dx > maxLength {
			dx = maxLength / float64(len(g.handleZone))
		}
		for i, card := range g.handleZone {
			card.move2Pos(cardHeapPos.x+68-float64(len(g.handleZone)-1-i)*dx, cardHeapPos.y, nil)
		}
	}
}

// 将牌从处理区移到弃牌堆
func (g *Games) moveCard2Drop(id data.CID) {
	for _, c := range g.handleZone {
		if c.getID() == id {
			if c.isOnFade() { //如果卡片已经在淡出了则跳过
				continue
			}
			c.setFadeOut(3*60, func() {
				g.handleZone, _ = getCardById(g.handleZone, id)
				g.calculateHandleZonePos()
				c.setPos(cardHeapPos.x, cardHeapPos.y)
				c.setVisibility(0)
				c.setInHandleZoon(false)
				if c.getID() == 0 {
					return
				}
				g.cards[c.getID()] = c
			})
			return
		}
	}
}

// 发牌
func (g *Games) addCard(rec cardReceive) {
	card := []cardI{}
	for i := 0; i < len(rec.cards); i++ {
		if c := g.cards[rec.cards[i]]; c != nil {
			card = append(card, c)
			g.cards[rec.cards[i]] = nil
			continue
		}
		for j := len(g.handleZone) - 1; j >= 0; j-- {
			c := g.handleZone[j]
			if c.getID() == rec.cards[i] {
				c.resetFadeOut()
				card = append(card, c)
				g.handleZone = append(g.handleZone[:j], g.handleZone[j+1:]...)
				g.calculateHandleZonePos()
				goto loop
			}
		}
		panic("在牌堆与处理区都找不到id=" + strconv.Itoa(int(rec.cards[i])) + "的牌")
	loop:
	}
	g.setLittleHeapNum(g.littleHeapNum - uint8(len(rec.cards)))
	g.getPlayer(rec.id).addCard(card...)
}

// 使用卡
func (g *Games) useCard(rec usecardInf) {
	c := g.getPlayer(rec.user).getCard(rec.card)
	c.deSelect(g.getPlayer(rec.user))
	g.getPlayer(rec.user).calculatePos()
	//如果是锦囊牌或者基本牌则将卡本体送入弃牌堆，否则复制一张送入弃牌堆
	if c.getCardType() == data.BaseCardType || c.getCardType() == data.TipsCardType {
		g.moveCard2Handle(c)
	} else {
		//bug
		//如果是闪电在转移或者生效则跳出
		if c.getCardName() == data.Lightning && len(rec.targets) == 1 && rec.targets[0] != rec.user {
			c.onUse(g, rec.user, rec.targets...)
			return
		}
		replica := copyNoIDCard(c.getID(), g)
		replica.setPos(c.getPos())
		g.moveCard2Handle(replica)
	}
	c.onUse(g, rec.user, rec.targets...)
}

func (g *Games) useTmpCard(inf useTmpCardInf) {
	user := g.getPlayer(inf.user)
	c := newCard(data.NewCard(inf.cname, inf.dec, inf.num), g)
	c.setPos(user.x+10, user.y+20)
	c.setSelectedAble(true)
	c.setVisibility(1)
	g.moveCard2Handle(c)
	c.onUse(g, inf.user, inf.targets...)
}

// 从玩家区域中移除牌
func (g *Games) removeCard(rec cardReceive) {
	for i := len(rec.cards) - 1; i >= 0; i-- {
		c := g.getPlayer(rec.id).getCard(rec.cards[i])
		c.deSelect(g.getPlayer(rec.id))
		c.setPromptText("[white]"+g.getPlayer(rec.id).dspName, " [yellow]弃牌")
		g.getPlayer(rec.id).calculatePos()
		g.moveCard2Handle(c)
		if g.state == data.DropCardState {
			g.moveCard2Drop(c.getID())
		}
	}
}

// 在玩家之间移动牌
func (g *Games) moveCard(inf cardMoveRec) {
	switch inf.src {
	case data.SpecialPIDGame:
		cards := []cardI{}
		for _, cid := range inf.cards {
			for i, c := range g.handleZone {
				if c.getID() == cid {
					cards = append(cards, c)
					g.handleZone = append(g.handleZone[:i], g.handleZone[i+1:]...)
					g.calculateHandleZonePos()
					goto loopEnd
				}
			}
			if g.cards[cid] != nil {
				cards = append(cards, g.cards[cid])
				g.cards[cid] = nil
				goto loopEnd
			}
			panic("ID为" + strconv.Itoa(int(cid)) + "的卡片不存在于处理区与弃牌堆")
		loopEnd:
		}
		g.getPlayer(inf.dst).addCard(cards...)
	default:
		src := g.getPlayer(inf.src)
		dst := g.getPlayer(inf.dst)
		cards := []cardI{}
		for _, cid := range inf.cards {
			c := src.getCard(cid)
			c.deSelect(src)
			cards = append(cards, c)
		}
		src.calculatePos()
		dst.addCard(cards...)
	}
}

// 接收hp信息
func (g *Games) setHp(inf setHpInf) {
	p := g.getPlayer(inf.pid)
	if inf.hp == p.hp {
		return
	}
	switch inf.hptype {
	case data.DffHPMax:
		sound.PlayAudio("assets/audio/bleeding.mp3")
		p.maxHp = inf.hp
		if p.hp > p.maxHp {
			p.hp = p.maxHp
		}
	case data.BleedingDmg:
		sound.PlayAudio("assets/audio/bleeding.mp3")
		p.hp = inf.hp
		p.dmgAnime.enable = true
		p.dmgAnime.hptype = inf.hptype
	case data.Recover:
		sound.PlayAudio("assets/audio/recover.mp3")
		p.hp = inf.hp
		if p.dying && p.hp > 0 {
			p.dying = false
		}
		p.recAnime.enable = true
	case data.NormalDmg:
		sound.PlayAudio("assets/audio/normalDmg.mp3")
		p.hp = inf.hp
		p.dmgAnime.enable = true
		p.dmgAnime.hptype = inf.hptype
	case data.FireDmg:
		sound.PlayAudio("assets/audio/fireDmg.mp3")
		p.hp = inf.hp
		p.dmgAnime.enable = true
		p.dmgAnime.hptype = inf.hptype
	case data.LightningDmg:
		sound.PlayAudio("assets/audio/lightnDmg.mp3")
		p.hp = inf.hp
		p.dmgAnime.enable = true
		p.dmgAnime.hptype = inf.hptype
	case data.AddHpMax:
		p.maxHp = inf.hp
	}
	if inf.hp < 1 {
		g.getPlayer(inf.pid).dying = true
	}
}

// 接受gsidcard信息
func (g *Games) handleGSIDCardsInf(inf gsidCardInf) {
	switch inf.id {
	case data.GSIDWGFD:
		if len(g.wgfdCards) == 0 {
			baseX := float64(1334/2 - (len(inf.cards)*136)/2)
			for i, c := range inf.cards {
				card := copyCardByID(c, g)
				card.setSelectedAble(true)
				card.setVisibility(1)
				card.setPos(float64(i)*136+baseX, 190)
				g.wgfdCards = append(g.wgfdCards, card)
			}
		} else {
			for i := 0; i < len(g.wgfdCards); i++ {
				if !isItemInList(inf.cards, g.wgfdCards[i].getID()) {
					g.wgfdCards[i].setSelectedAble(false)
				}
			}
		}
	case data.GSIDBurn:
		c := copyNoIDCard(inf.cards[0], g)
		for i := 0; i < len(g.playList); i++ {
			for _, pc := range g.playList[i].cards {
				if pc.getID() == inf.cards[0] {
					c.setPromptText("[white]"+g.playList[i].dspName, " [yellow]火攻展示")
					c.setPos(pc.getPos())
					goto out
				}
			}
		}
	out:
		g.moveCard2Handle(c)
	case data.GSIDSSQY:
		g.ssqyCards = nil
		//装备牌
		for i, cid := range inf.cards[:4] {
			if cid == 0 {
				continue
			}
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*140+146, 384)
			c.setSelectedAble(true)
			g.ssqyCards = append(g.ssqyCards, c)
		}
		//判定区卡牌
		for i, cid := range inf.cards[4:7] {
			if cid == 0 {
				continue
			}
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*140+792, 384)
			c.setSelectedAble(true)
			g.ssqyCards = append(g.ssqyCards, c)
		}
		//绘制手牌堆
		dx := 136.
		if len(inf.cards)-7 > 8 {
			dx = 1088. / float64(len(inf.cards)-7)
		}
		for i, cid := range inf.cards[7:] {
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*dx+148, 192)
			c.setSelectedAble(true)
			c.setShowBack(true)
			g.ssqyCards = append(g.ssqyCards, c)
		}
	case data.GSIDGHCQ:
		g.ghcqCards = nil
		//装备牌
		for i, cid := range inf.cards[:4] {
			if cid == 0 {
				continue
			}
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*140+146, 384)
			c.setSelectedAble(true)
			g.ghcqCards = append(g.ghcqCards, c)
		}
		//判定区卡牌
		for i, cid := range inf.cards[4:7] {
			if cid == 0 {
				continue
			}
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*140+792, 384)
			c.setSelectedAble(true)
			g.ghcqCards = append(g.ghcqCards, c)
		}
		//绘制手牌堆
		dx := 136.
		if len(inf.cards)-7 > 8 {
			dx = 1088. / float64(len(inf.cards)-7)
		}
		for i, cid := range inf.cards[7:] {
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*dx+148, 192)
			c.setSelectedAble(true)
			c.setShowBack(true)
			g.ghcqCards = append(g.ghcqCards, c)
		}
	case data.GSIDPerJudge:
		var c cardI
		switch data.JudgeSlot(inf.cards[0]) {
		case data.LBSSSlot:
			c = newCard(data.NewCard(data.LBSS, data.NoDec, 0), g)
		case data.BLCDSlot:
			c = newCard(data.NewCard(data.BLCD, data.NoDec, 0), g)
		case data.LightningSlot:
			c = newCard(data.NewCard(data.Lightning, data.NoDec, 0), g)
		}
		c.setPos(459, 138)
		c.setSelectedAble(true)
		g.judgeCard = c
	case data.GSIDJudge:
		c := copyNoIDCard(inf.cards[0], g)
		c.setPos(752, 138)
		c.setSelectedAble(true)
		g.judgeResult = c
	case data.GSIDPerSkillJudge:
		g.judgeSkill = data.SID(inf.cards[0])
	case data.GSIDSkillJudge:
		c := copyNoIDCard(inf.cards[0], g)
		c.setPos(633, 138)
		c.setSelectedAble(true)
		g.skillJudgeResult = c
	case data.GSIDQLG:
		if inf.cards[0] != 0 {
			c := copyCardByID(inf.cards[0], g)
			c.setPos(543, 206)
			c.setSelectedAble(true)
			g.qlgCards = append(g.qlgCards, c)
		}
		if inf.cards[1] != 0 {
			c := copyCardByID(inf.cards[1], g)
			c.setPos(685, 206)
			c.setSelectedAble(true)
			g.qlgCards = append(g.qlgCards, c)
		}
	case data.GSIDDropOtrCard:
		g.otherCards = nil
		//装备牌
		for i, cid := range inf.cards[:4] {
			if cid == 0 {
				continue
			}
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*140+146, 384)
			c.setSelectedAble(true)
			g.otherCards = append(g.otherCards, c)
		}
		//判定区卡牌
		for i, cid := range inf.cards[4:7] {
			if cid == 0 {
				continue
			}
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*140+792, 384)
			c.setSelectedAble(true)
			g.otherCards = append(g.otherCards, c)
		}
		//绘制手牌堆
		dx := 136.
		if len(inf.cards)-7 > 8 {
			dx = 1088. / float64(len(inf.cards)-7)
		}
		for i, cid := range inf.cards[7:] {
			c := copyCardByID(cid, g)
			c.setPos(float64(i)*dx+148, 192)
			c.setSelectedAble(true)
			c.setShowBack(true)
			g.otherCards = append(g.otherCards, c)
		}
	}
}

func (g *Games) useSkill(inf useSkillInf) {
	p := g.getPlayer(inf.user)
	switch inf.skill {
	case data.TurnBackSkill:
		p.turnBack = !p.turnBack
		return
	case data.TieSuoSkill:
		p.isLinked = !p.isLinked
		p.tslhAnime.enable = true
		return
	case data.CleanDrunk:
		p.isDrunk = false
		return
	case data.AddHeapNum:
		if inf.args == nil {
			return
		}
		g.setLittleHeapNum(g.littleHeapNum + inf.args[0])
		return
	case data.DrawLBSS:
		p.delayTipsAnime.enable = true
		p.delayTipsAnime.cardName = data.LBSS
		return
	case data.DrawBLCD:
		p.delayTipsAnime.enable = true
		p.delayTipsAnime.cardName = data.BLCD
		return
	case data.DrawLightning:
		p.delayTipsAnime.enable = true
		p.delayTipsAnime.cardName = data.Lightning
		return
		//以上为游戏特技
	case data.PoJunSkill:
		if len(inf.args) == 0 {
			break
		}
		t := g.getPlayer(data.PID(inf.args[0]))
		cards := []cardI{}
		for _, cid := range inf.args[1:] {
			cards = append(cards, t.getCard(data.CID(cid)))
		}
		t.addTSCard(data.PoJunSkill, cards...)
		t.calculatePos()
		box := newSkillBox(data.PoJunSkill)
		box.name.SetText(box.name.GetStr() + " " + strconv.Itoa(len(t.tsCard[data.PoJunSkill])))
		t.skillBoxs[data.PoJunSkill] = &box
		return
	case data.ChengLueSkill:
		if len(inf.args) == 0 {
			break
		}
		decs := make([]data.Decor, len(inf.args))
		for i, dec := range inf.args {
			decs[i] = data.Decor(dec)
		}
		if box := p.skillBoxs[data.ChengLueSkill]; box != nil {
			box.(*chenglueSkillBox).decs = decs
		} else {
			p.skillBoxs[data.ChengLueSkill] = &chenglueSkillBox{
				skillBox: newSkillBox(data.ChengLueSkill),
				decs:     decs,
			}
		}
		return
	case data.ShiCaiSkill:
		if len(inf.args) == 0 {
			break
		}
		if box := p.skillBoxs[data.ShiCaiSkill]; box != nil {
			box.(*shiCaiSkillBox).addType(data.CardType(inf.args[0]))
		} else {
			box := &shiCaiSkillBox{skillBox: newSkillBox(data.ShiCaiSkill)}
			box.addType(data.CardType(inf.args[0]))
			p.skillBoxs[data.ShiCaiSkill] = box
		}
		return
	case data.FeiYingSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if box := p.skillBoxs[data.FeiYingSkill]; box != nil {
			box.(*fengYinSkillBox).setColor(data.Decor(inf.args[0]))
		} else {
			box := &fengYinSkillBox{skillBox: newSkillBox(data.FeiYingSkill)}
			box.setColor(data.Decor(inf.args[0]))
			p.skillBoxs[data.FeiYingSkill] = box
		}
		return
	case data.JiLiSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if box := p.skillBoxs[data.JiLiSkill]; box != nil {
			box.(*jiLiSkillBox).setNum(int(inf.args[0]))
		} else {
			box := &jiLiSkillBox{skillBox: newSkillBox(data.JiLiSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.JiLiSkill] = box
		}
		return
	case data.WenJiSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if box := p.skillBoxs[data.WenJiSkill]; box != nil {
			box.(*wenJiSkillBox).setNum(data.CardName(inf.args[0]))
		} else {
			box := &wenJiSkillBox{skillBox: newSkillBox(data.WenJiSkill)}
			box.setNum(data.CardName(inf.args[0]))
			p.skillBoxs[data.WenJiSkill] = box
		}
		return
	case data.QiZhiSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if box := p.skillBoxs[data.QiZhiSkill]; box != nil {
			box.(*qiZhiSkillBox).setNum(int(inf.args[0]))
		} else {
			box := &qiZhiSkillBox{skillBox: newSkillBox(data.QiZhiSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.QiZhiSkill] = box
		}
		return
	case data.LiYuSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		t := g.getPlayer(data.PID(inf.args[0]))
		cards := []cardI{}
		for _, cid := range inf.args[1:] {
			cards = append(cards, t.getCard(data.CID(cid)))
		}
		t.addTSCard(data.LiYuSkill, cards...)
		t.calculatePos()
		if box := p.skillBoxs[data.LiYuSkill]; box != nil {
			box.(*liYuSkillBox).setNum(len(cards))
		} else {
			box := &liYuSkillBox{skillBox: newSkillBox(data.LiYuSkill)}
			box.setNum(len(cards))
			p.skillBoxs[data.LiYuSkill] = box
		}
		return
	case data.WuJiSkill:
		for i, s := range p.skills {
			if s.getID() == data.HuXiaoSkill {
				p.skills = append(p.skills[:i], p.skills[i+1:]...)
				goto drawAndSound
			}
		}
		goto drawAndSound
	case data.QianXiSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		t := inf.args[1]
		if box := g.getPlayer(data.PID(t)).skillBoxs[data.QianXiSkill]; box != nil {
			box.(*qianxiSkillbox).setCol(data.Decor(inf.args[0]))
		} else {
			box := &qianxiSkillbox{skillBox: newSkillBox(data.QianXiSkill)}
			box.setCol(data.Decor(inf.args[0]))
			g.getPlayer(data.PID(t)).skillBoxs[data.QianXiSkill] = box
		}
		return
	case data.WuKuSkill:
		if len(inf.args) == 0 {
			return
		}
		if inf.args[0] == 0 {
			delete(p.skillBoxs, data.WuKuSkill)
			return
		}
		if box := p.skillBoxs[data.WuKuSkill]; box != nil {
			box.(*wukuSkillBox).setNum(int(inf.args[0]))
		} else {
			box := &wukuSkillBox{skillBox: newSkillBox(data.WuKuSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.WuKuSkill] = box
		}
	case data.SanChenSkill:
		if len(inf.args) == 0 {
			return
		}
		p := g.getPlayer(data.PID(inf.args[0]))
		if p.pid != g.pid {
			break
		}
		for _, si := range p.skills {
			if s, ok := si.(*miewuSkill); ok {
				s.gaint = true
				break
			}
		}
		// p.sidList = append(p.sidList, data.MieWuSkill)
		// if p.pid != g.pid {
		// 	break
		// }
		// s := newSkillI(g, p.pid, data.MieWuSkill)
		// s.setPos(1020, 640)
		// if len(p.skills) == 3 {
		// 	p.skills[2].setPos(1020, 550)
		// }
		// p.skills = append(p.skills, s)
	case data.QuanJiSkill:
		if len(inf.args) < 2 {
			return
		}
		if inf.args[0] == 0 {
			delete(p.skillBoxs, data.QuanJiSkill)
			return
		}
		switch inf.args[1] {
		case 1:
			if box := p.skillBoxs[data.QuanJiSkill]; box != nil {
				box.(*quanJiSkillBox).setNum(int(inf.args[0]))
			} else {
				box := &quanJiSkillBox{skillBox: newSkillBox(data.QuanJiSkill)}
				box.setNum(int(inf.args[0]))
				p.skillBoxs[data.QuanJiSkill] = box
			}
			card := []cardI{p.getCard(data.CID(inf.args[2]))}
			p.addTSCard(data.QuanJiSkill, card...)
			p.selCard = nil
			p.calculatePos()
			goto drawAndSound
		case 0:
			box := &quanJiSkillBox{skillBox: newSkillBox(data.QuanJiSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.QuanJiSkill] = box
			return
		}
	case data.ZhenGuSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if inf.args[0] == 1 {
			if box := p.skillBoxs[data.ZhenGuSkill]; box == nil {
				p.skillBoxs[data.ZhenGuSkill] =
					&zhenGuSkillBox{skillBox: newSkillBox(data.ZhenGuSkill)}
			}
		} else {
			delete(p.skillBoxs, data.ZhenGuSkill)
		}
		return
	case data.XionHuoSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if inf.args[0] == 0 {
			delete(p.skillBoxs, data.XionHuoSkill)
			return
		}
		if box := p.skillBoxs[data.XionHuoSkill]; box != nil {
			box.(*baoliSkillBox).setNum(int(inf.args[0]))
		} else {
			box := &baoliSkillBox{skillBox: newSkillBox(data.XionHuoSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.XionHuoSkill] = box
		}
		return
	case data.JiQiaoSkill:
		if len(inf.args) == 0 {
			goto drawAndSound
		}
		if inf.args[0] == 0 {
			box := p.skillBoxs[data.JiQiaoSkill].(*jiQiaoSkillBox)
			box.count--
			if box.count == 0 {
				delete(p.skillBoxs, data.JiQiaoSkill)
				return
			}
			box.setNum(box.count)
		} else {
			if box := p.skillBoxs[data.JiQiaoSkill]; box != nil {
				box.(*jiQiaoSkillBox).count += 1
				box.(*jiQiaoSkillBox).setNum(box.(*jiQiaoSkillBox).count)
			} else {
				box := &jiQiaoSkillBox{skillBox: newSkillBox(data.JiQiaoSkill)}
				box.setNum(1)
				box.count = 1
				p.skillBoxs[data.JiQiaoSkill] = box
			}
		}
		return
	case data.JianYingSkill:
		if len(inf.args) < 2 {
			goto drawAndSound
		}
		if box := p.skillBoxs[data.JianYingSkill]; box != nil {
			box.(*jianYingSkillBox).dec = data.Decor(inf.args[0])
			box.(*jianYingSkillBox).num = data.CNum(inf.args[1])
		} else {
			p.skillBoxs[data.JianYingSkill] = &jianYingSkillBox{
				skillBox: newSkillBox(data.JianYingSkill),
				dec:      data.Decor(inf.args[0]),
				num:      data.CNum(inf.args[1]),
			}
		}
		return
	case data.YanYuSkill:
		if len(inf.args) < 1 {
			goto drawAndSound
		}
		if box := p.skillBoxs[data.YanYuSkill]; box != nil {
			box.(*yanYuSkillBox).setNum(int(inf.args[0]))
		} else {
			box := &yanYuSkillBox{skillBox: newSkillBox(data.YanYuSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.YanYuSkill] = box
		}
	case data.LueYingSkill:
		if len(inf.args) < 1 {
			goto drawAndSound
		}
		if inf.args[0] == 0 {
			delete(p.skillBoxs, data.LueYingSkill)
			return
		}
		if box := p.skillBoxs[data.LueYingSkill]; box != nil {
			box.(*lueYingSkillBox).setNum(int(inf.args[0]))
		} else {
			box := &lueYingSkillBox{skillBox: newSkillBox(data.LueYingSkill)}
			box.setNum(int(inf.args[0]))
			p.skillBoxs[data.LueYingSkill] = box
		}
		return
	case data.ShouXiSkill:
		if len(inf.args) < 2 {
			goto drawAndSound
		}
		if inf.args[0] == 0 {
			if box := p.skillBoxs[data.ShouXiSkill]; box != nil {
				box.(*shouXiSkillBox).setNum(int(inf.args[1]))
			} else {
				box := &shouXiSkillBox{skillBox: newSkillBox(data.ShouXiSkill)}
				box.setNum(int(inf.args[1]))
				p.skillBoxs[data.ShouXiSkill] = box
			}
		} else {
			if box := p.skillBoxs[data.ShouXiSkill]; box != nil {
				box.(*shouXiSkillBox).setAddSkill(data.SID(inf.args[1]))
			} else {
				box := &shouXiSkillBox{skillBox: newSkillBox(data.ShouXiSkill)}
				box.setAddSkill(data.SID(inf.args[1]))
				p.skillBoxs[data.ShouXiSkill] = box
			}
		}
		return
	}
drawAndSound:
	_, err := os.Stat(inf.skill.String())
	if err == nil {
		g.skillTitle = getSkillTitleImg(inf.skill.String())
	}
	//音效和动画
	sound.PlaySkillSound(inf.skill)
	p.skillAnime.index = 0
	p.skillAnime.counter = 0
	p.skillAnime.visibility = 1
	p.skillAnime.enable = true
	p.skillAnime.name = getSkillNameImg(inf.skill.String())
}

func (g *Games) updatePromptText() {
	var str string
	isLocal := g.curPlayer == g.pid
	switch g.state {
	case data.SkillSelectState:
		if isLocal {
			str = "[white]是否发动[blue]【" + g.selectSkill.Name() + "】"
		} else {
			str = "[yellow]" + g.selectSkill.Name() + " [white]思考中..."
		}
	case data.DropCardState:
		if isLocal {
			str = "[white]弃牌阶段，选" + strconv.Itoa(g.selNum) + "张手牌，点弃牌弃置"
		} else {
			str = "[yellow]弃牌 [white]思考中..."
		}
	case data.DyingState:
		var name string
		var num int
		for _, p := range g.playList {
			if p.hp <= 0 {
				name = p.dspName
				num = 1 - int(p.hp)
				break
			}
		}
		if isLocal {
			str = "[green]" + name + " [white]生命危急，需要" + strconv.Itoa(num) + "个[blue]桃。"
		} else {
			str = "[yellow]桃 [white]思考中..."
		}
	case data.WXKJState:
		rsp := <-g.useSkillRspRec
		cname := data.CardName(rsp.Args[0])
		if g.wxkjInf.cName == cname && g.wxkjInf.target == rsp.Targets[0] {
			g.wxkjInf.count++
		} else {
			g.wxkjInf.cName = cname
			g.wxkjInf.target = rsp.Targets[0]
			g.wxkjInf.count = 0
		}
		var enable string
		if rsp.Args[1]%2 == 0 {
			enable = "生效"
		} else {
			enable = "失效"
		}
		str = "[blue]" + g.wxkjInf.cName.ChnName() + "[white] 对 [green]" + g.getPlayer(rsp.Targets[0]).dspName +
			"[white]即将" + enable + "，是否出无懈可击？"
		g.promptText.SetText(str)
		g.promptText.SetPos((714-g.promptText.Width)/2+290, 545)
		g.promptTextOther.SetText("")
		return
	case data.ReConnState:
		g.promptText.SetText("[green]" + g.getPlayer(g.curPlayer).dspName + "[white]重新连接中")
		g.promptText.SetPos((714-g.promptText.Width)/2+290, 545)
		g.promptTextOther.SetText("")
		return
	default:
		if isLocal {
			str = g.state.GetText()
		} else {
			str = g.state.GetTextOther()
		}
	}
	if isLocal {
		g.promptText.SetText(str)
		g.promptText.SetPos((714-g.promptText.Width)/2+290, 545)
		g.promptTextOther.SetText("")
	} else {
		p := g.getPlayer(g.curPlayer)
		g.promptText.SetText("")
		g.promptTextOther.SetText(str)
		g.promptTextOther.SetPos(p.x+20, p.y+235)
	}
}

func (g *Games) Update(app *app.App) {
	//检查退出按钮
	g.quitBtn.Update(g)
	select {
	case <-g.closeSignal:
		g.app.CurMenu = NewMianMenu(g.app)
		g.bgm.Close()
		g.app.Bgm.Play()
		return
	default:
	}
	//检查点击以关闭图鉴
	if g.showRoleDetail {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.showRoleDetail = false
		}
	} else {
		//若当前玩家不为自己，检查点击玩家以显示图鉴
		if g.curPlayer != g.pid {
			for i := 0; i < len(g.playList); i++ {
				if g.playList[i].isClicked() {
					g.setRoleDetail(g.playList[i].sidList)
					g.roleDetailNum = data.PID(i)
					g.showRoleDetail = true
					break
				}
			}
		}
	}
	//更新进度条
	if g.processInf.remain > 0 {
		g.processInf.rate = g.processInf.remain / g.processInf.total
		g.processInf.remain--
	} else {
		g.processInf.rate = 0
	}
	//完成动画计算
	g.anime()
	switch g.state {
	case data.InitState:
		g.handleInit()
	case data.UseCardState:
		g.handleUseCard()
	case data.DropCardState:
		g.handleDropCard()
	case data.DyingState:
		g.handleDying()
	case data.WXKJState:
		g.handleWXKJ()
	case data.DodgeState:
		g.handleDodge()
	case data.DuelState:
		g.handleDuel()
	case data.NMRQState:
		g.handleNMRQ()
	case data.WJQFState:
		g.handleWJQF()
	case data.TYJYState:
		g.handleTYJY()
	case data.WGFDState:
		g.handleWGFD()
	case data.BurnShowState:
		g.handleBurnShow()
	case data.BurnDropState:
		g.handleBurnDrop()
	case data.SSQYState:
		g.handleSSQY()
	case data.GHCQState:
		g.handleGHCQ()
	case data.JDSRState:
		g.handleJDSR()
	case data.LuanWuState:
		g.handleLuanWu()
	case data.DropSelfAllCards:
		g.handleDropSelfAll()
	case data.GSFState:
		g.handleGSF()
	case data.QlYYDState:
		g.handleQLYYD()
	case data.SkillSelectState:
		g.handleSkillSelectState()
	case data.QLGState:
		g.handleQLG()
	case data.CXSGJState:
		g.handleCXSGJ()
	case data.QueDiState:
		g.handleQueDi()
	case data.DropOtherCardState:
		g.handleDropOtherCard()
	case data.DropSelfHandCard:
		g.handleDropCard()
	case data.PoJunState:
		g.handlePoJun()
	case data.FanKuiState:
		g.handleFanKui()
	case data.WenJiState:
		g.handleWenJi()
	case data.EnYuanState:
		g.handleBurnDrop()
	case data.GuanXingState:
		g.handleGuanXing()
	case data.QinYinState:
		g.handleQinYin()
	case data.YingHunState:
		g.handleYingHun()
	case data.LiyuState:
		g.handleLiYu()
	case data.ChangeRoleState:
		g.handleChangeRole()
	case data.FengPoState:
		g.handleFengPo()
	case data.LiangZhuState:
		g.handleLiangZhu()
	case data.GetSkillState:
		g.handleGetSkill()
	}
	//检查发牌
	select {
	case rec := <-g.cardReceiver:
		g.addCard(rec)
	default:
	}
	//检查用牌
	select {
	case rec := <-g.useReceiver:
		g.useCard(rec)
	default:
	}
	//检查游戏技能卡牌
	select {
	case inf := <-g.gsidCardReceiver:
		g.handleGSIDCardsInf(inf)
	default:
	}
	//检查弃牌
	select {
	case rec := <-g.removeReceiver:
		g.removeCard(rec)
	default:
	}
	//检查移动卡牌
	select {
	case inf := <-g.moveCardRec:
		g.moveCard(inf)
	default:
	}
	//检查使用技能
	select {
	case inf := <-g.useSkillRec:
		g.useSkill(inf)
	default:
	}
	//检查使用临时卡
	select {
	case inf := <-g.useTmpRec:
		g.useTmpCard(inf)
	default:
	}
	//检查当前回合拥有者
	select {
	case turnOwner := <-g.turnOwnerRec:
		g.getPlayer(g.turnOwner).isTurnowner = false
		g.turnOwner = turnOwner
		g.getPlayer(g.turnOwner).isTurnowner = true
	default:
	}
	//检查切阶段指令
	select {
	case inf := <-g.gameStateRec:
		g.switchState(inf)
	default:
	}
}

func (g *Games) handleInit() {
	if g.pid == -1 {
		g.pid = <-g.pidReceiver
		//初始化背景音乐
		g.bgm = sound.NewBgm("assets/audio/ingame.mp3")
		g.bgm.Play()
		//初始化牌堆
		g.cards = newCardList(g)
		g.promptText.SetText("[white]你是 [blue]" + strconv.Itoa(int(g.pid)+1) + "[white] 号位，正在等待其他玩家...")
		g.promptText.SetPos((1334-g.promptText.Width)/2, 545)
	}
	if g.roleList == nil {
		select {
		case g.roleList = <-g.availableRoleRec:
			var skillBgm *sound.Player
			getOnclick := func(b *roleSelBtn, role *data.Role) func(*Games) {
				return func(g *Games) {
					b.resetOthers(g)
					b.isSelected = true
					g.selRole = role
					//点击角色时播放音频
					for _, s := range role.SkillList {
						if a := sound.GetSkillAudio(s); a != nil {
							if skillBgm != nil {
								skillBgm.Close()
							}
							skillBgm = sound.NewAudioPlayer(a)
							skillBgm.Play()
							break
						}
					}
				}
			}
			x, y := 148.0, 630.0
			for i := 0; i < len(g.roleList); i++ {
				b := newRoleSelBtn(x, y, getPlayerImg("assets/role/smallChar/"+g.roleList[i].Name+".png"),
					g.roleList[i], nil)
				b.setOnClick(getOnclick(b, &g.roleList[i]))
				g.btnList = append(g.btnList, b)
				if i%8 == 7 {
					y -= 130
					x = 148
				} else {
					x += 130
				}
			}
			g.btnList = append(g.btnList, newSelConfirmBtn(func(g *Games) {
				g.roleInf <- *g.selRole
				g.hasSkip = true
				g.btnList = nil
				g.promptText.SetText("[white]正在等待其他玩家...")
				g.promptText.SetPos((1334-g.promptText.Width)/2, 545)
			}))
			g.selRole = &g.roleList[0]
			g.btnList[0].(*roleSelBtn).isSelected = true
			g.promptText.SetText("[white]你是 [blue]" + strconv.Itoa(int(g.pid)+1) + "[white] 号位，请选择武将")
			g.promptText.SetPos((1334-g.promptText.Width)/2, 545)
		default:
		}
	}
	if !g.hasSkip {
		//选将阶段的更新
		for i := 0; i < len(g.btnList); i++ {
			g.btnList[i].Update(g)
		}
	}
	//选完人后等待信息
	select {
	case pList := <-g.playerInfRec:
		posList := [][]struct{ x, y float64 }{
			{{x: 1120, y: 500},
				{x: 720, y: 20},
				{x: 300, y: 20}},
			{{x: 1120, y: 500},
				{x: 1100, y: 130},
				{x: 560, y: 20},
				{x: 20, y: 130}},
			{{x: 1120, y: 500},
				{x: 1100, y: 130},
				{x: 830, y: 30},
				{x: 560, y: 20},
				{x: 290, y: 30},
				{x: 20, y: 130}},
		}
		var index int
		switch len(pList) {
		case 3:
			index = 0
		case 4:
			index = 1
		case 6:
			index = 2
		}
		for i := 0; i < len(pList); i++ {
			g.playList = append(g.playList, newPlayer(g, pList[i].Role, pList[i].PID))
		}
		g.playList = append(g.playList[g.pid:], g.playList[:g.pid]...)
		g.playList[0].isLocal = true
		for i, p := range g.playList {
			p.setPos(posList[index][i].x, posList[index][i].y)
		}
		for i := 0; i < len(g.playList); i++ {
			g.addCard(<-g.cardReceiver)
		}
		g.promptText.SetText("")
		//g.switchState(<-g.gameStateRec)
	default:
	}
}

func (g *Games) handleUseCard() {
	if g.curPlayer != g.pid {
		return
	}
	g.getPlayer(g.pid).handleUseState(g)

}

func (g *Games) handleDropCard() {
	if g.curPlayer != g.pid {
		return
	}
	g.getPlayer(g.pid).handleDropState(g)
}

func (g *Games) handleDying() {
	if g.curPlayer != g.pid {
		return
	}
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleDying(g)
}

func (g *Games) handleWXKJ() {
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleWXKJ(g)
}

func (g *Games) handleDodge() {
	if g.curPlayer != g.pid {
		return
	}
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleDodge(g)
}

func (g *Games) handleDuel() {
	if g.curPlayer != g.pid {
		return
	}
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleDuel(g)
}

func (g *Games) handleNMRQ() {
	if g.curPlayer != g.pid {
		return
	}
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleNMRQ(g)
}

func (g *Games) handleWJQF() {
	if g.curPlayer != g.pid {
		return
	}
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleWJQF(g)
}

func (g *Games) handleTYJY() {}

func (g *Games) handleWGFD() {
	if g.curPlayer != g.pid {
		return
	}
	for _, c := range g.wgfdCards {
		c.handleWGFD(g)
	}
}

func (g *Games) handleBurnShow() {
	if g.curPlayer != g.pid {
		return
	}
	g.getPlayer(g.pid).handleBurnShow(g)
}

func (g *Games) handleBurnDrop() {
	if g.curPlayer != g.pid {
		return
	}
	if g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleBurnDrop(g)
}

func (g *Games) handleSSQY() {
	if g.curPlayer != g.pid {
		return
	}
	for _, c := range g.ssqyCards {
		c.handleSSQY(g)
	}
}

func (g *Games) handleGHCQ() {
	if g.curPlayer != g.pid {
		return
	}
	for _, c := range g.ghcqCards {
		c.handleGHCQ(g)
	}
}

func (g *Games) handleJDSR() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleJDSR(g)
}

func (g *Games) handleLuanWu() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleLuanWu(g)
}

func (g *Games) handleQLYYD() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleQLYYD(g)
}

func (g *Games) handleSkillSelectState() {
	if g.curPlayer != g.pid {
		return
	}
	g.getPlayer(g.pid).handleSkillSelectState(g)
}

func (g *Games) handleDropSelfAll() {
	if g.curPlayer != g.pid {
		return
	}
	g.getPlayer(g.pid).handleDropSelfAll(g)
}

func (g *Games) handleGSF() {
	if g.curPlayer != g.pid {
		return
	}
	g.getPlayer(g.pid).handleGSF(g)
}

func (g *Games) handleQLG() {
	if g.curPlayer != g.pid {
		return
	}
	for _, c := range g.qlgCards {
		c.handleQLG(g)
	}
}

func (g *Games) handleCXSGJ() {
	if g.curPlayer != g.pid {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		inf := <-g.useSkillRspRec
		if inf.Args[0] == 1 {
			g.btnList = append(g.btnList, newChooseLeftBtn("弃一张手牌", func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.CXSGJSkill, Args: []byte{0}}
				g.hasSkip = true
				g.btnList = nil
			}))
		} else {
			g.btnList = append(g.btnList, newFakeChooseBtn("弃一张手牌", 350, 480, func(g *Games) {
				g.hasSkip = true
				g.btnList = nil
			}))
		}
		g.btnList = append(g.btnList, newChooseRightBtn("让其摸一张牌", func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.CXSGJSkill, Args: []byte{1}}
			g.hasSkip = true
			g.btnList = nil
		}),
		)
	}
}

func (g *Games) handleQueDi() {
	if g.curPlayer != g.pid {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		inf := <-g.useSkillRspRec
		if inf.Args[0] == 1 {
			g.btnList = append(g.btnList, newChooseLeftBtn("获得其手牌", func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.QueDiSkill, Args: []byte{1, 0}}
				g.hasSkip = true
				g.btnList = nil
			}))
		} else {
			g.btnList = append(g.btnList, newFakeChooseBtn("获得其手牌", 350, 480, func(g *Games) {
				g.hasSkip = true
				g.btnList = nil
			}))
		}
		if inf.Args[1] == 1 {
			g.btnList = append(g.btnList, newChooseRightBtn("弃牌加伤害", func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.QueDiSkill, Args: []byte{0, 1}}
				g.hasSkip = true
				g.btnList = nil
			}))
		} else {
			g.btnList = append(g.btnList, newFakeChooseBtn("弃牌加伤害", 650, 480, func(g *Games) {
				g.hasSkip = true
				g.btnList = nil
			}))
		}
		if inf.Args[2] == 1 {
			g.btnList = append(g.btnList, newBeiShuiBtn("背水", true, func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.QueDiSkill, Args: []byte{1, 1}}
				g.hasSkip = true
				g.btnList = nil
			}))
		} else {
			g.btnList = append(g.btnList, newBeiShuiBtn("背水", false, func(g *Games) {
				g.hasSkip = true
				g.btnList = nil
			}))
		}
	}
}

func (g *Games) handleLiangZhu() {
	if g.curPlayer != g.pid {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList,
			newChooseLeftBtn("其摸两张牌", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: false}
				g.hasSkip = true
				g.btnList = nil
			}),
			newChooseRightBtn("你摸一张牌", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: true}
				g.hasSkip = true
				g.btnList = nil
			}),
		)
	}
}

func (g *Games) handleFengPo() {
	if g.curPlayer != g.pid {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList,
			newChooseLeftBtn("加摸牌", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: false}
				g.hasSkip = true
				g.btnList = nil
			}),
			newChooseRightBtn("加伤害", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: true}
				g.hasSkip = true
				g.btnList = nil
			}),
		)
	}
}

func (g *Games) handleQinYin() {
	if g.curPlayer != g.pid {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList,
			newChooseLeftBtn("流失一点体力", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: true}
				g.hasSkip = true
				g.btnList = nil
			}),
			newChooseRightBtn("回复一点体力", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: false}
				g.hasSkip = true
				g.btnList = nil
			}),
		)
	}
}

func (g *Games) handleYingHun() {
	if g.curPlayer != g.pid {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		p := g.getPlayer(g.pid)
		x := strconv.Itoa(int(p.maxHp - p.hp))
		g.btnList = append(g.btnList,
			newChooseLeftBtn("摸1弃"+x, func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: true}
				g.hasSkip = true
				g.btnList = nil
			}),
			newChooseRightBtn("摸"+x+"弃1", func(g *Games) {
				g.useCardInf <- data.UseCardInf{Skip: false}
				g.hasSkip = true
				g.btnList = nil
			}),
		)
	}
}

func (g *Games) handleDropOtherCard() {
	if g.curPlayer != g.pid {
		return
	}
	for _, c := range g.otherCards {
		c.handleDropOtherCard(g)
	}
}

func (g *Games) handlePoJun() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList, newFakeSkillConfirmBtn(nil))
	}
	if len(g.selcard) > 0 {
		if len(g.btnList) == 1 {
			g.btnList = append(g.btnList, newSkillConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.PoJunSkill, Cards: g.selcard}
				g.hasSkip = true
				g.btnList = nil
				g.selcard = nil
			}))
		}
	} else {
		g.btnList = g.btnList[:1]
	}
	for _, c := range g.otherCards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.setSelect(false)
			g.selcard = delFromList(g.selcard, func(cid data.CID) bool { return cid == c.getID() })
			continue
		}
		c.setSelect(true)
		g.selcard = append(g.selcard, c.getID())
		if len(g.selcard) > g.selNum {
			for _, c1 := range g.otherCards {
				if c1.getID() == g.selcard[0] {
					c1.setSelect(false)
					break
				}
			}
			g.selcard = g.selcard[1:]
		}
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
}

func (g *Games) handleFanKui() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList, newFakeSkillConfirmBtn(nil))
	}
	if len(g.selcard) == 1 {
		if len(g.btnList) == 1 {
			g.btnList = append(g.btnList, newSkillConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.FanKuiSkill, Cards: g.selcard}
				g.hasSkip = true
				g.btnList = nil
				g.selcard = nil
				return
			}))
		}
	} else {
		g.btnList = g.btnList[:1]
	}
	for _, c := range g.otherCards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.setSelect(false)
			g.selcard = nil
			continue
		}
		c.setSelect(true)
		g.selcard = []data.CID{c.getID()}
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
}

func (g *Games) handleWenJi() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	g.getPlayer(g.pid).handleWenJi(g)
}

func (g *Games) handleGuanXing() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	for i, c := range g.heapTop {
		c.anime()
		if !c.isClicked() {
			continue
		}
		g.heapTop = append(g.heapTop[:i], g.heapTop[i+1:]...)
		g.heapButtom = append(g.heapButtom, c)
		calculateCardPos(321, 172, 141, 705, false, g.heapTop...)
		calculateCardPos(321, 365, 141, 705, false, g.heapButtom...)
		return
	}
	for i, c := range g.heapButtom {
		c.anime()
		if !c.isClicked() {
			continue
		}
		g.heapButtom = append(g.heapButtom[:i], g.heapButtom[i+1:]...)
		g.heapTop = append(g.heapTop, c)
		calculateCardPos(321, 172, 141, 705, false, g.heapTop...)
		calculateCardPos(321, 365, 141, 705, false, g.heapButtom...)
		return
	}
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList, newSkillConfirmBtn(func(g *Games) {
			cards := []data.CID{}
			for _, c := range g.heapTop {
				cards = append(cards, c.getID())
			}
			for _, c := range g.heapButtom {
				cards = append(cards, c.getID())
			}
			g.useSkillInf <- data.UseSkillInf{ID: data.GuanXingSkill, Cards: cards, Args: []byte{byte(len(g.heapTop))}}
			g.hasSkip = true
			g.btnList = nil
			g.heapButtom = nil
			g.heapTop = nil

		}))
	}
	for _, b := range g.btnList {
		b.Update(g)
	}
}

func (g *Games) handleLiYu() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	p := g.getPlayer(g.pid)
	if len(g.btnList) == 0 {
		g.btnList = append(g.btnList, newFakeConfirmBtn(nil))
	}
	if len(g.selcard) > 0 {
		if len(g.btnList) == 1 {
			g.btnList = append(g.btnList, newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.LiYuSkill, Cards: g.selcard}
				g.hasSkip = true
				g.btnList = nil
				g.selcard = nil
			}))
		}
	} else {
		g.btnList = g.btnList[:1]
	}
	for _, c := range p.cards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.setSelect(false)
			g.selcard = delFromList(g.selcard, func(cid data.CID) bool { return cid == c.getID() })
			continue
		}
		c.setSelect(true)
		g.selcard = append(g.selcard, c.getID())
		if len(g.selcard) > g.selNum {
			for _, c1 := range p.cards {
				if c1.getID() == g.selcard[0] {
					c1.setSelect(false)
					break
				}
			}
			g.selcard = g.selcard[1:]
		}
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
}

func (g *Games) handleChangeRole() {
	if g.hasSkip {
		return
	}
	pList := <-g.playerInfRec
	for _, inf := range pList {
		index := int(inf.PID - g.pid)
		if index < 0 {
			index += len(g.playList)
		}
		p := newPlayer(g, inf.Role, inf.PID)
		p.setPos(g.playList[index].x, g.playList[index].y)
		g.playList[index] = p
	}
	g.hasSkip = true
}

func (g *Games) handleGetSkill() {
	if g.curPlayer != g.pid || g.hasSkip {
		return
	}
	for i := 0; i < len(g.btnList); i++ {
		g.btnList[i].Update(g)
	}
	if g.hasSkip {
		return
	}
	p := g.getPlayer(g.curPlayer)
	if len(g.btnList) == 0 {
		inf := <-g.useSkillRspRec
		g.btnList = append(g.btnList, newChooseLeftBtn("获得 "+data.SID(inf.Args[0]).Name(), func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.GetSkill, Args: []byte{0}}
			s := newSkillI(g, g.pid, data.SID(inf.Args[0]))
			skills := g.getPlayer(g.pid).skills
			_, y := skills[len(skills)-1].getPos()
			s.setPos(1050, y-30)
			p.skills = append(p.skills, s)
			g.hasSkip = true
			g.btnList = nil
		}))
		g.btnList = append(g.btnList, newChooseRightBtn("获得 "+data.SID(inf.Args[1]).Name(), func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.GetSkill, Args: []byte{1}}
			s := newSkillI(g, g.pid, data.SID(inf.Args[1]))
			skills := g.getPlayer(g.pid).skills
			_, y := skills[len(skills)-1].getPos()
			s.setPos(1050, y-30)
			p.skills = append(p.skills, s)
			g.hasSkip = true
			g.btnList = nil
		}),
		)
	}
}

func (g *Games) setLittleHeapNum(num uint8) {
	g.littleHeapNum = num
	g.littleHeaptext.SetText(strconv.Itoa(int(g.littleHeapNum)))
}
