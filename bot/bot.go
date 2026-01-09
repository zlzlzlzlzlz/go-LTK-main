package bot

import (
	"goltk/data"
	"math/rand"
	"time"
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

type setHpInf struct {
	pid    data.PID
	hp     data.HP
	hptype data.SetHpType
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

type gsidCardInf struct {
	id    data.GSID
	cards []data.CID
}

var cardHeap = data.GetCards()

func getCard(id data.CID) *data.Card {
	c := cardHeap[id-1]
	return &c
}

type Bot struct {
	state              data.GameState
	pid                data.PID
	curPlayer          data.PID
	turnOwner          data.PID
	cards              []data.Card //总牌堆
	players            []player
	wgfdCards          []data.CID
	ssqyCards          []data.CID
	ghcqCards          []data.CID
	qlgCards           []data.CID
	otherCards         []data.CID
	useCardInf         chan data.UseCardInf
	dropCardInf        chan data.DropCardInf
	useSkillInf        chan data.UseSkillInf
	cardReceiver       chan cardReceive             //接收发牌信息的chan
	removeReceiver     chan cardReceive             //接收弃牌信息的chan
	moveCardRec        chan cardMoveRec             //接收移动卡牌信息的chan
	useReceiver        chan usecardInf              //接受用牌信息的chan
	useTmpRec          chan useTmpCardInf           //接收使用临时卡信息的chan
	availableTargetrec chan data.AvailableTargetInf //接收可用目标信息的chan
	pidReceiver        chan data.PID                //接收主视角pid的chan
	playerInfRec       chan []data.PlayerInf        //接受玩家列表信息的chan
	gsidCardReceiver   chan gsidCardInf             //游戏技能的卡牌接收器
	useAbleReceiver    chan []data.CID              //可用的卡接收器
	useAbleSkillRec    chan []data.SID              //可用的主动技接收器
	dropAbleRec        chan dropAbleRec             //可丢弃的卡的接收器
	useSkillRspRec     chan data.UseSkillRsp        //用主动技能的回应的接收器
	turnOwnerRec       chan data.PID                //当前回合拥有者接收器
	skillSelectRec     chan data.SID                //问玩家要不要用技能的chan
	useSkillRec        chan useSkillInf             //接收使用技能信息的chan
	targetQuest        chan data.CID                //向服务端询问卡片可用目标
	availableRoleRec   chan []data.Role             //接收可选角色列表
	roleInf            chan data.Role               //向服务端发送选择的角色
	setHpReceiver      chan setHpInf
	gameStateRec       chan gameStateInf
	closeSignal        chan struct{}
	selectSkill        data.SID
	teamMate           []bool
	skills             []skillI
}

func NewBot(teamMate []bool) *Bot {
	b := Bot{
		useCardInf:         make(chan data.UseCardInf, 1),
		dropCardInf:        make(chan data.DropCardInf, 1),
		useSkillInf:        make(chan data.UseSkillInf, 4),
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
		setHpReceiver:      make(chan setHpInf, 4),
		gameStateRec:       make(chan gameStateInf, 1),
		closeSignal:        make(chan struct{}),
		teamMate:           teamMate,
	}
	return &b
}

func (b *Bot) GetUseCardInf() <-chan data.UseCardInf {
	return b.useCardInf
}

func (b *Bot) GetDropCardInf() <-chan data.DropCardInf {
	return b.dropCardInf
}

func (b *Bot) GetUseSkillInf() <-chan data.UseSkillInf {
	return b.useSkillInf
}

func (b *Bot) SetPid(id data.PID) {
	b.pidReceiver <- id
}

func (b *Bot) GetRole() <-chan data.Role {
	return b.roleInf
}

func (b *Bot) SendCard(id data.PID, cards ...data.CID) {
	b.cardReceiver <- cardReceive{id: id, cards: cards}
}

func (b *Bot) RemoveCard(id data.PID, cards ...data.CID) {
	b.removeReceiver <- cardReceive{id: id, cards: cards}
}

func (b *Bot) MoveCard(src, dst data.PID, cards ...data.CID) {
	b.moveCardRec <- cardMoveRec{src: src, dst: dst, cards: cards}
}

func (b *Bot) UseCard(user data.PID, c data.CID, targets ...data.PID) {
	b.useReceiver <- usecardInf{user: user, card: c, targets: targets}
}

func (b *Bot) UseTmpCard(user data.PID, name data.CardName, dec data.Decor, num data.CNum,
	tmpType data.TmpCardType, target ...data.PID) {
	b.useTmpRec <- useTmpCardInf{user: user, cname: name, dec: dec, num: num, tmpType: tmpType, targets: target}
}

func (b *Bot) SendUseAbleCards(cards []data.CID) {
	b.useAbleReceiver <- cards
}

func (b *Bot) SendUseAbleSkill(skills []data.SID) {
	b.useAbleSkillRec <- skills
}

func (b *Bot) SendDropAbleCard(cards []data.CID, dropNum uint8) {
	b.dropAbleRec <- dropAbleRec{cards: cards, dropNum: dropNum}
}

func (b *Bot) SendSkillSelect(sid data.SID) {
	b.skillSelectRec <- sid
}

func (b *Bot) UseSkill(user data.PID, skill data.SID, target []data.PID, args ...byte) {
	b.useSkillRec <- useSkillInf{user: user, skill: skill, target: target, args: args}
}

func (b *Bot) SendAvailableTarget(inf data.AvailableTargetInf) {
	b.availableTargetrec <- inf
}

func (b *Bot) SendPlayerInf(inf []data.PlayerInf) {
	b.playerInfRec <- inf
}

func (b *Bot) SendUseSkillRsp(rsp data.UseSkillRsp) {
	b.useSkillRspRec <- rsp
}

func (b *Bot) SendGSCards(id data.GSID, cards ...data.CID) {
	b.gsidCardReceiver <- gsidCardInf{id: id, cards: cards}
}

func (b *Bot) SetHP(pid data.PID, hp data.HP, hptype data.SetHpType) {
	b.setHpReceiver <- setHpInf{pid: pid, hp: hp, hptype: hptype}
}

func (b *Bot) SetGameState(state data.GameState, t time.Duration, curPlayer data.PID) {
	b.gameStateRec <- gameStateInf{state: state, t: t, curPlayer: curPlayer}
}

func (b *Bot) SetTurnOwner(pid data.PID) {
	b.turnOwnerRec <- pid
}

func (b *Bot) GetTargetQuest() <-chan data.CID {
	return b.targetQuest
}

func (b *Bot) SendAvailableRole(roles ...data.Role) {
	b.availableRoleRec <- roles
}

func (b *Bot) GetClientType() data.PlayerType {
	return data.BotPlayer
}

func (b *Bot) Close() {
	close(b.closeSignal)
}

func (b *Bot) Run() {
	b.handleInit()
	for {
		select {
		case <-b.closeSignal:
			return
		default:
		}
		if b.state != data.WXKJState && b.curPlayer != b.pid {
			b.normalUpdate() //若当前玩家不为自己则跳至普通更新
			continue
		}
		switch b.state {
		case data.UseCardState:
			b.handleUseCard()
		case data.DodgeState:
			b.handleDodge()
		case data.WXKJState:
			b.handleWXKJ()
		case data.DyingState:
			b.handleDying()
		case data.DuelState:
			b.handleDuel()
		case data.DropCardState:
			b.handleDrop()
		case data.NMRQState:
			b.handleNMRQ()
		case data.WJQFState:
			b.handleWJQF()
		case data.WGFDState:
			b.handleWGFD()
		case data.BurnShowState:
			b.handleBurnShow()
		case data.BurnDropState:
			b.handleBurnDrop()
		case data.SSQYState:
			b.handleSSQY()
		case data.GHCQState:
			b.handleGHCQ()
		case data.QLGState:
			b.handleQLG()
		case data.JDSRState:
			b.handleJDSRState()
		case data.DropSelfAllCards:
			b.handleDropAll()
		case data.QlYYDState:
			b.handleCtuAtk()
		case data.SkillSelectState:
			b.handleSkillSelectState()
		case data.SkillJudgeState:
			b.handleSkillJudgeState()
		case data.CXSGJState:
			b.handleCXSGJ()
		case data.QueDiState:
			b.handleQueDi()
		case data.DropOtherCardState:
			b.handleDropOtherCard()
		case data.DropSelfHandCard:
			b.handleDropSelfCard()
		case data.WenJiState:
			b.handleWenJi()
		case data.EnYuanState:
			b.handleEnYuan()
		case data.GuanXingState:
			b.handleGuanXing()
		case data.ChangeRoleState:
			b.handleChangeRole()
		case data.PoJunState:
			b.handlePoJun()
		case data.FanKuiState:
			b.handleFanKui()
		case data.QinYinState:
			b.handleCXSGJ()
		case data.FengPoState:
			b.handleFengPo()
		case data.DieState:
			b.handleDie()
		case data.LiangZhuState:
			b.handleLiangZhu()
		case data.GSFState:
			b.handleDropAll()
		case data.LuanWuState:
			b.handleLuanWu()
		case data.GetSkillState:
			b.handleGetSkill()
		default:
			b.waitForStateSwitch()
		}
		select {
		case inf := <-b.cardReceiver:
			b.addCard(inf)
		default:
		}
		select {
		case inf := <-b.removeReceiver:
			b.players[inf.id].removeCard(inf.cards...)
		default:
		}
		select {
		case rec := <-b.useReceiver:
			b.useCard(rec)
		default:
		}
		select {
		case inf := <-b.useTmpRec:
			b.useTmpCard(inf)
		default:
		}
		select {
		case <-b.useSkillRec:
		default:
		}
		select {
		case inf := <-b.gsidCardReceiver:
			b.handleGSIDCardsInf(inf)
		default:
		}
		select {
		case inf := <-b.setHpReceiver:
			b.players[inf.pid].hp = inf.hp
		default:
		}
		select {
		case inf := <-b.moveCardRec:
			b.moveCard(inf)
		default:
		}
	}
}

func (b *Bot) getSkill(id data.SID) skillI {
	for i := 0; i < len(b.skills); i++ {
		if b.skills[i].getID() == id {
			return b.skills[i]
		}
	}
	if id == data.ZBSMSkill {
		b.skills = append(b.skills, newSkillI(data.ZBSMSkill))
		return b.skills[len(b.skills)-1]
	}
	panic("该bot没有名为" + id.String() + "的技能")
}

func (b *Bot) isTeamMate(id data.PID) bool {
	return b.teamMate[id]
}

func (b *Bot) getTeamMate(targets ...data.PID) (teamMates []data.PID) {
	for _, t := range targets {
		if b.isTeamMate(t) {
			teamMates = append(teamMates, t)
		}
	}
	return
}

func (b *Bot) getEnemy(targets ...data.PID) (enemy []data.PID) {
	for _, t := range targets {
		if !b.isTeamMate(t) {
			enemy = append(enemy, t)
		}
	}
	return
}

// 获取存活队友数
func (b *Bot) getTeamMateCount() (num uint8) {
	for _, p := range b.players {
		if p.death {
			continue
		}
		if b.isTeamMate(p.pid) {
			num++
		}
	}
	return
}

// 获取存活敌人数
func (b *Bot) getEnemyCount() (num uint8) {
	for _, p := range b.players {
		if p.death {
			continue
		}
		if !b.isTeamMate(p.pid) {
			num++
		}
	}
	return
}

// 等待服务端发来的更新
func (b *Bot) normalUpdate() {
	select {
	case inf := <-b.gameStateRec:
		b.switchGameState(inf)
	case inf := <-b.cardReceiver:
		b.addCard(inf)
	case b.turnOwner = <-b.turnOwnerRec:
	case inf := <-b.removeReceiver:
		b.players[inf.id].removeCard(inf.cards...)
	case inf := <-b.moveCardRec:
		b.moveCard(inf)
	case rec := <-b.useReceiver:
		b.useCard(rec)
	case inf := <-b.useTmpRec:
		b.useTmpCard(inf)
	case <-b.useSkillRec:
	case inf := <-b.gsidCardReceiver:
		b.handleGSIDCardsInf(inf)
	case inf := <-b.setHpReceiver:
		b.players[inf.pid].hp = inf.hp
	case <-b.closeSignal:
	}
}

func (b *Bot) addCard(inf cardReceive) {
	b.players[inf.id].addCard(inf.cards...)
}

func (b *Bot) moveCard(inf cardMoveRec) {
	switch inf.src {
	case data.SpecialPIDGame:
		b.players[inf.dst].addCard(inf.cards...)
	default:
		b.players[inf.src].removeCard(inf.cards...)
		b.players[inf.dst].addCard(inf.cards...)
	}
}

func (b *Bot) useCard(inf usecardInf) {
	b.players[inf.user].removeCard(inf.card)
	//如果不为基本牌或锦囊牌则将牌放入对应区域
	c := b.cards[inf.card]
	switch c.CardType {
	case data.WeaponCardType:
		b.players[inf.user].equips[data.WeaponSlot] = getCard(c.ID)
	case data.ArmorCardType:
		b.players[inf.user].equips[data.ArmorSlot] = getCard(c.ID)
	case data.HorseUpCardType:
		b.players[inf.user].equips[data.HorseUpSlot] = getCard(c.ID)
	case data.HorseDownCardType:
		b.players[inf.user].equips[data.HorseDownSlot] = getCard(c.ID)
	case data.DealyTipsCardType:
		var slot *data.CID
		var t *player
		if len(inf.targets) > 0 {
			t = &b.players[inf.targets[0]]
		} else {
			t = &b.players[inf.user]
		}
		switch c.Name {
		case data.LBSS:
			slot = &t.judges[data.LBSSSlot]
		case data.BLCD:
			slot = &t.judges[data.BLCDSlot]
		case data.Lightning:
			slot = &t.judges[data.LightningSlot]
		}
		if len(inf.targets) > 0 {
			*slot = c.ID
		} else {
			*slot = 0
		}
	}
}

func (b *Bot) useTmpCard(inf useTmpCardInf) {
	tmpCard := data.NewCard(inf.cname, inf.dec, inf.num)
	switch tmpCard.CardType {
	case data.WeaponCardType:
		b.players[inf.user].equips[data.WeaponSlot] = &tmpCard
	case data.ArmorCardType:
		b.players[inf.user].equips[data.ArmorSlot] = &tmpCard
	case data.HorseUpCardType:
		b.players[inf.user].equips[data.HorseUpSlot] = &tmpCard
	case data.HorseDownCardType:
		b.players[inf.user].equips[data.HorseDownSlot] = &tmpCard
	}
}

// 等待切换阶段
func (b *Bot) waitForStateSwitch() {
	for {
		select {
		case inf := <-b.gameStateRec:
			b.switchGameState(inf)
			return
		case <-b.closeSignal:
			return
		case inf := <-b.removeReceiver:
			b.players[inf.id].removeCard(inf.cards...)
		case inf := <-b.useReceiver:
			b.useCard(inf)
		case inf := <-b.cardReceiver:
			b.players[inf.id].addCard(inf.cards...)
		case <-b.useSkillRec:
		case inf := <-b.useTmpRec:
			b.useTmpCard(inf)
		case inf := <-b.moveCardRec:
			b.moveCard(inf)
		case inf := <-b.gsidCardReceiver:
			b.handleGSIDCardsInf(inf)
		}
	}
}

func (b *Bot) switchGameState(inf gameStateInf) {
	//某一阶段结束时要做的事
	switch b.state {
	}
	b.state = inf.state
	b.curPlayer = inf.curPlayer
	//某一阶段开始时要做的事
	switch b.state {
	case data.SkillSelectState:
		b.selectSkill = <-b.skillSelectRec
	}
}

func (b *Bot) handleGSIDCardsInf(inf gsidCardInf) {
	switch inf.id {
	case data.GSIDWGFD:
		b.wgfdCards = inf.cards
	case data.GSIDSSQY:
		b.ssqyCards = inf.cards
	case data.GSIDGHCQ:
		b.ghcqCards = inf.cards
	case data.GSIDQLG:
		b.qlgCards = inf.cards
	case data.GSIDDropOtrCard:
		b.otherCards = inf.cards
	}
}

func (b *Bot) handleInit() {
	//初始化总牌堆
	b.cards = append([]data.Card{{}}, data.GetCards()...)
	//接受pid
	b.pid = <-b.pidReceiver
	roleList := <-b.availableRoleRec
	//发送选定的角色
	b.roleInf <- roleList[rand.Intn(len(roleList))]
	//接收玩家列表
	var pList []data.PlayerInf
	select {
	case pList = <-b.playerInfRec:
	case <-b.closeSignal:
		return
	}
	for _, inf := range pList {
		b.players = append(b.players, newPlayer(inf.Role, inf.PID))
		if inf.PID == b.pid {
			for _, s := range inf.Role.SkillList {
				b.skills = append(b.skills, newSkillI(s))
			}
		}
	}
	for i := 0; i < len(b.players); i++ {
		inf := <-b.cardReceiver
		b.players[inf.id].addCard(inf.cards...)
	}
	b.waitForStateSwitch()
}

func (b *Bot) handleUseCard() {
	//从服务器获取可用卡牌
	useAbleCard := <-b.useAbleReceiver
	//接受可用技能
	if skills := <-b.useAbleSkillRec; len(skills) > 0 {
		<-time.After((time.Duration(rand.Intn(500)) + 500) * time.Millisecond)
		if b.getSkill(skills[0]).handleActive(b) {
			b.waitForStateSwitch()
			return
		}
	}
	c := data.CID(0)
	targetList := []data.PID{}
	//优先使用桃
	for _, cid := range useAbleCard {
		if b.cards[cid].Name == data.Peach {
			c = cid
			goto useCard
		}
	}
	//没有桃则检查有没有合适的武器
	{
		//定义武器优先级
		weponePriority := map[data.CardName]int{data.GSF: 100, data.QLYYD: 90, data.QLG: 80, data.QGJ: 70, data.ZBSM: 60,
			data.FTHJ: 50, data.ZQYS: 40, data.GDD: 30, data.CXSGJ: 20, data.HBJ: 20, data.ZGLN: 10}
		bestWepone := b.players[b.pid].equips[data.WeaponSlot]
		for _, cid := range useAbleCard {
			if b.cards[cid].CardType != data.WeaponCardType {
				continue
			}
			if bestWepone == nil || weponePriority[bestWepone.Name] < weponePriority[b.cards[cid].Name] {
				bestWepone = getCard(cid)
			}
		}
		if bestWepone != b.players[b.pid].equips[data.WeaponSlot] {
			c = bestWepone.ID
			goto useCard
		}
	}
	//检查是否可以装备装甲或马
	for _, cid := range useAbleCard {
		switch b.cards[cid].CardType {
		case data.ArmorCardType:
			if b.players[b.pid].equips[data.ArmorSlot] == nil {
				c = cid
				goto useCard
			}
		case data.HorseUpCardType:
			if b.players[b.pid].equips[data.HorseUpSlot] == nil {
				c = cid
				goto useCard
			}
		case data.HorseDownCardType:
			if b.players[b.pid].equips[data.HorseDownSlot] == nil {
				c = cid
				goto useCard
			}
		}

	}
	//检查要不要出锦囊牌
	{
		//不需要目标的锦囊牌列表
		noTargetList := []data.CardName{data.WZSY, data.WJQF, data.NMRQ, data.TYJY, data.WGFD, data.TSLH}
		for _, cid := range useAbleCard {
			if b.cards[cid].CardType != data.TipsCardType {
				continue
			}
			if !isItemInList([]data.CardName{data.WZSY, data.WGFD, data.SSQY, data.GHCQ,
				data.Burn, data.Duel, data.NMRQ, data.WJQF, data.TSLH, data.LBSS, data.BLCD, data.Lightning},
				b.cards[cid].Name) {
				continue
			}
			if isItemInList(noTargetList, b.cards[cid].Name) {
				//五谷丰登
				if b.cards[cid].Name == data.WGFD {
					if b.getTeamMateCount() < b.getEnemyCount() {
						continue
					}
				}
				//万箭齐发和南蛮入侵
				if b.cards[cid].Name == data.WJQF || b.cards[cid].Name == data.NMRQ {
					if b.getTeamMateCount() >= b.getEnemyCount() {
						continue
					}
				}
				c = cid
				goto useCard
			}
			b.targetQuest <- cid
			targetInf := <-b.availableTargetrec
			if len(targetInf.TargetList) == 0 {
				continue
			}
			enemy := b.getEnemy(targetInf.TargetList...)
			if len(enemy) < int(targetInf.TargetNum) {
				break
			}
			//暂时只选一个
			randNum := rand.Intn(len(enemy))
			targetList = []data.PID{enemy[randNum]}
			c = cid
			goto useCard
		}
	}
	//检查要不要出杀与酒
	for _, cid := range useAbleCard {
		if b.cards[cid].Name == data.Attack || b.cards[cid].Name == data.FireAttack || b.cards[cid].Name == data.LightnAttack {
			//如果有杀检查可不可以打到敌人
			b.targetQuest <- cid
			inf := <-b.availableTargetrec
			if len(inf.TargetList) == 0 {
				break
			}
			enemy := b.getEnemy(inf.TargetList...)
			if b.players[b.pid].equips[data.WeaponSlot] != nil {
				if b.players[b.pid].equips[data.WeaponSlot].Name != data.QGJ {
					//检查藤甲
					if b.cards[cid].Name == data.Attack {
						for i := len(enemy) - 1; i >= 0; i-- {
							if b.players[enemy[i]].equips[data.ArmorSlot] != nil {
								if b.players[enemy[i]].equips[data.ArmorSlot].Name == data.TengJia {
									enemy = append(enemy[:i], enemy[i+1:]...)
								}
							}
						}
					}
					//检查仁王盾
					if b.cards[cid].Dec.ISBlack() {
						for i := len(enemy) - 1; i >= 0; i-- {
							if b.players[enemy[i]].equips[data.ArmorSlot] != nil {
								if b.players[enemy[i]].equips[data.ArmorSlot].Name == data.RWD {
									enemy = append(enemy[:i], enemy[i+1:]...)
								}
							}
						}
					}
				}
			}
			if len(enemy) == 0 {
				break
			}
			//检查有没有酒，有的话就喝
			for _, cid1 := range useAbleCard {
				if b.cards[cid1].Name == data.Drunk {
					c = cid1
					goto useCard
				}
			}
			//寻找血量最低的目标
			bestTarget := enemy[0]
			for _, target := range enemy {
				if b.players[target].hp < b.players[bestTarget].hp {
					bestTarget = target
				}
			}
			targetList = append(targetList, bestTarget)
			c = cid
			goto useCard
		}
	}
useCard:
	<-time.After(time.Duration(rand.Intn(1000)+1000) * time.Millisecond) //延时1-2s
	if c == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: c, TargetList: targetList}
	b.waitForStateSwitch()
}

func (b *Bot) handleDodge() {
	useAbleCard := <-b.useAbleReceiver
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(200)+200) * time.Millisecond) //延时1-2s
	if len(useAbleCard) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: useAbleCard[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleDie() {
	b.players[b.curPlayer].death = true
	b.waitForStateSwitch()
}

func (b *Bot) handleWXKJ() {
	cards := <-b.useAbleReceiver
	rsp := <-b.useSkillRspRec
	//接受可用技能
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-2s
	select {
	case inf := <-b.gameStateRec:
		b.switchGameState(inf)
		return
	default:
	}
	if len(cards) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	//根据卡名称决定要不要无懈
	isTeamMate := b.isTeamMate(rsp.Targets[0])
	enable := rsp.Args[1]%2 == 0 //true为生效
	if isItemInList([]data.CardName{data.WJQF, data.NMRQ}, data.CardName(rsp.Args[0])) {
		if isTeamMate && enable {
			goto useCard
		}
		if !isTeamMate && !enable {
			goto useCard
		}
	}
	if isItemInList([]data.CardName{data.TYJY, data.WGFD}, data.CardName(rsp.Args[0])) {
		if isTeamMate && !enable {
			goto useCard
		}
		if !isTeamMate && enable {
			goto useCard
		}
	}
	if isItemInList([]data.CardName{data.LBSS, data.BLCD}, data.CardName(rsp.Args[0])) {
		if isTeamMate && enable {
			goto useCard
		}
	}
	if !b.isTeamMate(b.curPlayer) {
		if isItemInList([]data.CardName{data.SSQY, data.GHCQ, data.Duel, data.Burn, data.WXKJ, data.WZSY}, data.CardName(rsp.Args[0])) {
			b.useCardInf <- data.UseCardInf{ID: cards[0]}
			b.waitForStateSwitch()
			return
		}
	}
	b.useCardInf <- data.UseCardInf{Skip: true}
	b.waitForStateSwitch()
	return
useCard:
	b.useCardInf <- data.UseCardInf{ID: cards[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleDying() {
	useAbleCard := <-b.useAbleReceiver
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(1500)+500) * time.Millisecond) //延时0.5-2s
	if len(useAbleCard) != 0 {
		for i := 0; i < len(b.players); i++ {
			if b.players[i].hp <= 0 && !b.players[i].death && b.isTeamMate(data.PID(i)) {
				b.useCardInf <- data.UseCardInf{ID: useAbleCard[0]}
				b.waitForStateSwitch()
				return
			}
		}
	}
	b.useCardInf <- data.UseCardInf{Skip: true}
	b.waitForStateSwitch()
}

func (b *Bot) handleDuel() {
	useAbleCard := <-b.useAbleReceiver
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(1500)+500) * time.Millisecond) //延时0.5-2s
	if len(useAbleCard) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: useAbleCard[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleNMRQ() {
	useAbleCard := <-b.useAbleReceiver
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(1500)+500) * time.Millisecond) //延时0.5-2s
	if len(useAbleCard) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: useAbleCard[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleWJQF() {
	useAbleCard := <-b.useAbleReceiver
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(1500)+500) * time.Millisecond) //延时0.5-2s
	if len(useAbleCard) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: useAbleCard[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleDrop() {
	dropInf := <-b.dropAbleRec
	dropNum := dropInf.dropNum
	cards := dropInf.cards
	if dropNum == 0 {
		b.waitForStateSwitch()
		return
	}
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	//对cards按照从大到小排序
	for end := len(cards) - 1; end > 1; end-- {
		for i := 0; i < end; i++ {
			if cards[i] < cards[i+1] {
				cards[i], cards[i+1] = cards[i+1], cards[i]
			}
		}
	}
	b.dropCardInf <- cards[:dropNum]
	b.waitForStateSwitch()
}

func (b *Bot) handleWGFD() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	for end := len(b.wgfdCards) - 1; end > 1; end-- {
		for i := 0; i < end; i++ {
			if b.wgfdCards[i] > b.wgfdCards[i+1] {
				b.wgfdCards[i], b.wgfdCards[i+1] = b.wgfdCards[i+1], b.wgfdCards[i]
			}
		}
	}
	b.useCardInf <- data.UseCardInf{ID: b.wgfdCards[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleBurnShow() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	b.useCardInf <- data.UseCardInf{ID: b.players[b.pid].cards[0]}
	b.waitForStateSwitch()
}

func (b *Bot) handleBurnDrop() {
	useAbleCard := <-b.useAbleReceiver
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	if len(useAbleCard) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: useAbleCard[rand.Intn(len(useAbleCard))]}
	b.waitForStateSwitch()
}

func (b *Bot) handleSSQY() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	cid := data.CID(0)
	if len(b.ssqyCards) == 0 {
		goto useCard
	}
	//先试图拿装备
	for i := 0; i < 4; i++ {
		if b.ssqyCards[i] != 0 {
			cid = b.ssqyCards[i]
			goto useCard
		}
	}
	//再试图拿手牌
	cid = b.ssqyCards[len(b.ssqyCards)-1]
useCard:
	b.ssqyCards = nil
	b.useCardInf <- data.UseCardInf{ID: cid}
	b.waitForStateSwitch()
}

func (b *Bot) handleGHCQ() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	cid := data.CID(0)
	if len(b.ghcqCards) == 0 {
		goto useCard
	}
	//先试图拿装备
	for i := 0; i < 4; i++ {
		if b.ghcqCards[i] != 0 {
			cid = b.ghcqCards[i]
			goto useCard
		}
	}
	//再试图拿手牌
	cid = b.ghcqCards[len(b.ghcqCards)-1]
useCard:
	b.ghcqCards = nil
	b.useCardInf <- data.UseCardInf{ID: cid}
	b.waitForStateSwitch()
}

