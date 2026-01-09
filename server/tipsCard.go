package server

import (
	"goltk/data"
)

// 无懈可击
type wxkjcard struct {
	card
}

func (c *wxkjcard) useAble(g *Games, user data.PID) bool {
	if isItemInList(g.players[user].unUseableCol, c.Dec) {
		return false
	}
	if _, ok := g.events.list[g.index].(*wxkjEvent); ok {
		return true
	}
	return false
}

// 无中生有
type wzsyCard struct {
	card
}

func (c *wzsyCard) use(g *Games, user data.PID, target ...data.PID) {
	//莺语
	if g.players[user].hasEffect(yingwuEffect) {
		g.players[user].findSkill(data.YingWuSkill).(*yingWuSkill).addCount(g, user)
		g.players[user].findSkill(data.LueYingSkill).(*lueYingSkill).check(g, user, data.WZSY)
	}
	if c.ID != 0 {
		g.useCard(user, c.ID)
		g.players[user].delCard(c.ID)
	}
	e := newWzsyEvent(user)
	g.events.insert(g.index, newWXKJEvent(c, user, user, e), e)
}

// 决斗
type duelCard struct {
	card
}

func (c *duelCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, target...)
		g.players[user].delCard(c.ID)
	}
	e := newDuelEvent(user, target[0], c)
	g.events.insert(g.index, newWXKJEvent(c, user, target[0], e), e)
}

func (c *duelCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	plist := g.getAllAliveOther(user)
	for i, p := range plist {
		//空城
		if g.players[p].hasEffect(kongChengEffect) && len(g.players[p].cards) == 0 {
			plist = append(plist[:i], plist[i+1:]...)
		}
	}
	return plist, 1
}

// 南蛮入侵
type nmrqCard struct {
	card
}

func (c *nmrqCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, g.getAllAliveOther(user)...)
		g.players[user].delCard(c.ID)
	}
	list := []eventI{}
	for pid := g.getNextPid(user); pid != user; pid = g.getNextPid(pid) {
		//帷幕
		if g.players[pid].hasEffect(weiMuEffect) && c.getDecor().ISBlack() {
			g.useSkill(pid, data.WeiMuSkill)
			continue
		}
		if g.players[pid].hasEffect(tengjiaEffect) {
			g.useSkill(pid, data.TengJiaSkill)
			continue
		}
		e := newNMRQEvent(user, pid, c)
		list = append(list, newWXKJEvent(c, user, pid, e), e)
	}
	g.events.insert(g.index, list...)
}

// 万箭齐发
type wjqfCard struct {
	card
}

func (c *wjqfCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, g.getAllAliveOther(user)...)
		g.players[user].delCard(c.ID)
	}
	list := []eventI{}
	for pid := g.getNextPid(user); pid != user; pid = g.getNextPid(pid) {
		//帷幕
		if g.players[pid].hasEffect(weiMuEffect) && c.getDecor().ISBlack() {
			g.useSkill(pid, data.WeiMuSkill)
			continue
		}
		if g.players[pid].hasEffect(tengjiaEffect) {
			g.useSkill(pid, data.TengJiaSkill)
			continue
		}
		e := newWJQFEvent(user, pid, c)
		wxkj := newWXKJEvent(c, user, pid, e)
		if g.players[pid].hasEffect(bgzEffect) {
			list = append(list, wxkj, newSkillSelectEvent(data.BGZSkill, pid, nil), e)
			continue
		}
		list = append(list, wxkj, e)
	}
	g.events.insert(g.index, list...)
}

// 桃园结义
type tyjyCard struct {
	card
}

func (c *tyjyCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, g.getAllAlivePlayer()...)
		g.players[user].delCard(c.ID)
	}
	list := []eventI{}
	if g.players[user].hp < g.players[user].maxHp {
		e := newTYJYEvent(user, user)
		list = append(list, newWXKJEvent(c, user, user, e), e)
	}
	for pid := g.getNextPid(user); pid != user; pid = g.getNextPid(pid) {
		if g.players[pid].hp < g.players[pid].maxHp {
			e := newTYJYEvent(user, pid)
			list = append(list, newWXKJEvent(c, user, pid, e), e)
		}
	}
	g.events.insert(g.index, list...)
}

// 五谷丰登
type wgfdCard struct {
	card
}

func (c *wgfdCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, g.getAllAlivePlayer()...)
		g.players[user].delCard(c.ID)
	}
	cards := g.getCardsFromTop(g.getAlivePlayerCount())
	pid := user
	events := []eventI{}
	for {
		wgfd := newWGFDEvent(user, pid, &cards)
		events = append(events, newWXKJEvent(c, user, pid, wgfd), wgfd)
		if pid = g.getNextPid(pid); pid == user {
			break
		}
	}
	g.events.insert(g.index, events...)
	iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDWGFD, cards...) })
}

// 火攻
type burnCard struct {
	card
}

func (c *burnCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, target...)
		g.players[user].delCard(c.ID)
	}
	burn := newBurnShowEvent(user, target[0], c)
	g.events.insert(g.index, newWXKJEvent(c, user, target[0], burn), burn)
}

func (c *burnCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	list := []data.PID{}
	for i := 0; i < len(g.players); i++ {
		if g.players[i].death || len(g.players[i].cards) == 0 {
			continue
		}
		list = append(list, g.players[i].pid)
	}
	return list, 1
}

// 铁索连环
type tslhCard struct {
	card
}

