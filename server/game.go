package server

import (
	"goltk/data"
	"math/rand"
	"strconv"
	"time"
)

type reconnInf struct {
	id      data.PID
	onStart func()
}

type Games struct {
	mode          data.GameMode
	events        eventList //事件列表
	turnOwner     data.PID
	index         int //事件索引
	priorityEvent []eventI
	priorityIndex int
	clients       []clientI  //客户端列表
	players       []*player  //玩家列表
	curPlayer     data.PID   //当前回合玩家pid
	cards         []cardI    //存储了所有牌的数组
	mainHeap      []data.CID //主牌堆
	dropHeap      []data.CID //弃牌堆
	qgzxState     uint8      //驱鬼逐邪第几阶段
	closeSignal   chan struct{}
	reconnInf     chan reconnInf
}

func NewGame(mode data.GameMode) *Games {
	return &Games{mode: mode, closeSignal: make(chan struct{}), reconnInf: make(chan reconnInf, 1)}
}

// 添加高优先级事件
func (g *Games) addPriorityEvent(e eventI) {
	g.priorityEvent = append(g.priorityEvent, e)
}

// 向玩家发牌
func (g *Games) sendCard2Player(pid data.PID, card ...data.CID) {
	g.players[pid].addCard(card...)
	//自书
	if g.players[pid].hasEffect(ziShuEffect) {
		g.players[pid].findSkill(data.ZiShuSkill).(*zishuSkill).check(g, pid, card...)
	}
	iterators(g.clients, func(c clientI) { c.SendCard(pid, card...) })
}

// 将玩家牌移除
func (g *Games) removePlayercard(pid data.PID, card ...data.CID) {
	cards := make([]data.CID, len(card))
	copy(cards, card)
	g.players[pid].delCard(cards...)
	iterators(g.clients, func(c clientI) { c.RemoveCard(pid, cards...) })
}

// 在玩家间移动卡牌
func (g *Games) moveCard(src, dst data.PID, card ...data.CID) {
	cards := make([]data.CID, len(card))
	copy(cards, card)
	switch src {
	case data.SpecialPIDGame:
		for _, cid := range cards {
			for i, c := range g.dropHeap {
				if cid == c {
					g.dropHeap = append(g.dropHeap[:i], g.dropHeap[i+1:]...)
					goto loopEnd
				}
			}
			for i, c := range g.mainHeap {
				if cid == c {
					g.mainHeap = append(g.mainHeap[:i], g.mainHeap[i+1:]...)
					goto loopEnd
				}
			}
			panic("ID=" + strconv.Itoa(int(cid)) + "的牌不在牌堆与弃牌堆")
		loopEnd:
		}
	default:
		g.players[src].delCard(cards...)
	}
	g.players[dst].addCard(cards...)
	//自书
	if g.players[dst].hasEffect(ziShuEffect) {
		g.players[dst].findSkill(data.ZiShuSkill).(*zishuSkill).check(g, dst, cards...)
	}
	iterators(g.clients, func(c clientI) { c.MoveCard(src, dst, cards...) })
}

// 广播玩家用牌操作
func (g *Games) useCard(user data.PID, card data.CID, targets ...data.PID) {
	if card == 0 {
		return
	}
	iterators(g.clients, func(c clientI) { c.UseCard(user, card, targets...) })
}

func (g *Games) useSkill(user data.PID, sid data.SID, args ...byte) {
	iterators(g.clients, func(c clientI) { c.UseSkill(user, sid, nil, args...) })
}

func (g *Games) useTmpCard(user data.PID, name data.CardName, dec data.Decor, num data.CNum,
	tmpType data.TmpCardType, target ...data.PID) {
	iterators(g.clients, func(c clientI) { c.UseTmpCard(user, name, dec, num, tmpType, target...) })
}

