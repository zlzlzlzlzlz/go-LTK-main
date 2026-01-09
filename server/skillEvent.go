package server

import (
	"goltk/data"
	"math/rand/v2"
	"time"
)

// 自己弃手牌
type dropSelfHandCardEvent struct {
	event
	target data.PID
	num    uint8
}

func newDropSelfHandCardEvent(target data.PID, num uint8) *dropSelfHandCardEvent {
	return &dropSelfHandCardEvent{target: target, num: num}
}

func (e *dropSelfHandCardEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.DropSelfHandCard, lastTime, e.target)
	g.clients[e.target].SendDropAbleCard(g.players[e.target].cards, e.num)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			g.dropCards(g.players[e.target].cards[:e.num]...)
			g.removePlayercard(e.target, g.players[e.target].cards[:e.num]...)
			return
		case cards := <-g.clients[e.target].GetDropCardInf():
			g.dropCards(cards...)
			g.removePlayercard(e.target, cards...)
			return
		}
	}
}

// 弃别人牌事件
type dropOtherCardEvent struct {
	event
	user, target data.PID
}

func newDropOtherCardEvent(user, target data.PID) *dropOtherCardEvent {
	return &dropOtherCardEvent{user: user, target: target}
}

func (e *dropOtherCardEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	cards := []data.CID{}
	t := g.players[e.target]
	//按照目标装备槽,判定区,手牌堆的顺序生成cards
	for _, c := range t.equipSlot {
		if c == nil {
			cards = append(cards, 0)
			continue
		}
		cards = append(cards, c.getID())
	}
	for _, c := range t.judgeSlot {
		if c == nil {
			cards = append(cards, 0)
			continue
		}
		cards = append(cards, c.getID())
	}
	cards = append(cards, t.cards...)
	//发送牌堆
	g.clients[e.user].SendGSCards(data.GSIDDropOtrCard, cards...)
	//设置游戏阶段
	g.setGameState(data.DropOtherCardState, lastTime, e.user)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			for i := len(cards) - 1; i >= 0; i-- {
				if cards[i] == 0 {
					continue
				}
				g.dropCards(cards[i])
				g.removePlayercard(e.target, cards[i])

				break
			}
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.ID == 0 || !isItemInList(cards, inf.ID) {
				continue
			}
			g.dropCards(inf.ID)
			g.removePlayercard(e.target, inf.ID)
			return
		}
	}
}

// 破军
type poJunEvent struct {
	event
	user, target data.PID
}

func newPoJunEvent(user, target data.PID) *poJunEvent {
	return &poJunEvent{user: user, target: target}
}

func (e *poJunEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	cards := []data.CID{}
	t := g.players[e.target]
	//依次添加装备区与手牌堆
	cardNum := len(t.cards)
	for i := 0; i < len(t.equipSlot); i++ {
		if t.equipSlot[i] == nil {
			cards = append(cards, 0)
			continue
		}
		cards = append(cards, t.equipSlot[i].getID())
		cardNum++
	}
	cards = append(cards, t.cards...)
	cardNum = min(int(t.hp), cardNum)
	//发送牌堆
	g.clients[e.user].SendUseSkillRsp(data.UseSkillRsp{ID: data.PoJunSkill, Cards: cards, Args: []byte{uint8(cardNum)}})
	//设置游戏阶段
	g.setGameState(data.PoJunState, lastTime, e.user)
	timer := time.After(lastTime)
	var selCard []data.CID
	for {
		select {
		case <-timer:
			for _, c := range cards {
				if c == 0 {
					continue
				}
				selCard = append(selCard, c)
				if len(selCard) == cardNum {
					goto finish
				}
			}
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if inf.ID != data.PoJunSkill || !isListContainAllTheItem(cards, inf.Cards...) ||
				len(inf.Cards) > cardNum || len(inf.Cards) == 0 {
				continue
			}
			selCard = inf.Cards
			goto finish
		}
	}
finish:
	arg := []byte{byte(e.target)}
	for _, c := range selCard {
		arg = append(arg, byte(c))
	}
	g.useSkill(e.user, data.PoJunSkill, arg...)
	t.delCard(selCard...)
	t.addTSCard(data.PoJunSkill, selCard...)
}

