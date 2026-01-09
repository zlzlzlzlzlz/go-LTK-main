package server

import (
	"goltk/data"
	"math/rand/v2"
)

type skillI interface {
	use(g *Games, user *player, inf useSkillInf, arg ...any) bool //使用技能,返回值代表是否结束当前事件
	getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf
	isUseAble(g *Games, user *player) bool //返回主动技在当前是否可用，如果不是主动技则总返回false
	getID() data.SID
	init(g *Games, p *player)
	handleTurnEnd(*Games) //处理回合结束
}

type useSkillInf struct {
	targets []data.PID
	cards   []data.CID
	args    []byte
}

type skill struct {
	id data.SID
}

func (s *skill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	return false
}

func (s *skill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 0}
}

func (s *skill) isUseAble(g *Games, user *player) bool {
	return false
}

func (s *skill) getID() data.SID {
	return s.id
}

func (s *skill) init(g *Games, p *player) {}

func (s *skill) handleTurnEnd(g *Games) {}

// 八卦阵类型
type bgzSkill struct {
	skill
}

func (s *bgzSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	var result data.CID
	g.events.insert(g.index, newSkillJudgeEvent(user.pid, data.BGZSkill, &result), newBGZEvent(user.pid, &result))
	return false
}

// 贯石斧
type gsfSkill struct {
	skill
}

func (s *gsfSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newGSFEvent(arg[0].([]eventI), user.pid))
	return false
}

// 青龙偃月刀
type qlyydSkill struct {
	skill
}

func (s *qlyydSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newCtuAtkEvent(user.pid, inf.targets[0]))
	return false
}

// 麒麟弓
type qlgSkill struct {
	skill
}

func (s *qlgSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	if g.players[inf.targets[0]].equipSlot[2] != nil || g.players[inf.targets[0]].equipSlot[3] != nil {
		g.events.insert(g.index, newQLGEvent(user.pid, inf.targets[0]))
	}
	return false
}

// 双股剑
type cxsgjSkill struct {
	skill
}

func (s *cxsgjSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newCXSGJEvent(user.pid, inf.targets[0]))
	return false
}

// 朱雀羽扇
type zqysSkill struct {
	skill
}

func (s *zqysSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	for _, e := range arg[0].([]eventI) {
		e.(*atkevent).dmgType = data.FireDmg
	}
	//朱雀
	if user.hasEffect(zhuQueEffect) {
		g.sendCard2Player(user.pid, g.getCardsFromTop(1)...)
		g.useSkill(user.pid, data.ZhuQueSkill)
	}
	return false
}

// 寒冰剑
type hbjSkill struct {
	skill
}

func (s *hbjSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	arg[0].(*damageEvent).setSkip(true)
	g.events.insert(g.index, newDropOtherCardEvent(user.pid, inf.targets[0]), newDropOtherCardEvent(user.pid, inf.targets[0]))
	return false
}

// 丈八蛇矛
type zbsmSkill struct {
	skill
}

func (s *zbsmSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		tmpcard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack})
		targets, _ := tmpcard.getAvailableTarget(g, user.pid)
		rsp := data.UseSkillRsp{ID: data.ZBSMSkill, Targets: targets, Cards: user.cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		dec := data.NoCol
		c1, c2 := g.cards[inf.cards[0]], g.cards[inf.cards[1]]
		if c1.getDecor().IsRed() && c2.getDecor().IsRed() {
			dec = data.RedCol
		} else if c1.getDecor().ISBlack() && c2.getDecor().ISBlack() {
			dec = data.BlackCol
		}
		if dec == data.NoCol && user.hasEffect(zhangbaEffect) {
			g.sendCard2Player(user.pid, g.getCardsFromTop(1)...)
		}
		g.useSkill(user.pid, data.ZBSMSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		tmpcard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack, Dec: dec})
		target := inf.targets
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			//应对出牌
			//应援
			if user.hasEffect(yingYuanEffect) {
				user.findSkill(data.YingYuanSkill).(*yingYuanSkill).check(g, user.pid, inf.targets, inf.cards...)
			}
			tmpcard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = inf.targets
		}
		if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
			//应对出牌
			tmpcard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = inf.targets
		}
		if _, ok := g.events.list[g.index].(*duelEvent); ok {
			//应对决斗
			target = []data.PID{arg[0].(data.PID)}
			*(arg[1].(*cardI)) = tmpcard
		}
		if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
			//应对南蛮入侵
			target = []data.PID{arg[0].(data.PID)}
		}
		g.useTmpCard(user.pid, data.Attack, dec, 0, data.ConvertCard, target...)
		return true
	}
	return false
}

func (s *zbsmSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if !user.hasAttack {
			return true
		}
		if user.hasEffect(unLimit) {
			return true
		}
		if user.hasEffect(liMuEffect) {
			return user.findSkill(data.LiMuSkill).(*liMuSkill).check(g, user.pid)
		}
	}
	if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*duelEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
		return true
	}
	return false
}

// 奋音
type fenYinSkill struct {
	skill
	enable bool
	col    bool //红为true,黑为false
}

// 检查技能能否使用
func (s *fenYinSkill) check(g *Games, card cardI, user data.PID) {
	if card.getDecor().IsRed() == card.getDecor().ISBlack() {
		return
	}
	g.useSkill(user, data.FeiYingSkill, byte(card.getDecor()))
	if !s.enable {
		s.enable = true
		s.col = card.getDecor().IsRed()
	} else {
		if s.col != card.getDecor().IsRed() {
			g.useSkill(user, data.FenYinSkill)
			s.col = card.getDecor().IsRed()
			g.sendCard2Player(user, g.getCardsFromTop(1)...)
		}
	}
}

func (s *fenYinSkill) handleTurnEnd(g *Games) {
	s.enable = false
}

// 蒺藜
type jiLiSkill struct {
	skill
	count distence
}

func (s *jiLiSkill) check(g *Games, dst distence, user data.PID) {
	s.count++
	g.useSkill(user, data.JiLiSkill, byte(s.count))
	if dst == -1 {
		if wep := g.players[user].equipSlot[data.WeaponSlot]; wep != nil {
			dst = wep.(*weaponCard).dst
		} else {
			dst = 1
		}
	}
	if s.count == dst {
		g.sendCard2Player(user, g.getCardsFromTop(int(dst))...)
		g.useSkill(user, data.JiLiSkill)
	}
}

func (s *jiLiSkill) handleTurnEnd(*Games) {
	s.count = 0
}

// 破军
type poJunSkill struct {
	skill
}

func (s *poJunSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newPoJunEvent(user.pid, inf.targets[0]))
	return false
}

func (s *poJunSkill) handleTurnEnd(g *Games) {
	for _, p := range g.players {
		cards := p.getTSCard(data.PoJunSkill)
		if len(cards) != 0 {
			p.addCard(cards...)
			iterators(g.clients, func(c clientI) { c.MoveCard(p.pid, p.pid, cards...) })
		}
	}
}

// 铁骑
type tieJiSkill struct {
	skill
}

func (s *tieJiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	var result data.CID
	g.events.insert(g.index, newSkillJudgeEvent(user.pid, data.TieJiSkill, &result), newTieJiEvent(arg[0].(*atkevent), &result))
	return false
}

// 问计
type wenJiSkill struct {
	skill
	name data.CardName
}

func (s *wenJiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newWenJiEvent(user.pid, inf.targets[0]))
	return false
}

func (s *wenJiSkill) handleTurnEnd(g *Games) {
	s.name = data.NoName
}

func (s *wenJiSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	for pid := g.getNextPid(user); pid != user; pid = g.getNextPid(pid) {
		if len(g.players[pid].cards) != 0 {
			inf.TargetList = append(inf.TargetList, pid)
		}
	}
	return inf
}

// 烈弓
type lieGongSkill struct {
	skill
}

func (s *lieGongSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	arg[0].(*atkevent).unResponsive = true
	return false
}

func (s *lieGongSkill) check(g *Games, user, target data.PID, event *atkevent) {
	p := g.players[user]
	distence := 0
	if p.equipSlot[data.WeaponSlot] != nil {
		distence = int(p.equipSlot[data.WeaponSlot].(*weaponCard).dst)
	} else {
		distence = 1
	}
	if len(g.players[target].cards) < int(distence) || len(g.players[target].cards) > int(p.hp) {
		g.events.insert(g.index, newSkillSelectEvent(data.LieGongSkill, user, nil, event))
	}
}

// 屯江
type tunJiangSkill struct {
	skill
	useState bool
	unable   bool
}

func (s *tunJiangSkill) check(g *Games, user data.PID, state data.GameState, card cardI, t []data.PID) {
	if state == data.UseCardState {
		if card == nil {
			s.useState = true
		} else {
			if isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack, data.Duel, data.GHCQ, data.SSQY, data.JDSR, data.NMRQ, data.WJQF,
				data.LBSS, data.BLCD, data.TYJY, data.WGFD}, card.getName()) {
				s.unable = true
			} else if card.getName() == data.TSLH {
				for _, pid := range t {
					if pid != user {
						s.unable = true
					}
				}
			} else if card.getName() == data.Burn {
				if user != t[0] {
					s.unable = true
				}
			}
		}
	} else if state == data.EndState {
		if s.useState && !s.unable {
			m := map[data.RoleSide]uint8{}
			for _, p := range g.players {
				m[p.side] = 1
			}
			g.sendCard2Player(user, g.getCardsFromTop(len(m))...)
			g.useSkill(user, data.TunJiangSkill)
		}
	}
}

func (s *tunJiangSkill) handleTurnEnd(*Games) {
	s.unable = false
	s.useState = false
}

// 狼袭
type langXiSkill struct {
	skill
}

func (s *langXiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmg := rand.IntN(3)
	if dmg == 0 {
		return false
	}
	dmgEvent := newDamageEvent(user.pid, inf.targets[0], data.NormalDmg, nil, data.HP(dmg))
	g.events.insert(g.index, newBeforeDamageEvent(dmgEvent), dmgEvent, newAfterDmgEvent(dmgEvent))
	return false
}

func (s *langXiSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	for pid := g.getNextPid(user); pid != user; pid = g.getNextPid(pid) {
		if g.players[pid].hp <= g.players[user].hp {
			inf.TargetList = append(inf.TargetList, pid)
		}
	}
	return inf
}

func (s *langXiSkill) init(g *Games, p *player) {
	g.addHpMax(p.pid, 2)
}

type yisuanSkill struct {
	skill
	hasUsed   bool
	eventList []eventI //技能选择阶段
}

