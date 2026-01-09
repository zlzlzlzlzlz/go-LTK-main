package server

import (
	"goltk/data"
	"math/rand"
	"time"
)

type prepareEvent struct {
	event
	user data.PID
}

func newPrePareEvent(user data.PID) *prepareEvent {
	return &prepareEvent{user: user}
}

func (e *prepareEvent) trigger(g *Games) {
	const waitTime = time.Millisecond * 200 //等待时间200ms
	g.turnOwner = e.user
	g.setGameState(data.PrepareState, waitTime, e.user)
	iterators(g.clients, func(c clientI) { c.SetTurnOwner(e.user) })
	p := g.players[e.user]
	//魔道
	if p.hasEffect(modaoEffect) {
		g.sendCard2Player(e.user, g.getCardsFromTop(2)...)
	}
	//狼袭
	if p.hasEffect(langXiEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.LangXiSkill, e.user, nil))
	}
	//英魂
	if p.hasEffect(yingHunEffect) && p.hp < p.maxHp {
		g.events.insert(g.index, newSkillSelectEvent(data.YingHunSkill, e.user, nil))
	}
	//观星
	if p.hasEffect(guanXingEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.GuanXingSkill, e.user, nil))
	}
	//诡计
	if p.hasEffect(guijiEffect) {
		p.findSkill(data.GuiJiSkill).(*guijiSkill).check(g, e.user)
	}
	//吸星
	if p.hasEffect(xixingEffect) {
		p.findSkill(data.XiXingSkill).(*xiXingSkill).check(g, e.user)
	}
	//机巧
	if p.hasEffect(jiqiaoEffect) {
		p.findSkill(data.JiQiaoSkill).(*jiQiaoSkill).check(g, e.user)
	}
	//兽袭
	if p.hasEffect(shouxiEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.ShouXiSkill, e.user, nil))
	}
	//龙旋
	if p.hasEffect(longxuanEffect) {
		skill := p.findSkill(data.LongXuanSkill).(*longXuanSkill)
		g.sendCard2Player(e.user, g.getCards(skill.check(g.players[e.user]), e.user)...)
	}
	//神兽
	if p.hasEffect(shenshouEffect) {
		g.sendCard2Player(e.user, g.getCardsFromTop(2)...)
		p.enableEffect(extraAtkEffect)
		g.useSkill(e.user, data.ShenShouSkill)
	}
	<-time.After(waitTime)
}

type judgeEvent struct {
	event
	target data.PID
}

func newJudgeEvent(target data.PID) *judgeEvent {
	return &judgeEvent{target: target}
}

func (e *judgeEvent) trigger(g *Games) {
	const waitTime = time.Millisecond * 200 //等待时间200ms
	g.setGameState(data.JudgedState, waitTime, e.target)
	p := g.players[e.target]
	if p.judgeSlot[data.LBSSSlot] != nil {
		var result data.CID
		iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDPerJudge, data.CID(data.LBSSSlot)) })
		judge := newCardJudgeEvent(&result, e.target)
		lbss := newLBSSEvent(&result)
		card := p.judgeSlot[data.LBSSSlot].(*lbssCard)
		g.events.insert(g.index, newWXKJEvent(card, card.user, e.target, judge, lbss), judge, lbss,
			newDealyTipsEndEvent(e.target, data.LBSSSlot), newJudgeEvent(e.target))
		return
	}
	if p.judgeSlot[data.BLCDSlot] != nil {
		var result data.CID
		iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDPerJudge, data.CID(data.BLCDSlot)) })
		judge := newCardJudgeEvent(&result, e.target)
		blcd := newBLCDEvent(&result)
		card := p.judgeSlot[data.BLCDSlot].(*blcdCard)
		g.events.insert(g.index, newWXKJEvent(card, card.user, e.target, judge, blcd), judge, blcd,
			newDealyTipsEndEvent(e.target, data.BLCDSlot), newJudgeEvent(e.target))
		return
	}
	if p.judgeSlot[data.LightningSlot] != nil {
		var result data.CID
		iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDPerJudge, data.CID(data.LightningSlot)) })
		judge := newCardJudgeEvent(&result, e.target)
		end := newDealyTipsEndEvent(e.target, data.LightningSlot)
		card := p.judgeSlot[data.LightningSlot].(*lightnCard)
		linghtning := newLightningEvent(&result, e.target, end, card)
		g.events.insert(g.index, newWXKJEvent(card, card.user, e.target, judge, linghtning),
			judge, linghtning, end)
		return
	}
}

type sendCardEvent struct {
	event
	user data.PID
}

func newSendCardEvent(target data.PID) *sendCardEvent {
	return &sendCardEvent{user: target}
}

func (e *sendCardEvent) trigger(g *Games) {
	<-time.After(200 * time.Millisecond)
	cards := g.getCards(2, e.user)
	if g.cards[cards[0]].getID() == g.cards[cards[1]].getID() {
		panic("")
	}
	p := g.players[e.user]
	//朱雀
	if p.hasEffect(zhuQueEffect) {
		cards = append(cards, g.getCardsFromTop(1)...)
		g.useSkill(e.user, data.ZhuQueSkill)
	}
	//独进
	if p.hasEffect(duJinEffect) {
		count := 0
		for _, c := range p.equipSlot {
			if c != nil {
				count++
			}
		}
		cards = append(cards, g.getCardsFromTop(count/2+1)...)
		g.useSkill(e.user, data.DuJinskill)
	}
	//太平
	if p.hasEffect(taipingEffect) {
		cards = append(cards, g.getCardsFromTop(2)...)
		g.useSkill(e.user, data.TaiPingSkill)
	}
	//制衡
	if p.hasEffect(tongyeEffect) {
		num := p.findSkill(data.ZhiHengSkill).(*zhiHengSkill).check(g)
		cards = append(cards, g.getCardsFromTop(num)...)
		g.useSkill(e.user, data.TongYeSkill)
	}
	g.sendCard2Player(e.user, cards...)
}