// 问计
type wenJiEvent struct {
	event
	user, target data.PID
}

func newWenJiEvent(user, target data.PID) *wenJiEvent {
	return &wenJiEvent{user: user, target: target}
}

func (e *wenJiEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	t := g.players[e.target]
	g.setGameState(data.WenJiState, lastTime, e.target)
	cards := []data.CID{}
	for _, equip := range g.players[e.target].equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		cards = append(cards, equip.getID())
	}
	cards = append(cards, t.cards...)
	g.clients[e.target].SendDropAbleCard(cards, 1)
	timer := time.After(lastTime)
	var recCard data.CID
	for {
		select {
		case <-timer:
			for _, c := range cards {
				if c != 0 {
					recCard = c
				}
			}
			goto finish
		case inf := <-g.clients[e.target].GetUseCardInf():
			if !isItemInList(cards, inf.ID) {
				continue
			}
			recCard = inf.ID
			goto finish
		}
	}
finish:
	g.moveCard(e.target, e.user, recCard)
	g.useSkill(e.user, data.WenJiSkill, byte(g.cards[recCard].getName()))
	g.players[e.user].findSkill(data.WenJiSkill).(*wenJiSkill).name = g.cards[recCard].getName()
}

// 铁骑
type tieJiEvent struct {
	event
	atkevent *atkevent
	result   *data.CID
}

func newTieJiEvent(atkevent *atkevent, result *data.CID) *tieJiEvent {
	return &tieJiEvent{atkevent: atkevent, result: result}
}

func (e *tieJiEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor().IsRed() {
		e.atkevent.unResponsive = true
	}
}

// 恩怨
type enYuanEvent struct {
	event
	user, target data.PID
}

func newEnYuanEvent(user, target data.PID) *enYuanEvent {
	return &enYuanEvent{target: target, user: user}
}

func (e *enYuanEvent) trigger(g *Games) {
	if !g.players[e.user].hasEffect(enYuanEffect) || g.players[e.target].death || g.players[e.user].death {
		return
	}
	g.useSkill(e.user, data.EnYuanSkill)
	const lastTime = 20 * time.Second
	t := g.players[e.target]
	list := []data.CID{}
	for _, c := range t.cards {
		if g.cards[c].getDecor() == data.HeartDec {
			list = append(list, c)
		}
	}
	g.setGameState(data.EnYuanState, lastTime, e.target)
	g.clients[e.target].SendUseAbleCards(list)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			e := newDamageEvent(e.target, e.target, data.BleedingDmg, nil, 1)
			g.events.insert(g.index, newBeforeDamageEvent(e), e, newAfterDmgEvent(e))
			return
		case inf := <-g.clients[e.target].GetUseCardInf():
			if inf.Skip {
				e := newDamageEvent(e.target, e.target, data.BleedingDmg, nil, 1)
				g.events.insert(g.index, newBeforeDamageEvent(e), e, newAfterDmgEvent(e))
				return
			}
			if !isItemInList(t.cards, inf.ID) || g.cards[inf.ID].getDecor() != data.HeartDec {
				continue
			}
			g.moveCard(e.target, e.user, inf.ID)
			return
		}
	}
}

// 观星
type guanxingEvent struct {
	event
	user data.PID
}

func newGuanXingEvent(user data.PID) *guanxingEvent {
	return &guanxingEvent{user: user}
}

func (e *guanxingEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.GuanXingState, lastTime, e.user)
	cards := g.getCardsFromTop(min(g.getAlivePlayerCount(), 5))
	g.clients[e.user].SendUseSkillRsp(data.UseSkillRsp{ID: data.GuanXingSkill, Cards: cards})
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			g.mainHeap = append(cards, g.mainHeap...)
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if len(inf.Args) != 1 || int(inf.Args[0]) > len(cards) || !isListContainAllTheItem(cards, inf.Cards...) {
				continue
			}
			//使用arg[0]=n作为分隔，[0,n)放入牌堆顶，其余放入底部
			n := inf.Args[0]
			top := inf.Cards[:n]
			buttom := append([]data.CID{}, inf.Cards[n:]...)
			g.mainHeap = append(top, g.mainHeap...)
			g.mainHeap = append(g.mainHeap, buttom...)
			return
		}
	}
}