func (s *yisuanSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmg := newDamageEvent(user.pid, user.pid, data.DffHPMax, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	for i := len(g.dropHeap) - 1; i >= 0; i-- {
		if g.dropHeap[i] == arg[0].(data.CID) {
			g.dropHeap = append(g.dropHeap[:i], g.dropHeap[i+1:]...)
			g.sendCard2Player(user.pid, arg[0].(data.CID))
			s.hasUsed = true
			for _, e := range s.eventList {
				e.setSkip(true)
			}
			return false
		}
	}
	return false
}

func (s *yisuanSkill) check(g *Games, card data.CID, user data.PID, targets []data.PID) {
	if s.hasUsed || card == 0 || (g.turnOwner != user) {
		return
	}
	if g.cards[card].getType() != data.TipsCardType {
		return
	}
	if g.cards[card].getName() == data.TSLH && targets != nil && len(targets) == 0 {
		return
	}
	for i := len(g.dropHeap) - 1; i >= 0; i-- {
		if g.dropHeap[i] == card {
			selectEvent := newSkillSelectEvent(data.YiSuanSkill, user, nil, card)
			s.eventList = append(s.eventList, selectEvent)
			g.events.insert(g.index, selectEvent)
			return
		}
	}
}

func (s *yisuanSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
}

// 枭首
type xiaoShouSkill struct {
	skill
}

func (s *xiaoShouSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmgEvent := newDamageEvent(user.pid, inf.targets[0], data.NormalDmg, nil, 3)
	g.events.insert(g.index, newBeforeDamageEvent(dmgEvent), dmgEvent, newAfterDmgEvent(dmgEvent))
	return false
}

func (s *xiaoShouSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	for pid := g.getNextPid(user); pid != user; pid = g.getNextPid(pid) {
		if g.players[pid].hp >= g.players[user].hp {
			inf.TargetList = append(inf.TargetList, pid)
		}
	}
	return inf
}

// 仁德
type rendeSkill struct {
	skill
	hasUsed bool
}

func (s *rendeSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		rsp := data.UseSkillRsp{ID: data.RenDeSkill, Targets: g.getAllAliveOther(user.pid), Cards: user.cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.targets) == 0 {
			return false
		}
		if len(inf.cards) <= 1 || !isListContainAllTheItem(user.cards, inf.cards...) {
			return false
		}
		s.hasUsed = true
		g.moveCard(user.pid, inf.targets[0], inf.cards...)
		g.useSkill(user.pid, data.RenDeSkill)
		if len(inf.cards) >= 2 {
			g.recover(user.pid, user.pid, 1)
		}
		return true
	}
	return false
}

func (s *rendeSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

func (s *rendeSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

type jieyingSkill struct {
	skill
	user data.PID
}

func (s *jieyingSkill) init(g *Games, p *player) {
	s.user = p.pid
	num := uint8(0)
	for _, p := range g.players {
		if p.side == data.Shu {
			num++
		}
	}
	num = min(num, 3)
	g.addHpMax(s.user, 2*num)
}

func (s *jieyingSkill) handleTurnEnd(g *Games) {
	if g.players[s.user].hp < g.players[s.user].maxHp && g.turnOwner == s.user {
		g.recover(s.user, s.user, 1)
		g.useSkill(s.user, data.JieYingSkill)
	}
}

// 龙胆
type longDanSkill struct {
	skill
}

func (s *longDanSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		tmpcard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack})
		targets, _ := tmpcard.getAvailableTarget(g, user.pid)
		cards := []data.CID{}
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			for _, c := range user.cards {
				if g.cards[c].getName() == data.Dodge {
					cards = append(cards, c)
				}
			}
		}
		if _, ok := g.events.list[g.index].(*duelEvent); ok {
			for _, c := range user.cards {
				if g.cards[c].getName() == data.Dodge {
					cards = append(cards, c)
				}
			}
		}
		if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
			for _, c := range user.cards {
				if g.cards[c].getName() == data.Dodge {
					cards = append(cards, c)
				}
			}
		}
		if _, ok := g.events.list[g.index].(*jdsrEvent); ok {
			for _, c := range user.cards {
				if g.cards[c].getName() == data.Dodge {
					cards = append(cards, c)
				}
			}
		}
		if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
			for _, c := range user.cards {
				if g.cards[c].getName() == data.Dodge {
					cards = append(cards, c)
				}
			}
		}
		if _, ok := g.events.list[g.index].(*dodgeEvent); ok {
			for _, c := range user.cards {
				if isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack}, g.cards[c].getName()) {
					cards = append(cards, c)
				}
			}
		}
		if _, ok := g.events.list[g.index].(*wjqfEvent); ok {
			for _, c := range user.cards {
				if isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack}, g.cards[c].getName()) {
					cards = append(cards, c)
				}
			}
		}
		rsp := data.UseSkillRsp{ID: data.LongDanSkill, Targets: targets, Cards: cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 1 {
			return false
		}
		originalCard := g.cards[inf.cards[0]]
		g.useSkill(user.pid, data.LongDanSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		tmpcard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack, Dec: originalCard.getDecor(),
			Num: originalCard.getNum()})
		target := inf.targets
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			//应对出牌
			tmpcard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = inf.targets
			g.useTmpCard(user.pid, data.Attack, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
			tmpcard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = inf.targets
			g.useTmpCard(user.pid, data.Attack, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*duelEvent); ok {
			//应对决斗
			target = []data.PID{arg[0].(data.PID)}
			*(arg[1].(*cardI)) = tmpcard
			g.useTmpCard(user.pid, data.Attack, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
			//应对南蛮入侵
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Attack, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*wjqfEvent); ok {
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Dodge, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*dodgeEvent); ok {
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Dodge, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		}
		if len(g.players[target[0]].cards) != 0 {
			num := rand.IntN(len(g.players[target[0]].cards))
			g.moveCard(target[0], user.pid, g.players[target[0]].cards[num])
		}
		return true
	}
	return false
}

func (s *longDanSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if !user.hasAttack {
			return true
		}
		if user.hasEffect(unLimit) {
			return true
		}
	}
	if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*duelEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*wjqfEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*dodgeEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
		return true
	}
	return false
}

// 武圣
type wushengSkill struct {
	skill
}

func (s *wushengSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		tmpcard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack})
		var targets []data.PID
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			targets, _ = tmpcard.getAvailableTarget(g, user.pid)
		} else {
			targets = g.getAllAlivePlayer()
		}
		cards := []data.CID{}
		if isItemInList(user.unUseableCol, data.DiamondDec) {
			cards = append(cards, []data.CID{0, 0, 0, 0}...)
		} else {
			for _, c := range user.equipSlot {
				if c == nil || !c.getDecor().IsRed() {
					cards = append(cards, 0)
					continue
				}
				cards = append(cards, c.getID())
			}
			for _, c := range user.cards {
				if g.cards[c].getDecor().IsRed() {
					cards = append(cards, c)
				}
			}
		}
		rsp := data.UseSkillRsp{ID: data.WuShengSkill, Targets: targets, Cards: cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 1 {
			return false
		}
		originalCard := g.cards[inf.cards[0]]
		g.useSkill(user.pid, data.WuShengSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		tmpcard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack, Dec: originalCard.getDecor(),
			Num: originalCard.getNum()})
		target := inf.targets
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			//应对出牌
			tmpcard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = inf.targets
		}
		if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
			tmpcard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = inf.targets
		}
		if _, ok := g.events.list[g.index].(*duelEvent); ok {
			//应对决斗
			target = []data.PID{arg[0].(data.PID)}
			*(arg[1].(*cardI)) = tmpcard
		}
		if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
			//应对南蛮入侵
			target = []data.PID{arg[0].(data.PID)}
		}
		if _, ok := g.events.list[g.index].(*ctuAtkEvent); ok {
			//应对追杀
			target = []data.PID{arg[0].(data.PID)}
			tmpcard.use(g, user.pid, target...)
		}
		g.useTmpCard(user.pid, data.Attack, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, target...)
		return true
	}
	return false
}

func (s *wushengSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if !user.hasAttack {
			return true
		}
		if user.hasEffect(unLimit) {
			return true
		}
	}
	if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*duelEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*ctuAtkEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
		return true
	}
	return false
}

// 图射
type tuSheSkill struct {
	skill
	num int
}

func (s *tuSheSkill) check(g *Games, user data.PID, card cardI, targets []data.PID) {
	p := g.players[user]
	s.num = 0
	if isItemInList([]data.CardType{data.WeaponCardType, data.ArmorCardType,
		data.HorseDownCardType, data.HorseUpCardType}, card.getType()) {
		return
	}
	for _, c := range p.cards {
		if g.cards[c].getType() == data.BaseCardType {
			return
		}
	}
	if targets != nil {
		if card.getName() == data.JDSR {
			s.num = 1
		} else {
			s.num = len(targets)
		}
	} else {
		if isItemInList([]data.CardName{data.NMRQ, data.WJQF}, card.getName()) {
			s.num = len(g.getAllAliveOther(user))
		} else if isItemInList([]data.CardName{data.WGFD, data.TYJY}, card.getName()) {
			s.num = len(g.getAllAlivePlayer())
		} else {
			// if isItemInList([]data.CardName{data.Peach, data.Drunk, data.WZSY, data.Lightning, data.JDSR}, card.getName())
			s.num = 1
		}
	}
	g.events.insert(g.index, newSkillSelectEvent(data.TuSheSkill, user, nil))
}

func (s *tuSheSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.sendCard2Player(user.pid, g.getCardsFromTop(s.num)...)
	return false
}

// 立牧
type liMuSkill struct {
	skill
	hasUsed bool
}