func (b *Bot) handleJDSRState() {
	<-b.useAbleSkillRec
	<-b.useAbleReceiver
	<-time.After(time.Duration(rand.Intn(500)+2500) * time.Millisecond) //延时0.5-1s
	b.useCardInf <- data.UseCardInf{Skip: true}
	b.waitForStateSwitch()
}

func (b *Bot) handleCtuAtk() {
	cards := <-b.useAbleReceiver
	<-b.useAbleSkillRec
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	if len(cards) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
	} else {
		b.useCardInf <- data.UseCardInf{ID: cards[0]}
	}
	b.waitForStateSwitch()
}

func (b *Bot) handleSkillSelectState() {
	targets := <-b.availableTargetrec
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	if targets.TargetNum > 1 || len(targets.TargetList) < int(targets.TargetNum) {
		b.useSkillInf <- data.UseSkillInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	friendlySkill := []data.SID{data.YingYuanSkill, data.JiQiaoSkill}
	unUseSkill := []data.SID{data.LiYuSkill, data.ZhenGuSkill, data.HBJSkill, data.JinQuSkill, data.YingHunSkill, data.QinYinSkill}
	if isItemInList(unUseSkill, b.selectSkill) {
		b.useSkillInf <- data.UseSkillInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	var realTarget []data.PID
	if isItemInList(friendlySkill, b.selectSkill) {
		realTarget = b.getTeamMate(targets.TargetList...)
	} else {
		realTarget = b.getEnemy(targets.TargetList...)
	}
	if len(realTarget) < 1 {
		if targets.TargetNum == 0 {
			b.useSkillInf <- data.UseSkillInf{ID: b.selectSkill}
		} else {
			b.useSkillInf <- data.UseSkillInf{Skip: true}
		}
		b.waitForStateSwitch()
		return
	}
	b.useSkillInf <- data.UseSkillInf{ID: b.selectSkill, TargetList: []data.PID{realTarget[rand.Intn(len(realTarget))]}}
	b.waitForStateSwitch()
}

func (b *Bot) handleSkillJudgeState() {
	for {
		select {
		case inf := <-b.gsidCardReceiver:
			b.handleGSIDCardsInf(inf)
		case inf := <-b.gameStateRec:
			b.switchGameState(inf)
			return
		}
	}
}

func (b *Bot) handleDropAll() {
	dropInf := <-b.dropAbleRec
	dropNum := dropInf.dropNum
	cards := dropInf.cards
	if dropNum == 0 {
		b.waitForStateSwitch()
		return
	}
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	//对cards按照从大到小排序
	for end := len(cards) - 1; end > 1; end-- {
		for i := 0; i < end; i++ {
			if cards[i] < cards[i+1] {
				cards[i], cards[i+1] = cards[i+1], cards[i]
			}
		}
	}
	b.dropCardInf <- cards[:dropNum]
	b.waitForStateSwitch()
}

func (b *Bot) handleQLG() {
	<-time.After(time.Duration(rand.Intn(500)+2500) * time.Millisecond) //延时0.5-1s
	cid := data.CID(0)
	if b.qlgCards[0] != 0 {
		cid = data.CID(b.qlgCards[0])
	} else {
		cid = data.CID(b.qlgCards[1])
	}
	b.useCardInf <- data.UseCardInf{ID: data.CID(cid)}
	b.waitForStateSwitch()
}

func (b *Bot) handleCXSGJ() {
	<-b.useSkillRspRec
	<-time.After(time.Duration(rand.Intn(500)+2500) * time.Millisecond) //延时0.5-1s
	b.useSkillInf <- data.UseSkillInf{ID: data.CXSGJSkill, Args: []byte{1}}
	b.waitForStateSwitch()
}

func (b *Bot) handleQueDi() {
	rsp := <-b.useSkillRspRec
	<-time.After(time.Duration(rand.Intn(500)+1500) * time.Millisecond) //延时0.5-1s
	var args []byte
	if rsp.Args[0] == 0 {
		args = []byte{0, 1}
	} else {
		args = []byte{1, 0}
	}
	b.useSkillInf <- data.UseSkillInf{ID: data.QueDiSkill, Args: args}
	b.waitForStateSwitch()
}

func (b *Bot) handleFengPo() {
	<-time.After(time.Duration(rand.Intn(500)+2500) * time.Millisecond) //延时0.5-1s
	b.useCardInf <- data.UseCardInf{Skip: false}
	b.waitForStateSwitch()
}

func (b *Bot) handleLiangZhu() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	b.useCardInf <- data.UseCardInf{Skip: true}
	b.waitForStateSwitch()
}

func (b *Bot) handleDropOtherCard() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	cid := data.CID(0)
	if len(b.otherCards) == 0 {
		goto useCard
	}
	//先试图拿装备
	for i := 0; i < 4; i++ {
		if b.otherCards[i] != 0 {
			cid = b.otherCards[i]
			goto useCard
		}
	}
	//再试图拿手牌
	cid = b.otherCards[len(b.otherCards)-1]
useCard:
	b.otherCards = nil
	b.useCardInf <- data.UseCardInf{ID: cid}
	b.waitForStateSwitch()
}