// 琴音
type qinYinevent struct {
	event
	user data.PID
}

func newQinYinevent(user data.PID) *qinYinevent {
	return &qinYinevent{user: user}
}

func (e *qinYinevent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.QinYinState, lastTime, e.user)
	timer := time.After(lastTime)
	list := []eventI{}
	dmgevent := newDamageEvent(e.user, e.user, data.BleedingDmg, nil, 1)
	for {
		select {
		case <-timer:
			goto rec
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				goto dmg
			} else {
				goto rec
			}
		}
	}
dmg:
	list = append(list, newBeforeDamageEvent(dmgevent), dmgevent, newAfterDmgEvent(dmgevent))
	for pid := g.getNextPid(e.user); pid != e.user; pid = g.getNextPid(pid) {
		dmgevent := newDamageEvent(pid, pid, data.BleedingDmg, nil, 1)
		list = append(list, newBeforeDamageEvent(dmgevent), dmgevent, newAfterDmgEvent(dmgevent))
	}
	g.events.insert(g.index, list...)
	return
rec:
	g.recover(e.user, e.user, 1)
	for pid := g.getNextPid(e.user); pid != e.user; pid = g.getNextPid(pid) {
		g.recover(e.user, pid, 1)
	}
}

// 英魂
type yingHunEvent struct {
	event
	user, target data.PID
}

func newYingHunEvent(user, target data.PID) *yingHunEvent {
	return &yingHunEvent{user: user, target: target}
}

func (e *yingHunEvent) trigger(g *Games) {
	if g.players[e.user].death {
		return
	}
	const lastTime = 20 * time.Second
	g.setGameState(data.YingHunState, lastTime, e.user)
	timer := time.After(lastTime)
	num := uint8(g.players[e.user].maxHp - g.players[e.user].hp)
	for {
		select {
		case <-timer:
			//摸一弃x
			g.sendCard2Player(e.target, g.getCards(1, e.target)...)
			g.events.insert(g.index, newDropSelfHandCardEvent(e.target, num))
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				//摸一弃x
				g.sendCard2Player(e.target, g.getCards(1, e.target)...)
				g.events.insert(g.index, newDropSelfHandCardEvent(e.target, num))
				return
			} else {
				//摸x弃一
				g.sendCard2Player(e.target, g.getCards(int(num), e.target)...)
				g.events.insert(g.index, newDropSelfHandCardEvent(e.target, 1))
				return
			}
		}
	}
}

// 奇制
type qizhiEvent struct {
	event
	user, target data.PID
}

func newQiZhiEvent(user, target data.PID) *qizhiEvent {
	return &qizhiEvent{user: user, target: target}
}