type useCardEvent struct {
	event
	user data.PID
}

func newUseCardEvent(user data.PID) *useCardEvent {
	return &useCardEvent{user: user}
}

func (e *useCardEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	p := g.players[e.user]
	nextUseCard := newUseCardEvent(e.user)
	g.events.insert(g.index, nextUseCard)
	//魔箭
	if p.hasEffect(mojianEffect) && p.findSkill(data.MoJianSkill).(*mojianSkill).check() {
		g.events.insert(g.index, newSkillSelectEvent(data.MoJianSkill, e.user, nil))
		return
	}
	//驭兽
	if p.hasEffect(yushouEffect) && p.findSkill(data.YuShouSkill).(*yushouSkill).check() {
		g.events.insert(g.index, newSkillSelectEvent(data.YuShouSkill, e.user, nil))
		return
	}
	//凶祸
	if p.hasEffect(baoliEffect) {
		for _, p := range g.players {
			if p.hasEffect(xionhuoEffect) {
				p.findSkill(data.XionHuoSkill).(*xionHuoSkill).check(g, e.user)
				return
			}
		}
		return
	}
	g.setGameState(data.UseCardState, lastTime, e.user)
	useAbleCards := g.players[e.user].getUseAbleCards(g)
	g.clients[e.user].SendUseAbleCards(useAbleCards)
	useAbleSkill := g.players[e.user].getUseAbleSkill(g)
	g.clients[e.user].SendUseAbleSkill(useAbleSkill)
	//攻击范围
	var dst distence
	if p.equipSlot[data.WeaponSlot] != nil {
		dst = p.equipSlot[data.WeaponSlot].(*weaponCard).dst
	} else {
		dst = distence(1)
	}
	//屯江
	if p.hasEffect(tunJinagEffect) {
		p.findSkill(data.TunJiangSkill).(*tunJiangSkill).check(g, e.user, data.UseCardState, nil, nil)
	}
	var useCard cardI      //使用的卡
	var targets []data.PID //目标
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			goto useFail
		//回应玩家寻找可用目标的请求
		case c := <-g.clients[e.user].GetTargetQuest():
			list, num := g.cards[c].getAvailableTarget(g, e.user)
			g.clients[e.user].SendAvailableTarget(data.AvailableTargetInf{TargetNum: num, TargetList: list})
		case inf := <-g.clients[e.user].GetUseSkillInf():
			//回应使用技能请求
			if p.findSkill(inf.ID).use(g, p, useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args},
				&useCard, &targets) {
				goto useSuccess
			}
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				goto useFail
			}
			if !isItemInList(useAbleCards, inf.ID) { //检查玩家发送的卡牌的正确性
				continue
			}
			useCard = g.cards[inf.ID]
			targets = inf.TargetList
			if g.cards[inf.ID].getType() == data.BaseCardType || g.cards[inf.ID].getType() == data.TipsCardType {
				g.dropCards(inf.ID)
				//应援
				if p.hasEffect(yingYuanEffect) {
					p.findSkill(data.YingYuanSkill).(*yingYuanSkill).check(g, e.user, targets, useCard.getID())
				}
			}
			//恃才
			if p.hasEffect(shicaiEffect) {
				p.findSkill(data.ShiCaiSkill).(*shiCaiSkill).check(g, useCard, e.user, targets)
			}
			//亦算
			if p.hasEffect(yisuanEffect) {
				p.findSkill(data.YiSuanSkill).(*yisuanSkill).check(g, useCard.getID(), e.user, targets)
			}
			g.cards[inf.ID].use(g, e.user, inf.TargetList...)
			if useCard.getName() == data.TSLH && len(targets) == 0 {
				return
			}
			goto useSuccess
		}
	}
useSuccess:
	if useCard == nil {
		return
	}
	//青龙
	if p.hasEffect(qinglongEffect) {
		p.findSkill(data.QinglongSkill).(*qingLongSkill).check(e.user, useCard.getName(), targets)
	}
	//渐营
	if p.hasEffect(jianyingEffect) {
		p.findSkill(data.JianYingSkill).(*jianyingSkill).check(g, e.user, useCard)
	}
	//奋音技能
	if p.hasEffect(fenYinEffect) {
		p.findSkill(data.FenYinSkill).(*fenYinSkill).check(g, useCard, e.user)
	}
	if p.hasEffect(caichonEffect) {
		g.sendCard2Player(e.user, g.getCardsFromTop(2)...)
	}
	//蒺藜技能
	if p.hasEffect(jiLiEffect) {
		p.findSkill(data.JiLiSkill).(*jiLiSkill).check(g, dst, e.user)
	}
	//屯江
	if p.hasEffect(tunJinagEffect) {
		p.findSkill(data.TunJiangSkill).(*tunJiangSkill).check(g, e.user, data.UseCardState, useCard, targets)
	}
	//图射
	if p.hasEffect(tuSheEffect) {
		p.findSkill(data.TuSheSkill).(*tuSheSkill).check(g, e.user, useCard, targets)
	}
	//空城
	if p.hasEffect(kongChengEffect) && len(p.cards) == 0 {
		g.useSkill(e.user, data.KongChengSkill)
	}
	//奇制
	if p.hasEffect(qizhiEffect) && (useCard.getType() == data.BaseCardType ||
		useCard.getType() == data.TipsCardType || useCard.getType() == data.DealyTipsCardType) {
		p.findSkill(data.QiZhiSkill).(*qiZhiSkill).check(g, e.user, useCard, targets)
	}
	//集智
	if p.hasEffect(jizhiEffect) {
		if p.findSkill(data.JiZhiSkill).(*jiZhiSkill).check(useCard, targets) {
			g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
			g.useSkill(e.user, data.JiZhiSkill)
		}
	}
	if len(targets) == 0 {
		return
	}
	//无双
	if p.hasEffect(wuShuangEffect) && (useCard.getName() == data.Attack || useCard.getName() == data.Duel) {
		g.useSkill(e.user, data.WuShuangSkill)
	} else if g.players[targets[0]].hasEffect(wuShuangEffect) && useCard.getName() == data.Duel {
		g.useSkill(targets[0], data.WuShuangSkill)
	}
	//检查凤魄
	if p.hasEffect(fengpoEffect) && useCard.getName() == data.Duel {
		if p.findSkill(data.FengPoSkill).(*fengPoSkill).check() {
			g.events.insert(g.index, newSkillSelectEvent(data.FengPoSkill, e.user, []data.PID{targets[0]}))
		}
	}
	//却敌
	if p.hasEffect(quediEffect) {
		p.findSkill(data.QueDiSkill).(*quediSkill).check(g, e.user, useCard, targets)
	}
	return