// 设置当前阶段
func (g *Games) setGameState(state data.GameState, t time.Duration, curPlayer data.PID) {
	g.curPlayer = curPlayer
	iterators(g.clients, func(c clientI) { c.SetGameState(state, t, curPlayer) })
}

// 取牌
func (g *Games) getCards(n int, p data.PID) []data.CID {
	if g.players[p].hasEffect(cunmuEffect) {
		g.useSkill(p, data.CunMuSkill)
		return g.getCardsFromBottom(n)
	}
	return g.getCardsFromTop(n)
}

// 从主牌堆取n张牌
func (g *Games) getCardsFromTop(n int) []data.CID {
	if len(g.mainHeap) < n {
		rand.Shuffle(len(g.dropHeap), func(i, j int) { g.dropHeap[i], g.dropHeap[j] = g.dropHeap[j], g.dropHeap[i] })
		g.mainHeap = append(g.mainHeap, g.dropHeap...)
		// for i := 0; i < len(g.mainHeap); i++ {
		// 	for j := i + 1; j < len(g.mainHeap); j++ {
		// 		if g.mainHeap[i] == g.mainHeap[j] {
		// 			panic("")
		// 		}
		// 	}
		// }
		g.useSkill(0, data.AddHeapNum, byte(len(g.dropHeap)))
		g.dropHeap = nil
	}
	cards := make([]data.CID, n)
	copy(cards, g.mainHeap[:n])
	g.mainHeap = g.mainHeap[n:]
	return cards
}

// 从牌堆底取n张牌
func (g *Games) getCardsFromBottom(n int) []data.CID {
	if len(g.mainHeap) < n {
		rand.Shuffle(len(g.dropHeap), func(i, j int) { g.dropHeap[i], g.dropHeap[j] = g.dropHeap[j], g.dropHeap[i] })
		g.mainHeap = append(g.mainHeap, g.dropHeap...)
		g.useSkill(0, data.AddHeapNum, byte(len(g.dropHeap)))
		g.dropHeap = nil
	}
	cards := make([]data.CID, n)
	copy(cards, g.mainHeap[len(g.mainHeap)-n:])
	g.mainHeap = g.mainHeap[:len(g.mainHeap)-n]
	return cards
}

// 将牌置入弃牌堆
func (g *Games) dropCards(cards ...data.CID) {
	for _, c := range cards {
		if c == 0 {
			panic("")
		}
	}
	testMap := map[data.CID]struct{}{}
	//为cards本身查重
	for _, c := range cards {
		testMap[c] = struct{}{}
	}
	if len(testMap) != len(cards) {
		panic("")
	}
	//对弃牌堆与主牌堆查重
	testMap = map[data.CID]struct{}{}
	for _, c := range g.dropHeap {
		testMap[c] = struct{}{}
	}
	for _, c := range g.mainHeap {
		testMap[c] = struct{}{}
	}
	for _, c := range cards {
		if _, ok := testMap[c]; ok {
			panic("")
		}
	}
	g.dropHeap = append(g.dropHeap, cards...)
}

// 获取存活的玩家个数
func (g *Games) getAlivePlayerCount() (count int) {
	for i := 0; i < len(g.players); i++ {
		if g.players[i].death {
			continue
		}
		count++
	}
	return
}

func (g *Games) Run() {
	g.init()
	for {
		select {
		case <-g.closeSignal:
			return
		case inf := <-g.reconnInf:
			g.addPriorityEvent(newReConnEvent(inf.id, inf.onStart))
			continue
		default:
		}
		//检查高优先级事件
		if g.priorityIndex < len(g.priorityEvent) {
			if g.priorityEvent[g.priorityIndex].isSkipAble() {
				g.priorityIndex++
				continue
			}
			g.priorityEvent[g.priorityIndex].trigger(g)
			g.priorityIndex++
			if g.priorityIndex == len(g.priorityEvent) {
				g.priorityIndex = 0
				g.priorityEvent = nil
				g.index++
			}
			continue
		}
		//检查普通事件
		if g.events.list[g.index].isSkipAble() {
			g.index++
			continue
		}
		g.events.list[g.index].trigger(g)
		if len(g.priorityEvent) == 0 {
			g.index++
		}
	}
}

