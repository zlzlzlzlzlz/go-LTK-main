package server

import "goltk/data"

type lbssCard struct {
	card
	user data.PID
}

func (c *lbssCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, target[0])
		g.players[user].delCard(c.ID)
	}
	g.players[target[0]].judgeSlot[data.LBSSSlot] = c
	c.user = user
}

func (c *lbssCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	// list := g.getAllAliveOther(user)
	list := g.getAllAlivePlayer()
	for i := 0; i < len(list); i++ {
		t := g.players[list[i]]
		if t.judgeSlot[data.LBSSSlot] != nil || t.hasEffect(guiMeiEffect) {
			list = append(list[:i], list[i+1:]...)
		}
		//帷幕
		if t.hasEffect(weiMuEffect) && c.getDecor().ISBlack() {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list, 1
}

type blcdCard struct {
	card
	user data.PID
}

func (c *blcdCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, target[0])
		g.players[user].delCard(c.ID)
	}
	g.players[target[0]].judgeSlot[data.BLCDSlot] = c
	c.user = user
}

func (c *blcdCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
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
	for i := 0; i < len(pList); i++ {
		t := g.players[pList[i]]
		if t.judgeSlot[data.BLCDSlot] != nil || t.hasEffect(guiMeiEffect) {
			continue
		}
		//帷幕
		if t.hasEffect(weiMuEffect) && c.getDecor().ISBlack() {
			continue
		}
		pList = append(pList, pList[i])
	}
	return pList, 1
}

type lightnCard struct {
	card
	user data.PID
}

func (c *lightnCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID, user)
		g.players[user].delCard(c.ID)
	}
	g.players[user].judgeSlot[data.LightningSlot] = c
	c.user = user
}

func (c *lightnCard) useAble(g *Games, pid data.PID) bool {
	p := g.players[pid]
	if p.judgeSlot[data.LightningSlot] != nil || p.hasEffect(guiMeiEffect) {
		return false
	}
	if p.hasEffect(weiMuEffect) {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return true
	}
	return false
}