func (s *liMuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		cards := []data.CID{}
		for _, c := range user.equipSlot {
			if c == nil || c.getDecor() != data.DiamondDec {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		for _, c := range user.cards {
			if g.cards[c].getDecor() == data.DiamondDec {
				cards = append(cards, c)
			}
		}
		rsp := data.UseSkillRsp{ID: data.LiMuSkill, Cards: cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 1 {
			return false
		}
		originalCard := g.cards[inf.cards[0]]
		g.useSkill(user.pid, data.LiMuSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		tmpcard := newCard(data.Card{CardType: data.TipsCardType, Name: data.LBSS, Dec: originalCard.getDecor(),
			Num: originalCard.getNum()})
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			//应对出牌
			tmpcard.use(g, user.pid, user.pid)
			*(arg[0].(*cardI)) = tmpcard
			*(arg[1].(*[]data.PID)) = []data.PID{user.pid}
		}
		g.recover(user.pid, user.pid, 1)
		g.useTmpCard(user.pid, data.LBSS, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard, user.pid)
		s.hasUsed = true
		return true
	}
	return false
}

func (s *liMuSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if user.judgeSlot[data.LBSSSlot] == nil && !s.hasUsed {
			return true
		}
	}
	return false
}

func (s *liMuSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
}

func (s *liMuSkill) check(g *Games, user data.PID) bool {
	for _, s := range g.players[user].judgeSlot {
		if s != nil {
			return true
		}
	}
	return false
}

// 地动
type diDongSkill struct {
	skill
}

func (s *diDongSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.players[inf.targets[0]].turnBack = !g.players[inf.targets[0]].turnBack
	g.useSkill(inf.targets[0], data.TurnBackSkill)
	return false
}

func (s *diDongSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	inf.TargetList = g.getAllAliveOther(user)
	return inf
}

// 归心
type guiXinSkill struct {
	skill
}

func (s *guiXinSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	for _, p := range g.getAllAliveOther(user.pid) {
		list := g.players[p].cards
		for _, c := range g.players[p].equipSlot {
			if c != nil && c.getID() == 0 {
				list = append(list, c.getID())
			}
		}
		if len(list) == 0 {
			continue
		}
		num := rand.IntN(len(list))
		g.moveCard(p, user.pid, list[num])
	}
	g.useSkill(user.pid, data.TurnBackSkill)
	user.turnBack = !user.turnBack
	return false
}

func (s *guiXinSkill) check(g *Games) (n int) {
	for _, p := range g.getAllAlivePlayer() {
		if g.players[p].side == data.Wei {
			n++
		}
	}
	return 2 * n
}

type luoLeiSkill struct {
	skill
}

func (s *luoLeiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	num := rand.IntN(len(inf.targets))
	e := newDamageEvent(user.pid, inf.targets[num], data.LightningDmg, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(e), e, newAfterDmgEvent(e))
	return false
}

func (s *luoLeiSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAliveOther(user)}
}

type guiHuoSkill struct {
	skill
}

func (s *guiHuoSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	e := newDamageEvent(user.pid, inf.targets[0], data.FireDmg, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(e), e, newAfterDmgEvent(e))
	return false
}

func (s *guiHuoSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAliveOther(user)}
}

// 丈八
type zhangbaSkill struct {
	skill
}

func (s *zhangbaSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *zhangbaSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.ZBSM, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.ZBSM, data.NoDec, 0, data.VirtualCard)
}

type yanyueSkill struct {
	skill
	enable bool
}

func (s *yanyueSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *yanyueSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.QLYYD, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.QLYYD, data.NoDec, 0, data.VirtualCard)
}

func (s *yanyueSkill) handleTurnEnd(g *Games) {
	s.enable = false
}

// 青钢
type qinggangSkill struct {
	skill
}

func (s *qinggangSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *qinggangSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.QGJ, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.QGJ, data.NoDec, 0, data.VirtualCard)
}

func (s *qinggangSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmg := newDamageEvent(user.pid, inf.targets[0], data.NormalDmg, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	return false
}

func (s *qinggangSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	inf.TargetList = g.getAllAliveOther(user)
	for i, p := range inf.TargetList {
		if p == arg[0].(data.PID) {
			inf.TargetList = append(inf.TargetList[:i], inf.TargetList[i+1:]...)
		}
	}
	return inf
}

// 世仇
type ShiChouSkill struct {
	skill
}

func (s *ShiChouSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmg := arg[0].(*damageEvent)
	arg[1].(*beforeDamageEvent).setSkip(true)
	bleed := newDamageEvent(dmg.target, dmg.target, data.BleedingDmg, nil, 1)
	bleed.dmg = dmg.dmg
	g.events.insert(g.index, bleed)
	dmg.damageType = data.DffHPMax
	return false
}

// 麒麟
type qilinSkill struct {
	skill
}

func (s *qilinSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *qilinSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.QLG, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.QLG, data.NoDec, 0, data.VirtualCard)
}

// 观星
type guanxingSkill struct {
	skill
}

func (s *guanxingSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newGuanXingEvent(user.pid))
	return false
}

// 八阵
type bazhenSkill struct {
	skill
}

func (s *bazhenSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *bazhenSkill) check(g *Games, p *player) {
	if p.equipSlot[data.ArmorSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.BGZ, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.ArmorSlot] = c
	g.useTmpCard(p.pid, data.BGZ, data.NoDec, 0, data.VirtualCard)
}

// 孔明八阵
type newbazhenSkill struct {
	bazhenSkill
	enable bool
}

func (s *newbazhenSkill) count() data.HP {
	if s.enable {
		return 0
	} else {
		s.enable = true
		return 1
	}
}

func (s *newbazhenSkill) handleTurnEnd(g *Games) {
	s.enable = false
}

// 业炎
type yeyanSkill struct {
	skill
	hasUsed bool
}

func (s *yeyanSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		cards := user.cards
		rsp := data.UseSkillRsp{ID: data.YeYanSkill, Cards: cards, Targets: g.getAllAlivePlayer()}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 4 {
			return false
		}
		g.useSkill(user.pid, data.YeYanSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		dmguser := newDamageEvent(user.pid, user.pid, data.BleedingDmg, nil, 3)
		dmgtar := newDamageEvent(user.pid, inf.targets[0], data.FireDmg, nil, 3)
		g.events.insert(g.index, newBeforeDamageEvent(dmguser), dmguser, newAfterDmgEvent(dmguser),
			newBeforeDamageEvent(dmgtar), dmgtar, newAfterDmgEvent(dmgtar))
		s.hasUsed = true
		return true
	}
	return false
}

func (s *yeyanSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if user.judgeSlot[data.LBSSSlot] == nil && !s.hasUsed {
			return true
		}
	}
	return false
}

func (s *yeyanSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
}

// 琴音
type qinYinSkill struct {
	skill
}

func (s *qinYinSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newQinYinevent(user.pid))
	return false
}

// 朱雀
type zhuQueSkill struct {
	skill
}

func (s *zhuQueSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *zhuQueSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.ZQYS, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.ZQYS, data.NoDec, 0, data.VirtualCard)
}

// 古锭
type gudingSkill struct {
	skill
}

func (s *gudingSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *gudingSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.GDD, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.GDD, data.NoDec, 0, data.VirtualCard)
}

// 英魂
type yingHunSkill struct {
	skill
}

func (s *yingHunSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newYingHunEvent(user.pid, inf.targets[0]))
	return false
}

func (s *yingHunSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAlivePlayer()}
}

// 无双
type wuShuangSkill struct {
	skill
	hasUsed bool
}

func (s *wuShuangSkill) check() bool {
	if s.hasUsed {
		s.hasUsed = false
		return false
	}
	s.hasUsed = true
	return true
}

// 方天
type fangTianSkill struct {
	skill
	count uint8
}

func (s *fangTianSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *fangTianSkill) check(g *Games, p *player) {
	if p.equipSlot[data.WeaponSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.FTHJ, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.WeaponSlot] = c
	g.useTmpCard(p.pid, data.FTHJ, data.NoDec, 0, data.VirtualCard)
}

func (s *fangTianSkill) counter(g *Games, user data.PID) {
	s.count++
	if s.count == 3 {
		num := g.players[user].findSkill(data.LiYuSkill).(*LiYuSkill).count
		g.sendCard2Player(user, g.getCardsFromTop(int(num))...)
		s.count = 0
		g.useSkill(user, data.FangTianSkill)
		dmg := newDamageEvent(user, user, data.BleedingDmg, nil, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	}
}

// 利驭
type LiYuSkill struct {
	skill
	count uint8
}

func (s *LiYuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newLiYuEvent(user.pid))
	return false
}

func (s *LiYuSkill) handleTurnEnd(g *Games) {
	for _, p := range g.players {
		cards := p.getTSCard(data.LiYuSkill)
		if len(cards) != 0 {
			p.addCard(cards...)
			iterators(g.clients, func(c clientI) { c.MoveCard(p.pid, p.pid, cards...) })
			g.useSkill(p.pid, data.LiYuSkill)
		}
	}
	s.count = 0
}

// 奇制
type qiZhiSkill struct {
	skill
	count uint8
}

func (s *qiZhiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newQiZhiEvent(user.pid, inf.targets[0]))
	s.count++
	g.useSkill(user.pid, data.QiZhiSkill, s.count)
	return false
}

func (s *qiZhiSkill) check(g *Games, user data.PID, card cardI, t []data.PID) {
	targets := []data.PID{}
	if isItemInList([]data.CardName{data.TYJY, data.WGFD}, card.getName()) {
		return
	} else if isItemInList([]data.CardName{data.Drunk, data.Peach, data.WZSY, data.Lightning}, card.getName()) {
		targets = append(targets, user)
	} else {
		targets = t
	}
	g.events.insert(g.index, newSkillSelectEvent(data.QiZhiSkill, user, nil, targets))
}

func (s *qiZhiSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	list := g.getAllAlivePlayer()
	m := arg[0]
	for _, p := range list {
		if !isItemInList(m.([]data.PID), p) {
			inf.TargetList = append(inf.TargetList, p)
		}
	}
	return inf
}

// 进趋
type jinQuSkill struct {
	skill
	count uint8
}

func (s *jinQuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	s.count = user.findSkill(data.QiZhiSkill).(*qiZhiSkill).count
	user.findSkill(data.QiZhiSkill).(*qiZhiSkill).count = 0
	g.sendCard2Player(user.pid, g.getCardsFromTop(2)...)
	num := min(uint8(len(user.cards))-s.count, 0)
	if num > 0 {
		g.events.insert(g.index, newDropSelfHandCardEvent(user.pid, num))
	}
	return false
}

// 镇骨
type zhenGuSkill struct {
	skill
	user, target data.PID
}

func (s *zhenGuSkill) init(g *Games, p *player) {
	s.user, s.target = p.pid, p.pid
}

func (s *zhenGuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	s.target = inf.targets[0]
	g.players[inf.targets[0]].enableEffect(zhenGuEffect)
	g.useSkill(inf.targets[0], data.ZhenGuSkill, byte(1))
	return false
}

func (s *zhenGuSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAliveOther(user)}
}

func (s *zhenGuSkill) handleTurnEnd(g *Games) {
	if g.turnOwner != s.user && g.turnOwner != s.target {
		return
	}
	if s.user == s.target {
		return
	}
	p, t := g.players[s.user], g.players[s.target]
	if !t.hasEffect(zhenGuEffect) {
		return
	}
	pnum, tnum := len(p.cards), len(t.cards)
	if pnum > tnum {
		g.sendCard2Player(s.target, g.getCards(pnum-tnum, s.target)...)
	} else if pnum < tnum {
		g.events.insert(g.index, newDropSelfHandCardEvent(s.target, uint8(tnum)-uint8(pnum)))
	}
	g.useSkill(s.user, data.ZhenGuSkill)
	if g.turnOwner == s.target {
		g.useSkill(s.target, data.ZhenGuSkill, byte(0))
		t.disableEffect(zhenGuEffect)
		s.target = s.user
	}
}

// 应援
type yingYuanSkill struct {
	skill
	list  []data.CardName
	cards []data.CID
}