func (e *qizhiEvent) trigger(g *Games) {
	if g.players[e.user].death {
		return
	}
	const lastTime = 20 * time.Second
	timer := time.After(lastTime)
	t := g.players[e.target]
	cards := t.cards
	if e.user == e.target {
		g.setGameState(data.DropSelfAllCards, lastTime, e.user)
		g.clients[e.user].SendDropAbleCard(g.players[e.user].cards, 1)
	} else {
		cards = nil
		g.setGameState(data.DropOtherCardState, lastTime, e.user)
		//依次添加装备区与手牌堆
		for i := 0; i < len(t.equipSlot); i++ {
			if t.equipSlot[i] == nil {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, t.equipSlot[i].getID())
		}
		for _, c := range t.judgeSlot {
			if c == nil {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		cards = append(cards, t.cards...)
		//发送牌堆
		g.clients[e.user].SendGSCards(data.GSIDDropOtrCard, cards...)
	}
	for {
		select {
		case <-timer:
			g.dropCards(cards[0])
			g.removePlayercard(e.target, cards[0])
			g.sendCard2Player(e.target, g.getCards(1, e.target)...)
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			g.dropCards(inf.ID)
			g.removePlayercard(e.target, inf.ID)
			g.sendCard2Player(e.target, g.getCards(1, e.target)...)
			return
		case cards := <-g.clients[e.user].GetDropCardInf():
			g.dropCards(cards...)
			g.removePlayercard(e.user, cards...)
			g.sendCard2Player(e.target, g.getCards(1, e.target)...)
			return
		}
	}
}

// 利驭
type liyuEvent struct {
	event
	user data.PID
}

func newLiYuEvent(user data.PID) *liyuEvent {
	return &liyuEvent{user: user}
}

func (e *liyuEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.LiyuState, lastTime, e.user)
	g.clients[e.user].SendUseSkillRsp(data.UseSkillRsp{ID: data.LiYuSkill, Args: []byte{byte(g.players[e.user].hp)}})
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			g.players[e.user].delCard(inf.Cards...)
			g.players[e.user].addTSCard(data.LiYuSkill, inf.Cards...)
			g.players[e.user].findSkill(data.LiYuSkill).(*LiYuSkill).count = uint8(len(inf.Cards))
			arg := []byte{byte(e.user)}
			for _, c := range inf.Cards {
				arg = append(arg, byte(c))
			}
			g.useSkill(e.user, data.LiYuSkill, arg...)
			return
		}
	}
}

// 潜袭
type qianXiEvent struct {
	event
	user, target data.PID
}

func newQianXiEvent(user, target data.PID) *qianXiEvent {
	return &qianXiEvent{user: user, target: target}
}

func (e *qianXiEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.DropSelfAllCards, lastTime, e.user)
	g.clients[e.user].SendDropAbleCard(append([]data.CID{}, g.players[e.user].cards...), 1)
	timer := time.After(lastTime)
	p, t := g.players[e.user], g.players[e.target]
	for {
		select {
		case <-timer:
			card := p.cards[0]
			g.dropCards(card)
			g.removePlayercard(e.user, card)
			if g.cards[card].getDecor().IsRed() {
				t.unUseableCol = append(t.unUseableCol, data.HeartDec, data.DiamondDec)
				g.useSkill(e.user, data.QianXiSkill, byte(data.RedCol), byte(t.pid))
			} else {
				t.unUseableCol = append(t.unUseableCol, data.ClubDec, data.SpadeDec)
				g.useSkill(e.user, data.QianXiSkill, byte(data.BlackCol), byte(t.pid))
			}
			return
		case cards := <-g.clients[e.user].GetDropCardInf():
			if len(cards) > 1 {
				continue
			}
			card := cards[0]
			g.dropCards(card)
			g.removePlayercard(e.user, card)
			if g.cards[card].getDecor().IsRed() {
				t.unUseableCol = append(t.unUseableCol, data.HeartDec, data.DiamondDec)
				g.useSkill(e.user, data.QianXiSkill, byte(data.RedCol), byte(t.pid))
			} else {
				t.unUseableCol = append(t.unUseableCol, data.ClubDec, data.SpadeDec)
				g.useSkill(e.user, data.QianXiSkill, byte(data.BlackCol), byte(t.pid))
			}
			return
		}
	}
}

// 追击
type zhuiJiEvent struct {
	event
	dmgEvent, beforeDmg eventI
	result              *data.CID
}

func newZhuiJiEvent(damageEvent eventI, beforeDmg eventI, result *data.CID) *zhuiJiEvent {
	return &zhuiJiEvent{dmgEvent: damageEvent, result: result, beforeDmg: beforeDmg}
}

func (e *zhuiJiEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor() != data.HeartDec {
		e.dmgEvent.(*damageEvent).damageType = data.DffHPMax
		e.dmgEvent.(*damageEvent).dmg = 1
		e.beforeDmg.setSkip(true)
	}
}