func (c *tslhCard) use(g *Games, user data.PID, target ...data.PID) {
	if len(target) == 0 {
		if c.ID != 0 {
			g.useCard(user, c.ID)
			g.players[user].delCard(c.ID)
		}
		g.sendCard2Player(user, g.getCards(1, user)...)
		return
	}
	if len(target) == 1 {
		//莺语
		if g.players[user].hasEffect(yingwuEffect) {
			g.players[user].findSkill(data.YingWuSkill).(*yingWuSkill).addCount(g, user)
			g.players[user].findSkill(data.LueYingSkill).(*lueYingSkill).check(g, user, data.TSLH)
		}
	}
	if c.ID != 0 {
		g.useCard(user, c.ID, target...)
		g.players[user].delCard(c.ID)
	}
	for i := len(target) - 1; i >= 0; i-- {
		e := newTSLHEvent(user, target[i])
		g.events.insert(g.index, newWXKJEvent(c, user, target[i], e), e)
	}
}

func (c *tslhCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	list := g.getAllAlivePlayer()
	for i, p := range list {
		if g.players[p].hasEffect(weiMuEffect) {
			list = append(list[i:], list[:i+1]...)
		}
	}
	return list, 2
}

// 顺手牵羊
type ssqyCard struct {
	card
}

func (c *ssqyCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, target[0])
		g.players[user].delCard(c.ID)
	}
	//莺语
	if g.players[user].hasEffect(yingwuEffect) {
		g.players[user].findSkill(data.YingWuSkill).(*yingWuSkill).addCount(g, user)
		g.players[user].findSkill(data.LueYingSkill).(*lueYingSkill).check(g, user, data.SSQY)
	}
	e := newSSQYEvent(user, target[0])
	g.events.insert(g.index, newWXKJEvent(c, user, target[0], e), e)
}

func (c *ssqyCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	var pList []data.PID
	if g.players[user].hasEffect(liMuEffect) {
		if g.players[user].findSkill(data.LiMuSkill).(*liMuSkill).check(g, user) {
			pList = g.getPlayerInDst(user, g.players[user].getAtkDst())
		} else {
			pList = g.getPlayerInDst(user, 1)
		}
	} else if g.players[user].hasEffect(chenglueEffect) {
		if isItemInList(g.players[user].findSkill(data.ChengLueSkill).(*chengLueSkill).decList, c.getDecor()) {
			pList = g.getAllAliveOther(user)
		} else {
			pList = g.getPlayerInDst(user, 1)
		}
	} else if g.players[user].hasEffect(qicaiEffect) {
		pList = g.getAllAliveOther(user)
	} else {
		pList = g.getPlayerInDst(user, 1)
	}
	for i := len(pList) - 1; i >= 0; i-- {
		p := g.players[pList[i]]
		if len(p.cards) != 0 {
			continue
		}
		if p.hasEffect(weiMuEffect) {
			continue
		}
		hasEquip := false
		for j := 0; j < len(p.equipSlot); j++ {
			if p.equipSlot[j] != nil && p.equipSlot[j].getID() != 0 {
				hasEquip = true
				break
			}
		}
		if hasEquip {
			continue
		}
		hasJudge := false
		for j := 0; j < len(p.judgeSlot); j++ {
			if p.judgeSlot[j] != nil && p.judgeSlot[j].getID() != 0 {
				hasJudge = true
				break
			}
		}
		if hasJudge {
			continue
		}
		pList = append(pList[:i], pList[i+1:]...)
	}
	return pList, 1
}

// 过河拆桥
type ghcqCard struct {
	card
}

func (c *ghcqCard) use(g *Games, user data.PID, target ...data.PID) {
	//莺语
	if g.players[user].hasEffect(yingwuEffect) {
		g.players[user].findSkill(data.YingWuSkill).(*yingWuSkill).addCount(g, user)
		g.players[user].findSkill(data.LueYingSkill).(*lueYingSkill).check(g, user, data.GHCQ)
	}
	if c.ID != 0 {
		g.useCard(user, c.ID, target[0])
		g.players[user].delCard(c.ID)
	}
	e := newGHCQEvent(user, target[0])
	g.events.insert(g.index, newWXKJEvent(c, user, target[0], e), e)
}

func (c *ghcqCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	pList := g.getAllAliveOther(user)
	for i := len(pList) - 1; i >= 0; i-- {
		p := g.players[pList[i]]
		if len(p.cards) != 0 {
			continue
		}
		hasEquip := false
		for j := 0; j < len(p.equipSlot); j++ {
			if p.equipSlot[j] != nil && p.equipSlot[j].getID() != 0 {
				hasEquip = true
				break
			}
		}
		if hasEquip {
			continue
		}
		hasJudge := false
		for j := 0; j < len(p.judgeSlot); j++ {
			if p.judgeSlot[j] != nil && p.judgeSlot[j].getID() != 0 {
				hasJudge = true
				break
			}
		}
		if hasJudge {
			continue
		}
		pList = append(pList[:i], pList[i+1:]...)
	}
	return pList, 1
}

type jdsrCard struct {
	card
}

func (c *jdsrCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, target[0], target[1])
		g.players[user].delCard(c.ID)
	}
	e := newJDSREvent(user, target[0], target[1])
	g.events.insert(g.index, newWXKJEvent(c, user, target[0], e), e)
}

func (c *jdsrCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	list := []data.PID{}
	for killer := g.getNextPid(user); killer != user; killer = g.getNextPid(killer) {
		wep := g.players[killer].equipSlot[data.WeaponSlot]
		if wep == nil || wep.getID() == 0 {
			continue
		}
		dst := wep.(*weaponCard).dst
		list = append(list, killer)
		list = append(list, g.getPlayerInDst(killer, dst)...)
		list = append(list, -1) //加分隔符
	}
	return list, 1
}