func (s *yingYuanSkill) check(g *Games, user data.PID, targets []data.PID, cards ...data.CID) {
	if user != g.turnOwner {
		return
	}
	s.cards = []data.CID{}
	for _, c := range cards {
		if g.cards[c].getName() == data.TSLH && len(targets) == 0 && len(cards) == 1 {
			return
		}
		if isItemInList(s.list, g.cards[c].getName()) {
			return
		}
		//确保在弃牌堆里
		for i := len(g.dropHeap) - 1; i >= 0; i-- {
			if g.cards[g.dropHeap[i]].getID() == c {
				//添加要应援的卡
				s.cards = append(s.cards, c)
				break
			}
		}
	}
	g.events.insert(g.index, newSkillSelectEvent(data.YingYuanSkill, user, nil))
}

func (s *yingYuanSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAliveOther(user)}
}

func (s *yingYuanSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	for _, c := range s.cards {
		for i := len(g.dropHeap) - 1; i >= 0; i-- {
			if g.dropHeap[i] == c {
				g.moveCard(data.SpecialPIDGame, inf.targets[0], c)
				s.list = append(s.list, g.cards[c].getName())
				goto next
			}
		}
		for i := len(g.mainHeap) - 1; i >= 0; i-- {
			if g.mainHeap[i] == c {
				g.moveCard(data.SpecialPIDGame, inf.targets[0], c)
				s.list = append(s.list, g.cards[c].getName())
			}
		}
	next:
	}
	return false
}

func (s *yingYuanSkill) handleTurnEnd(g *Games) {
	s.list = nil
}

// 自书
type zishuSkill struct {
	skill
	list []data.CID
	user data.PID
}

func (s *zishuSkill) check(g *Games, user data.PID, cards ...data.CID) {
	if g.turnOwner == user {
		c := g.getCardsFromTop(1)
		g.players[user].addCard(c...)
		iterators(g.clients, func(cc clientI) { cc.SendCard(user, c...) })
		g.useSkill(user, data.ZiShuSkill)
		return
	}
	s.list = append(s.list, cards...)
}

func (s *zishuSkill) handleTurnEnd(g *Games) {
	if len(s.list) != 0 {
		clist := []data.CID{}
		for _, c := range s.list {
			if isItemInList(g.players[s.user].cards, c) {
				clist = append(clist, c)
			}
		}
		g.removePlayercard(s.user, clist...)
		s.list = nil
		g.useSkill(s.user, data.ZiShuSkill)
	}
}

func (s *zishuSkill) init(g *Games, p *player) {
	s.user = p.pid
}

// 潜袭
type qianXiSkill struct {
	skill
}

func (s *qianXiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.sendCard2Player(user.pid, g.getCardsFromTop(1)...)
	g.events.insert(g.index, newQianXiEvent(user.pid, inf.targets[0]))
	return false
}

func (s *qianXiSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	list := g.getPlayerInDst(user, 1)
	return data.AvailableTargetInf{TargetNum: 1, TargetList: list}
}

// 追击
type zhuiJiSkill struct {
	skill
}

func (s *zhuiJiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	var result data.CID
	g.events.insert(g.index, newSkillJudgeEvent(user.pid, data.ZhuiJiSkill, &result),
		newZhuiJiEvent(arg[0].(*damageEvent), arg[1].(*beforeDamageEvent), &result))
	return false
}

// 诡计
type guijiSkill struct {
	skill
}

func (s *guijiSkill) check(g *Games, user data.PID) {
	p := g.players[user]
	list := []int{}
	for i := 0; i < len(p.judgeSlot); i++ {
		if p.judgeSlot[i] != nil {
			list = append(list, i)
		}
	}
	if len(list) == 0 {
		return
	}
	num := rand.IntN(len(list))
	c := p.judgeSlot[list[num]]
	if c.getID() != 0 {
		g.useCard(user, c.getID())
		g.players[user].delCard(c.getID())
		g.dropCards(c.getID())
	} else {
		g.useTmpCard(user, c.getName(), c.getDecor(), data.CNum(c.getDecor()), data.VirtualCard)
	}
	g.useSkill(user, data.GuiJiSkill)
}

// 索命
type suoMinfSkill struct {
	skill
}

func (s *suoMinfSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	plist := g.getAllAliveOther(user.pid)
	for i := 0; i < len(plist); i++ {
		if g.players[plist[i]].isLinked {
			continue
		}
		g.events.insert(g.index, newTSLHEvent(user.pid, plist[i]))
	}
	return false
}

// 吸星
type xiXingSkill struct {
	skill
}