// 反馈
type fankuiEvent struct {
	event
	user, target data.PID
}

func newfanKuiEvent(user, target data.PID) *fankuiEvent {
	return &fankuiEvent{user: user, target: target}
}

func (e *fankuiEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	t := g.players[e.target]
	if len(t.cards)+t.getEquipCount() == 0 {
		return
	}
	cards := []data.CID{}
	if t.death || !g.players[e.user].hasEffect(fankuiEffect) {
		return
	}
	//按照目标装备槽,手牌堆的顺序生成cards
	for _, c := range t.equipSlot {
		if c == nil {
			cards = append(cards, 0)
			continue
		}
		cards = append(cards, c.getID())
	}
	cards = append(cards, t.cards...)
	g.clients[e.user].SendUseSkillRsp(data.UseSkillRsp{ID: data.FanKuiSkill, Targets: []data.PID{e.target}, Cards: cards})
	//设置游戏阶段
	g.setGameState(data.FanKuiState, lastTime, e.user)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			for {
				num := rand.IntN(len(cards))
				if num == 0 {
					continue
				}
				g.moveCard(e.target, e.user, cards[num])
				return
			}
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if len(inf.Cards) != 1 || !isItemInList(cards, inf.Cards[0]) {
				continue
			}
			g.moveCard(e.target, e.user, inf.Cards[0])
			return
		}
	}
}

// 丹术
type danShuEvent struct {
	event
	user   data.PID
	result *data.CID
}

func newDanShuEvent(user data.PID, result *data.CID) *danShuEvent {
	return &danShuEvent{user: user, result: result}
}

func (e *danShuEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor().IsRed() {
		g.recover(e.user, e.user, 1)
	}
}

// 魔炎
type moYanEvent struct {
	event
	user, target data.PID
	result       *data.CID
}

func newMoYanEvent(user, target data.PID, result *data.CID) *moYanEvent {
	return &moYanEvent{user: user, target: target, result: result}
}

func (e *moYanEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor().IsRed() {
		dmg := newDamageEvent(e.user, e.target, data.FireDmg, nil, 2)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	}
}

// 良助
type liangzhuEvent struct {
	event
	user, target data.PID
}

func newLiangZhuEvent(user, target data.PID) *liangzhuEvent {
	return &liangzhuEvent{user: user, target: target}
}

func (e *liangzhuEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.LiangZhuState, lastTime, e.user)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
				return
			} else {
				g.sendCard2Player(e.target, g.getCards(2, e.target)...)
				return
			}
		}
	}
}

// 成略
type chengLueEvent struct {
	event
	user   data.PID
	enable bool //阴阳
}

func newChengLueEvent(user data.PID, enable bool) *chengLueEvent {
	return &chengLueEvent{user: user, enable: enable}
}

func (e *chengLueEvent) trigger(g *Games) {
	p := g.players[e.user]
	const lastTime = 20 * time.Second
	getNum, dropNum := 1, 2
	if e.enable {
		getNum, dropNum = dropNum, getNum
	}
	g.setGameState(data.DropSelfHandCard, lastTime, e.user)
	g.sendCard2Player(e.user, g.getCardsFromBottom(getNum)...)
	if len(p.cards) < dropNum {
		dropNum = 1
	}
	cards := p.cards
	g.clients[e.user].SendDropAbleCard(cards, uint8(dropNum))
	decList := []data.Decor{}
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			if len(cards) == 1 {
				g.dropCards(cards[0])
				decList = append(decList, g.cards[cards[0]].getDecor())
				g.removePlayercard(e.user, cards[0])
				goto end
			}
			g.dropCards(cards[dropNum])
			g.removePlayercard(e.user, cards[dropNum])
			for i := 0; i < dropNum; i++ {
				decList = append(decList, g.cards[cards[i]].getDecor())
			}
			goto end
		case inf := <-g.clients[e.user].GetDropCardInf():
			g.dropCards(inf...)
			g.removePlayercard(e.user, inf...)
			for i := 0; i < dropNum; i++ {
				decList = append(decList, g.cards[inf[i]].getDecor())
			}
			goto end
		}
	}