useFail:
	nextUseCard.setSkip(true)
	//燕语
	if p.hasEffect(yanyuEffect) {
		if p.findSkill(data.YanYuSkill).(*yanYuSkill).count > 0 {
			g.events.insert(g.index, newSkillSelectEvent(data.YanYuSkill, e.user, nil, 0))
		}
	}
}

type dropCardEvent struct {
	event
	user data.PID
}

func newDropCardEvent(user data.PID) *dropCardEvent {
	return &dropCardEvent{user: user}
}

func (e *dropCardEvent) trigger(g *Games) {
	p := g.players[e.user]
	//克己
	if p.hasEffect(kejiEffect) {
		if !p.findSkill(data.KeJiSkill).(*keJiSkill).useAtk {
			g.useSkill(e.user, data.KeJiSkill)
			return
		}
	}
	//青龙
	if p.hasEffect(qinglongEffect) {
		if !p.findSkill(data.QinglongSkill).(*qingLongSkill).unable {
			g.useSkill(e.user, data.QinglongSkill)
			return
		}
	}
	dropNum := len(p.cards) - int(p.hp)
	cards := []data.CID{}
	//急救
	if p.hasEffect(jijiuEffect) {
		for _, c := range p.cards {
			if g.cards[c].getDecor() != data.HeartDec {
				cards = append(cards, c)
			} else {
				dropNum--
			}
		}
	} else {
		cards = append([]data.CID{}, p.cards...)
	}
	//权计
	if p.hasEffect(quanjiEffect) {
		num := min(p.findSkill(data.QuanJiSkill).(*quanJiSkill).count, 7)
		dropNum -= int(num)
	}
	dropNum += int(p.upDropNum)
	// /归心
	if p.hasEffect(guiXinEffect) {
		dropNum -= p.findSkill(data.GuiXinSkill).(*guiXinSkill).check(g)
	}
	if dropNum <= 0 {
		return
	}
	// 琴音
	if p.hasEffect(qinYineffect) && dropNum >= 2 {
		g.events.insert(g.index, newSkillSelectEvent(data.QinYinSkill, e.user, nil))
	}
	const lastTime = 20 * time.Second
	g.setGameState(data.DropCardState, lastTime, e.user)
	g.clients[e.user].SendDropAbleCard(cards, uint8(dropNum))
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			n := len(p.cards) - int(p.hp)
			g.dropCards(p.cards[:n]...)
			g.removePlayercard(e.user, p.cards[:n]...)
			return
		case cards := <-g.clients[e.user].GetDropCardInf():
			if len(cards) != dropNum || !isListContainAllTheItem(p.cards, cards...) {
				continue
			}
			g.dropCards(cards...)
			g.removePlayercard(e.user, cards...)
			return
		}
	}
}

type endEvent struct {
	event
	user data.PID
}

func newEndEvent(user data.PID) *endEvent {
	return &endEvent{user: user}
}

func (e *endEvent) trigger(g *Games) {
	const waitTime = time.Millisecond * 200 //等待时间200ms
	g.setGameState(data.EndState, waitTime, e.user)
	p := g.players[e.user]
	//三陈and灭吴
	if p.hasEffect(sanchenEffect) {
		p.findSkill(data.SanChenSkill).(*sanChenSkill).check(g, e.user)
	}
	//强征
	if p.hasEffect(qiangzhengEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.QiangZhengSkill, e.user, nil))
	}
	//屯江
	if p.hasEffect(tunJinagEffect) {
		p.findSkill(data.TunJiangSkill).(*tunJiangSkill).check(g, e.user, data.EndState, nil, nil)
	}
	//枭首
	if p.hasEffect(xiaoShouEffect) {
		for _, t := range g.getAllAliveOther(e.user) {
			if g.players[t].hp >= p.hp {
				g.events.insert(g.index, newSkillSelectEvent(data.XiaoShouSkill, e.user, nil))
				break
			}
		}
	}
	//进趋
	if p.hasEffect(jinQuEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.JinQuSkill, e.user, []data.PID{e.user}))
	}
	//镇骨
	if p.hasEffect(zhenGuEffect) && p.hasSkill(data.ZhenGuSkill) {
		g.events.insert(g.index, newSkillSelectEvent(data.ZhenGuSkill, e.user, nil))
	}
	//索命
	if p.hasEffect(suomingEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.SuoMingSkill, e.user, nil))
	}
	//暴敛
	if p.hasEffect(baolianEffect) {
		g.sendCard2Player(e.user, g.getCardsFromTop(2)...)
		g.useSkill(e.user, data.BaoLianSkill)
	}
	//神兽
	if p.hasEffect(shenshouEffect) {
		g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
		g.useSkill(e.user, data.ShenShouSkill)
	}
	//炼狱
	if p.hasEffect(lianyuEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.LianYuSkill, e.user, nil))
	}
	//机巧
	if p.hasEffect(jiqiaoEffect) {
		p.findSkill(data.JiQiaoSkill).(*jiQiaoSkill).check(g, e.user)
	}
	//烈袭
	if p.hasEffect(liexiEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.LiexiSkill, e.user, nil))
	}
	<-time.After(waitTime)
}

