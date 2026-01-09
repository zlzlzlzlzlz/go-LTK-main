package server

import "goltk/data"

type horseCard struct {
	card
	isUpHorse bool
}

func (c *horseCard) use(g *Games, user data.PID, t ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID)
		g.players[user].delCard(c.ID)
	}
	var slot *cardI
	if c.isUpHorse {
		slot = &g.players[user].equipSlot[data.HorseUpSlot]
	} else {
		slot = &g.players[user].equipSlot[data.HorseDownSlot]
	}
	if *slot != nil {
		if id := (*slot).getID(); id != 0 {
			g.dropCards(id)
		}
		//枭姬
		if g.players[user].hasEffect(xiaojiEffect) {
			g.sendCard2Player(user, g.getCardsFromTop(2)...)
			g.useSkill(user, data.XiaoJiSkill)
		}
	}
	*slot = g.cards[c.ID]
	//武库
	for _, p := range g.players {
		if p.hasEffect(wukuEffect) {
			p.findSkill(data.WuKuSkill).(*wuKuSkill).check(g, p.pid)
		}
	}
}

//防具卡
type armorCard struct {
	card
}

func (c *armorCard) use(g *Games, user data.PID, t ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID)
		g.players[user].delCard(c.ID)
	}
	slot := &g.players[user].equipSlot[data.ArmorSlot]
	if *slot != nil {
		if g.players[user].hasEffect(byszEffect) && (*slot).getName() == data.BYSZ {
			g.recover(user, user, 1)
		}
		g.players[user].disableEffect(equipEffectMap[(*slot).getName()])
		if skill, ok := equipSkillMap[(*slot).getName()]; ok {
			g.players[user].removeSkill(skill)
		}
		if id := (*slot).getID(); id != 0 {
			g.dropCards(id)
		}
		//枭姬
		if g.players[user].hasEffect(xiaojiEffect) {
			g.sendCard2Player(user, g.getCardsFromTop(2)...)
			g.useSkill(user, data.XiaoJiSkill)
		}
	}
	*slot = g.cards[c.ID]
	g.players[user].enableEffect(equipEffectMap[c.Name])
	if skill, ok := equipSkillMap[c.Name]; ok {
		g.players[user].addSkill(skill)
	}
	//武库
	for _, p := range g.players {
		if p.hasEffect(wukuEffect) {
			p.findSkill(data.WuKuSkill).(*wuKuSkill).check(g, p.pid)
		}
	}
}

//武器卡
type weaponCard struct {
	card
	dst distence
}

func newWeaponCard(c data.Card) *weaponCard {
	if dst, ok := data.WeaponDstMap[c.Name]; ok {
		return &weaponCard{card: card{Card: c}, dst: distence(dst)}
	}
	panic("没有找到" + c.Name.String() + "的武器距离")
}

func (c *weaponCard) use(g *Games, user data.PID, t ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID)
		g.players[user].delCard(c.ID)
	}
	slot := &g.players[user].equipSlot[data.WeaponSlot]
	if *slot != nil {
		g.players[user].disableEffect(equipEffectMap[(*slot).getName()])
		if skill, ok := equipSkillMap[(*slot).getName()]; ok {
			g.players[user].removeSkill(skill)
		}
		if id := (*slot).getID(); id != 0 {
			g.dropCards(id)
		}
		//枭姬
		if g.players[user].hasEffect(xiaojiEffect) {
			g.sendCard2Player(user, g.getCardsFromTop(2)...)
			g.useSkill(user, data.XiaoJiSkill)
		}
	}
	*slot = g.cards[c.ID]
	if eff, ok := equipEffectMap[c.Name]; ok {
		g.players[user].enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.Name]; ok {
		g.players[user].addSkill(skill)
	}
	//武库
	for _, p := range g.players {
		if p.hasEffect(wukuEffect) {
			p.findSkill(data.WuKuSkill).(*wuKuSkill).check(g, p.pid)
		}
	}
}

var equipSkillMap = map[data.CardName]data.SID{
	data.ZBSM: data.ZBSMSkill, data.ZQYS: data.ZQYSSkill, data.QLYYD: data.QLYYDSkill, data.QLG: data.QLGSkill,
	data.CXSGJ: data.CXSGJSkill, data.HBJ: data.HBJSkill, data.GSF: data.GSFSkill, data.BGZ: data.BGZSkill,
}