end:
	p.findSkill(data.ChengLueSkill).(*chengLueSkill).decList = decList
	arg := make([]byte, len(decList))
	for i, dec := range decList {
		arg[i] = byte(dec)
	}
	g.useSkill(e.user, data.ChengLueSkill, arg...)
}

type fengPoEvent struct {
	event
	user, target data.PID
	skill        *fengPoSkill
}

func newFengPoEvent(user, target data.PID, skill *fengPoSkill) *fengPoEvent {
	return &fengPoEvent{user: user, target: target, skill: skill}
}

func (e *fengPoEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.FengPoState, lastTime, e.user)
	timer := time.After(lastTime)
	num := uint8(0)
	for _, c := range g.players[e.target].cards {
		if g.cards[c].getDecor() == data.DiamondDec {
			num++
		}
	}
	for {
		select {
		case <-timer:
			e.skill.conunt = num
			g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				e.skill.conunt = num
				g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
				return
			} else {
				e.skill.conunt = 1
				g.sendCard2Player(e.user, g.getCardsFromTop(int(num))...)
				return
			}
		}
	}
}

// 给卡
type skillSendCardEvent struct {
	event
	target data.PID
	num    uint8
}

func newSkillSendCardEvent(target data.PID, num uint8) *skillSendCardEvent {
	return &skillSendCardEvent{target: target, num: num}
}

func (s *skillSendCardEvent) trigger(g *Games) {
	g.sendCard2Player(s.target, g.getCards(int(s.num), s.target)...)
}

type quanJiEvent struct {
	event
	user data.PID
}

func newQuanJiEvent(user data.PID) *quanJiEvent {
	return &quanJiEvent{user: user}
}

func (e *quanJiEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.DropSelfAllCards, lastTime, e.user)
	cards := []data.CID{}
	for _, equip := range g.players[e.user].equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		cards = append(cards, equip.getID())
	}
	g.clients[e.user].SendDropAbleCard(append(cards, g.players[e.user].cards...), 1)
	timer := time.After(lastTime)
	p := g.players[e.user]
	s := p.findSkill(data.QuanJiSkill).(*quanJiSkill)
	for {
		select {
		case <-timer:
			c := p.cards[0]
			p.addTSCard(data.QuanJiSkill, c)
			p.delCard(p.cards[0])
			s.count++
			g.useSkill(e.user, data.QuanJiSkill, byte(s.count), byte(1))
			return
		case cards := <-g.clients[e.user].GetDropCardInf():
			if len(cards) > 1 {
				continue
			}
			p.addTSCard(data.QuanJiSkill, cards...)
			p.delCard(cards...)
			s.count++
			g.useSkill(e.user, data.QuanJiSkill, byte(s.count), byte(1), byte(cards[0]))
			return
		}
	}
}

type lonyinEvent struct {
	event
	user, target data.PID
	card         data.CID
}

func newLongYinEvent(user, target data.PID, card data.CID) *lonyinEvent {
	return &lonyinEvent{user: user, target: target, card: card}
}

func (e *lonyinEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.DropSelfAllCards, lastTime, e.user)
	cards := []data.CID{}
	for _, equip := range g.players[e.user].equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		cards = append(cards, equip.getID())
	}
	g.clients[e.user].SendDropAbleCard(append(cards, g.players[e.user].cards...), 1)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case card := <-g.clients[e.user].GetDropCardInf():
			g.dropCards(card...)
			g.removePlayercard(e.user, card...)
			g.players[e.target].hasAttack = false
			if g.cards[e.card].getDecor().IsRed() {
				g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
			}
			return
		}
	}
}

type queDiEvent struct {
	event
	user, target data.PID
}

func newQueDiEvent(user, target data.PID) *queDiEvent {
	return &queDiEvent{user: user, target: target}
}