type go2NextEvent struct {
	event
	user data.PID
}

func newGo2NextEvent(user data.PID) *go2NextEvent {
	return &go2NextEvent{user: user}
}

func (e *go2NextEvent) trigger(g *Games) {
	//在结束阶段重置玩家回合内状态
	g.players[e.user].hasAttack = false
	g.players[e.user].hasUseDrunk = false
	g.players[e.user].isDrunk = false
	g.players[e.user].unUseableCol = []data.Decor{}
	g.players[e.user].upDropNum = 0
	g.events.list = nil
	g.events.append(g.newGameEventList(g.getNextTurnOwner(e.user))...)
	g.index = -1
	//调用技能的handleEnd
	iterators(g.players, func(p *player) { iterators(p.skills, func(s skillI) { s.handleTurnEnd(g) }) })
}

// 濒死事件
type dyingEvent struct {
	event
	dyingPlayer data.PID
	user        data.PID
}

func newDyingEvent(dyingpid data.PID, user data.PID) *dyingEvent {
	return &dyingEvent{dyingPlayer: dyingpid, user: user}
}

func (e *dyingEvent) trigger(g *Games) {
	const waitTime = time.Second * 20 //等待时间20s
	g.setGameState(data.DyingState, waitTime, e.user)
	g.clients[e.user].SendUseAbleCards(g.players[e.user].getUseAbleCards(g))
	useAbleSkill := g.players[e.user].getUseAbleSkill(g)
	g.clients[e.user].SendUseAbleSkill(useAbleSkill)
	timer := time.After(waitTime)
	var useCard cardI
	p := g.players[e.user]
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseSkillInf():
			//回应使用技能请求
			if p.findSkill(inf.ID).use(g, p, useSkillInf{cards: inf.Cards, args: inf.Args}, e.dyingPlayer) {
				goto useSuccess
			}
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				return
			}
			useCard = g.cards[inf.ID]
			g.useCard(e.user, inf.ID)
			g.players[e.user].delCard(inf.ID)
			g.dropCards(inf.ID)
			goto useSuccess
		}
	}
useSuccess:
	g.recover(e.user, e.dyingPlayer, 1)
	//蒺藜技能
	if g.players[e.user].hasEffect(jiLiEffect) {
		g.players[e.user].findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, e.user)
	}
	for i := g.index + 1; i < len(g.events.list); i++ {
		event, ok := g.events.list[i].(*dyingEvent)
		if ok {
			event.setSkip(true)
		}
	}
	if g.players[e.dyingPlayer].hp < 1 {
		list := []eventI{
			newDyingEvent(e.dyingPlayer, e.user),
		}
		for i := g.getNextPid(e.user); i != e.user; i = g.getNextPid(i) {
			list = append(list, newDyingEvent(e.dyingPlayer, i))
		}
		g.events.insert(g.index, list...)
		return
	}
	if useCard != nil {
		//恃才
		if g.players[e.user].hasEffect(shicaiEffect) {
			g.players[e.user].findSkill(data.ShiCaiSkill).(*shiCaiSkill).check(g, useCard, e.user, nil)
		}
		//图射
		if g.players[e.user].hasEffect(tuSheEffect) {
			g.players[e.user].findSkill(data.TuSheSkill).(*tuSheSkill).check(g, e.user, useCard, []data.PID{e.dyingPlayer})
		}
	}
	for i := g.index + 1; i < len(g.events.list); i++ {
		if event, ok := g.events.list[i].(*beforeDieEvent); ok {
			event.setSkip(true)
		}
		if event, ok := g.events.list[i].(*dieEvent); ok {
			event.setSkip(true)
		}
		if event, ok := g.events.list[i].(*changeRoleEvent); ok {
			event.setSkip(true)
		}
	}
}

// 死亡前事件
type beforeDieEvent struct {
	event
	dieEvent *dieEvent
}

func newBeforeDieEvent(dieEvent *dieEvent) *beforeDieEvent {
	return &beforeDieEvent{dieEvent: dieEvent}
}

func (e *beforeDieEvent) trigger(g *Games) {
	t := e.dieEvent.target
	//涅槃
	if g.players[t].hasEffect(niepanEffect) &&
		!g.players[t].findSkill(data.NiePanSkill).(*niePanSkill).hasUsed {
		g.events.insert(g.index, newSkillSelectEvent(data.NiePanSkill, t, []data.PID{t}, e.dieEvent))
		return
	}
	switch g.mode {
	case data.QGZXEasyMode:
		if e.dieEvent.target == 3 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	case data.QGZXVeryHardMode:
		if e.dieEvent.target == 4 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	case data.QGZXHardMode:
		if e.dieEvent.target == 4 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	case data.QGZXNormalMode:
		if e.dieEvent.target == 3 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	case data.QGZXDoubleMode:
		if e.dieEvent.target == 3 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	case data.QGZXFreeMode:
		if e.dieEvent.target == 2 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	case data.NianShouMode:
		if e.dieEvent.target == 3 {
			g.events.insert(g.index+1, newChangeRoleEvent())
		}
	}

}

// 换角色事件
type changeRoleEvent struct {
	event
}

func newChangeRoleEvent() *changeRoleEvent {
	return &changeRoleEvent{}
}