func (g *Games) AddClient(c clientI) {
	c.SetPid(data.PID(len(g.clients)))
	g.clients = append(g.clients, c)
}

func (g *Games) Close() {
	close(g.closeSignal)
	for _, c := range g.clients {
		c.Close()
	}
}

func (g *Games) ClientReConn(id data.PID, onStart func()) {
	g.reconnInf <- reconnInf{id: id, onStart: onStart}
}

// 获取下一个玩家id
func (g *Games) getNextPid(pid data.PID) data.PID {
	for {
		pid++
		if pid == data.PID(len(g.players)) {
			pid = 0
		}
		if g.players[pid].death {
			continue
		}
		return pid
	}
}

// 获取下一个回合者
func (g *Games) getNextTurnOwner(pid data.PID) data.PID {
	for {
		pid++
		if pid == data.PID(len(g.players)) {
			pid = 0
		}
		if g.players[pid].death {
			continue
		}
		if g.players[pid].turnBack {
			g.players[pid].turnBack = false
			g.useSkill(pid, data.TurnBackSkill)
			continue
		}
		return pid
	}
}

// 获取上一个玩家id
func (g *Games) getPrvPid(pid data.PID) data.PID {
	for {
		pid--
		if pid == -1 {
			pid = data.PID(len(g.players) - 1)
		}
		if g.players[pid].death {
			continue
		}
		return pid
	}
}

// 获取所有存活玩家
func (g *Games) getAllAlivePlayer() (list []data.PID) {
	for i := 0; i < len(g.players); i++ {
		if g.players[i].death {
			continue
		}
		list = append(list, g.players[i].pid)
	}
	return
}

// 获取除了自己以外的所有存活玩家
func (g *Games) getAllAliveOther(user data.PID) (list []data.PID) {
	for i := 0; i < len(g.players); i++ {
		if g.players[i].death || g.players[i].pid == user {
			continue
		}
		list = append(list, g.players[i].pid)
	}
	return
}

type distence data.Distence

// 返回p1到p2的距离
func (g *Games) getDst(p1 data.PID, p2 data.PID) distence {
	if p1 == p2 {
		return 0
	}
	tmpDstN, tmpDstP := 0, 0
	for tmpp := p1; tmpp != p2; tmpp = g.getNextPid(tmpp) {
		tmpDstN++
	}
	for tmpp := p1; tmpp != p2; tmpp = g.getPrvPid(tmpp) {
		tmpDstP++
	}
	dst := min(tmpDstN, tmpDstP)
	dst -= int(g.players[p1].getEffectCount(dstDffEffect))
	dst += int(g.players[p2].getEffectCount(dstUpEffect))
	if g.players[p1].equipSlot[data.HorseDownSlot] != nil {
		dst -= 1
	}
	if g.players[p2].equipSlot[data.HorseUpSlot] != nil {
		dst++
	}
	return distence(max(1, dst))
}

// 获取与p距离不超过dst的玩家列表
func (g *Games) getPlayerInDst(p data.PID, dst distence) (pList []data.PID) {
	for p2 := g.getNextPid(p); p2 != p; p2 = g.getNextPid(p2) {
		if g.getDst(p, p2) <= dst {
			pList = append(pList, p2)
		}
	}
	return
}