func (e *queDiEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.QueDiState, lastTime, e.user)
	//arg [2]byte,其中arg[0]表示第一个选项(获得其手牌)是否可选,arg[1]表示第二个选项(弃牌)是否可选,2为背水
	arg := [3]byte{0, 0, 0}
	if len(g.players[e.target].cards) != 0 {
		arg[0] = 1
	}
	for _, c := range g.players[e.user].cards {
		if g.cards[c].getType() == data.BaseCardType {
			arg[1] = 1
			break
		}
	}
	if arg[0] == 1 && arg[1] == 1 {
		arg[2] = 1
	}
	g.clients[e.user].SendUseSkillRsp(data.UseSkillRsp{ID: data.QueDiSkill, Args: arg[:]})
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if len(inf.Args) != 2 || inf.ID != data.QueDiSkill {
				continue
			}
			num := 0
			//inf的arg[0],arg[1]=代表两个选项发动
			if inf.Args[0] == 1 {
				if arg[0] == 0 {
					continue
				}
				cards := g.players[e.target].cards
				card := cards[rand.IntN(len(cards))]
				g.moveCard(e.target, e.user, card)
				num++
			}
			if inf.Args[1] == 1 {
				g.events.insert(g.index, newQueDiDropEvent(e.user))
				g.players[e.user].findSkill(data.QueDiSkill).(*quediSkill).addDmg = true
				num++
			}
			if num == 2 {
				dmg := newDamageEvent(e.user, e.user, data.DffHPMax, nil, 1)
				g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
			}
			return
		}
	}
}

type jiQiaoEvent struct {
	event
	user, target data.PID
}

func newJiQiaoEvent(user, target data.PID) *jiQiaoEvent {
	return &jiQiaoEvent{user: user, target: target}
}

func (e *jiQiaoEvent) trigger(g *Games) {
	g.players[e.target].enableEffect(changJiEffect, data.JiQiaoSkill)
	g.useSkill(e.target, data.JiQiaoSkill, byte(1))
	tmpcard := newCard(data.Card{CardType: data.TipsCardType, Name: data.WZSY})
	tmpcard.use(g, e.user)
	g.useTmpCard(e.user, data.WZSY, data.NoCol, 0, data.VirtualCard)
	g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
	g.useSkill(e.user, data.JiZhiSkill)
}

type keJiEvent struct {
	event
	user   data.PID
	result *data.CID
}

func newKeJiEvent(user data.PID, result *data.CID) *keJiEvent {
	return &keJiEvent{user: user, result: result}
}

func (e *keJiEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor().IsRed() {
		g.recover(e.user, e.user, 1)
	} else {
		g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
	}
}

type luanWuEvent struct {
	event
	user    data.PID
	targets []data.PID
}

func newLuanWuEvent(user data.PID, targets []data.PID) *luanWuEvent {
	return &luanWuEvent{user: user, targets: targets}
}

func (e *luanWuEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.LuanWuState, lastTime, e.user)
	useAbel := g.players[e.user].getUseAbleCards(g)
	g.clients[e.user].SendUseAbleCards(useAbel)
	useAbleSkill := g.players[e.user].getUseAbleSkill(g)
	g.clients[e.user].SendUseAbleSkill(useAbleSkill)
	timer := time.After(lastTime)
	var useCard cardI      //使用的卡
	var targets []data.PID //目标
	for {
		select {
		case <-timer:
			dmg := newDamageEvent(g.turnOwner, e.user, data.NormalDmg, nil, 1)
			g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
			return
		case <-g.clients[e.user].GetTargetQuest():
			g.clients[e.user].SendAvailableTarget(data.AvailableTargetInf{TargetNum: 1, TargetList: e.targets})
		case inf := <-g.clients[e.user].GetUseSkillInf():
			//回应使用技能请求
			if g.players[e.user].findSkill(inf.ID).use(g, g.players[e.user], useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args},
				&useCard, &targets) {
				goto finish
			}
			continue
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				dmg := newDamageEvent(g.turnOwner, e.user, data.NormalDmg, nil, 1)
				g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
				return
			}
			if !g.cards[inf.ID].useAble(g, inf.TargetList[0]) {
				continue
			}
			useCard = g.cards[inf.ID]
			targets = inf.TargetList
			g.dropCards(inf.ID)
			g.cards[inf.ID].use(g, e.user, inf.TargetList[0])
			goto finish
		}
	}