func (e *changeRoleEvent) trigger(g *Games) {
	getLmp := func() data.Role {
		return data.GhostList[rand.Intn(4)]
	}
	getDevil := func(state uint8) data.Role {
		return data.GhostList[2+rand.Intn(2)+(2*int(state))]
	}
	g.qgzxState++
	var pList []data.PlayerInf
	switch g.mode {
	case data.QGZXNormalMode:
		if g.qgzxState == 4 {
			g.events.insert(g.index, newSettleEvent(0, 1, 2))
			return
		}
		pList = []data.PlayerInf{
			{Role: getDevil(g.qgzxState), PID: 3},
		}
	case data.QGZXEasyMode:
		if g.qgzxState == 3 {
			g.events.insert(g.index, newSettleEvent(0, 1, 2))
			return
		}
		pList = []data.PlayerInf{
			{Role: getDevil(g.qgzxState), PID: 3},
		}
	case data.QGZXHardMode:
		if g.qgzxState == 6 {
			g.events.insert(g.index, newSettleEvent(0, 1, 2))
			return
		}
		pList = []data.PlayerInf{
			{Role: getLmp(), PID: 3},
			{Role: data.GhostList[g.qgzxState+4], PID: 4},
			{Role: getLmp(), PID: 5},
		}
	case data.QGZXVeryHardMode:
		if g.qgzxState == 5 {
			g.events.insert(g.index, newSettleEvent(0, 1, 2))
			return
		}
		pList = []data.PlayerInf{
			{Role: data.GhostList[2*(g.qgzxState/2)], PID: 3},
			{Role: data.GhostList[g.qgzxState+4], PID: 4},
			{Role: data.GhostList[2*(g.qgzxState/2)+1], PID: 5},
		}
	case data.QGZXDoubleMode:
		if g.qgzxState == 5 {
			g.events.insert(g.index, newSettleEvent(0, 1, 2))
			return
		}
		pList = []data.PlayerInf{
			{Role: data.GroupGhostList[g.qgzxState], PID: 3},
		}
	case data.QGZXFreeMode:
		if g.qgzxState == 5 {
			g.events.insert(g.index, newSettleEvent(0, 1))
			return
		}
		pList = []data.PlayerInf{
			{Role: getDevil(g.qgzxState), PID: 2},
		}
	case data.NianShouMode:
		g.events.insert(g.index, newSettleEvent(0, 1, 2))
		return
	}
	for i, c := range g.clients {
		c.SetGameState(data.ChangeRoleState, 0, data.PID(i))
	}
	iterators(g.clients, func(c clientI) { c.SendPlayerInf(pList) })
	<-time.After(1 * time.Second)
	for _, inf := range pList {
		p := newPlayer(g, inf.Role, inf.PID)
		g.players[inf.PID] = &p
		for _, sid := range inf.Role.SkillList {
			p.addSkill(sid)
		}
		g.sendCard2Player(inf.PID, g.getCardsFromTop(4)...)
	}
}

// 死亡事件
type dieEvent struct {
	event
	atker  data.PID
	target data.PID
}

func newDieEvent(atker, target data.PID) *dieEvent {
	return &dieEvent{target: target, atker: atker}
}

func (e *dieEvent) trigger(g *Games) {
	g.setGameState(data.DieState, 0, e.target)
	p := g.players[e.atker]
	t := g.players[e.target]
	//凶祸
	if t.hasEffect(xionhuoEffect) {
		for _, p := range g.getAllAlivePlayer() {
			g.players[p].disableEffect(baoliEffect)
		}
	}
	//山崩
	if t.hasEffect(shanBengEffect) {
		g.useSkill(e.target, data.ShanBengSkill)
		for _, p := range g.players {
			for _, c := range p.equipSlot {
				if c != nil && c.getID() != 0 {
					g.dropCards(c.getID())
					g.removePlayercard(p.pid, c.getID())
				}
			}
		}
	}
	//悲鸣
	if t.hasEffect(beiMinEffect) {
		g.useSkill(e.target, data.BeiMingSkill)
		g.dropCards(p.cards...)
		g.removePlayercard(e.atker, p.cards...)
	}
	//挥泪
	if t.hasEffect(huiHanEffect) {
		g.useSkill(e.target, data.HuiLeiSkill)
		p.dropAllCards(g)
	}
	//冥爆
	if t.hasEffect(mingBaoEffect) {
		g.useSkill(e.target, data.MingBaoSkill)
		for pid := g.getPrvPid(e.target); pid != e.target; pid = g.getPrvPid(pid) {
			e := newDamageEvent(e.target, pid, data.FireDmg, nil, 1)
			g.events.insert(g.index, newBeforeDamageEvent(e), e, newAfterDmgEvent(e))
		}
	}
	//仇决
	if p.hasEffect(choujueEffect) {
		g.addHpMax(e.atker, 1)
		g.sendCard2Player(e.atker, g.getCardsFromTop(2)...)
		p.findSkill(data.QueDiSkill).(*quediSkill).hasUsed = false
		g.useSkill(e.atker, data.ChouJueSkill)
	}
	t.cleanCards(g)
	if e.target == g.turnOwner {
		for i := g.index + 1; i < len(g.events.list); i++ {
			if e, ok := g.events.list[i].(*judgeEvent); ok {
				e.setSkip(true)
			}
			if e, ok := g.events.list[i].(*sendCardEvent); ok {
				e.setSkip(true)
			}
			if e, ok := g.events.list[i].(*useCardEvent); ok {
				e.setSkip(true)
			}
			if e, ok := g.events.list[i].(*dropCardEvent); ok {
				e.setSkip(true)
			}
			if e, ok := g.events.list[i].(*endEvent); ok {
				e.setSkip(true)
			}
		}
	}
	for _, event := range g.events.list[g.index+1:] {
		if befDmg, ok := event.(*beforeDamageEvent); ok {
			if befDmg.dmgEvent.target == e.target {
				befDmg.setSkip(true)
			}
		}
		if dmg, ok := event.(*damageEvent); ok {
			if dmg.target == e.target {
				dmg.setSkip(true)
			}
		}
		if aftDmg, ok := event.(*afterDamageEvent); ok {
			if aftDmg.dmgEvent.target == e.target {
				aftDmg.setSkip(true)
			}
		}
		if skillSel, ok := event.(*skillSelectEvent); ok {
			if skillSel.user == e.target {
				skillSel.setSkip(true)
			}
		}
	}
	t.death = true
	//清空effect
	t.effectMap = make(map[effect]*struct {
		counter        uint8
		skillEffectMap map[data.SID]uint8
	})
	<-time.After(1 * time.Second)
}

