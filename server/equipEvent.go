package server

import (
	"goltk/data"
	"time"
)

// 贯石斧
type GSFEvent struct {
	event
	events []eventI //要取消跳过的伤害事件
	user   data.PID
}

func newGSFEvent(events []eventI, user data.PID) *GSFEvent {
	return &GSFEvent{events: events, user: user}
}

func (e *GSFEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	if g.players[e.user].death {
		return
	}
	g.setGameState(data.GSFState, lastTime, e.user)
	cards := []data.CID{}
	for i, equip := range g.players[e.user].equipSlot {
		if i == int(data.WeaponSlot) || equip == nil || equip.getID() == 0 {
			continue
		}
		cards = append(cards, equip.getID())
	}
	g.clients[e.user].SendDropAbleCard(append(cards, g.players[e.user].cards...), 2)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case cards := <-g.clients[e.user].GetDropCardInf():
			if len(cards) == 0 {
				return
			}
			for _, event := range e.events {
				event.setSkip(false)
			}
			g.dropCards(cards...)
			g.removePlayercard(e.user, cards...)
			return
		}
	}
}

// 青龙偃月刀
type ctuAtkEvent struct {
	event
	user, target data.PID
}

func newCtuAtkEvent(user data.PID, target data.PID) *ctuAtkEvent {
	return &ctuAtkEvent{user: user, target: target}
}

func (e *ctuAtkEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	p := g.players[e.user]
	if p.death {
		return
	}
	useAbleCards := p.getUseAbleCards(g)
	g.clients[e.user].SendUseAbleCards(useAbleCards)
	g.setGameState(data.QlYYDState, lastTime, e.user)
	useAbleSkill := p.getUseAbleSkill(g)
	g.clients[e.user].SendUseAbleSkill(useAbleSkill)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			if p.findSkill(inf.ID).use(g, p, useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args}, e.target) {
				goto useSuccess
			}
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				return
			}
			if !isItemInList(useAbleCards, inf.ID) { //检查玩家发送的卡牌的正确性
				continue
			}
			g.dropCards(inf.ID)
			g.cards[inf.ID].use(g, e.user, e.target)
			goto useSuccess
		}
	}
useSuccess:
	if p.hasEffect(yanyueEffect) {
		p.findSkill(data.YanYueSkill).(*yanyueSkill).enable = true
	}
}

// 麒麟弓
type qlgEvent struct {
	event
	user, target data.PID
}

func newQLGEvent(user, target data.PID) *qlgEvent {
	return &qlgEvent{user: user, target: target}
}

func (e *qlgEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	horses := []data.CID{}
	t := g.players[e.target]
	//目标双马槽
	for _, c := range t.equipSlot[2:] {
		if c == nil {
			horses = append(horses, 0)
			continue
		}
		horses = append(horses, c.getID())
	}
	//发送牌堆
	g.clients[e.user].SendGSCards(data.GSIDQLG, horses...)
	//设置游戏阶段
	g.setGameState(data.QLGState, lastTime, e.user)
	timer := time.After(lastTime)
	execQiLin := func() {
		if !g.players[e.user].hasEffect(qilinEffect) {
			return
		}
		cards := append([]data.CID{}, t.cards...)
		g.useSkill(e.user, data.QiLinSkill)
		g.dropCards(cards...)
		g.removePlayercard(e.target, cards...)
	}
	for {
		select {
		case <-timer:
			for _, c := range horses {
				if c != 0 {
					g.removePlayercard(e.target, c)
				}
			}
			execQiLin()
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.ID == 0 || !isItemInList(horses, inf.ID) {
				continue
			}
			g.removePlayercard(e.target, inf.ID)
			execQiLin()
			return
		}
	}
}

// 雌雄双股剑
type csxgjEvent struct {
	event
	user, target data.PID
}

func newCXSGJEvent(user, target data.PID) *csxgjEvent {
	return &csxgjEvent{user: user, target: target}
}

func (e *csxgjEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.CXSGJState, lastTime, e.target)
	//arg [2]byte,其中arg[0]表示第一个选项(目标弃一张牌)是否可选,arg[1]表示第二个选项(使用者摸一张)是否可选
	arg := [2]byte{0, 1}
	if len(g.players[e.target].cards) != 0 {
		arg[0] = 1
	}
	g.clients[e.target].SendUseSkillRsp(data.UseSkillRsp{ID: data.CXSGJSkill, Args: arg[:]})
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			g.sendCard2Player(e.user, g.getCards(1, e.user)...)
			return
		case inf := <-g.clients[e.target].GetUseSkillInf():
			if len(inf.Args) != 1 || inf.ID != data.CXSGJSkill {
				continue
			}
			if inf.Args[0] == 0 {
				if arg[0] == 0 {
					continue
				}
				g.events.insert(g.index, newDropSelfHandCardEvent(e.target, 1))
				return
			} else {
				g.sendCard2Player(e.user, g.getCards(1, e.user)...)
				return
			}
		}
	}
}

// 八卦阵
type bgzEvent struct {
	event
	user   data.PID
	result *data.CID
}

func newBGZEvent(user data.PID, result *data.CID) *bgzEvent {
	return &bgzEvent{user: user, result: result}
}

func (e *bgzEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor().IsRed() {
		g.players[e.user].enableEffect(virtualDodgeEffect)
	}
}