func (b *Bot) handleDropSelfCard() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	dropInf := <-b.dropAbleRec
	dropNum := dropInf.dropNum
	cards := dropInf.cards
	//对cards按照从大到小排序
	for end := len(cards) - 1; end > 1; end-- {
		for i := 0; i < end; i++ {
			if cards[i] < cards[i+1] {
				cards[i], cards[i+1] = cards[i+1], cards[i]
			}
		}
	}
	b.dropCardInf <- cards[:dropNum]
	b.waitForStateSwitch()
}

func (b *Bot) handleWenJi() {
	inf := <-b.dropAbleRec
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	b.useCardInf <- data.UseCardInf{ID: inf.cards[len(inf.cards)-1]}
	b.waitForStateSwitch()
}

func (b *Bot) handleEnYuan() {
	<-b.useAbleReceiver
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	b.useCardInf <- data.UseCardInf{Skip: true}
	b.waitForStateSwitch()
}

func (b *Bot) handleGuanXing() {
	inf := <-b.useSkillRspRec
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	b.useSkillInf <- data.UseSkillInf{ID: data.GuanXingSkill, Cards: inf.Cards, Args: []byte{0}}
	b.waitForStateSwitch()
}

func (b *Bot) handleChangeRole() {
	//接收玩家列表
	pList := <-b.playerInfRec
	for _, inf := range pList {
		b.players[inf.PID] = newPlayer(inf.Role, inf.PID)
	}
	b.waitForStateSwitch()
}