func (g *Games) newGameEventList(pid data.PID) []eventI {
	p := g.players[pid]
	e := []eventI{}
	//落雷
	if p.hasEffect(luoLeiEffect) {
		e = append(e, newSkillSelectEvent(data.LuoLeiSkill, pid, nil))
	}
	//准备
	e = append(e, newPrePareEvent(pid))
	//判定
	e = append(e, newJudgeEvent(pid))
	//摸牌
	e = append(e, newSendCardEvent(pid))
	//检查问计
	if p.hasEffect(wenJiEffect) {
		e = append(e, newSkillSelectEvent(data.WenJiSkill, pid, nil))
	}
	//利驭
	if p.hasEffect(liyuEffect) {
		e = append(e, newSkillSelectEvent(data.LiYuSkill, pid, nil))
	}
	//潜袭
	if p.hasEffect(qianXiEffect) {
		e = append(e, newSkillSelectEvent(data.QianXiSkill, pid, nil))
	}
	//出牌
	e = append(e, newUseCardEvent(pid))
	//弃牌
	e = append(e, newDropCardEvent(pid))
	//地动
	if p.hasEffect(diDongEffect) {
		e = append(e, newSkillSelectEvent(data.DiDongSkill, pid, nil))
	}
	if p.hasEffect(guiHuoEffect) {
		e = append(e, newSkillSelectEvent(data.GuiHuoSkill, pid, nil))
	}
	//结束
	e = append(e, newEndEvent(pid))
	//下一个
	e = append(e, newGo2NextEvent(pid))
	return e
}

func (g *Games) init() {
	//初始化cards与主牌堆
	g.cards = newCardList()
	g.mainHeap = make([]data.CID, len(g.cards)-1)
	for i := 1; i < len(g.cards); i++ {
		g.mainHeap[i-1] = g.cards[i].getID()
	}
	rand.Shuffle(len(g.mainHeap), func(i, j int) { g.mainHeap[i], g.mainHeap[j] = g.mainHeap[j], g.mainHeap[i] })
	//向客户端发送可选角色
	roleList := data.GetRoleList()
	rand.Shuffle(len(roleList), func(i, j int) { roleList[i], roleList[j] = roleList[j], roleList[i] })
	switch g.mode {
	case data.QGZXEasyMode:
		for _, c := range g.clients[:3] {
			c.SendAvailableRole(roleList[:8]...)
			roleList = roleList[8:]
		}
		// g.clients[0].SendAvailableRole(roleList[:8]...)
		// g.clients[0].SendAvailableRole(data.GetRoleList()[2])
		// for _, c := range g.clients[1:3] {
		// 	c.SendAvailableRole(data.GetRoleList()[2])
		// }
		// g.clients[3].SendAvailableRole(data.GhostList[rand.Intn(4)])
		g.clients[3].SendAvailableRole(data.GhostList[0])
	case data.QGZXNormalMode:
		for _, c := range g.clients[:3] {
			c.SendAvailableRole(roleList[:6]...)
			roleList = roleList[6:]
		}
		// g.clients[2].SendAvailableRole(data.GetRoleList()[11])
		g.clients[3].SendAvailableRole(data.GhostList[rand.Intn(4)])
	case data.QGZXHardMode:
		for _, c := range g.clients[:3] {
			c.SendAvailableRole(roleList[:8]...)
			roleList = roleList[8:]
		}
		g.clients[3].SendAvailableRole(data.GhostList[rand.Intn(4)])
		g.clients[4].SendAvailableRole(data.GhostList[4])
		g.clients[5].SendAvailableRole(data.GhostList[rand.Intn(4)])
	case data.QGZXVeryHardMode:
		for _, c := range g.clients[:3] {
			c.SendAvailableRole(roleList[:8]...)
			roleList = roleList[8:]
		}
		g.clients[3].SendAvailableRole(data.GhostList[rand.Intn(4)])
		g.clients[4].SendAvailableRole(data.GhostList[4])
		g.clients[5].SendAvailableRole(data.GhostList[rand.Intn(4)])
	case data.QGZXDoubleMode:
		for _, c := range g.clients[:3] {
			c.SendAvailableRole(roleList[:8]...)
			roleList = roleList[8:]
		}
		// g.clients[2].SendAvailableRole(data.GetRoleList()[11])
		g.clients[3].SendAvailableRole(data.GroupGhostList[0])
	case data.QGZXFreeMode:
		for _, c := range g.clients[:2] {
			c.SendAvailableRole(roleList[:8]...)
			roleList = roleList[8:]
		}
		g.clients[2].SendAvailableRole(data.GhostList[rand.Intn(4)])
	case data.NianShouMode:
		for _, c := range g.clients[:3] {
			c.SendAvailableRole(roleList[:8]...)
			roleList = roleList[8:]
		}
		g.clients[3].SendAvailableRole(data.MonsterList[2])
	default:
		iterators(g.clients, func(c clientI) { c.SendAvailableRole(roleList[:6]...); roleList = roleList[6:] })
	}
	//等待客户端选择角色
	playerInfList := []data.PlayerInf{}
	playerSkillList := [][]data.SID{}
	for i := 0; i < len(g.clients); i++ {
		select {
		case role := <-g.clients[i].GetRole():
			playerSkillList = append(playerSkillList, role.SkillList)
			p := newPlayer(g, role, data.PID(i))
			g.players = append(g.players, &p)
			playerInfList = append(playerInfList, data.PlayerInf{Role: role, PID: data.PID(i)})
		case <-g.closeSignal:
			return
		}
	}
	//广播玩家列表
	iterators(g.clients, func(c clientI) { c.SendPlayerInf(playerInfList) })
	//向玩家发牌
	for pid := 0; pid < len(g.players); pid++ {
		cards := g.getCards(4, data.PID(pid))
		g.sendCard2Player(data.PID(pid), cards...)
	}
	//初始化玩家技能
	for i := 0; i < len(g.players); i++ {
		for _, sid := range playerSkillList[i] {
			g.players[i].addSkill(sid)
		}
	}
	//初始化事件列表
	g.events.list = g.newGameEventList(0)
	switch g.mode {
	case data.NianShouMode:
		for i := 0; i != 3; i++ {
			g.addPriorityEvent(newGetSkillEvent(data.PID(i)))
		}
	}
}

