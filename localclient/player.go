package localclient

import (
	"goltk/data"
	"goltk/front"
	"image"
	"image/color"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

var playerImgBuf = map[string]*ebiten.Image{}

// 获取角色图像
func getPlayerImg(path string) *ebiten.Image {
	img, ok := playerImgBuf[path]
	if ok {
		return img
	}
	i := loadImg(path)
	playerImgBuf[path] = i
	return i
}

var playerImgs = struct {
	bgFaction      [10]*ebiten.Image
	smallFaction   [10]*ebiten.Image
	dragonImg      [4]*ebiten.Image
	smallHandle    *ebiten.Image
	hpImg          [3]*ebiten.Image
	hpxxImg        *ebiten.Image
	hpmaxImg       *ebiten.Image
	frame          *ebiten.Image
	isSelectImg    [12]*ebiten.Image
	inturnImg      [12]*ebiten.Image
	isDrunkImg     *ebiten.Image
	dyingImg       *ebiten.Image
	dyingImg2      *ebiten.Image
	dieImg         *ebiten.Image
	linkImg        *ebiten.Image
	bleendImg      *ebiten.Image
	nordmgImg      [10]*ebiten.Image
	firedmgImg     [10]*ebiten.Image
	lightndmgImg   [10]*ebiten.Image
	bleedingdmgImg [10]*ebiten.Image
	recoverImg     [7]*ebiten.Image
	lbssItemImg    *ebiten.Image
	blcdItemImg    *ebiten.Image
	lightnItemImg  *ebiten.Image
	turnBackImg    *ebiten.Image
	skillBoxImg    *ebiten.Image
	otherStateBg   *ebiten.Image
	lbssImg        [12]*ebiten.Image
	blcdImg        [12]*ebiten.Image
	lightningImg   [3]*ebiten.Image
	skillImg       [10]*ebiten.Image
	atkImg         [13]*ebiten.Image
	dodgeImg       [13]*ebiten.Image
	peachImg       [13]*ebiten.Image
	drunkImg       [13]*ebiten.Image
}{
	smallHandle:   loadImg("assets/role/smallHandle.png"),
	hpmaxImg:      loadImg("assets/player/hp/hpmax.png"),
	hpxxImg:       loadImg("assets/player/hp/hpxx.png"),
	frame:         loadImg("assets/role/frame.png"),
	isDrunkImg:    loadImg("assets/player/drunk.png"),
	dyingImg:      loadImg("assets/player/dying.png"),
	dyingImg2:     loadImg("assets/player/dying2.png"),
	dieImg:        loadImg("assets/player/die.png"),
	linkImg:       loadImg("assets/player/tiesuo.png"),
	lbssItemImg:   loadImg("assets/player/lbssItem.png"),
	blcdItemImg:   loadImg("assets/player/blcditem.png"),
	lightnItemImg: loadImg("assets/player/lightnItem.png"),
	turnBackImg:   loadImg("assets/player/turnBack.png"),
	skillBoxImg:   loadImg("assets/player/skillBox.png"),
	bleendImg:     loadImg("assets/player/bleed.png"),
	otherStateBg:  loadImg("assets/player/otherStateBg.png"),
}

func init() {
	for i := 0; i < 10; i++ {
		playerImgs.bgFaction[i] = loadImg("assets/role/faction/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < 10; i++ {
		playerImgs.smallFaction[i] = loadImg("assets/role/smallfaction/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < 4; i++ {
		playerImgs.dragonImg[i] = loadImg("assets/role/dragon/" + strconv.Itoa(i+1) + ".png")
	}
	for i := 0; i < len(playerImgs.isSelectImg); i++ {
		playerImgs.isSelectImg[i] = loadImg("assets/player/isSelected/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.inturnImg); i++ {
		playerImgs.inturnImg[i] = loadImg("assets/player/useState/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.hpImg); i++ {
		playerImgs.hpImg[i] = loadImg("assets/player/hp/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.nordmgImg); i++ {
		playerImgs.nordmgImg[i] = loadImg("assets/player/nordmg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.firedmgImg); i++ {
		playerImgs.firedmgImg[i] = loadImg("assets/player/firedmg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.lightndmgImg); i++ {
		playerImgs.lightndmgImg[i] = loadImg("assets/player/lightndmg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.bleedingdmgImg); i++ {
		playerImgs.bleedingdmgImg[i] = loadImg("assets/player/bleedingdmg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.recoverImg); i++ {
		playerImgs.recoverImg[i] = loadImg("assets/player/recover/lizi" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.lbssImg); i++ {
		playerImgs.lbssImg[i] = loadImg("assets/player/lbssanime/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.blcdImg); i++ {
		playerImgs.blcdImg[i] = loadImg("assets/player/blcdanime/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.lightningImg); i++ {
		playerImgs.lightningImg[i] = loadImg("assets/player/lightning/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.skillImg); i++ {
		playerImgs.skillImg[i] = loadImg("assets/player/skill/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.atkImg); i++ {
		playerImgs.atkImg[i] = loadImg("assets/player/cardImg/atkImg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.dodgeImg); i++ {
		playerImgs.dodgeImg[i] = loadImg("assets/player/cardImg/dodgeImg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.peachImg); i++ {
		playerImgs.peachImg[i] = loadImg("assets/player/cardImg/peachImg/" + strconv.Itoa(i) + ".png")
	}
	for i := 0; i < len(playerImgs.drunkImg); i++ {
		playerImgs.drunkImg[i] = loadImg("assets/player/cardImg/drunkImg/" + strconv.Itoa(i) + ".png")
	}
}

type player struct {
	rect
	g           *Games
	img         *ebiten.Image
	unSelAble   bool
	isSelected  bool
	isTurnowner bool
	pid         data.PID
	isLocal     bool //是否是本地玩家
	x, y        float64
	hp          data.HP
	maxHp       data.HP
	name        string
	dspName     string
	side        data.RoleSide
	dragon      uint8
	male        bool
	judgeSlot   [3]cardI
	equipSlot   [4]*equipCard
	skills      []skillI
	sidList     []data.SID
	cards       []cardI
	selCard     []data.CID
	tsCard      map[data.SID][]cardI //技能暂存的卡
	//动画元素
	stateAnime struct {
		index   int
		counter int
	}
	dyingAnime struct {
		index   float32
		counter float32
	}
	tslhAnime struct {
		index  float32
		enable bool
	}
	dmgAnime struct {
		enable  bool
		index   int
		counter int
		hptype  data.SetHpType
	}
	recAnime struct {
		enable bool
		index  int
		count  int
	}
	delayTipsAnime struct {
		enable     bool
		index      int
		counter    int
		visibility float32
		cardName   data.CardName
	}
	skillAnime struct {
		enable     bool
		index      int
		counter    int
		visibility float32
		name       *ebiten.Image
	}
	basicCardAnime struct {
		enable       bool
		index, count int
		name         data.CardName
	}
	skillBoxs    map[data.SID]skillBoxI
	btnGrup      []buttonI
	cardsText    *front.TextItem //debug 卡片文本
	pidText      *front.TextItem2
	isDrunk      bool
	dying        bool
	death        bool
	isLinked     bool
	turnBack     bool //翻面
	enableSkills map[data.SID]struct{}
	cardNumText  *front.TextItem2
}

func newPlayer(g *Games, r data.Role, pid data.PID) *player {
	p := &player{
		g:            g,
		img:          getPlayerImg("assets/role/charImg/" + r.Name + ".png"),
		rect:         rect{wide: 175, height: 233},
		pid:          pid,
		hp:           r.MaxHP,
		maxHp:        r.MaxHP,
		name:         r.Name,
		dspName:      r.DspName,
		side:         r.Side,
		male:         !r.Female,
		sidList:      r.SkillList,
		dragon:       r.Dragon,
		tsCard:       map[data.SID][]cardI{},
		enableSkills: map[data.SID]struct{}{},
		cardsText:    front.NewTextItem("", 0, 0, 32, false, 24, color.White), //debug 卡片文本
		skillBoxs:    map[data.SID]skillBoxI{},
		cardNumText:  front.NewTextItem2("[#ffffff]0", 0, 0, 32, 4, 32),
		pidText:      front.NewTextItem2("[black]"+pid.Name(), 0, 0, 28, 2, 0),
	}
	if pid != g.pid {
		return p
	}
	x, y := 1000., 730.
	for i, sid := range r.SkillList {
		s := newSkillI(g, pid, sid)
		if s.isActiveSkill() {
			x = 1020
			y -= 50
			if i == 0 {
				y = 655
			}
			s.setPos(x, y)
			p.skills = append(p.skills, s)
			continue
		}
		if i%2 == 0 {
			y -= 34
			x = 1050
		} else {
			x -= 65
		}
		s.setPos(x, y)
		p.skills = append(p.skills, s)
	}
	return p
}

func (p *player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	//如果当前为自己的回合则绘制率框
	if p.isTurnowner {
		op.GeoM.Reset()
		op.GeoM.Translate(p.x-14, p.y-10)
		screen.DrawImage(playerImgs.inturnImg[p.stateAnime.index], op)
	}
	//绘制被选中
	if p.isSelected {
		op.GeoM.Reset()
		op.GeoM.Translate(p.x-14, p.y-10)
		screen.DrawImage(playerImgs.isSelectImg[p.stateAnime.index], op)
	}
	if p.unSelAble {
		op.ColorScale.Scale(0.5, 0.5, 0.5, 1)
	}
	if p.death {
		op.ColorScale.Scale(0.4, 0.4, 0.4, 1)
	}
	//绘制阴影
	op.GeoM.Reset()
	op.GeoM.Translate(p.x-3, p.y-3)
	screen.DrawImage(playerImgs.frame, op)
	//绘制背景
	op.GeoM.Translate(3, 3)
	screen.DrawImage(playerImgs.bgFaction[p.side], op)
	//计算卡面坐标并绘制
	op.GeoM.Reset()
	op.GeoM.Translate(p.getSubItemPos(100, 101, p.img.Bounds().Size()))
	screen.DrawImage(p.img, op)
	//画龙
	if p.dragon != 0 {
		op.GeoM.Reset()
		op.GeoM.Translate(p.x-8, p.y-40)
		screen.DrawImage(playerImgs.dragonImg[p.dragon-1], op)
	}
	//绘制小阵营
	op.GeoM.Reset()
	op.GeoM.Translate(p.getSubItemPos(164, 18, playerImgs.smallFaction[p.side].Bounds().Size()))
	screen.DrawImage(playerImgs.smallFaction[p.side], op)
	if p.death {
		goto drawName
	}
	//绘制小卡组
	op.GeoM.Reset()
	op.GeoM.Translate(p.getSubItemPos(22, 218, playerImgs.smallHandle.Bounds().Size()))
	screen.DrawImage(playerImgs.smallHandle, op)
	p.cardNumText.Draw(screen)
	//绘制其他玩家阶段文本背景
	if p.g.curPlayer == p.pid && p.g.pid != p.pid {
		op.GeoM.Reset()
		op.GeoM.Translate(p.x, p.y+233)
		screen.DrawImage(playerImgs.otherStateBg, op)
	}
	//绘制进度条
	if (p.pid == p.g.curPlayer || p.g.state == data.WXKJState) && p.g.processInf.rate > 0 {
		var x, y, dx, dy float64
		var timerBar, timerBg *ebiten.Image
		if p.isLocal {
			x, y = 290, 550
			dx, dy = 6, 1
			timerBar = gameImg.processBar
			timerBg = gameImg.processBg
		} else {
			x, y = p.x+15, p.y+232
			timerBar = gameImg.processBarOther
			timerBg = gameImg.processBarotherBg
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		screen.DrawImage(timerBg, op)
		op.GeoM.Reset()
		op.GeoM.Scale(p.g.processInf.rate, 1)
		op.GeoM.Translate(x+dx, y+dy)
		screen.DrawImage(timerBar, op)
	}
	//绘制血上限
	if p.maxHp < 7 {
		op.GeoM.Reset()
		op.GeoM.Translate(p.x+8, p.y+199)
		for i := 0; i < int(p.maxHp); i++ {
			op.GeoM.Translate(0, float64(-20))
			screen.DrawImage(playerImgs.hpmaxImg, op)
		}
	}
	//绘制血
	if p.maxHp < 7 {
		op.GeoM.Reset()
		op.GeoM.Translate(p.x+8, p.y+200)
		var hpimg *ebiten.Image
		switch p.hp {
		case 1:
			hpimg = playerImgs.hpImg[0]
		case 2:
			hpimg = playerImgs.hpImg[1]
		default:
			hpimg = playerImgs.hpImg[2]
		}
		for i := 0; i < int(p.hp); i++ {
			op.GeoM.Translate(0, float64(-20))
			screen.DrawImage(hpimg, op)
		}
	} else {
		hpimg := playerImgs.hpImg[2]
		op.GeoM.Reset()
		op.GeoM.Translate(p.x+8, p.y+130)
		screen.DrawImage(hpimg, op)
		op.GeoM.Translate(-4, 18)
		screen.DrawImage(playerImgs.hpxxImg, op)
		txet := strconv.Itoa(int(p.hp))
		if p.hp > 9 {
			front.DrawText(screen, txet, p.x+6, p.y+170, 24, false)
		} else {
			front.DrawText(screen, txet, p.x+9, p.y+170, 24, false)
		}
	}
drawName:
	//绘制名字
	front.DrawText(screen, p.dspName, p.x+17, p.y+34, 20, true)
	for i := 0; i < len(p.btnGrup); i++ {
		p.btnGrup[i].Draw(screen)
	}
	//绘制号位
	if p.pid == p.g.pid {
		p.pidText.SetPos(p.x+80, p.y+210)
	} else {
		p.pidText.SetPos(p.x+80, p.y+234)
	}
	p.pidText.Draw(screen)
	//绘制游戏元素
	{
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Reset()
		var dx float64
		if p.isLocal {
			op.GeoM.Translate(p.x-50, p.y+70)
			dx = -16
		} else {
			op.GeoM.Translate(p.x+10, p.y+235)
			dx = 16
		}
		//绘制延时锦囊小标识
		if p.judgeSlot[data.LBSSSlot] != nil {
			screen.DrawImage(playerImgs.lbssItemImg, op)
			op.GeoM.Translate(dx, 0)
		}
		if p.judgeSlot[data.BLCDSlot] != nil {
			screen.DrawImage(playerImgs.blcdItemImg, op)
			op.GeoM.Translate(dx, 0)
		}
		if p.judgeSlot[data.LightningSlot] != nil {
			screen.DrawImage(playerImgs.lightnItemImg, op)
			op.GeoM.Translate(dx, 0)
		}
		//绘制酒效果
		if p.isDrunk {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x-8, p.y)
			screen.DrawImage(playerImgs.isDrunkImg, op)
		}
		//绘制翻面效果
		if p.turnBack {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x+30, p.y)
			screen.DrawImage(playerImgs.turnBackImg, op)
		}
		//绘制连环
		if p.tslhAnime.enable {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x-10, p.y+60)
			wide := playerImgs.linkImg.Bounds().Dx()
			x1 := int(float32(wide) * p.tslhAnime.index)
			img := playerImgs.linkImg.SubImage(image.Rect(0, 0, x1, 56)).(*ebiten.Image)
			screen.DrawImage(img, op)
		}
		//绘制技能计数槽
		{
			skills := make([]data.SID, len(p.skillBoxs))
			i := 0
			for k := range p.skillBoxs {
				skills[i] = k
				i++
			}
			sort.Slice(skills, func(i, j int) bool { return skills[i] < skills[j] })
			x, y := p.x+31, p.y+100
			for _, s := range skills {
				p.skillBoxs[s].Draw(x, y, screen)
				y += 35
			}
		}
		//绘制延时锦囊效果
		if p.delayTipsAnime.enable {
			op.GeoM.Reset()
			op.ColorScale.SetA(p.skillAnime.visibility)
			op.GeoM.Translate(p.x, p.y+20)
			if p.delayTipsAnime.cardName == data.LBSS {
				screen.DrawImage(playerImgs.lbssImg[p.delayTipsAnime.index], op)

			} else if p.delayTipsAnime.cardName == data.BLCD {
				screen.DrawImage(playerImgs.blcdImg[p.delayTipsAnime.index], op)
			} else {
				if p.delayTipsAnime.index > 2 {
					p.delayTipsAnime.enable = false
				} else {
					op.GeoM.Translate(-40, -140)
					screen.DrawImage(playerImgs.lightningImg[p.delayTipsAnime.index], op)
				}
			}
		}
		//绘制伤害
		if p.dmgAnime.enable {
			op.GeoM.Reset()
			op.ColorScale.SetA(1 - float32(p.dmgAnime.index)/10)
			op.GeoM.Translate(p.x, p.y)
			screen.DrawImage(playerImgs.bleendImg, op)
			var img *ebiten.Image
			switch p.dmgAnime.hptype {
			case data.NormalDmg:
				img = playerImgs.nordmgImg[p.dmgAnime.index]
			case data.FireDmg:
				img = playerImgs.firedmgImg[p.dmgAnime.index]
			case data.LightningDmg:
				img = playerImgs.lightndmgImg[p.dmgAnime.index]
			case data.BleedingDmg:
				img = playerImgs.bleedingdmgImg[p.dmgAnime.index]
			default:
			}
			op.GeoM.Reset()
			op.GeoM.Translate(p.x, p.y)
			screen.DrawImage(img, op)
		}
		//回血
		if p.recAnime.enable {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x-30, p.y+162)
			if p.maxHp < 7 {
				op.GeoM.Translate(0, float64(p.hp)*float64(-20))
			} else {
				op.GeoM.Translate(0, -76)
			}
			screen.DrawImage(playerImgs.recoverImg[p.recAnime.index], op)
		}
		//绘制濒死
		if p.dying {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x-8, p.y)
			screen.DrawImage(playerImgs.dyingImg2, op)
			op.GeoM.Reset()
			op.GeoM.Translate(p.x+70, p.y+30)
			op.ColorScale.ScaleAlpha(p.dyingAnime.counter)
			screen.DrawImage(playerImgs.dyingImg, op)
		}
		//绘制死亡
		if p.death {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x+30, p.y+22)
			screen.DrawImage(playerImgs.dieImg, op)
		}
		//绘制基本牌特效
		if p.basicCardAnime.enable {
			op.GeoM.Reset()
			op.GeoM.Translate(p.x-26, p.y-9)
			if isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack}, p.basicCardAnime.name) {
				screen.DrawImage(playerImgs.atkImg[p.basicCardAnime.index], op)
			} else {
				switch p.basicCardAnime.name {
				case data.Dodge:
					screen.DrawImage(playerImgs.dodgeImg[p.basicCardAnime.index], op)
				case data.Peach:
					screen.DrawImage(playerImgs.peachImg[p.basicCardAnime.index], op)
				case data.Drunk:
					screen.DrawImage(playerImgs.drunkImg[p.basicCardAnime.index], op)
				}
			}
		}
		//绘制技能使用效果
		if p.skillAnime.enable {
			op.GeoM.Reset()
			op.ColorScale.SetA(p.skillAnime.visibility)
			op.GeoM.Translate(p.x+24, p.y+53)
			screen.DrawImage(playerImgs.skillImg[p.skillAnime.index], op)
			x := float64(175-p.skillAnime.name.Bounds().Dx()) / 2
			op.GeoM.Reset()
			op.GeoM.Translate(p.x+x, p.y+100)
			screen.DrawImage(p.skillAnime.name, op)
			op.ColorScale.Reset()
		}
		//绘制技能列表
		for i := 0; i < len(p.skills); i++ {
			p.skills[i].Draw(screen)
		}
	}
	//画卡
	p.drawCards(screen)
	//p.cardsText.Draw(screen) //debug 卡片文本
}

func (p *player) drawCards(screen *ebiten.Image) {
	//绘制牌堆
	for i := 0; i < len(p.cards); i++ {
		p.cards[i].Draw(screen)
	}
	//绘制装备区
	for i := 0; i < len(p.equipSlot); i++ {
		if p.equipSlot[i] == nil {
			continue
		}
		p.equipSlot[i].Draw(screen)
	}
	//绘制判定区
	for i := 0; i < len(p.judgeSlot); i++ {
		if p.judgeSlot[i] == nil {
			continue
		}
		p.judgeSlot[i].Draw(screen)
	}
}

func (p *player) anime() {
	//阶段外框
	if p.isTurnowner || p.isSelected {
		p.stateAnime.counter++
		if p.stateAnime.counter == 3 {
			p.stateAnime.counter = 0
			p.stateAnime.index++
			if p.stateAnime.index > 11 {
				p.stateAnime.index = 0
			}
		}
	}
	//濒死
	if p.dying {
		p.dyingAnime.index += 0.01
		if p.dyingAnime.index > 1 {
			p.dyingAnime.index = -1
		}
		if p.dyingAnime.index > 0 {
			p.dyingAnime.counter = p.dyingAnime.index
		} else {
			p.dyingAnime.counter = -p.dyingAnime.index
		}
	}
	//血量更改
	if p.dmgAnime.enable {
		//伤害动画
		p.dmgAnime.counter++
		if p.dmgAnime.counter == 2 {
			p.dmgAnime.counter = 0
			p.dmgAnime.index++
			if p.dmgAnime.index < 3 {
				p.x -= 1
			} else if p.dmgAnime.index < 6 {
				p.x += 1
			}
			if p.dmgAnime.index >= 10 {
				p.dmgAnime.index = 0
				p.dmgAnime.counter = 0
				p.dmgAnime.enable = false
			}
		}
	}
	//恢复
	if p.recAnime.enable {
		p.recAnime.count++
		if p.recAnime.count == 2 {
			p.recAnime.index++
			p.recAnime.count = 0
		}
		if p.recAnime.index == 7 {
			p.recAnime.enable = false
			p.recAnime.index = 0
			p.recAnime.count = 0
		}
	}
	//连环
	if p.tslhAnime.enable {
		if !p.isLinked {
			p.tslhAnime.index -= 0.05
			if p.tslhAnime.index < 0 {
				p.tslhAnime.enable = false
			}
		}
		if p.isLinked && p.tslhAnime.index < 1 {
			p.tslhAnime.index += 0.05
		}
	}
	//延时锦囊
	if p.delayTipsAnime.enable {
		p.delayTipsAnime.counter++
		p.delayTipsAnime.visibility -= 1. / 120
		if p.delayTipsAnime.counter == 6 {
			p.delayTipsAnime.counter = 0
			p.delayTipsAnime.index++
		}
		if p.delayTipsAnime.index == 12 {
			p.delayTipsAnime.counter = 0
			p.delayTipsAnime.index = 0
			p.delayTipsAnime.visibility = 1
			p.delayTipsAnime.enable = false
		}
	}
	//基本牌
	if p.basicCardAnime.enable {
		p.basicCardAnime.count++
		if p.basicCardAnime.count >= 2 {
			p.basicCardAnime.index++
			p.basicCardAnime.count = 0
		}
		if p.basicCardAnime.index >= 13 {
			p.basicCardAnime.index = 0
			p.basicCardAnime.count = 0
			p.basicCardAnime.enable = false
		}
	}
	//技能
	if p.skillAnime.enable {
		p.skillAnime.counter++
		if p.skillAnime.index == 9 {
			p.skillAnime.visibility -= 1. / 60
			if p.skillAnime.counter == 60 {
				p.skillAnime.counter = 0
				p.skillAnime.index = 0
				p.skillAnime.visibility = 1
				p.skillAnime.enable = false
			}
		} else if p.skillAnime.counter == 3 {
			p.skillAnime.counter = 0
			p.skillAnime.index++
		}
	}
	for i := 0; i < len(p.cards); i++ {
		p.cards[i].anime()
	}
	//更新装备区
	for i := 0; i < len(p.equipSlot); i++ {
		if p.equipSlot[i] == nil {
			continue
		}
		p.equipSlot[i].anime()
	}
	//更新判定区
	for i := 0; i < len(p.judgeSlot); i++ {
		if p.judgeSlot[i] == nil {
			continue
		}
		p.judgeSlot[i].anime()
	}
}

// 向玩家手牌堆中加牌
func (p *player) addCard(cards ...cardI) {
	for i := 0; i < len(cards); i++ {
		card := cards[i]
		x, y := card.getPos()
		card.resetFadeOut()
		card.setPos(x+10+20*float64(i-len(cards)/2), y)
		card.setVisibility(1)
		card.setInHandleZoon(false)
		p.cards = append(p.cards, card)
		if !p.isLocal {
			p.cardsText.SetStr(p.cardsText.Str + "\n" + card.getCardName().String()) //debug 卡片文本
			p.cardsText.SetPos(p.x, p.y-20)                                          //debug 卡片文本
			card.setShowBack(true)
			card.move2Pos(p.x+10+20+20*float64(i-len(cards)/2), p.y+40,
				func() { card.setFadeOut(1*60, func() { card.setVisibility(0); card.setPos(p.x+20, p.y+40) }) })
			continue
		}
		card.setShowBack(false)
	}
	p.sortCards()
	p.calculatePos()
	p.setCardNumText(len(p.cards))
}

// 获取子元素坐标,输入子元素的中心坐标与尺寸，返回子元素的左上角坐标
func (p *player) getSubItemPos(x, y float64, size image.Point) (float64, float64) {
	return x - float64(size.X/2) + p.x, y - float64(size.Y/2) + p.y
}

// 设置player坐标
func (p *player) setPos(x, y float64) {
	p.rect.x, p.rect.y = int(x), int(y)
	p.x, p.y = x, y
	p.cardNumText.SetPos(x+8+(23-p.cardNumText.Width)/2, y+200)
}

func (p *player) setCardNumText(n int) {
	p.cardNumText.SetText("[#ffffff]" + strconv.Itoa(n))
	p.cardNumText.SetPos(p.x+8+(23-p.cardNumText.Width)/2, p.y+200)
}

// 对手牌进行排序
func (p *player) sortCards() {
	for end := len(p.cards); end > 0; end-- {
		for i := 0; i < end-1; i++ {
			if p.cards[i].getID() > p.cards[i+1].getID() {
				p.cards[i], p.cards[i+1] = p.cards[i+1], p.cards[i]
			}
		}
	}
}

func (p *player) calculatePos() {
	if !p.isLocal {
		return
	}
	c := p.cards
	if len(c) <= 6 {
		for i := 0; i < len(c); i++ {
			y := 570.
			if c[i].getSelectState() {
				y -= 8
			}
			x := float64(i)*128. + 100.
			c[i].move2Pos(x, y, nil)
		}
	} else {
		for i := 0; i < len(c); i++ {
			y := 570.
			if c[i].getSelectState() {
				y -= 8
			}
			x := float64(i)*(128-float64(128*len(c)-900)/float64(len(c)-1)) + 100
			c[i].move2Pos(x, y, nil)
		}
	}
}

// 从玩家区域中取走指定的卡
func (p *player) getCard(id data.CID) cardI {
	if id == 0 {
		panic("不能取走id=0的卡牌")
	}
	//检查手牌堆
	for i := 0; i < len(p.cards); i++ {
		if p.cards[i].getID() == id {
			c := p.cards[i]
			p.cards = append(p.cards[:i], p.cards[i+1:]...)
			if !p.isLocal { //debug 卡片文本
				cardsStr := ""
				for j := 0; j < len(p.cards); j++ {
					cardsStr += p.cards[j].getCardName().String() + "\n"
				}
				p.cardsText.SetStr(cardsStr)
				p.cardsText.SetPos(p.x-p.cardsText.Width+100, p.y-20)
			}
			p.setCardNumText(len(p.cards))
			return c
		}
	}
	//检查判定区
	for i := 0; i < len(p.judgeSlot); i++ {
		if p.judgeSlot[i] == nil {
			continue
		}
		if p.judgeSlot[i].getID() == id {
			c := p.judgeSlot[i]
			p.judgeSlot[i] = nil
			return c
		}
	}
	//检查装备区
	for i := 0; i < len(p.equipSlot); i++ {
		if p.equipSlot[i] == nil {
			continue
		}
		if p.equipSlot[i].getID() == id {
			if p.equipSlot[i].getCardName() == data.ZBSM {
				for j := 0; j < len(p.skills); j++ {
					if p.skills[j].getID() == data.ZBSMSkill {
						p.skills = append(p.skills[:j], p.skills[j+1:]...)
						break
					}
				}
			}
			c := p.equipSlot[i]
			p.equipSlot[i] = nil
			c.isEquip = false
			return c
		}
	}
	//检查技能暂存卡牌
	for sid, cards := range p.tsCard {
		for i := 0; i < len(cards); i++ {
			if cards[i].getID() != id {
				continue
			}
			c := cards[i]
			p.tsCard[sid] = append(cards[:i], cards[i+1:]...)
			if len(p.tsCard[sid]) == 0 {
				delete(p.tsCard, sid)
			}
			return c
		}
	}
	panic("玩家区域中没有id=" + strconv.Itoa(int(id)) + "的卡")
}

// 添加技能暂存卡牌
func (p *player) addTSCard(sid data.SID, cards ...cardI) {
	p.tsCard[sid] = append(p.tsCard[sid], cards...)
}

// 启用技能
func (p *player) enableSkill(g *Games, sid data.SID) {
	p.enableSkills[sid] = struct{}{}
}

// 检查技能是否启动
func (p *player) isSkillEnable(sid data.SID) bool {
	_, ok := p.enableSkills[sid]
	return ok
}

// 检查skills中是否至少有一个启用
func (p *player) oneOfSkillsEnable(sid ...data.SID) bool {
	for i := 0; i < len(sid); i++ {
		if _, ok := p.enableSkills[sid[i]]; ok {
			return true
		}
	}
	return false
}

// 关闭技能
func (p *player) disableSkill(sid data.SID) {
	delete(p.enableSkills, sid)
}

// 从玩家手牌堆中寻找指定的卡
func (p *player) findCard(id data.CID) cardI {
	for i := 0; i < len(p.cards); i++ {
		if p.cards[i].getID() == id {
			return p.cards[i]
		}
	}
	for i := 0; i < len(p.equipSlot); i++ {
		if p.equipSlot[i] != nil && p.equipSlot[i].getID() == id {
			return p.equipSlot[i]
		}
	}
	panic("玩家手牌堆中没有id=" + strconv.Itoa(int(id)) + "的卡")
}

func (p *player) updateBtn(g *Games) {
	for i := 0; i < len(p.btnGrup); i++ {
		p.btnGrup[i].Update(g)
	}
}

func (p *player) updateSkill(g *Games) {
	for i := 0; i < len(p.skills); i++ {
		p.skills[i].update(g, p)
	}
}

func (p *player) handleUseState(g *Games) {
	if p.oneOfSkillsEnable(data.ZBSMSkill, data.RenDeSkill, data.WuShengSkill,
		data.LongDanSkill, data.LiMuSkill, data.YeYanSkill, data.QingNangSkill,
		data.JieYinSkill, data.ChengLueSkill, data.XueHenSkill, data.MieWuSkill,
		data.PaiYiSkill, data.XionHuoSkill, data.ZhuiFengSkill, data.LuanWuSkill,
		data.ZhiHengSkill, data.YanYuSkill, data.JianYingSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	for _, card := range p.cards {
		card.handleUseState(g, p)
	}
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newEndStateBtn(func(g *Games) { g.useCardInf <- data.UseCardInf{Skip: true}; g.hasSkip = true; p.btnGrup = nil }))
		}
	} else {
		p.btnGrup = nil
	}
	p.updateSkill(g)
}

func (p *player) handleDropState(g *Games) {
	p.updateBtn(g)
	for i := len(p.cards) - 1; i >= 0; i-- {
		if p.cards[i].handleDropState(g, p) {
			break
		}
	}
	if len(p.btnGrup) == 0 {
		p.btnGrup = append(p.btnGrup, newDropBtn(func(g *Games) {
			g.dropCardInf <- append(data.DropCardInf{}, p.selCard...)
		}))
	}
}

func (p *player) handleDying(g *Games) {
	if p.oneOfSkillsEnable(data.JiJiuSkill, data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleDying(g, p)
	}
	p.updateSkill(g)
}

func (p *player) handleWXKJ(g *Games) {
	if p.oneOfSkillsEnable(data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleWXKJ(g, p)
	}
	p.updateSkill(g)
}

func (p *player) handleDodge(g *Games) {
	if p.oneOfSkillsEnable(data.LongDanSkill, data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleDodge(g, p)
	}
	p.updateSkill(g)
}

func (p *player) handleDuel(g *Games) {
	if p.oneOfSkillsEnable(data.ZBSMSkill, data.WuShengSkill, data.LongDanSkill, data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleDuel(g, p)
	}
	p.updateSkill(g)
}

func (p *player) handleNMRQ(g *Games) {
	if p.oneOfSkillsEnable(data.ZBSMSkill, data.WuShengSkill, data.LongDanSkill, data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleNMRQ(g, p)
	}
	p.updateSkill(g)
}

func (p *player) handleJDSR(g *Games) {
	if p.oneOfSkillsEnable(data.MieWuSkill, data.LongDanSkill, data.WuShengSkill, data.ZBSMSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	for _, card := range p.cards {
		card.handleJDSR(g, p)
	}
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) { g.useCardInf <- data.UseCardInf{Skip: true}; g.hasSkip = true; p.btnGrup = nil }))
		}
	} else {
		p.btnGrup = nil
	}
	p.updateSkill(g)
}

func (p *player) handleLuanWu(g *Games) {
	if p.oneOfSkillsEnable(data.MieWuSkill, data.LongDanSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	for _, card := range p.cards {
		card.handleLuanWu(g, p)
	}
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) { g.useCardInf <- data.UseCardInf{Skip: true}; g.hasSkip = true; p.btnGrup = nil }))
		}
	} else {
		p.btnGrup = nil
	}
	p.updateSkill(g)
}