func (s *xiXingSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmg := newDamageEvent(user.pid, inf.targets[0], data.LightningDmg, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	return false
}

func (s *xiXingSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	plist := g.getAllAliveOther(user)
	pid := []data.PID{}
	for i := 0; i < len(plist); i++ {
		if g.players[plist[i]].isLinked {
			pid = append(pid, plist[i])
		}
	}
	return data.AvailableTargetInf{TargetNum: 1, TargetList: pid}
}

func (s *xiXingSkill) check(g *Games, user data.PID) {
	plist := g.getAllAliveOther(user)
	for i := 0; i < len(plist); i++ {
		if g.players[plist[i]].isLinked {
			g.events.insert(g.index, newSkillSelectEvent(data.XiXingSkill, user, nil))
			return
		}
	}
}

// 炼狱
type lianYuSkill struct {
	skill
}

func (s *lianYuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	for p := g.getPrvPid(user.pid); p != user.pid; p = g.getPrvPid(p) {
		if g.players[p].death {
			continue
		}
		dmg := newDamageEvent(user.pid, p, data.FireDmg, nil, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	}
	return false
}

// 强征
type qiangZhengSkill struct {
	skill
}

func (s *qiangZhengSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	t := g.players[inf.targets[0]]
	num := rand.IntN(len(t.cards))
	g.moveCard(inf.targets[0], user.pid, t.cards[num])
	return false
}

func (s *qiangZhengSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	plist := g.getAllAliveOther(user)
	for i := 0; i < len(plist); i++ {
		if len(g.players[plist[i]].cards) == 0 {
			plist = append(plist[:i], plist[i+1:]...)
		}
	}
	return data.AvailableTargetInf{TargetNum: 1, TargetList: plist}
}

// 蛮甲
type manjiaSkill struct {
	skill
}

func (s *manjiaSkill) init(g *Games, p *player) {
	s.check(g, p)
}

func (s *manjiaSkill) check(g *Games, p *player) {
	if p.equipSlot[data.ArmorSlot] != nil {
		return
	}
	c := newCard(data.NewCard(data.TengJia, data.NoDec, 0))
	if eff, ok := equipEffectMap[c.getName()]; ok {
		p.enableEffect(eff)
	}
	if skill, ok := equipSkillMap[c.getName()]; ok {
		p.addSkill(skill)
	}
	p.equipSlot[data.ArmorSlot] = c
	g.useTmpCard(p.pid, data.TengJia, data.NoDec, 0, data.VirtualCard)
}

// 涅槃
type niePanSkill struct {
	skill
	hasUsed bool
}

func (s *niePanSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	s.hasUsed = true
	user.cleanCards(g)
	user.hasUseDrunk = false
	if user.isDrunk {
		user.isDrunk = false
		g.useSkill(user.pid, data.CleanDrunk)
	}
	if user.isLinked {
		g.events.insert(g.index, newTSLHEvent(user.pid, user.pid))
	}
	if user.turnBack {
		user.turnBack = false
		g.useSkill(user.pid, data.TurnBackSkill)
	}
	num := 3 - user.hp
	g.recover(user.pid, user.pid, num)
	g.sendCard2Player(user.pid, g.getCardsFromTop(3)...)
	for _, s := range user.skills {
		s.init(g, user)
	}
	arg[0].(*dieEvent).setSkip(true)
	return false
}

// 魔箭
type mojianSkill struct {
	skill
	hasUsed bool
}

func (s *mojianSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	tmpcard := newCard(data.Card{CardType: data.TipsCardType, Name: data.WJQF, Dec: data.NoDec})
	tmpcard.use(g, user.pid)
	g.useTmpCard(user.pid, data.WJQF, data.NoDec, 0, data.VirtualCard, g.getAllAliveOther(user.pid)...)
	return false
}

func (s *mojianSkill) check() bool {
	if s.hasUsed {
		return false
	}
	s.hasUsed = true
	return true
}

func (s *mojianSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

// 御兽
type yushouSkill struct {
	skill
	hasUsed bool
}

func (s *yushouSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	tmpcard := newCard(data.Card{CardType: data.TipsCardType, Name: data.NMRQ, Dec: data.NoDec})
	tmpcard.use(g, user.pid)
	g.useTmpCard(user.pid, data.NMRQ, data.NoDec, 0, data.VirtualCard, g.getAllAliveOther(user.pid)...)
	return false
}

func (s *yushouSkill) check() bool {
	if s.hasUsed {
		return false
	}
	s.hasUsed = true
	return true
}

func (s *yushouSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

// 决策
type jueCeSkill struct {
	skill
}

func (s *jueCeSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	dmg := newDamageEvent(user.pid, inf.targets[0], data.NormalDmg, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
	return false
}

// 反馈
type fanKuiSkill struct {
	skill
}

func (s *fanKuiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newfanKuiEvent(user.pid, inf.targets[0]))
	return false
}

// 丹术
type danshuSkill struct {
	skill
}

func (s *danshuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	var result data.CID
	g.events.insert(g.index, newSkillJudgeEvent(user.pid, data.DanShuSkill, &result), newDanShuEvent(user.pid, &result))
	return false
}

// 魔炎
type moYanSkill struct {
	skill
}

func (s *moYanSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	var result data.CID
	g.events.insert(g.index, newSkillJudgeEvent(user.pid, data.MoYanSkill, &result), newMoYanEvent(user.pid, inf.targets[0], &result))
	return false
}

func (s *moYanSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAliveOther(user)}
}

// 急救
type jijiuSkill struct {
	skill
}

func (s *jijiuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		cards := []data.CID{}
		for _, c := range user.equipSlot {
			if c == nil || !c.getDecor().IsRed() {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		for _, c := range user.cards {
			if g.cards[c].getDecor().IsRed() {
				cards = append(cards, c)
			}
		}
		rsp := data.UseSkillRsp{ID: data.JiJiuSkill, Cards: cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 1 {
			return false
		}
		originalCard := g.cards[inf.cards[0]]
		g.useSkill(user.pid, data.JiJiuSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		g.useTmpCard(user.pid, data.Peach, originalCard.getDecor(), originalCard.getNum(), data.ConvertCard)
		return true
	}
	return false
}

func (s *jijiuSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*dyingEvent); ok {
		if g.turnOwner != user.pid {
			return true
		}
	}
	return false
}

// 青囊
type qingnangSkill struct {
	skill
	hasUsed bool
}

func (s *qingnangSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		plist := []data.PID{}
		for _, p := range g.players {
			if p.hp < p.maxHp {
				plist = append(plist, p.pid)
			}
		}
		rsp := data.UseSkillRsp{ID: data.QingNangSkill, Cards: user.cards, Targets: plist}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 1 {
			return false
		}
		g.useSkill(user.pid, data.QingNangSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		g.recover(user.pid, inf.targets[0], 1)
		s.hasUsed = true
		return true
	}
	return false
}

func (s *qingnangSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if !s.hasUsed {
			return true
		}
	}
	return false
}

func (s *qingnangSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

// 结姻
type jieyinSkill struct {
	skill
	hasUsed bool
}

func (s *jieyinSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		plist := []data.PID{}
		for _, p := range g.getAllAliveOther(user.pid) {
			if g.players[p].hp < g.players[p].maxHp {
				plist = append(plist, p)
			}
		}
		rsp := data.UseSkillRsp{ID: data.JieYinSkill, Targets: plist, Cards: user.cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.targets) == 0 {
			return false
		}
		if len(inf.cards) <= 1 || !isListContainAllTheItem(user.cards, inf.cards...) {
			return false
		}
		s.hasUsed = true
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		g.useSkill(user.pid, data.JieYinSkill)
		g.recover(user.pid, inf.targets[0], 1)
		g.recover(user.pid, user.pid, 1)
		return true
	}
	return false
}

func (s *jieyinSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

func (s *jieyinSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

// 良助
type liangzhuSkill struct {
	skill
}

func (s *liangzhuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newLiangZhuEvent(user.pid, inf.targets[0]))
	return false
}

// 灭吴
type miewuSkill struct {
	skill
	hasUsed bool
	gaint   bool
}

func (s *miewuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌类型,为1时为发送使用技能请求,为2时为请求可用目标
	if len(inf.args) < 1 || !s.gaint {
		return false
	}
	switch inf.args[0] {
	case 0:
		nameList := []byte{}
		for cname := data.Attack; cname <= data.Lightning; cname++ {
			tmpCard := newCard(data.NewCard(cname, data.NoDec, 0))
			if tmpCard.useAble(g, user.pid) {
				nameList = append(nameList, byte(cname))
			}
		}
		cards := []data.CID{}
		for _, c := range user.equipSlot {
			if c == nil || isItemInList(user.unUseableCol, c.getDecor()) {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		for _, c := range user.cards {
			if isItemInList(user.unUseableCol, g.cards[c].getDecor()) {
				continue
			}
			cards = append(cards, c)
		}
		g.clients[user.pid].SendUseSkillRsp(data.UseSkillRsp{ID: data.MieWuSkill, Cards: cards, Args: nameList})
		return false
	case 1:
		if len(inf.args) < 2 || len(inf.cards) != 1 || inf.args[1] > byte(data.Lightning) {
			return false
		}
		srcCard := g.cards[inf.cards[0]]
		var tmpCard cardI
		g.useSkill(user.pid, data.MieWuSkill)
		g.dropCards(inf.cards[0])
		g.removePlayercard(user.pid, inf.cards[0])
		target := inf.targets
		if _, ok := g.events.list[g.index].(*useCardEvent); ok {
			//应对出牌
			tmpCard = newCard(data.NewCard(data.CardName(inf.args[1]), srcCard.getDecor(), srcCard.getNum()))
			tmpCard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpCard
			*(arg[1].(*[]data.PID)) = inf.targets
			g.useTmpCard(user.pid, data.CardName(inf.args[1]), srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
			tmpCard = newCard(data.NewCard(data.CardName(inf.args[1]), srcCard.getDecor(), srcCard.getNum()))
			tmpCard.use(g, user.pid, inf.targets...)
			*(arg[0].(*cardI)) = tmpCard
			*(arg[1].(*[]data.PID)) = inf.targets
			g.useTmpCard(user.pid, data.CardName(inf.args[1]), srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*duelEvent); ok {
			//应对决斗
			target = []data.PID{arg[0].(data.PID)}
			*(arg[1].(*cardI)) = tmpCard
			g.useTmpCard(user.pid, data.Attack, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
			//应对南蛮入侵
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Attack, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*wjqfEvent); ok {
			//万箭齐发
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Dodge, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*dyingEvent); ok {
			//濒死
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Peach, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*dodgeEvent); ok {
			//闪
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Dodge, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*jdsrEvent); ok {
			//借刀杀人
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Attack, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*ctuAtkEvent); ok {
			//追杀
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.Attack, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		if _, ok := g.events.list[g.index].(*wxkjEvent); ok {
			//无懈可击
			target = []data.PID{arg[0].(data.PID)}
			g.useTmpCard(user.pid, data.WXKJ, srcCard.getDecor(), srcCard.getNum(), data.ConvertCard, target...)
		}
		g.sendCard2Player(user.pid, g.getCardsFromTop(1)...)
		num := &user.findSkill(data.WuKuSkill).(*wuKuSkill).count
		(*num)--
		g.useSkill(user.pid, data.WuKuSkill, byte(*num))
		s.hasUsed = true
		return true
	case 2:
		if len(inf.args) != 2 || len(inf.cards) != 1 || inf.args[1] > byte(data.Lightning) {
			return false
		}
		srcCard := g.cards[inf.cards[0]]
		tmpCard := newCard(data.NewCard(data.CardName(inf.args[1]), srcCard.getDecor(), srcCard.getNum()))
		avableTarget, tNum := tmpCard.getAvailableTarget(g, user.pid)
		g.clients[user.pid].SendUseSkillRsp(data.UseSkillRsp{ID: data.MieWuSkill, Targets: avableTarget, Args: []byte{tNum}})
		return false
	case 3:
		//借刀杀人专用
		if len(inf.args) != 1 || len(inf.targets) == 0 {
			return false
		}
		target := inf.targets[0]
		slot := g.players[target].equipSlot[data.WeaponSlot]
		if slot == nil || slot.getID() == 0 {
			return false
		}
		plist := g.getPlayerInDst(target, slot.(*weaponCard).dst)
		g.clients[user.pid].SendUseSkillRsp(data.UseSkillRsp{ID: data.MieWuSkill, Targets: plist})
		return false
	}
	return false
}

func (s *miewuSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
}

func (s *miewuSkill) isUseAble(g *Games, user *player) bool {
	if !s.gaint {
		return false
	}
	if user.findSkill(data.WuKuSkill).(*wuKuSkill).count == 0 {
		return false
	}
	if s.hasUsed {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*luanWuEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*wxkjEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*wjqfEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*nmrqEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*duelEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*dyingEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*jdsrEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*dodgeEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*ctuAtkEvent); ok {
		return true
	}
	return false
}

// 武库
type wuKuSkill struct {
	skill
	count uint8
}

func (s *wuKuSkill) check(g *Games, user data.PID) {
	if s.count < 3 {
		s.count++
		g.useSkill(user, data.WuKuSkill, byte(s.count))
	}
}

// 三陈
type sanChenSkill struct {
	skill
	enable bool //是否觉醒获得灭吴
}

func (s *sanChenSkill) check(g *Games, user data.PID) {
	if s.enable {
		return
	}
	if g.players[user].findSkill(data.WuKuSkill).(*wuKuSkill).count == 3 {
		g.addHpMax(user, 1)
		g.recover(user, user, 1)
		g.players[user].findSkill(data.MieWuSkill).(*miewuSkill).gaint = true
		g.useSkill(user, data.SanChenSkill, byte(user))
		s.enable = true
	}
}

// 成略
type chengLueSkill struct {
	skill
	enable  bool         //false 为阳
	decList []data.Decor //生效花色
	hasUsed bool
}

func (s *chengLueSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newChengLueEvent(user.pid, s.enable))
	s.enable = !s.enable
	s.hasUsed = true
	g.useSkill(user.pid, data.ChengLueSkill)
	return true
}

func (s *chengLueSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

func (s *chengLueSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
	s.decList = nil
}

func (s *chengLueSkill) check(c cardI) bool {
	return isItemInList(s.decList, c.getDecor())
}

// 恃才
type shiCaiSkill struct {
	skill
	useType []data.CardType
}

func (s *shiCaiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	card := arg[0].(cardI)
	if !isItemInList([]data.CardType{data.BaseCardType, data.TipsCardType}, card.getType()) {
		g.removePlayercard(user.pid, card.getID())
		g.mainHeap = append([]data.CID{card.getID()}, g.mainHeap...)
		g.sendCard2Player(user.pid, g.getCardsFromBottom(1)...)
		g.useSkill(user.pid, data.AddHeapNum, byte(1))
		return false
	}
	for i := len(g.dropHeap) - 1; i >= 0; i-- {
		if g.dropHeap[i] == card.getID() {
			g.dropHeap = append(g.dropHeap[:i], g.dropHeap[i+1:]...)
			break
		}
	}
	g.mainHeap = append([]data.CID{card.getID()}, g.mainHeap...)
	g.sendCard2Player(user.pid, g.getCardsFromBottom(1)...)
	g.useSkill(user.pid, data.AddHeapNum, byte(1))
	return false
}

func (s *shiCaiSkill) check(g *Games, c cardI, user data.PID, target []data.PID) {
	if c.getType() == data.DealyTipsCardType {
		return
	}
	if c.getID() == 0 {
		return
	}
	if c.getName() == data.TSLH && len(target) == 0 {
		return
	}
	if !isItemInList(s.useType, c.getType()) {
		g.events.insert(g.index, newSkillSelectEvent(data.ShiCaiSkill, user, nil, c))
		g.useSkill(user, data.ShiCaiSkill, byte(c.getType()))
		if !isItemInList([]data.CardType{data.BaseCardType, data.TipsCardType}, c.getType()) {
			s.useType = append(s.useType,
				data.WeaponCardType, data.ArmorCardType, data.HorseDownCardType, data.HorseUpCardType)
		} else {
			s.useType = append(s.useType, c.getType())
		}
	}
}

func (s *shiCaiSkill) handleTurnEnd(g *Games) {
	s.useType = nil
}

// 凤魄
type fengPoSkill struct {
	skill
	hasUsed bool
	enable  bool
	conunt  uint8
}

func (s *fengPoSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newFengPoEvent(user.pid, inf.targets[0], s))
	s.hasUsed = true
	s.enable = true
	return false
}

func (s *fengPoSkill) check() bool {
	return !s.hasUsed
}

func (s *fengPoSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
	s.conunt = 0
	s.enable = false
}

// 武继
type wuJiSkill struct {
	skill
	count   uint8
	hasUsed bool
}

func (s *wuJiSkill) handleTurnEnd(g *Games) {
	if !g.players[g.turnOwner].hasEffect(wujiEffect) {
		return
	}
	p := g.players[g.turnOwner]
	if s.count >= 3 {
		g.useSkill(g.turnOwner, data.WuJiSkill)
		g.addHpMax(g.turnOwner, 1)
		g.recover(g.turnOwner, g.turnOwner, 1)
		s.hasUsed = true
		p.removeSkill(data.HuXiaoSkill)
	}
	s.count = 0
}

func (s *wuJiSkill) check() {
	s.count++
}

// 雪恨
type xueHenSkill struct {
	skill
	hasUsed bool
}

func (s *xueHenSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		cards := []data.CID{}
		for _, c := range user.equipSlot {
			if c == nil || !c.getDecor().IsRed() {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		for _, c := range user.cards {
			if g.cards[c].getDecor().IsRed() {
				cards = append(cards, c)
			}
		}
		targets := g.getAllAliveOther(user.pid)
		arg := byte(user.maxHp - user.hp)
		rsp := data.UseSkillRsp{ID: data.XueHenSkill, Targets: targets, Cards: cards, Args: []byte{arg}}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		if len(inf.cards) != 1 {
			return false
		}
		g.useSkill(user.pid, data.XueHenSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		list := []eventI{}
		for _, p := range inf.targets {
			dmg := newDamageEvent(user.pid, p, data.NormalDmg, nil, 1)
			list = append(list, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		}
		for _, p := range inf.targets {
			list = append(list, newSkillSendCardEvent(p, 1))
		}
		g.events.insert(g.index, list...)
		s.hasUsed = true
		return true
	}
	return false
}

func (s *xueHenSkill) isUseAble(g *Games, user *player) bool {
	if user.maxHp-user.hp <= 0 {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

type quanJiSkill struct {
	skill
	count uint8
}

// useskill arg:0为权计数，1为（1）增加数量或（0）减少，2为盖的牌
func (s *quanJiSkill) check(g *Games, user data.PID, cname data.CardName, up bool, atk bool) {
	p := g.players[user]
	if up {
		if !atk {
			g.events.insert(g.index, newSkillSelectEvent(data.QuanJiSkill, user, nil))
			return
		}
		if isItemInList([]data.CardName{data.Burn, data.Duel, data.Attack, data.LightnAttack, data.FireAttack}, cname) ||
			(isItemInList([]data.CardName{data.WJQF, data.NMRQ}, cname) && g.getAlivePlayerCount() == 2) {
			g.events.insert(g.index, newSkillSelectEvent(data.QuanJiSkill, user, nil))
			return
		}
		return
	}
	card := p.delTsCardBottom(data.QuanJiSkill)
	g.dropCards(card)
	p.addCard(card)
	g.removePlayercard(user, card)
	s.count--
	g.useSkill(user, data.QuanJiSkill, byte(s.count), byte(0))
}

func (s *quanJiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.sendCard2Player(user.pid, g.getCardsFromTop(1)...)
	g.events.insert(g.index, newQuanJiEvent(user.pid))
	return false
}

type paiYiSkill struct {
	skill
	hasUsed bool
}

func (s *paiYiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	skill := user.findSkill(data.QuanJiSkill).(*quanJiSkill)
	num := min(skill.count, 7)
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		targets := g.getAllAlivePlayer()
		rsp := data.UseSkillRsp{ID: data.PaiYiSkill, Targets: targets}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		g.useSkill(user.pid, data.PaiYiSkill)
		user.findSkill(data.QuanJiSkill).(*quanJiSkill).check(g, user.pid, data.NoName, false, false)
		t := inf.targets[0]
		g.sendCard2Player(t, g.getCards(int(num), t)...)
		if len(g.players[t].cards) > len(user.cards) {
			dmg := newDamageEvent(user.pid, t, data.NormalDmg, nil, 1)
			g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		}
		s.hasUsed = true
		return true
	}
	return false
}

func (s *paiYiSkill) isUseAble(g *Games, user *player) bool {
	if user.findSkill(data.QuanJiSkill).(*quanJiSkill).count == 0 {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

func (s *paiYiSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
}

type longyinSkill struct {
	skill
}

func (s *longyinSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newLongYinEvent(user.pid, arg[0].(data.PID), arg[1].(data.CID)))
	return false
}

type xionHuoSkill struct {
	skill
	user   data.PID
	num    uint8
	target []data.PID //凶祸1不能对你出杀
}

func (s *xionHuoSkill) init(g *Games, p *player) {
	s.user = p.pid
	s.num = 3
	g.useSkill(s.user, data.XionHuoSkill, s.num)
}

func (s *xionHuoSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		plist := []data.PID{}
		for _, p := range g.getAllAliveOther(user.pid) {
			if !g.players[p].hasEffect(baoliEffect) {
				plist = append(plist, p)
			}
		}
		rsp := data.UseSkillRsp{ID: data.XionHuoSkill, Targets: plist}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		g.players[inf.targets[0]].enableEffect(baoliEffect, data.XionHuoSkill)
		g.useSkill(inf.targets[0], data.XionHuoSkill, 1)
		s.num--
		g.useSkill(s.user, data.XionHuoSkill, s.num)
		g.useSkill(s.user, data.XionHuoSkill)
		return true
	}
	return false
}

func (s *xionHuoSkill) check(g *Games, target data.PID) {
	if target == s.user {
		return
	}
	//不是自己执行一项
	n := rand.IntN(3)
	// n := 2
	t := g.players[target]
	switch n {
	case 0:
		dmg := newDamageEvent(target, target, data.FireDmg, nil, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		s.target = []data.PID{target}
	case 1:
		dmg := newDamageEvent(target, target, data.BleedingDmg, nil, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		t.upDropNum++
	case 2:
		cards := []data.CID{}
		num := rand.IntN(len(t.cards))
		cards = append(cards, t.cards[num])
		equipList := []data.CID{}
		for _, c := range t.equipSlot {
			if c != nil && c.getID() != 0 {
				equipList = append(equipList, c.getID())
			}
		}
		if len(equipList) > 0 {
			cards = append(cards, equipList[rand.IntN(len(equipList))])
		}
		g.moveCard(target, s.user, cards...)
	}
	t.disableEffect(baoliEffect, data.XionHuoSkill)
	g.useSkill(target, data.XionHuoSkill, 0)
}

func (s *xionHuoSkill) addNum(g *Games) {
	if s.num < 3 {
		s.num++
		g.useSkill(s.user, data.XionHuoSkill, s.num)
	}
}

func (s *xionHuoSkill) handleTurnEnd(g *Games) {
	s.target = nil
}

func (s *xionHuoSkill) isUseAble(g *Games, user *player) bool {
	if user.findSkill(data.XionHuoSkill).(*xionHuoSkill).num == 0 {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return true
	}
	return false
}

type quediSkill struct {
	skill
	hasUsed bool
	addDmg  bool
}

func (s *quediSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newQueDiEvent(user.pid, arg[0].(data.PID)))
	s.hasUsed = true
	return false
}

func (s *quediSkill) handleTurnEnd(g *Games) {
	s.hasUsed = false
	s.addDmg = false
}

func (s *quediSkill) useAddDmg() bool {
	if s.addDmg {
		s.addDmg = false
		return true
	}
	return false
}

func (s *quediSkill) check(g *Games, user data.PID, card cardI, targets []data.PID) {
	if !isItemInList([]data.CardName{data.Duel, data.Attack, data.FireAttack, data.LightnAttack},
		card.getName()) {
		return
	}
	if len(targets) > 1 || s.hasUsed {
		return
	}
	var useAble bool
	if len(g.players[targets[0]].cards) != 0 {
		useAble = true
	}
	for _, c := range g.players[user].cards {
		if g.cards[c].getType() == data.BaseCardType {
			useAble = true
			break
		}
	}
	if !useAble {
		return
	}
	g.events.insert(g.index, newSkillSelectEvent(data.QueDiSkill, user, nil, targets[0]))
}

type zhuiFengSkill struct {
	skill
	count  uint8
	enable bool
}

func (s *zhuiFengSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		rsp := data.UseSkillRsp{ID: data.ZhuiFengSkill, Targets: g.getAllAliveOther(user.pid)}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		g.useSkill(user.pid, data.ZhuiFengSkill)
		tmpcard := newCard(data.Card{CardType: data.TipsCardType, Name: data.Duel})
		tmpcard.use(g, user.pid, inf.targets...)
		*(arg[0].(*cardI)) = tmpcard
		*(arg[1].(*[]data.PID)) = inf.targets
		s.count++
		s.enable = true
		g.useTmpCard(user.pid, data.Duel, data.NoCol, 0, data.VirtualCard, inf.targets...)
		dmg := newDamageEvent(user.pid, user.pid, data.BleedingDmg, nil, 1)
		g.addPriorityEvent(dmg)
		return true
	}
	return false
}

func (s *zhuiFengSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		if s.count < 2 {
			return true
		}
	}
	return false
}

type jiZhiSkill struct {
	skill
}

func (s *jiZhiSkill) check(useCard cardI, targets []data.PID) bool {
	if useCard.getType() != data.TipsCardType {
		return false
	}
	if useCard.getName() == data.TSLH && len(targets) == 0 {
		return false
	}
	return true
}

type jiQiaoSkill struct {
	skill
}

func (s *jiQiaoSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	g.events.insert(g.index, newJiQiaoEvent(user.pid, inf.targets[0]))
	g.events.insert(g.index, newDropSelfHandCardEvent(user.pid, 2))
	return false
}

func (s *jiQiaoSkill) check(g *Games, user data.PID) {
	if len(g.players[user].cards) < 2 {
		return
	}
	g.events.insert(g.index, newSkillSelectEvent(data.JiQiaoSkill, user, nil))
}

func (s *jiQiaoSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetList: g.getAllAlivePlayer(), TargetNum: 1}
}

type keJiSkill struct {
	skill
	useAtk bool
	count  uint8
}

func (s *keJiSkill) handleTurnEnd(*Games) {
	s.useAtk = false
	s.count = 0
}

func (s *keJiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	var result data.CID
	s.count++
	g.events.insert(g.index, newSkillJudgeEvent(user.pid, data.KeJiSkill, &result), newKeJiEvent(user.pid, &result))
	return false
}

type luanWuSkill struct {
	skill
	hasUsed bool
}

func (s *luanWuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	s.hasUsed = true
	for i := g.getPrvPid(user.pid); i != user.pid; i = g.getPrvPid(i) {
		n := distence(20)
		var plist []data.PID
		for _, p := range g.getAllAliveOther(i) {
			//空城
			if g.players[p].hasEffect(kongChengEffect) && len(g.players[p].cards) == 0 {
				continue
			}
			//凶祸
			if g.players[p].hasEffect(xionhuoEffect) {
				skill := g.players[p].findSkill(data.XionHuoSkill).(*xionHuoSkill)
				if skill.target != nil {
					if skill.target[0] == i {
						continue
					}
				}
			}
			dst := g.getDst(i, p)
			if dst < n {
				plist = []data.PID{p}
				n = dst
			} else if dst == n {
				plist = append(plist, p)
			} else {
				continue
			}
		}
		g.events.insert(g.index, newLuanWuEvent(i, plist))
	}
	return true
}

func (s *luanWuSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

func (s *luanWuSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

type zhiHengSkill struct {
	skill
	hasUsed bool
}

func (s *zhiHengSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		cards := []data.CID{}
		for _, c := range user.equipSlot {
			if c == nil {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		cards = append(cards, user.cards...)
		rsp := data.UseSkillRsp{ID: data.ZhiHengSkill, Cards: cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		num := len(inf.cards)
		g.useSkill(user.pid, data.ZhiHengSkill)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		g.sendCard2Player(user.pid, g.getCardsFromTop(num)...)
		s.hasUsed = true
		return true
	}
	return false
}

func (s *zhiHengSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

func (s *zhiHengSkill) check(g *Games) (num int) {
	for _, p := range g.getAllAlivePlayer() {
		if g.players[p].side == data.Wu {
			num = rand.IntN(3)
		}
	}
	return
}

func (s *zhiHengSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

type jianyingSkill struct {
	skill
	dec     data.Decor
	num     data.CNum
	enable  bool
	hasUsed bool
}

func (s *jianyingSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌类型,为1时为发送使用技能请求,为2时为请求可用目标
	if len(inf.args) < 1 {
		return false
	}
	switch inf.args[0] {
	case 0:
		nameList := []byte{}
		for cname := data.Attack; cname <= data.Peach; cname++ {
			tmpCard := newCard(data.NewCard(cname, data.NoDec, 0))
			if tmpCard.useAble(g, user.pid) {
				nameList = append(nameList, byte(cname))
			}
		}
		cards := []data.CID{}
		for _, c := range user.equipSlot {
			if c == nil || isItemInList(user.unUseableCol, c.getDecor()) {
				cards = append(cards, 0)
				continue
			}
			cards = append(cards, c.getID())
		}
		for _, c := range user.cards {
			if isItemInList(user.unUseableCol, g.cards[c].getDecor()) {
				continue
			}
			cards = append(cards, c)
		}
		g.clients[user.pid].SendUseSkillRsp(data.UseSkillRsp{ID: data.JianYingSkill, Cards: cards, Args: nameList})
		return false
	case 1:
		if len(inf.args) < 2 || len(inf.cards) != 1 || inf.args[1] > byte(data.Peach) {
			return false
		}
		srcCard := g.cards[inf.cards[0]]
		var tmpCard cardI
		g.useSkill(user.pid, data.JianYingSkill)
		g.dropCards(inf.cards[0])
		g.removePlayercard(user.pid, inf.cards[0])
		target := inf.targets
		var dec data.Decor
		if s.enable {
			dec = s.dec
		} else {
			dec = srcCard.getDecor()
		}
		tmpCard = newCard(data.NewCard(data.CardName(inf.args[1]), dec, srcCard.getNum()))
		tmpCard.use(g, user.pid, inf.targets...)
		*(arg[0].(*cardI)) = tmpCard
		*(arg[1].(*[]data.PID)) = inf.targets
		g.useTmpCard(user.pid, data.CardName(inf.args[1]), dec, srcCard.getNum(), data.ConvertCard, target...)
		s.hasUsed = true
		return true
	case 2:
		if len(inf.args) != 2 || len(inf.cards) != 1 || inf.args[1] > byte(data.LightnAttack) {
			return false
		}
		srcCard := g.cards[inf.cards[0]]
		tmpCard := newCard(data.NewCard(data.CardName(inf.args[1]), srcCard.getDecor(), srcCard.getNum()))
		avableTarget, tNum := tmpCard.getAvailableTarget(g, user.pid)
		g.clients[user.pid].SendUseSkillRsp(data.UseSkillRsp{ID: data.JianYingSkill, Targets: avableTarget, Args: []byte{tNum}})
		return false
	}
	return false
}

func (s *jianyingSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return !s.hasUsed
	}
	return false
}

func (s *jianyingSkill) check(g *Games, p data.PID, card cardI) {
	if !s.enable {
		s.dec = card.getDecor()
		s.num = card.getNum()
		g.useSkill(p, data.JianYingSkill, byte(s.dec), byte(s.num))
		s.enable = true
		return
	}
	if card.getDecor() == s.dec || card.getNum() == s.num {
		g.useSkill(p, data.JianYingSkill)
		g.sendCard2Player(p, g.getCardsFromTop(1)...)
	}
	s.dec = card.getDecor()
	s.num = card.getNum()
	g.useSkill(p, data.JianYingSkill, byte(s.dec), byte(s.num))
}

func (s *jianyingSkill) handleTurnEnd(*Games) {
	s.dec = data.NoDec
	s.num = 0
	s.enable = false
	s.hasUsed = false
}

type shiBeiSkill struct {
	skill
	enable bool
}

func (s *shiBeiSkill) check(g *Games, p data.PID) {
	g.useSkill(p, data.ShiBeiSkill)
	if !s.enable {
		s.enable = true
		g.recover(p, p, 1)
		return
	}
	dmg := newDamageEvent(p, p, data.BleedingDmg, nil, 1)
	g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
}

func (s *shiBeiSkill) handleTurnEnd(*Games) {
	s.enable = false
}

type qiaoShiSkill struct {
	skill
	hasUsed bool
}

func (s *qiaoShiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	s.hasUsed = true
	g.sendCard2Player(user.pid, g.getCards(2, user.pid)...)
	g.useSkill(inf.targets[0], data.QiaoShiSkill)
	g.recover(user.pid, inf.targets[0], arg[0].(data.HP))
	return false
}

func (s *qiaoShiSkill) handleTurnEnd(*Games) {
	s.hasUsed = false
}

type yanYuSkill struct {
	skill
	count uint8
}

func (s *yanYuSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	//inf.Arg[0]为0时为请求可用卡牌与可用目标,为1时为发送使用技能请求
	//arg【0】为0表示出牌结束触发
	if len(inf.args) < 1 {
		if len(arg) < 1 {
			return false
		}
		g.sendCard2Player(inf.targets[0], g.getCards(3*int(s.count), inf.targets[0])...)
		return false
	}
	switch inf.args[0] {
	case 0:
		cards := []data.CID{}
		for _, c := range user.cards {
			if isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack}, g.cards[c].getName()) {
				cards = append(cards, c)
			}
		}
		rsp := data.UseSkillRsp{ID: data.YanYuSkill, Cards: cards}
		g.clients[user.pid].SendUseSkillRsp(rsp)
		return false
	case 1:
		s.count++
		g.useSkill(user.pid, data.YanYuSkill, s.count)
		g.dropCards(inf.cards...)
		g.removePlayercard(user.pid, inf.cards...)
		g.sendCard2Player(user.pid, g.getCardsFromTop(1)...)
		return true
	}
	return false
}

func (s *yanYuSkill) isUseAble(g *Games, user *player) bool {
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return s.count < 2
	}
	return false
}

func (s *yanYuSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	return data.AvailableTargetInf{TargetNum: 1, TargetList: g.getAllAliveOther(user)}
}

func (s *yanYuSkill) handleTurnEnd(*Games) {
	s.count = 0
}

type lueYingSkill struct {
	skill
	count   uint8
	hasUsed uint8
}

func (s *lueYingSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	if arg[0] == 0 {
		tmpCard := newCard(data.Card{CardType: data.TipsCardType, Name: data.GHCQ})
		tmpCard.use(g, user.pid, inf.targets[0])
		g.useTmpCard(user.pid, data.GHCQ, data.NoDec, 0, data.VirtualCard, inf.targets[0])
	} else {
		hasAtk := user.hasAttack
		tmpCard := newCard(data.Card{CardType: data.BaseCardType, Name: data.Attack})
		tmpCard.use(g, user.pid, inf.targets[0])
		user.hasAttack = hasAtk
		g.useTmpCard(user.pid, data.Attack, data.NoDec, 0, data.VirtualCard, inf.targets[0])
	}
	return false
}

func (s *lueYingSkill) addCount(g *Games, user data.PID) {
	if s.hasUsed == 2 || g.turnOwner != user {
		return
	}
	s.count++
	g.useSkill(user, data.LueYingSkill, s.count)
	s.hasUsed++
}

func (s *lueYingSkill) check(g *Games, user data.PID, cname data.CardName) {
	if s.count < 2 {
		return
	}
	s.count -= 2
	g.useSkill(user, data.LueYingSkill, s.count)
	g.sendCard2Player(user, g.getCardsFromTop(1)...)
	if isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack}, cname) {
		//arg 0为使用过河拆桥
		g.events.insert(g.index, newSkillSelectEvent(data.LueYingSkill, user, nil, 0))
	} else {
		g.events.insert(g.index, newSkillSelectEvent(data.LueYingSkill, user, nil, 1))

	}
}

func (s *lueYingSkill) handleTurnEnd(*Games) {
	s.hasUsed = 0
}

func (s *lueYingSkill) getAvailableTarget(g *Games, user data.PID, arg ...any) data.AvailableTargetInf {
	inf := data.AvailableTargetInf{TargetNum: 1}
	if arg[0] == 0 {
		inf.TargetList = g.getAllAliveOther(user)
	} else {
		inf.TargetList = g.getPlayerInDst(user, g.players[user].getAtkDst())
	}
	return inf
}

type yingWuSkill struct {
	skill
	hasUsed uint8
}

func (s *yingWuSkill) addCount(g *Games, user data.PID) {
	if s.hasUsed == 2 || g.turnOwner != user {
		return
	}
	skill := g.players[user].findSkill(data.LueYingSkill).(*lueYingSkill)
	skill.count++
	g.useSkill(user, data.LueYingSkill, skill.count)
	s.hasUsed++
}

func (s *yingWuSkill) handleTurnEnd(*Games) {
	s.hasUsed = 0
}

type shouXiSkill struct {
	skill
	count int
	user  data.PID
	sid   data.SID
}

func (s *shouXiSkill) init(g *Games, p *player) {
	s.user = p.pid
}

func (s *shouXiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	num := rand.IntN(len(data.AtkSkillList))
	s.sid = data.AtkSkillList[num]
	user.addSkill(data.AtkSkillList[num])
	//arg[0]为0为修改兽盾，为1为获得技能:
	g.useSkill(user.pid, data.ShouXiSkill, 1, byte(data.AtkSkillList[num]))
	return false
}

func (s *shouXiSkill) handleTurnEnd(g *Games) {
	if g.turnOwner != s.user {
		return
	}
	g.players[s.user].removeSkill(s.sid)
	g.useSkill(s.user, data.ShouXiSkill, 1, 0)
	count := g.players[s.user].getEffectCount(shoudunEffect)
	num := min(s.count, 2)
	num = min(num, 2-int(count))
	for i := 0; i < num; i++ {
		g.players[s.user].enableEffect(shoudunEffect, data.ShouXiSkill)
	}
	g.useSkill(s.user, data.ShouXiSkill, 0, g.players[s.user].getEffectCount(shoudunEffect))
	s.count = 0
}

type moJiaSkill struct {
	skill
}

func (s *moJiaSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	t := g.players[arg[0].(data.PID)]
	if user.hp < 30 {
		if !user.hasEffect(baoliEffect) {
			user.addSkill(data.BaoLianSkill)
		}
	}
	if user.hp < 20 {
		if !user.hasEffect(fankuiEffect) {
			user.addSkill(data.FanKuiSkill)
		}
	}
	if user.hp < 15 {
		for _, c := range user.judgeSlot {
			if c != nil && c.getID() != 0 {
				g.dropCards(c.getID())
				g.removePlayercard(user.pid, c.getID())
			}
		}
	}
	if user.hp < 10 {
		if !t.turnBack {
			t.turnBack = true
			g.useSkill(t.pid, data.TurnBackSkill)
		}
	}
	if user.hp < 5 {
		if !user.hasEffect(danshueffect) {
			user.addSkill(data.DanShuSkill)
		}
	}
	return false
}

// 龙旋
type longXuanSkill struct {
	skill
}

func (s *longXuanSkill) check(p *player) (n int) {
	for _, c := range p.judgeSlot {
		if c == nil {
			continue
		}
		n += 2
	}
	return
}

// 烈袭
type liexiSkill struct {
	skill
}

func (s *liexiSkill) use(g *Games, user *player, inf useSkillInf, arg ...any) bool {
	for _, c := range user.cards {
		if g.cards[c].getDecor() == data.ClubDec {
			j := rand.IntN(g.getAlivePlayerCount())
			dmg := newDamageEvent(user.pid, g.getAllAliveOther(user.pid)[j], data.NormalDmg, nil, 1)
			g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		}
	}
	return false
}

// 青龙
type qingLongSkill struct {
	skill
	unable bool
}

func (s *qingLongSkill) check(user data.PID, cname data.CardName, target []data.PID) {
	if isItemInList([]data.CardName{data.WZSY, data.Peach, data.Drunk, data.Lightning}, cname) {
		return
	} else if isItemInList([]data.CardName{data.Burn, data.TSLH}) {
		if target == nil || len(target) == 0 {
			return
		} else if len(target) > 1 {
			s.unable = true
		} else if target[0] != user {
			s.unable = true
		}
	} else {
		s.unable = true
	}
}

func (s *qingLongSkill) handleTurnEnd(*Games) {
	s.unable = false
}

func newSkill(id data.SID) skillI {
	s := skill{id: id}
	switch id {
	case data.EmptySkill:
		return &s
	case data.BGZSkill:
		return &bgzSkill{skill: s}
	case data.QLYYDSkill:
		return &qlyydSkill{skill: s}
	case data.GSFSkill:
		return &gsfSkill{skill: s}
	case data.QLGSkill:
		return &qlgSkill{skill: s}
	case data.CXSGJSkill:
		return &cxsgjSkill{skill: s}
	case data.ZQYSSkill:
		return &zqysSkill{skill: s}
	case data.HBJSkill:
		return &hbjSkill{skill: s}
	case data.ZBSMSkill:
		return &zbsmSkill{skill: s}
	case data.FenYinSkill:
		return &fenYinSkill{skill: s}
	case data.JiLiSkill:
		return &jiLiSkill{skill: s}
	case data.PoJunSkill:
		return &poJunSkill{skill: s}
	case data.TunJiangSkill:
		return &tunJiangSkill{skill: s}
	case data.LangXiSkill:
		return &langXiSkill{skill: s}
	case data.WenJiSkill:
		return &wenJiSkill{skill: s}
	case data.RenDeSkill:
		return &rendeSkill{skill: s}
	case data.WuShengSkill:
		return &wushengSkill{skill: s}
	case data.LieGongSkill:
		return &lieGongSkill{skill: s}
	case data.TuSheSkill:
		return &tuSheSkill{skill: s}
	case data.LiMuSkill:
		return &liMuSkill{skill: s}
	case data.LongDanSkill:
		return &longDanSkill{skill: s}
	case data.TieJiSkill:
		return &tieJiSkill{skill: s}
	case data.DiDongSkill:
		return &diDongSkill{skill: s}
	case data.GuiXinSkill:
		return &guiXinSkill{skill: s}
	case data.LuoLeiSkill:
		return &luoLeiSkill{skill: s}
	case data.GuiHuoSkill:
		return &guiHuoSkill{skill: s}
	case data.ZhangBaSkill:
		return &zhangbaSkill{skill: s}
	case data.ShiChouSkill:
		return &ShiChouSkill{skill: s}
	case data.QingGangSkill:
		return &qinggangSkill{skill: s}
	case data.YanYueSkill:
		return &yanyueSkill{skill: s}
	case data.QiLinSkill:
		return &qilinSkill{skill: s}
	case data.GuanXingSkill:
		return &guanxingSkill{skill: s}
	case data.YeYanSkill:
		return &yeyanSkill{skill: s}
	case data.QinYinSkill:
		return &qinYinSkill{skill: s}
	case data.ZhuQueSkill:
		return &zhuQueSkill{skill: s}
	case data.YingHunSkill:
		return &yingHunSkill{skill: s}
	case data.GuDingSkill:
		return &gudingSkill{skill: s}
	case data.WuShuangSkill:
		return &wuShuangSkill{skill: s}
	case data.FangTianSkill:
		return &fangTianSkill{skill: s}
	case data.QiZhiSkill:
		return &qiZhiSkill{skill: s}
	case data.JinQuSkill:
		return &jinQuSkill{skill: s}
	case data.LiYuSkill:
		return &LiYuSkill{skill: s}
	case data.ZhenGuSkill:
		return &zhenGuSkill{skill: s}
	case data.YingYuanSkill:
		return &yingYuanSkill{skill: s}
	case data.QianXiSkill:
		return &qianXiSkill{skill: s}
	case data.ZhuiJiSkill:
		return &zhuiJiSkill{skill: s}
	case data.BaZhenSkill:
		return &bazhenSkill{skill: s}
	case data.GuiJiSkill:
		return &guijiSkill{skill: s}
	case data.SuoMingSkill:
		return &suoMinfSkill{skill: s}
	case data.XiXingSkill:
		return &xiXingSkill{skill: s}
	case data.ManJiaSkill:
		return &manjiaSkill{skill: s}
	case data.XiaoShouSkill:
		return &xiaoShouSkill{skill: s}
	case data.LianYuSkill:
		return &lianYuSkill{skill: s}
	case data.QiangZhengSkill:
		return &qiangZhengSkill{skill: s}
	case data.NiePanSkill:
		return &niePanSkill{skill: s}
	case data.MoJianSkill:
		return &mojianSkill{skill: s}
	case data.YuShouSkill:
		return &yushouSkill{skill: s}
	case data.NewBaZhenSkill:
		return &newbazhenSkill{bazhenSkill: bazhenSkill{skill: s}}
	case data.JueCeSkill:
		return &jueCeSkill{skill: s}
	case data.FanKuiSkill:
		return &fanKuiSkill{skill: s}
	case data.MoYanSkill:
		return &moYanSkill{skill: s}
	case data.DanShuSkill:
		return &danshuSkill{skill: s}
	case data.JiJiuSkill:
		return &jijiuSkill{skill: s}
	case data.QingNangSkill:
		return &qingnangSkill{skill: s}
	case data.JieYinSkill:
		return &jieyinSkill{skill: s}
	case data.LiangZhuSkill:
		return &liangzhuSkill{skill: s}
	case data.ChengLueSkill:
		return &chengLueSkill{skill: s}
	case data.ShiCaiSkill:
		return &shiCaiSkill{skill: s}
	case data.WuKuSkill:
		return &wuKuSkill{skill: s}
	case data.SanChenSkill:
		return &sanChenSkill{skill: s}
	case data.MieWuSkill:
		return &miewuSkill{skill: s}
	case data.FengPoSkill:
		return &fengPoSkill{skill: s}
	case data.ZiShuSkill:
		return &zishuSkill{skill: s}
	case data.XueHenSkill:
		return &xueHenSkill{skill: s}
	case data.WuJiSkill:
		return &wuJiSkill{skill: s}
	case data.QuanJiSkill:
		return &quanJiSkill{skill: s}
	case data.PaiYiSkill:
		return &paiYiSkill{skill: s}
	case data.LongYinSkill:
		return &longyinSkill{skill: s}
	case data.XionHuoSkill:
		return &xionHuoSkill{skill: s}
	case data.YiSuanSkill:
		return &yisuanSkill{skill: s}
	case data.QueDiSkill:
		return &quediSkill{skill: s}
	case data.ZhuiFengSkill:
		return &zhuiFengSkill{skill: s}
	case data.JiZhiSkill:
		return &jiZhiSkill{skill: s}
	case data.JiQiaoSkill:
		return &jiQiaoSkill{skill: s}
	case data.KeJiSkill:
		return &keJiSkill{skill: s}
	case data.JieYingSkill:
		return &jieyingSkill{skill: s}
	case data.LuanWuSkill:
		return &luanWuSkill{skill: s}
	case data.ZhiHengSkill:
		return &zhiHengSkill{skill: s}
	case data.JianYingSkill:
		return &jianyingSkill{skill: s}
	case data.ShiBeiSkill:
		return &shiBeiSkill{skill: s}
	case data.QiaoShiSkill:
		return &qiaoShiSkill{skill: s}
	case data.YanYuSkill:
		return &yanYuSkill{skill: s}
	case data.LueYingSkill:
		return &lueYingSkill{skill: s}
	case data.YingWuSkill:
		return &yingWuSkill{skill: s}
	case data.ShouXiSkill:
		return &shouXiSkill{skill: s}
	case data.MoJiaSkill:
		return &moJiaSkill{skill: s}
	case data.LongXuanSkill:
		return &longXuanSkill{skill: s}
	case data.LiexiSkill:
		return &liexiSkill{skill: s}
	case data.QinglongSkill:
		return &qingLongSkill{skill: s}
	}
	panic("技能id=" + id.String() + "的技能不存在")
}

func newSkillList() []skillI {
	list := make([]skillI, data.SkillListEndPos)
	for sid := data.EmptySkill; sid < data.SkillListEndPos; sid++ {
		list[sid] = newSkill(sid)
	}
	return list
}