finish:
	p := g.players[e.user]
	//恃才
	if p.hasEffect(shicaiEffect) {
		p.findSkill(data.ShiCaiSkill).(*shiCaiSkill).check(g, useCard, e.user, targets)
	}
	//蒺藜技能
	if p.hasEffect(jiLiEffect) {
		p.findSkill(data.JiLiSkill).(*jiLiSkill).check(g, p.getAtkDst(), e.user)
	}
	//却敌
	if p.hasEffect(quediEffect) {
		p.findSkill(data.QueDiSkill).(*quediSkill).check(g, e.user, useCard, targets)
	}
	//无双
	if p.hasEffect(wuShuangEffect) {
		g.useSkill(e.user, data.WuShuangSkill)
	}
	//图射
	if p.hasEffect(tuSheEffect) {
		p.findSkill(data.TuSheSkill).(*tuSheSkill).check(g, e.user, useCard, targets)
	}
}

type queDiDropEvent struct {
	event
	user data.PID
}

func newQueDiDropEvent(user data.PID) *queDiDropEvent {
	return &queDiDropEvent{user: user}
}

func (e *queDiDropEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.DropSelfHandCard, lastTime, e.user)
	cards := []data.CID{}
	for _, c := range g.players[e.user].cards {
		if g.cards[c].getType() == data.BaseCardType {
			cards = append(cards, c)
		}
	}
	g.clients[e.user].SendDropAbleCard(cards, 1)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			g.dropCards(cards[0])
			g.removePlayercard(e.user, cards[0])
			return
		case cards := <-g.clients[e.user].GetDropCardInf():
			g.dropCards(cards...)
			g.removePlayercard(e.user, cards...)
			return
		}
	}
}

// 选择用不用樵拾技能的事件
type qiaoShiSelectEvent struct {
	event
	user, target data.PID
	num          int
}

func newQiaoShiSelectEvent(user data.PID, target data.PID, num int) *qiaoShiSelectEvent {
	return &qiaoShiSelectEvent{target: target, user: user}
}

func (e *qiaoShiSelectEvent) trigger(g *Games) {
	if g.players[e.user].death {
		return
	}
	const lastTime = 20 * time.Second
	iterators(g.clients, func(c clientI) { c.SendSkillSelect(data.QiaoShiSkill) })
	p := g.players[e.user]
	t := g.players[e.target]
	g.clients[e.user].SendAvailableTarget(data.AvailableTargetInf{})
	g.setGameState(data.SkillSelectState, lastTime, e.user)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if inf.Skip {
				return
			}
			if inf.ID != data.QiaoShiSkill {
				continue
			}
			t.findSkill(data.QiaoShiSkill).use(g, p, useSkillInf{targets: []data.PID{e.target}}, data.HP(e.num))
			return
		}
	}
}

type getSkillEvent struct {
	event
	user data.PID
}

func newGetSkillEvent(user data.PID) *getSkillEvent {
	return &getSkillEvent{user: user}
}

func (e *getSkillEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	num1 := rand.IntN(len(data.AtkSkillList))
check:
	num2 := rand.IntN(len(data.AtkSkillList))
	if num1 == num2 {
		goto check
	}
	args := []byte{byte(data.AtkSkillList[num1]), byte(data.AtkSkillList[num2])}
	g.clients[e.user].SendUseSkillRsp(data.UseSkillRsp{ID: data.GetSkill, Args: args})
	g.setGameState(data.GetSkillState, lastTime, e.user)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if inf.ID != data.GetSkill {
				continue
			}
			if inf.Args[0] == 0 {
				g.players[e.user].addSkill(data.AtkSkillList[num1])
			} else {
				g.players[e.user].addSkill(data.AtkSkillList[num2])
			}
			return
		}
	}
}