func (p *player) handleQLYYD(g *Games) {
	if p.oneOfSkillsEnable(data.WuShengSkill, data.LongDanSkill, data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	for _, card := range p.cards {
		card.handleQLYYD(g, p)
	}
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) { g.useCardInf <- data.UseCardInf{Skip: true}; g.hasSkip = true; p.btnGrup = nil }))
		}
	} else {
		p.btnGrup = nil
	}
	p.updateSkill(g)
}

func (p *player) handleWJQF(g *Games) {
	if p.oneOfSkillsEnable(data.LongDanSkill, data.MieWuSkill) {
		p.updateSkill(g)
		return
	}
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleWJQF(g, p)
	}
	p.updateSkill(g)
}

func (p *player) handleBurnShow(g *Games) {
	p.updateBtn(g)
	for _, card := range p.cards {
		card.handleBurnShow(g, p)
	}
}

func (p *player) handleBurnDrop(g *Games) {
	p.updateBtn(g)
	if len(p.selCard) == 0 {
		if len(p.btnGrup) == 0 && !g.hasSkip {
			p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
				newCancleBtn(func(g *Games) {
					g.useCardInf <- data.UseCardInf{Skip: true}
					g.hasSkip = true
					p.btnGrup = nil
				}))
		}
	} else {
		p.btnGrup = nil
	}
	for _, card := range p.cards {
		card.handleBurnDrop(g, p)
	}
}