// 伤害前事件
type beforeDamageEvent struct {
	event
	dmgEvent *damageEvent
	count    uint8 //为0：第一次生成，检测人物技能，为1：检测樵拾，为2：检测寒冰剑
}

func newBeforeDamageEvent(dmgevent *damageEvent) *beforeDamageEvent {
	return &beforeDamageEvent{dmgEvent: dmgevent}
}

func (e *beforeDamageEvent) trigger(g *Games) {
	p := g.players[e.dmgEvent.atker]
	t := g.players[e.dmgEvent.target]
	beforeDmg := newBeforeDamageEvent(e.dmgEvent)
	beforeDmg.count = e.count + 1
	g.events.insert(g.index, beforeDmg)
check:
	switch e.count {
	case 0:
		//绝情
		if p.hasEffect(jueQingeffect) {
			e.dmgEvent.damageType = data.BleedingDmg
			g.useSkill(p.pid, data.JueQingSkill)
			beforeDmg.setSkip(true)
			return
		}
		//誓仇
		if p.hasEffect(shiChouEffect) && g.getDst(p.pid, t.pid) == 1 {
			g.events.insert(g.index, newSkillSelectEvent(data.ShiChouSkill, p.pid, []data.PID{t.pid}, e.dmgEvent, beforeDmg))
			return
		}
		//追击
		if p.hasEffect(zhuiJiEffect) && g.getDst(p.pid, t.pid) == 1 && e.dmgEvent.card != nil {
			if e.dmgEvent.card.getName() == data.Attack || e.dmgEvent.card.getName() == data.FireAttack ||
				e.dmgEvent.card.getName() == data.LightnAttack {
				g.events.insert(g.index, newSkillSelectEvent(data.ZhuiJiSkill, e.dmgEvent.atker,
					[]data.PID{e.dmgEvent.target}, e.dmgEvent, beforeDmg))
				return
			}
		}
		e.count = 1
		beforeDmg.setSkip(true)
		goto check
	case 1:
		if e.dmgEvent.card != nil {
			if e.dmgEvent.card.getName() == data.Attack || e.dmgEvent.card.getName() == data.FireAttack ||
				e.dmgEvent.card.getName() == data.LightnAttack {
				//寒冰剑
				if p.hasEffect(hbjEffect) && len(t.cards)+t.getEquipCount() > 0 {
					HBJSelect := newSkillSelectEvent(data.HBJSkill, e.dmgEvent.atker,
						[]data.PID{e.dmgEvent.target}, e.dmgEvent)
					g.events.insert(g.index, HBJSelect)
				}
			}
		}
		beforeDmg.setSkip(true)
	}
}

// 伤害事件
type damageEvent struct {
	event
	atker      data.PID
	target     data.PID
	damageType data.SetHpType
	card       cardI
	dmg        data.HP
	realDmg    data.HP //结算完成后的真实伤害
}

func newDamageEvent(atker data.PID, target data.PID, damageType data.SetHpType, card cardI, dmg data.HP) *damageEvent {
	return &damageEvent{atker: atker, target: target, damageType: damageType, card: card, dmg: dmg}
}

