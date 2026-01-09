package server

import (
	"goltk/data"
)

// 杀
type atkCard struct {
	card
	dmgType data.SetHpType
}

func (c *atkCard) use(g *Games, user data.PID, target ...data.PID) {
	p, t := g.players[user], g.players[target[0]]
	if p.hasAttack {
		if p.equipSlot[data.WeaponSlot] != nil && p.equipSlot[data.WeaponSlot].getName() == data.ZGLN {
			g.useSkill(user, data.ZGLNSkill)
		} else if p.hasEffect(unLimit) {
			g.useSkill(user, p.getEffectSkill(unLimit))
		}
	}
	if c.ID != 0 {
		g.useCard(user, c.ID, target...)
		p.delCard(c.ID)
	}
	p.hasAttack = true
	atkEventList := []eventI{}
	for i := len(target) - 1; i >= 0; i-- {
		event := newAtkEvent(user, target[i], c.dmgType, c)
		atkEventList = append(atkEventList, event)
		g.events.insert(g.index, event)
		//检查破军
		if p.hasEffect(poJunEffect) && len(t.cards)+t.getEquipCount() != 0 {
			g.events.insert(g.index, newSkillSelectEvent(data.PoJunSkill, user, []data.PID{target[i]}))
		}
		//检测是否有雌雄双股剑
		if p.hasEffect(cxsgjEffect) && (p.isfemale != g.players[target[i]].isfemale) {
			g.events.insert(g.index, newSkillSelectEvent(data.CXSGJSkill, user, []data.PID{target[i]}))
		}
		//烈弓
		if p.hasEffect(lieGongEffect) {
			p.findSkill(data.LieGongSkill).(*lieGongSkill).check(g, user, target[i], event)
		}
		//铁骑
		if p.hasEffect(tieJiEffect) {
			g.events.insert(g.index, newSkillSelectEvent(data.TieJiSkill, user, nil, event))
		}
	}
	//略影
	if p.hasEffect(lueyingEffect) {
		p.findSkill(data.LueYingSkill).(*lueYingSkill).addCount(g, user)
	}
	//检查龙吟
	for _, p := range g.getAllAlivePlayer() {
		if g.players[p].hasEffect(longyinEffect) &&
			(len(g.players[p].cards)+g.players[p].getEquipCount()) > 0 {
			g.events.insert(g.index, newSkillSelectEvent(data.LongYinSkill, p, nil, user, c.getID()))
		}
	}
	//检查凤魄
	if p.hasEffect(fengpoEffect) && len(target) == 1 {
		if p.findSkill(data.FengPoSkill).(*fengPoSkill).check() {
			g.events.insert(g.index, newSkillSelectEvent(data.FengPoSkill, user, target))
		}
	}
	//朱雀羽扇
	if p.hasEffect(zqysEffect) && c.dmgType == data.NormalDmg {
		g.events.insert(g.index, newSkillSelectEvent(data.ZQYSSkill, user, nil, atkEventList))
	}
	//克己
	if t.hasEffect(kejiEffect) && g.turnOwner == user {
		t.findSkill(data.KeJiSkill).(*keJiSkill).useAtk = true
	}
}

func (c *atkCard) useAble(g *Games, user data.PID) bool {
	p := g.players[user]
	if isItemInList(p.unUseableCol, c.Dec) {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok { //对于出牌事件
		//如果没打过杀或者有无限杀效果则允许用杀
		if !p.hasAttack || p.hasEffect(unLimit) {
			return true
		}
		//成略无视次数
		if p.hasEffect(chenglueEffect) {
			return p.findSkill(data.ChengLueSkill).(*chengLueSkill).check(c)
		}
		//立木
		if p.hasEffect(liMuEffect) {
			return p.findSkill(data.LiMuSkill).(*liMuSkill).check(g, user)
		}
		//额外出杀
		if p.hasEffect(extraAtkEffect) {
			p.disableEffect(extraAtkEffect)
			return true
		}
	}
	if _, ok := g.events.list[g.index].(*duelEvent); ok { //决斗阶段
		return true
	}
	if _, ok := g.events.list[g.index].(*nmrqEvent); ok { //南蛮入侵阶段
		return true
	}
	if _, ok := g.events.list[g.index].(*jdsrEvent); ok { //借刀杀人阶段
		return true
	}
	if _, ok := g.events.list[g.index].(*ctuAtkEvent); ok { //追杀阶段
		return true
	}
	if _, ok := g.events.list[g.index].(*luanWuEvent); ok { //追杀阶段
		return true
	}
	return false
}

func (c *atkCard) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	p := g.players[user]
	dst := distence(1)
	tarNum := uint8(1)
	if wep := p.equipSlot[data.WeaponSlot]; wep != nil {
		dst = wep.(*weaponCard).dst
	}
	//成略无视距离
	if p.hasEffect(chenglueEffect) {
		if p.findSkill(data.ChengLueSkill).(*chengLueSkill).check(c) {
			dst = 99
		}
	}
	//方天画戟
	if p.hasEffect(fthjEffect) && len(p.cards) == 1 {
		tarNum = 3
	}
	plist := g.getPlayerInDst(user, dst)
	for i, p := range plist {
		//空城
		if g.players[p].hasEffect(kongChengEffect) && len(g.players[p].cards) == 0 {
			plist = append(plist[:i], plist[i+1])
		}
		//凶祸
		if g.players[p].hasEffect(xionhuoEffect) {
			skill := g.players[p].findSkill(data.XionHuoSkill).(*xionHuoSkill)
			if skill.target != nil {
				if skill.target[0] == user {
					plist = append(plist[:i], plist[i+1])
				}
			}
		}
	}
	return plist, tarNum
}

// 桃
type peachCard struct {
	card
}

func (c *peachCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID)
		g.players[user].delCard(c.ID)
	}
	g.recover(user, user, 1)
}

func (c *peachCard) useAble(g *Games, user data.PID) bool {
	if isItemInList(g.players[user].unUseableCol, c.Dec) {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if g.players[user].hp < g.players[user].maxHp {
			return true
		}
	}
	if _, ok := g.events.list[g.index].(*dyingEvent); ok {
		return true
	}
	return false
}

// 酒
type drunkCard struct {
	card
}

func (c *drunkCard) use(g *Games, user data.PID, target ...data.PID) {
	if c.ID != 0 {
		g.useCard(user, c.ID)
		g.players[user].delCard(c.ID)
	}
	g.players[user].isDrunk = true
	g.players[user].hasUseDrunk = true
}

func (c *drunkCard) useAble(g *Games, user data.PID) bool {
	p := g.players[user]
	if isItemInList(p.unUseableCol, c.Dec) {
		return false
	}
	if p.isDrunk {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if p.hasUseDrunk {
			if p.hasEffect(liMuEffect) {
				return g.players[user].findSkill(data.LiMuSkill).(*liMuSkill).check(g, user)
			}
			//成略无视次数
			if g.players[user].hasEffect(chenglueEffect) {
				return g.players[user].findSkill(data.ChengLueSkill).(*chengLueSkill).check(c)
			}
			return false
		} else {
			return true
		}
	}
	if e, ok := g.events.list[g.index].(*dyingEvent); ok {
		if e.dyingPlayer == e.user {
			return true
		}
	}
	return false
}

// 闪
type dodgeCard struct {
	card
}

func (c *dodgeCard) useAble(g *Games, user data.PID) bool {
	if isItemInList(g.players[user].unUseableCol, c.Dec) {
		return false
	}
	if _, ok := g.events.list[g.index].(*dodgeEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*wjqfEvent); ok {
		return true
	}
	return false
}