func (p *player) handleDropSelfAll(g *Games) {
	p.updateBtn(g)
	if len(p.btnGrup) == 0 && !g.hasSkip {
		p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil))
	}
	if len(p.selCard) == g.selNum {
		if len(p.btnGrup) == 1 {
			p.btnGrup = append(p.btnGrup, newConfirmBtn(func(g *Games) {
				g.dropCardInf <- p.selCard
			}))
		}
	} else {
		p.btnGrup = p.btnGrup[:1]
	}
	for i := len(p.cards) - 1; i >= 0; i-- {
		if p.cards[i].handleDropSelfAll(g, p) {
			break
		}
	}
	for _, equip := range p.equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		equip.handleDropSelfAll(g, p)
	}
}

func (p *player) handleGSF(g *Games) {
	p.updateBtn(g)
	if len(p.btnGrup) == 0 && !g.hasSkip {
		p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil),
			newCancleBtn(func(g *Games) {
				g.dropCardInf <- data.DropCardInf{}
				p.btnGrup = nil
			}))
	}
	if len(p.selCard) == g.selNum {
		if len(p.btnGrup) == 2 {
			p.btnGrup = append(p.btnGrup, newConfirmBtn(func(g *Games) {
				g.dropCardInf <- p.selCard
			}))
		}
	} else {
		p.btnGrup = p.btnGrup[:2]
	}
	for i := len(p.cards) - 1; i >= 0; i-- {
		if p.cards[i].handleGSF(g, p) {
			break
		}
	}
	for _, equip := range p.equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		equip.handleGSF(g, p)
	}
}