func (b *Bot) handlePoJun() {
	inf := <-b.useSkillRspRec
	cards := []data.CID{}
	for _, c := range inf.Cards {
		if c != 0 {
			cards = append(cards, c)
		}
		if len(cards) == int(inf.Args[0]) {
			break
		}
	}
	b.useSkillInf <- data.UseSkillInf{ID: data.PoJunSkill, Cards: cards}
	b.waitForStateSwitch()
}

func (b *Bot) handleFanKui() {
	<-time.After(time.Duration(rand.Intn(500)+500) * time.Millisecond) //延时0.5-1s
	inf := <-b.useSkillRspRec
	for _, c := range inf.Cards {
		if c != 0 {
			b.useSkillInf <- data.UseSkillInf{ID: data.FanKuiSkill, Cards: []data.CID{c}}
			b.waitForStateSwitch()
			return
		}
	}
	b.waitForStateSwitch()
}

func (b *Bot) handleLuanWu() {
	//从服务器获取可用卡牌
	useAbleCard := <-b.useAbleReceiver
	//接受可用技能
	<-b.useAbleSkillRec
	if len(useAbleCard) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.targetQuest <- 0
	inf := <-b.availableTargetrec
	enemy := b.getEnemy(inf.TargetList...)
	if len(enemy) == 0 {
		b.useCardInf <- data.UseCardInf{Skip: true}
		b.waitForStateSwitch()
		return
	}
	b.useCardInf <- data.UseCardInf{ID: useAbleCard[0], TargetList: []data.PID{enemy[0]}}
	b.waitForStateSwitch()
}

func (b *Bot) handleGetSkill() {
	<-b.useSkillRspRec
	<-time.After(time.Duration(500) * time.Millisecond) //延时0.5-1s
	b.useSkillInf <- data.UseSkillInf{ID: data.GetSkill, Args: []byte{1}}
	b.waitForStateSwitch()
}