func (e *damageEvent) trigger(g *Games) {
	p := g.players[e.atker]
	t := g.players[e.target]
	if t.death {
		return
	}
	//机巧
	if t.hasEffect(changJiEffect) {
		t.disableEffect(changJiEffect)
		g.useSkill(t.pid, data.JiQiaoSkill, byte(0))
		e.dmg = 0
		for _, p := range g.players {
			if p.hasEffect(jiqiaoEffect) {
				g.useSkill(p.pid, data.JiQiaoSkill)
				break
			}
		}
		return
	}
	g.setGameState(data.SetHpState, 0, e.target)
	dmg := e.dmg
	if e.card != nil {
		if e.card.getName() == data.Attack || e.card.getName() == data.LightnAttack || e.card.getName() == data.FireAttack {
			//麒麟弓
			if p.hasEffect(qlgEffect) && (t.equipSlot[2] != nil || t.equipSlot[3] != nil) {
				g.events.insert(g.index, newSkillSelectEvent(data.QLGSkill, e.atker, []data.PID{e.target}))
			}
			//古锭刀
			if p.hasEffect(gddEffect) && len(t.cards) == 0 {
				dmg += 1
				g.useSkill(e.atker, data.GDDSkill)
				if p.hasEffect(guDingEffect) {
					g.events.insert(g.index, newSkillSelectEvent(data.YingHunSkill, e.atker, nil))
				}
			}
			//破军
			if p.hasEffect(poJunEffect) && len(p.cards) >= len(t.cards) {
				pcount, tcount := 0, 0
				for _, s := range p.equipSlot {
					if s != nil {
						pcount++
					}
				}
				for _, s := range t.equipSlot {
					if s != nil {
						tcount++
					}
				}
				if pcount >= tcount {
					dmg++
					g.useSkill(e.atker, data.PoJunSkill)
				}
			}
		}
	}
	{ //骄姿
		jiaozi := false
		owner := data.PID(0)
		if p.hasEffect(jiaoZiEffect) {
			owner = p.pid
			jiaozi = true
			for _, p1 := range g.players {
				if len(p1.cards) >= len(p.cards) && p1 != p {
					jiaozi = false
				}
			}
		} else if t.hasEffect(jiaoZiEffect) {
			owner = p.pid
			jiaozi = true
			for _, p1 := range g.players {
				if len(p1.cards) >= len(t.cards) && p1 != t {
					jiaozi = false
				}
			}
		}
		if jiaozi {
			dmg += 1
			g.useSkill(owner, data.JiaoZiSkill)
		}
	}
	//凶祸
	if t.hasEffect(baoliEffect) {
		if p.hasEffect(xionhuoEffect) {
			dmg++
			g.useSkill(e.atker, data.XionHuoSkill)
		}
	}
	switch e.damageType {
	//火伤藤甲
	case data.FireDmg:
		if t.hasEffect(tengjiaEffect) {
			dmg += 1
		}
		//扣上限
	case data.DffHPMax:
		t.maxHp -= dmg
		if t.hp > t.maxHp {
			t.hp = t.maxHp
		}
		iterators(g.clients, func(c clientI) { c.SetHP(e.target, t.maxHp, e.damageType) })
		if t.maxHp < 1 {
			die := newDieEvent(e.atker, e.target)
			g.events.insert(g.index, newBeforeDieEvent(die), die)
			return
		}
		return
	}
	//白银狮子
	if t.hasEffect(byszEffect) && dmg > 1 {
		dmg = 1
	}
	//八阵
	if t.hasEffect(newBaZhenEffect) {
		g.useSkill(e.target, data.NewBaZhenSkill)
		dmg = t.findSkill(data.NewBaZhenSkill).(*newbazhenSkill).count()
	}
	//帷幕
	if t.hasEffect(weiMuEffect) && g.turnOwner == e.target {
		g.useSkill(e.target, data.WeiMuSkill)
		g.sendCard2Player(e.target, g.getCardsFromTop(2*int(dmg))...)
		return
	}
	//兽盾
checkShouDun:
	if t.hasEffect(shoudunEffect) && dmg > 0 {
		t.disableEffect(shoudunEffect)
		dmg--
		g.useSkill(e.target, data.ShouXiSkill)
		g.useSkill(e.target, data.ShouXiSkill, 0, t.getEffectCount(shoudunEffect))
		goto checkShouDun
	}
	//检查铁索连环
	if t.isLinked && (e.damageType == data.FireDmg || e.damageType == data.LightningDmg) {
		for pid := g.getPrvPid(e.target); pid != e.target; pid = g.getPrvPid(pid) {
			if g.players[pid].isLinked {
				tslhEvent := newTSLHEvent(pid, pid)
				dmgEvent := newDamageEvent(e.atker, pid, e.damageType, nil, dmg)
				g.events.insert(g.index, tslhEvent, newBeforeDamageEvent(dmgEvent), dmgEvent, newAfterDmgEvent(dmgEvent))
			}
		}
		g.events.insert(g.index, newTSLHEvent(t.pid, t.pid))
	}
	e.realDmg = dmg
	t.hp -= dmg
	iterators(g.clients, func(c clientI) { c.SetHP(e.target, t.hp, e.damageType) })
	if t.hp < 0 {
		for _, p := range g.getAllAlivePlayer() {
			//杀绝
			if g.players[p].hasEffect(shajueEffect) && e.card != nil {
				if g.getDmgCard(e.card.getID()) {
					g.sendCard2Player(p, e.card.getID())
				}
				g.players[p].findSkill(data.XionHuoSkill).(*xionHuoSkill).addNum(g)
				break
			}
		}
	}
	<-time.After(200 * time.Millisecond)
	//权计造成伤害check
	if p.hasEffect(quanjiEffect) && dmg > 0 && e.card != nil {
		p.findSkill(data.QuanJiSkill).(*quanJiSkill).check(g, e.atker, e.card.getName(), true, true)
	}
	//魔甲
	if t.hasEffect(mojiaEffect) {
		if t.hp < 30 && dmg > 0 {
			g.events.insert(g.index, newSkillSelectEvent(data.MoJiaSkill, t.pid, nil, p.pid))
		}
	}
	//菜虫
	if t.hasEffect(caichonEffect) {
		g.recover(e.target, e.target, dmg)
	}
	if t.hp < 1 {
		//生成濒死事件
		list := []eventI{
			newDyingEvent(e.target, e.target),
		}
		//完杀
		for _, p := range g.getAllAlivePlayer() {
			if g.players[p].hasEffect(wanShaEffect) && g.turnOwner == p {
				list = append(list, newDyingEvent(e.target, p))
				die := newDieEvent(e.atker, e.target)
				list = append(list, newBeforeDieEvent(die), die)
				g.events.insert(g.index, list...)
				g.useSkill(p, data.WanShaSkill)
				return
			}
		}
		for i := g.getNextPid(e.target); i != e.target; i = g.getNextPid(i) {
			list = append(list, newDyingEvent(e.target, i))
		}
		//生成死亡事件
		die := newDieEvent(e.atker, e.target)
		list = append(list, newBeforeDieEvent(die), die)
		g.events.insert(g.index, list...)
	}
}

type afterDamageEvent struct {
	event
	dmgEvent *damageEvent
}

func newAfterDmgEvent(dmgEvent *damageEvent) *afterDamageEvent {
	return &afterDamageEvent{dmgEvent: dmgEvent}
}