func (p *player) handleSkillSelectState(g *Games) {
	p.updateBtn(g)
	if g.hasSkip {
		return
	}
	//如果不需要选目标
	if g.skillTargetInf.TargetNum == 0 {
		if len(p.btnGrup) == 0 {
			p.btnGrup = append(p.btnGrup, newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: g.selectSkill}
				g.hasSkip = true
				p.btnGrup = nil
			}), newCancleBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{Skip: true}
				g.hasSkip = true
				p.btnGrup = nil
			}))
		}
		return
	}
	//生成取消按钮
	if len(p.btnGrup) == 0 {
		p.btnGrup = append(p.btnGrup, newCancleBtn(func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{Skip: true}
			g.hasSkip = true
			p.btnGrup = nil
		}))
	}
	//检测玩家点击
	for _, t := range g.playList {
		if t.unSelAble || !t.isClicked() {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			delFromList(g.skilltargetList, func(id data.PID) bool { return id == t.pid })
		}
		g.skilltargetList = append(g.skilltargetList, t.pid)
		t.isSelected = true
		if len(g.skilltargetList) > int(g.skillTargetInf.TargetNum) {
			g.getPlayer(g.skilltargetList[0]).isSelected = false
			g.skilltargetList = g.skilltargetList[1:]
		}
	}
	if len(g.skilltargetList) > 0 {
		if len(p.btnGrup) == 1 {
			p.btnGrup = append(p.btnGrup, newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: g.selectSkill, TargetList: g.skilltargetList}
				g.hasSkip = true
				p.btnGrup = nil
			}))
		}
	} else {
		p.btnGrup = p.btnGrup[:1]
	}
}

func (p *player) handleWenJi(g *Games) {
	if len(p.btnGrup) == 0 {
		p.btnGrup = append(p.btnGrup, newFakeConfirmBtn(nil))
	}
	for i := 0; i < len(p.cards); i++ {
		p.cards[i].handleWenJi(g, p)
	}
	for _, equip := range p.equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		equip.handleWenJi(g, p)
	}
}