func (g *Games) recover(user data.PID, target data.PID, num data.HP) {
	_, t := g.players[user], g.players[target]
	num = min(num, t.maxHp-t.hp)
	if num <= 0 {
		return
	}
	t.hp += num
	g.setGameState(data.SetHpState, 0, target)
	iterators(g.clients, func(c clientI) { c.SetHP(target, t.hp, data.Recover) })
	//恩怨
	if t.hasEffect(enYuanEffect) && user != target {
		g.sendCard2Player(user, g.getCards(int(num), user)...)
	}
	//良助
	for _, p := range g.players {
		if p.hasEffect(liangzhuEffect) && g.turnOwner == target {
			g.events.insert(g.index, newSkillSelectEvent(data.LiangZhuSkill, p.pid, []data.PID{target}))
		}
	}
	<-time.After(100 * time.Millisecond)
}

func (g *Games) addHpMax(user data.PID, num uint8) {
	g.players[user].maxHp += data.HP(num)
	g.setGameState(data.SetHpState, 0, user)
	iterators(g.clients, func(c clientI) { c.SetHP(user, g.players[user].maxHp, data.AddHpMax) })
	<-time.After(100 * time.Millisecond)
}

// 获得造成伤害的牌
func (g *Games) getDmgCard(cid data.CID) bool {
	for i := len(g.dropHeap) - 1; i >= 0; i-- {
		if g.cards[g.dropHeap[i]].getID() == cid {
			g.dropHeap = append(g.dropHeap[:i], g.dropHeap[i+1:]...)
			return true
		}
	}
	for i := 0; i < len(g.mainHeap); i++ {
		if g.cards[g.mainHeap[i]].getID() == cid {
			g.mainHeap = append(g.mainHeap[:i], g.mainHeap[i+1:]...)
			return true
		}
	}
	print("弃牌堆和主牌堆都没有造成伤害的牌")
	return false
}