func (e *afterDamageEvent) trigger(g *Games) {
	if e.dmgEvent.damageType == data.DffHPMax || e.dmgEvent.damageType == data.BleedingDmg {
		return
	}
	if e.dmgEvent.dmg == 0 || e.dmgEvent.skip {
		return
	}
	p := g.players[e.dmgEvent.atker]
	t := g.players[e.dmgEvent.target]
	// 兽袭
	if p.hasEffect(shouxiEffect) {
		p.findSkill(data.ShouXiSkill).(*shouXiSkill).count++
	}
	//归心
	if t.hasEffect(guiXinEffect) {
		for i := 0; i < int(e.dmgEvent.realDmg); i++ {
			g.events.insert(g.index, newSkillSelectEvent(data.GuiXinSkill, e.dmgEvent.target, nil))
		}
	}
	//市北
	if t.hasEffect(shibeiEffect) {
		t.findSkill(data.ShiBeiSkill).(*shiBeiSkill).check(g, t.pid)
	}
	//反馈
	if t.hasEffect(fankuiEffect) && len(p.cards)+p.getEquipCount() != 0 && t != p {
		for i := 0; i < int(e.dmgEvent.realDmg); i++ {
			g.events.insert(g.index, newSkillSelectEvent(data.FanKuiSkill, e.dmgEvent.target, []data.PID{e.dmgEvent.atker}))
		}
	}
	//恩怨
	if t.hasEffect(enYuanEffect) && t != p {
		g.events.insert(g.index, newEnYuanEvent(e.dmgEvent.target, e.dmgEvent.atker))
	}
	//武继
	if p.hasEffect(wujiEffect) && g.turnOwner == e.dmgEvent.atker {
		p.findSkill(data.WuJiSkill).(*wuJiSkill).check()
	}
	//权计受伤check
	if t.hasEffect(quanjiEffect) {
		t.findSkill(data.QuanJiSkill).(*quanJiSkill).check(g, e.dmgEvent.target, data.NoName, true, false)
	}
	//克己
	if t.hasEffect(kejiEffect) && g.turnOwner != e.dmgEvent.target {
		if t.findSkill(data.KeJiSkill).(*keJiSkill).count < 2 {
			g.events.insert(g.index, newSkillSelectEvent(data.KeJiSkill, e.dmgEvent.target, nil))
		}
	}
	//樵拾
	if t.hasEffect(qiaoshiEffect) {
		if !t.findSkill(data.QiaoShiSkill).(*qiaoShiSkill).hasUsed {
			g.events.insert(g.index, newQiaoShiSelectEvent(p.pid, t.pid, int(e.dmgEvent.realDmg)))
		}
	}
}

// 选择用不用技能的事件
type skillSelectEvent struct {
	event
	skillId data.SID
	arg     []any
	user    data.PID
	target  []data.PID
}

func newSkillSelectEvent(skillId data.SID, user data.PID, target []data.PID, arg ...any) *skillSelectEvent {
	return &skillSelectEvent{skillId: skillId, arg: arg, user: user, target: target}
}

func (e *skillSelectEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	iterators(g.clients, func(c clientI) { c.SendSkillSelect(e.skillId) })
	p := g.players[e.user]
	g.clients[e.user].SendAvailableTarget(p.findSkill(e.skillId).getAvailableTarget(g, e.user, e.arg...))
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
			if inf.ID != e.skillId {
				continue
			}
			if len(e.target) > 0 {
				p.findSkill(e.skillId).use(g, p, useSkillInf{targets: e.target}, e.arg...)
				g.useSkill(e.user, e.skillId)
			} else {
				p.findSkill(e.skillId).use(g, p, useSkillInf{targets: inf.TargetList}, e.arg...)
				g.useSkill(e.user, e.skillId)
			}
			return
		}
	}
}

type cardJudgeEvent struct {
	event
	result *data.CID
	user   data.PID
}

func newCardJudgeEvent(result *data.CID, user data.PID) *cardJudgeEvent {
	return &cardJudgeEvent{result: result, user: user}
}

func (e *cardJudgeEvent) trigger(g *Games) {
	g.setGameState(data.MakeJudgeState, 0, g.curPlayer)
	cid := g.getCardsFromTop(1)
	*e.result = cid[0]
	g.dropCards(cid[0])
	iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDJudge, cid...) })
	<-time.After(1 * time.Second)
}

type skillJudgeEvent struct {
	event
	user   data.PID
	sid    data.SID
	result *data.CID
}

func newSkillJudgeEvent(user data.PID, sid data.SID, result *data.CID) *skillJudgeEvent {
	return &skillJudgeEvent{user: user, sid: sid, result: result}
}

func (e *skillJudgeEvent) trigger(g *Games) {
	g.setGameState(data.SkillJudgeState, 0, e.user)
	cid := g.getCardsFromTop(1)
	*e.result = cid[0]
	g.dropCards(cid[0])
	iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDPerSkillJudge, data.CID(e.sid)) })
	iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDSkillJudge, *e.result) })
	<-time.After(1 * time.Second)
}

// 结算事件
type settleEvent struct {
	event
	winner []data.PID
}

func newSettleEvent(winner ...data.PID) *settleEvent {
	return &settleEvent{winner: winner}
}

func (e *settleEvent) trigger(g *Games) {
	for i, c := range g.clients {
		if isItemInList(e.winner, data.PID(i)) {
			c.SetGameState(data.WinState, 0, data.PID(i))
		} else {
			c.SetGameState(data.LoseState, 0, data.PID(i))
		}
	}
	<-time.After(5 * time.Second)
	g.Close()
}

// 客户端重新连接事件
type reConnEvent struct {
	event
	user    data.PID
	onStart func()
}

func newReConnEvent(user data.PID, onstart func()) *reConnEvent {
	return &reConnEvent{
		user:    user,
		onStart: onstart,
	}
}

func (e *reConnEvent) trigger(g *Games) {
	e.onStart()
	for i, c := range g.clients {
		if i == int(e.user) {
			continue
		}
		c.SetGameState(data.ReConnState, 0, e.user)
	}
	c := g.clients[e.user]
	c.SetPid(e.user)
	c.SendAvailableRole(g.players[e.user].originRole)
	select {
	case <-c.GetRole():
	case <-g.closeSignal:
		return
	}
	pList := make([]data.PlayerInf, len(g.players))
	for i, p := range g.players {
		pList[i].PID = p.pid
		pList[i].Role = p.originRole
		pList[i].Role.MaxHP = p.maxHp
	}
	c.SendPlayerInf(pList)
	for _, p := range g.players {
		cards := make([]data.CID, len(p.cards))
		copy(cards, p.cards)
		c.SendCard(p.pid, cards...)
	}
	//未完成
}
