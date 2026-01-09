package server

import (
	"goltk/data"
	"time"
)

// 杀事件
type atkevent struct {
	event
	user, target data.PID
	dmgType      data.SetHpType
	card         cardI //杀本体卡
	unResponsive bool
	addDmg       uint8
}

func newAtkEvent(user, target data.PID, dmgType data.SetHpType, card cardI) *atkevent {
	return &atkevent{user: user, target: target, dmgType: dmgType, card: card}
}

// 无时间长度杀事件
func (e *atkevent) trigger(g *Games) {
	p, t := g.players[e.user], g.players[e.target]
	//储存额外加在最前面事件
	list := []eventI{}
	dmgEvent := newDamageEvent(e.user, e.target, e.dmgType, e.card, 1)
	befDmgEvent := newBeforeDamageEvent(dmgEvent)
	aftDmgEvent := newAfterDmgEvent(dmgEvent)
	//检查伏骑
	if p.hasEffect(fuJiEffect) && g.getDst(e.target, e.user) == 1 {
		e.unResponsive = true
		g.useSkill(e.user, data.FuJiSkill)
	}
	//检查问计
	if p.hasEffect(wenJiEffect) &&
		isItemInList([]data.CardName{data.Attack, data.FireAttack, data.LightnAttack},
			p.findSkill(data.WenJiSkill).(*wenJiSkill).name) {
		e.unResponsive = true
		g.useSkill(e.user, data.WenJiSkill)
	}
	//酒
	if p.isDrunk {
		dmgEvent.dmg += 1
		p.isDrunk = false
	}
	//醉酒
	if p.hasEffect(zuijiuEffect) {
		dmgEvent.dmg += 1
	}
	//却敌
	if p.hasEffect(quediEffect) {
		if p.findSkill(data.QueDiSkill).(*quediSkill).useAddDmg() {
			dmgEvent.dmg++
		}
	}
	dmgEvent.dmg += data.HP(e.addDmg)
	//凤魄
	if p.hasEffect(fengpoEffect) {
		skill := p.findSkill(data.FengPoSkill).(*fengPoSkill)
		dmgEvent.dmg += data.HP(skill.conunt)
		skill.enable = false
	}
	//仁王
	if t.hasEffect(renwangEffect) && t.equipSlot[data.ArmorSlot] == nil {
		if e.card.getDecor().ISBlack() {
			g.useSkill(e.target, data.RenWangSkill)
			//略影
			if p.hasEffect(lueyingEffect) {
				p.findSkill(data.LueYingSkill).(*lueYingSkill).check(g, e.user, e.card.getName())
			}
			return
		}
	}
	//检查是否无视防具
	if !p.hasEffect(ignorArmor) {
		//检查是否有仁王盾
		if t.hasEffect(rwdEffect) {
			if e.card.getDecor().ISBlack() {
				g.useSkill(e.target, data.RWDSkill)
				//略影
				if p.hasEffect(lueyingEffect) {
					p.findSkill(data.LueYingSkill).(*lueYingSkill).check(g, e.user, e.card.getName())
				}
				return
			}
		}
		//检测是否有藤甲
		if t.hasEffect(tengjiaEffect) && e.dmgType == data.NormalDmg {
			g.useSkill(e.target, data.TengJiaSkill)
			//略影
			if p.hasEffect(lueyingEffect) {
				p.findSkill(data.LueYingSkill).(*lueYingSkill).check(g, e.user, e.card.getName())
			}
			return
		}
		//是否有八卦阵
		if t.hasEffect(bgzEffect) && !e.unResponsive {
			list = append(list, newSkillSelectEvent(data.BGZSkill, t.pid, nil))
		}
	}
	//是否不可响应
	if e.unResponsive {
		//略影
		if p.hasEffect(lueyingEffect) {
			p.findSkill(data.LueYingSkill).(*lueYingSkill).check(g, e.user, e.card.getName())
		}
		g.events.insert(g.index, befDmgEvent, dmgEvent, aftDmgEvent)
		return
	}
	//正常
	g.events.insert(g.index, newDodgeEvent(e.target, e.user, []eventI{befDmgEvent, dmgEvent, aftDmgEvent}, e.card),
		befDmgEvent, dmgEvent, aftDmgEvent)
	g.events.insert(g.index, list...)
}

// 闪事件
type dodgeEvent struct {
	event
	atker      data.PID
	target     data.PID
	skipEvents []eventI
	card       cardI //杀本体卡
}

func newDodgeEvent(target data.PID, atker data.PID, skipEvents []eventI, card cardI) *dodgeEvent {
	return &dodgeEvent{target: target, atker: atker, skipEvents: skipEvents, card: card}
}

func (e *dodgeEvent) trigger(g *Games) {
	p, t := g.players[e.atker], g.players[e.target]
	//检查玩家是否有虚拟闪效果
	var usecard cardI //使用的卡
	if g.players[e.target].hasEffect(virtualDodgeEffect) {
		g.players[e.target].disableEffect(virtualDodgeEffect)
		g.useTmpCard(e.target, data.Dodge, data.NoDec, 0, data.VirtualCard, e.atker)
		usecard = newCard(data.Card{Name: data.Dodge, CardType: data.BaseCardType, Dec: data.NoDec})
		goto dodgeSuccess
	}
	{
		const lastTime = 20 * time.Second
		g.setGameState(data.DodgeState, lastTime, e.target)
		useAbaleCards := g.players[e.target].getUseAbleCards(g)
		g.clients[e.target].SendUseAbleCards(useAbaleCards)
		useAbleSkill := g.players[e.target].getUseAbleSkill(g)
		g.clients[e.target].SendUseAbleSkill(useAbleSkill)
		timer := time.After(lastTime)
		t := g.players[e.target]
		for {
			select {
			case <-timer:
				goto dodgeFail
			case inf := <-g.clients[e.target].GetUseSkillInf():
				//回应使用技能请求
				if t.findSkill(inf.ID).use(g, t, useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args}, e.atker) {
					goto dodgeSuccess
				}
			case inf := <-g.clients[e.target].GetUseCardInf():
				if inf.Skip {
					goto dodgeFail
				}
				if !isItemInList(useAbaleCards, inf.ID) { //检查玩家发送的牌是否可用
					continue
				}
				g.useCard(e.target, inf.ID, e.atker)
				g.players[e.target].delCard(inf.ID)
				usecard = g.cards[inf.ID]
				g.dropCards(inf.ID)
				goto dodgeSuccess
			}
		}
	}
dodgeSuccess:
	for _, event := range e.skipEvents {
		event.setSkip(true)
	}
	//奋音技能
	if t.hasEffect(fenYinEffect) && e.target == g.turnOwner {
		t.findSkill(data.FenYinSkill).(*fenYinSkill).check(g, usecard, e.target)
	}
	//蒺藜技能
	if t.hasEffect(jiLiEffect) {
		t.findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, e.target)
	}
	//无双
	if p.hasEffect(wuShuangEffect) {
		if p.findSkill(data.WuShuangSkill).(*wuShuangSkill).check() {
			g.events.insert(g.index, newAtkEvent(e.atker, e.target, e.card.(*atkCard).dmgType, e.card))
			goto wushuanSkip
		}
	}
	//青釭
	if p.hasEffect(qingGangEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.QingGangSkill, e.atker, nil, e.target))
	} else if t.hasEffect(qingGangEffect) {
		g.recover(e.target, e.target, 1)
	}
	//贯石斧
	if p.hasEffect(gsfEffect) {
		g.events.insert(g.index, newSkillSelectEvent(data.GSFSkill, e.atker, []data.PID{e.target}, e.skipEvents))
	}
	//追杀
	if p.hasEffect(continueAtk) {
		g.events.insert(g.index, newSkillSelectEvent(data.QLYYDSkill, e.atker, []data.PID{e.target}))
	}
	//虎啸
	if p.hasEffect(huxiaoEffect) {
		p.hasAttack = false
		g.useSkill(e.atker, data.HuXiaoSkill)
	}
wushuanSkip:
	//恃才
	if t.hasEffect(shicaiEffect) {
		t.findSkill(data.ShiCaiSkill).(*shiCaiSkill).check(g, usecard, e.target, nil)
	}
	//渐营
	if t.hasEffect(jianyingEffect) && g.turnOwner == e.target {
		t.findSkill(data.JianYingSkill).(*jianyingSkill).check(g, e.target, usecard)
	}
	//略影
	if p.hasEffect(lueyingEffect) {
		p.findSkill(data.LueYingSkill).(*lueYingSkill).check(g, e.atker, e.card.getName())
	}
	//应援
	if t.hasEffect(yingYuanEffect) {
		t.findSkill(data.YingYuanSkill).(*yingYuanSkill).check(g, e.target, nil, usecard.getID())
	}
	return
dodgeFail:
	//青龙
	if p.hasEffect(yanyueEffect) {
		if p.findSkill(data.YanYueSkill).(*yanyueSkill).enable {
			e.skipEvents[1].(*damageEvent).dmg += 1
			p.findSkill(data.YanYueSkill).(*yanyueSkill).enable = false
		} else {
			p.hasAttack = false
		}
	}
	//无双
	if p.hasEffect(wuShuangEffect) {
		p.findSkill(data.WuShuangSkill).(*wuShuangSkill).hasUsed = false
	}
	//利驭
	if p.hasEffect(fangTianEffect) {
		p.findSkill(data.FangTianSkill).(*fangTianSkill).counter(g, e.atker)
	}
	//略影
	if p.hasEffect(lueyingEffect) {
		p.findSkill(data.LueYingSkill).(*lueYingSkill).check(g, e.atker, e.card.getName())
	}
}

// 无懈可击事件
type wxkjEvent struct {
	event
	user         data.PID
	card         cardI
	targetEvents []eventI //将要被无懈可击的事件列表
	target       data.PID //对谁生效
	count        uint8    //无懈翻转，被无懈一次为true
}

func newWXKJEvent(card cardI, user, target data.PID, events ...eventI) *wxkjEvent {
	return &wxkjEvent{targetEvents: events, user: user, card: card, target: target}
}

func (e *wxkjEvent) trigger(g *Games) {
	p := g.players[e.user]
	//问计
	if p.hasEffect(wenJiEffect) && p.findSkill(data.WenJiSkill).(*wenJiSkill).name == e.card.getName() {
		return
	}
	const lastTime = 10 * time.Second
	//检查是否有玩家拥有无懈可击
	hasWXKJ := false
	for _, p := range g.players {
		if isListContain(p.cards, func(cid data.CID) bool { return g.cards[cid].getName() == data.WXKJ }) {
			hasWXKJ = true
			break
		}
		if p.hasEffect(miewuEffect) || p.hasEffect(yisuanEffect) {
			hasWXKJ = true
			break
		}
	}
	if !hasWXKJ {
		return
	}
	rsp := data.UseSkillRsp{ID: 0, Targets: []data.PID{e.target}, Args: []byte{byte(e.card.getName()), e.count}}
	//手动为每个玩家SetGameState
	for i := 0; i < len(g.clients); i++ {
		if g.players[i].death {
			continue
		}
		g.clients[i].SetGameState(data.WXKJState, lastTime, e.user)
		g.clients[i].SendUseSkillRsp(rsp)
		//伏骑
		if g.players[e.user].hasEffect(fuJiEffect) && g.getDst(data.PID(i), e.user) == 1 {
			g.clients[i].SendUseAbleCards(nil)
			g.clients[i].UseSkill(e.user, data.FuJiSkill, nil)
			g.clients[i].SendUseAbleSkill(nil)
		} else {
			g.clients[i].SendUseAbleCards(g.players[i].getUseAbleCards(g))
			useAbleSkill := g.players[i].getUseAbleSkill(g)
			g.clients[i].SendUseAbleSkill(useAbleSkill)
		}
	}
	timer := time.After(lastTime)
	//合并所有玩家的用牌信息通道
	usecardInf := make(chan struct {
		data.UseCardInf
		pid data.PID
	}, 1)
	useskillInf := make(chan struct {
		data.UseSkillInf
		pid data.PID
	})
	isClose := make(chan struct{})
	for i, c := range g.clients {
		if g.players[i].death {
			continue
		}
		go func() {
			for {
				select {
				case inf := <-c.GetUseCardInf():
					usecardInf <- struct {
						data.UseCardInf
						pid data.PID
					}{UseCardInf: inf, pid: data.PID(i)}
				case inf := <-c.GetUseSkillInf():
					useskillInf <- struct {
						data.UseSkillInf
						pid data.PID
					}{UseSkillInf: inf, pid: data.PID(i)}
				case <-isClose:
					return
				}
			}
		}()
	}
	skipCount := 0
	for {
		select {
		case <-timer:
			close(isClose)
			return
		case inf := <-useskillInf:
			//回应使用技能请求
			if g.players[inf.pid].findSkill(inf.ID).use(g, g.players[inf.pid],
				useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args}, e.user) {
				close(isClose)
				for _, event := range e.targetEvents {
					event.setSkip(!event.isSkipAble())
				}
				wxkj := newWXKJEvent(g.cards[inf.ID], inf.pid, e.target, e.targetEvents...)
				wxkj.count = e.count + 1
				g.events.insert(g.index, wxkj)
				return
			}
			continue
			//高危险
		case inf := <-usecardInf:
			if inf.Skip {
				skipCount++
				if skipCount == g.getAlivePlayerCount() {
					close(isClose)
					return
				}
				continue
			}
			close(isClose)
			g.useCard(inf.pid, inf.ID)
			g.players[inf.pid].delCard(inf.ID)
			//奋音技能
			if g.players[inf.pid].hasEffect(fenYinEffect) && inf.pid == g.turnOwner {
				g.players[inf.pid].findSkill(data.FenYinSkill).(*fenYinSkill).check(g, g.cards[inf.ID], inf.pid)
			}
			//蒺藜技能
			if g.players[inf.pid].hasEffect(jiLiEffect) {
				g.players[inf.pid].findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, inf.pid)
			}
			g.dropCards(inf.ID)
			for _, event := range e.targetEvents {
				event.setSkip(!event.isSkipAble())
			}
			wxkj := newWXKJEvent(g.cards[inf.ID], inf.pid, e.target, e.targetEvents...)
			wxkj.count = e.count + 1
			g.events.insert(g.index, wxkj)
			//恃才
			if g.players[inf.pid].hasEffect(shicaiEffect) {
				g.players[inf.pid].findSkill(data.ShiCaiSkill).(*shiCaiSkill).check(g, g.cards[inf.ID], inf.pid, nil)
			}
			//亦算
			if g.players[inf.pid].hasEffect(yisuanEffect) && g.turnOwner == inf.pid {
				g.players[inf.pid].findSkill(data.YiSuanSkill).(*yisuanSkill).check(g, inf.ID, inf.pid, nil)
			}
			//渐营
			if g.players[inf.pid].hasEffect(jianyingEffect) && g.turnOwner == inf.pid {
				g.players[inf.pid].findSkill(data.JianYingSkill).(*jianyingSkill).check(g, e.user, g.cards[inf.ID])
			}
			//应援
			if g.players[inf.pid].hasEffect(yingYuanEffect) {
				g.players[inf.pid].findSkill(data.YingYuanSkill).(*yingYuanSkill).check(g, inf.pid, nil, inf.ID)
			}
			//集智
			if p.hasEffect(jizhiEffect) {
				if p.findSkill(data.JiZhiSkill).(*jiZhiSkill).check(g.cards[inf.ID], nil) {
					g.sendCard2Player(e.user, g.getCardsFromTop(1)...)
					g.useSkill(e.user, data.JiZhiSkill)
				}
			}
			return
		}
	}
}

// 无中生有事件
type wzsyEvent struct {
	event
	user data.PID
}

func newWzsyEvent(user data.PID) *wzsyEvent {
	return &wzsyEvent{user: user}
}

func (e *wzsyEvent) trigger(g *Games) {
	cards := g.getCards(2, e.user)
	g.sendCard2Player(e.user, cards...)
}

// 决斗事件
type duelEvent struct {
	event
	user   data.PID
	target data.PID
	card   cardI //决斗本体牌
}

func newDuelEvent(user, target data.PID, card cardI) *duelEvent {
	return &duelEvent{user: user, target: target, card: card}
}

func (e *duelEvent) trigger(g *Games) {
	t := g.players[e.target]
	p := g.players[e.user]
	//伏骑,问计
	if (g.players[g.turnOwner].hasEffect(fuJiEffect) && g.getDst(e.target, g.turnOwner) == 1) ||
		(g.players[g.turnOwner].hasEffect(wenJiEffect) &&
			g.players[g.turnOwner].findSkill(data.WenJiSkill).(*wenJiSkill).name == data.Duel) {
		dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		return
	}
	//帷幕
	if t.hasEffect(weiMuEffect) && e.card.getDecor().ISBlack() {
		g.useSkill(e.target, data.WeiMuSkill)
		return
	}
	const lastTime = 20 * time.Second
	g.setGameState(data.DuelState, lastTime, e.target)
	useAbaleCards := g.players[e.target].getUseAbleCards(g)
	g.clients[e.target].SendUseAbleCards(useAbaleCards)
	useAbleSkill := g.players[e.target].getUseAbleSkill(g)
	g.clients[e.target].SendUseAbleSkill(useAbleSkill)
	duelevent := newDuelEvent(e.target, e.user, e.card)
	g.events.insert(g.index, duelevent)
	timer := time.After(lastTime)
	var useCard cardI //使用的牌
	for {
		select {
		case <-timer:
			goto usefail
		case inf := <-g.clients[e.target].GetUseSkillInf():
			//回应使用技能请求
			if t.findSkill(inf.ID).use(g, t, useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args}, e.user, &useCard) {
				goto useSuccess
			}
		case inf := <-g.clients[e.target].GetUseCardInf():
			if inf.Skip {
				goto usefail
			}
			if !isItemInList(useAbaleCards, inf.ID) { //检查玩家发送的牌是否可用
				continue
			}
			g.useCard(e.target, inf.ID, e.user)
			g.players[e.target].delCard(inf.ID)
			useCard = g.cards[inf.ID]
			g.dropCards(inf.ID)
			goto useSuccess
		}
	}
useSuccess:
	//奋音技能
	if t.hasEffect(fenYinEffect) && e.target == g.turnOwner {
		t.findSkill(data.FenYinSkill).(*fenYinSkill).check(g, useCard, e.target)
	}
	//蒺藜技能
	if t.hasEffect(jiLiEffect) {
		t.findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, e.target)
	}
	//克己
	if t.hasEffect(kejiEffect) && g.turnOwner == e.target {
		t.findSkill(data.KeJiSkill).(*keJiSkill).useAtk = true
	}
	//无双(高危险)
	if p.hasEffect(wuShuangEffect) {
		if p.findSkill(data.WuShuangSkill).(*wuShuangSkill).check() {
			g.events.insert(g.index, newDuelEvent(e.user, e.target, e.card))
			duelevent.setSkip(true)
		}
	}
	return
usefail:
	duelevent.setSkip(true)
	dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
	//却敌
	if g.players[e.user].hasEffect(quediEffect) {
		if g.players[e.user].findSkill(data.QueDiSkill).(*quediSkill).useAddDmg() {
			dmg.dmg++
		}
	}
	//却敌
	if g.players[e.target].hasEffect(quediEffect) {
		if g.players[e.target].findSkill(data.QueDiSkill).(*quediSkill).useAddDmg() {
			dmg.dmg++
		}
	}
	//椎锋
	if g.players[e.target].hasEffect(zhuiFengEffect) {
		skill := g.players[e.target].findSkill(data.ZhuiFengSkill).(*zhuiFengSkill)
		if skill.enable {
			skill.enable = false
			skill.count = 2
			g.useSkill(e.target, data.ZhuiFengSkill)
			return
		}
	}
	//椎锋
	if g.players[e.user].hasEffect(zhuiFengEffect) {
		skill := g.players[e.user].findSkill(data.ZhuiFengSkill).(*zhuiFengSkill)
		if skill.enable {
			skill.enable = false
		}
	}
	//凤魄
	if p.hasEffect(fengpoEffect) {
		skill := p.findSkill(data.FengPoSkill).(*fengPoSkill)
		dmg.dmg += data.HP(skill.conunt)
		skill.enable = false
	}
	g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
}

type nmrqEvent struct {
	event
	user   data.PID
	target data.PID
	card   cardI //南蛮入侵卡
}

func newNMRQEvent(user, target data.PID, card cardI) *nmrqEvent {
	return &nmrqEvent{user: user, target: target, card: card}
}

func (e *nmrqEvent) trigger(g *Games) {
	//伏骑,问计
	if (g.players[e.user].hasEffect(fuJiEffect) && g.getDst(e.target, g.turnOwner) == 1) ||
		(g.players[e.user].hasEffect(wenJiEffect) &&
			g.players[e.user].findSkill(data.WenJiSkill).(*wenJiSkill).name == data.NMRQ) {
		dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		return
	}
	const lastTime = 20 * time.Second
	t := g.players[e.target]
	g.setGameState(data.NMRQState, lastTime, e.target)
	useAbaleCards := t.getUseAbleCards(g)
	g.clients[e.target].SendUseAbleCards(useAbaleCards)
	useAbleSkill := t.getUseAbleSkill(g)
	g.clients[e.target].SendUseAbleSkill(useAbleSkill)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
			g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
			return
		case inf := <-g.clients[e.target].GetUseSkillInf():
			//回应使用技能请求
			if t.findSkill(inf.ID).use(g, t, useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args}, e.user) {
				goto success
			}
		case inf := <-g.clients[e.target].GetUseCardInf():
			if inf.Skip {
				dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
				g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
				return
			}
			if !isItemInList(useAbaleCards, inf.ID) { //检查玩家发送的牌是否可用
				continue
			}
			g.useCard(e.target, inf.ID, e.user)
			g.players[e.target].delCard(inf.ID)
			g.dropCards(inf.ID)
			goto success
		}
	}
success:
	//蒺藜技能
	if t.hasEffect(jiLiEffect) {
		t.findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, e.target)
	}
}

type wjqfEvent struct {
	event
	user   data.PID
	target data.PID
	card   cardI //万箭齐发卡
}

func newWJQFEvent(user, target data.PID, card cardI) *wjqfEvent {
	return &wjqfEvent{user: user, target: target, card: card}
}

func (e *wjqfEvent) trigger(g *Games) {
	//伏骑,问计
	if (g.players[e.user].hasEffect(fuJiEffect) && g.getDst(e.target, g.turnOwner) == 1) ||
		(g.players[e.user].hasEffect(wenJiEffect) &&
			g.players[e.user].findSkill(data.WenJiSkill).(*wenJiSkill).name == data.WJQF) {
		dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		return
	}
	if g.players[e.target].hasEffect(virtualDodgeEffect) {
		g.players[e.target].disableEffect(virtualDodgeEffect)
		g.useTmpCard(e.target, data.Dodge, data.NoDec, 0, data.VirtualCard, e.user)
		return
	}
	t := g.players[e.target]
	const lastTime = 20 * time.Second
	g.setGameState(data.WJQFState, lastTime, e.target)
	useAbaleCards := t.getUseAbleCards(g)
	g.clients[e.target].SendUseAbleCards(useAbaleCards)
	useAbleSkill := t.getUseAbleSkill(g)
	g.clients[e.target].SendUseAbleSkill(useAbleSkill)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
			g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
			return
		case inf := <-g.clients[e.target].GetUseSkillInf():
			//回应使用技能请求
			if t.findSkill(inf.ID).use(g, t, useSkillInf{targets: inf.TargetList, cards: inf.Cards, args: inf.Args}, e.user) {
				goto success
			}
		case inf := <-g.clients[e.target].GetUseCardInf():
			if inf.Skip {
				dmg := newDamageEvent(e.user, e.target, data.NormalDmg, e.card, 1)
				g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
				return
			}
			if !isItemInList(useAbaleCards, inf.ID) { //检查玩家发送的牌是否可用
				continue
			}
			g.useCard(e.target, inf.ID, e.user)
			g.players[e.target].delCard(inf.ID)
			g.dropCards(inf.ID)
			goto success
		}
	}
success:
	//蒺藜技能
	if g.players[e.target].hasEffect(jiLiEffect) {
		g.players[e.target].findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, e.target)
	}
}

type tyjyEvent struct {
	event
	user   data.PID
	target data.PID
}

func newTYJYEvent(user, target data.PID) *tyjyEvent {
	return &tyjyEvent{user: user, target: target}
}

func (e *tyjyEvent) trigger(g *Games) {
	g.setGameState(data.TYJYState, 0, e.target)
	g.recover(e.user, e.target, 1)
	<-time.After(1 * time.Second)
}

// 五谷丰登事件
type wgfdEvent struct {
	event
	user   data.PID
	target data.PID
	cards  *[]data.CID
}

func newWGFDEvent(user, target data.PID, cards *[]data.CID) *wgfdEvent {
	return &wgfdEvent{user: user, target: target, cards: cards}
}

func (e *wgfdEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	<-time.After(20 * time.Millisecond)
	g.setGameState(data.WGFDState, lastTime, e.target)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			c := (*e.cards)[0]
			*e.cards = (*e.cards)[1:]
			g.sendCard2Player(e.target, c)
			iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDWGFD, (*e.cards)...) })
			return
		case inf := <-g.clients[e.target].GetUseCardInf():
			if !isItemInList(*e.cards, inf.ID) {
				continue
			}
			for i := 0; i < len(*e.cards); i++ {
				if (*e.cards)[i] != inf.ID {
					continue
				}
				c := (*e.cards)[i]
				*e.cards = append((*e.cards)[:i], (*e.cards)[i+1:]...)
				g.sendCard2Player(e.target, c)
				iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDWGFD, (*e.cards)...) })
				break
			}
			return
		}
	}
}

// 火攻展示事件
type burnShowEvent struct {
	event
	user   data.PID
	target data.PID
	card   cardI
}

func newBurnShowEvent(user, target data.PID, card cardI) *burnShowEvent {
	return &burnShowEvent{user: user, target: target, card: card}
}

func (e *burnShowEvent) trigger(g *Games) {
	const lastTime = 10 * time.Second
	g.setGameState(data.BurnShowState, lastTime, e.target)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			g.events.insert(g.index, newburnDropEvent(e.user, e.target, g.players[e.target].cards[0], e.card))
			return
		case inf := <-g.clients[e.target].GetUseCardInf():
			if !isItemInList(g.players[e.target].cards, inf.ID) {
				continue
			}
			g.events.insert(g.index, newburnDropEvent(e.user, e.target, inf.ID, e.card))
			return
		}
	}
}

// 火攻弃置卡牌阶段
type burnDropEvent struct {
	event
	user, target data.PID
	card         data.CID
	burnCard     cardI //火攻本体牌
}

func newburnDropEvent(user, target data.PID, card data.CID, burnCard cardI) *burnDropEvent {
	return &burnDropEvent{user: user, target: target, card: card, burnCard: burnCard}
}

func (e *burnDropEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	g.setGameState(data.BurnDropState, lastTime, e.user)
	//发送目标展示卡牌
	iterators(g.clients, func(c clientI) { c.SendGSCards(data.GSIDBurn, e.card) })
	//计算并发送可用牌
	useAbleCard := []data.CID{}
	for _, c := range g.players[e.user].cards {
		if g.cards[c].getDecor() == g.cards[e.card].getDecor() {
			useAbleCard = append(useAbleCard, c)
		}
	}
	g.clients[e.user].SendUseAbleCards(useAbleCard)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.Skip {
				return
			}
			if !isItemInList(useAbleCard, inf.ID) {
				continue
			}
			g.dropCards(inf.ID)
			g.removePlayercard(e.user, inf.ID)
			dmg := newDamageEvent(e.user, e.target, data.FireDmg, e.burnCard, 1)
			g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
			return
		}
	}
}

// 铁索连环
type tslhEvent struct {
	event
	user, target data.PID
}

func newTSLHEvent(user, target data.PID) *tslhEvent {
	return &tslhEvent{user: user, target: target}
}

func (e *tslhEvent) trigger(g *Games) {
	g.players[e.target].isLinked = !g.players[e.target].isLinked
	g.useSkill(e.target, data.TieSuoSkill)
	<-time.After(100 * time.Millisecond)
}

// 顺手牵羊
type ssqyEvent struct {
	event
	user, target data.PID
}

func newSSQYEvent(user, target data.PID) *ssqyEvent {
	return &ssqyEvent{user: user, target: target}
}

func (e *ssqyEvent) trigger(g *Games) {
	const lastTime = 10 * time.Second
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
	if len(cards) == 0 {
		return
	}
	//发送顺手牵羊牌堆
	g.clients[e.user].SendGSCards(data.GSIDSSQY, cards...)
	//设置游戏阶段
	g.setGameState(data.SSQYState, lastTime, e.user)
	timer := time.After(lastTime)
	for {
		select {
		case <-timer:
			for i := len(cards) - 1; i >= 0; i-- {
				if cards[i] == 0 {
					continue
				}
				g.moveCard(e.target, e.user, cards[i])
				//魔炎
				if t.hasEffect(moYanEffect) {
					g.events.insert(g.index, newSkillSelectEvent(data.MoYanSkill, e.target, nil, 0))
				}
				break
			}
			return
		case inf := <-g.clients[e.user].GetUseCardInf():
			if inf.ID == 0 || !isItemInList(cards, inf.ID) {
				continue
			}
			g.moveCard(e.target, e.user, inf.ID)
			return
		}
	}
}

// 过河拆桥
type ghcqEvent struct {
	event
	user, target data.PID
}

func newGHCQEvent(user, target data.PID) *ghcqEvent {
	return &ghcqEvent{user: user, target: target}
}

func (e *ghcqEvent) trigger(g *Games) {
	const lastTime = 10 * time.Second
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
	if len(cards) == 0 {
		return
	}
	//发送牌堆
	g.clients[e.user].SendGSCards(data.GSIDGHCQ, cards...)
	//设置游戏阶段
	g.setGameState(data.GHCQState, lastTime, e.user)
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
				//魔炎
				if t.hasEffect(moYanEffect) {
					g.events.insert(g.index, newSkillSelectEvent(data.MoYanSkill, e.target, nil, 0))
				}
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

// 借刀杀人
type jdsrEvent struct {
	event
	user, killer, target data.PID
}

func newJDSREvent(user, killer, target data.PID) *jdsrEvent {
	return &jdsrEvent{user: user, killer: killer, target: target}
}

func (e *jdsrEvent) trigger(g *Games) {
	const lastTime = 20 * time.Second
	killer := g.players[e.killer]
	g.setGameState(data.JDSRState, lastTime, e.killer)
	useAbleSkill := killer.getUseAbleSkill(g)
	g.clients[e.killer].SendUseAbleSkill(useAbleSkill)
	useAble := killer.getUseAbleCards(g)
	g.clients[e.killer].SendUseAbleCards(useAble)
	timer := time.After(lastTime)
	var useCard cardI
	for {
		select {
		case <-timer:
			g.moveCard(e.killer, e.user, g.players[e.killer].equipSlot[data.WeaponSlot].getID())
			return
		case inf := <-g.clients[e.killer].GetUseSkillInf():
			if !isItemInList(useAbleSkill, inf.ID) {
				continue
			}
			if killer.findSkill(inf.ID).use(g, killer, useSkillInf{inf.TargetList, inf.Cards, inf.Args}) {
				goto useAtk
			}
		case inf := <-g.clients[e.killer].GetUseCardInf():
			if inf.Skip {
				g.moveCard(e.killer, e.user, g.players[e.killer].equipSlot[data.WeaponSlot].getID())
				return
			}
			if !isItemInList(useAble, inf.ID) {
				continue
			}
			g.cards[inf.ID].use(g, e.killer, e.target)
			useCard = g.cards[inf.ID]
			goto useAtk
		}
	}
useAtk:
	//蒺藜技能
	if g.players[e.target].hasEffect(jiLiEffect) {
		g.players[e.target].findSkill(data.JiLiSkill).(*jiLiSkill).check(g, -1, e.target)
	}
	//恃才
	if g.players[e.killer].hasEffect(shicaiEffect) {
		g.players[e.killer].findSkill(data.ShiCaiSkill).(*shiCaiSkill).check(g, useCard, e.target, nil)
	}
	//却敌
	if g.players[e.killer].hasEffect(quediEffect) {
		g.players[e.killer].findSkill(data.QueDiSkill).(*quediSkill).check(g, e.user, useCard, []data.PID{e.target})
	}
}

// 乐不思蜀
type lbssEvent struct {
	event
	result *data.CID //判定结果
}

func newLBSSEvent(result *data.CID) *lbssEvent {
	return &lbssEvent{result: result}
}

func (e *lbssEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor() == data.HeartDec {
		return
	}
	g.useSkill(g.turnOwner, data.DrawLBSS)
	for i := g.index + 1; i < g.events.size(); i++ {
		if e, ok := g.events.list[i].(*useCardEvent); ok {
			e.setSkip(true)
			return
		}
	}
}

// 兵粮寸断
type blcdEvent struct {
	event
	result *data.CID //判定结果
}

func newBLCDEvent(result *data.CID) *blcdEvent {
	return &blcdEvent{result: result}
}

func (e *blcdEvent) trigger(g *Games) {
	if g.cards[*e.result].getDecor() == data.ClubDec {
		return
	}
	g.useSkill(g.turnOwner, data.DrawBLCD)
	for i := g.index + 1; i < g.events.size(); i++ {
		if e, ok := g.events.list[i].(*sendCardEvent); ok {
			e.setSkip(true)
			return
		}
	}
}

// 闪电
type lightningEvent struct {
	event
	result *data.CID
	target data.PID
	end    eventI
	card   cardI //闪电卡
}

func newLightningEvent(result *data.CID, target data.PID, end eventI, card cardI) *lightningEvent {
	return &lightningEvent{result: result, target: target, end: end, card: card}
}

func (e *lightningEvent) trigger(g *Games) {
	result := g.cards[*e.result]
	if result.getDecor() == data.SpadeDec && (result.getNum() >= 2 && result.getNum() <= 9) {
		dmg := newDamageEvent(e.target, e.target, data.LightningDmg, e.card, 3)
		g.events.insert(g.index, newBeforeDamageEvent(dmg), dmg, newAfterDmgEvent(dmg))
		g.useSkill(e.target, data.DrawLightning)
		c := g.players[e.target].judgeSlot[data.LightningSlot]
		g.players[e.target].judgeSlot[data.LightningSlot] = nil
		if c.getID() != 0 {
			g.useCard(e.target, c.getID())
			g.dropCards(c.getID())
		} else {
			g.useTmpCard(e.target, c.getName(), c.getDecor(), c.getNum(), data.VirtualCard)
		}
		e.end.setSkip(true)
	}
}

// 延时锦囊牌结束阶段
type dealyTipsEndEvent struct {
	event
	target data.PID
	slot   data.JudgeSlot
}

func newDealyTipsEndEvent(target data.PID, slot data.JudgeSlot) *dealyTipsEndEvent {
	return &dealyTipsEndEvent{target: target, slot: slot}
}

func (e *dealyTipsEndEvent) trigger(g *Games) {
	t := g.players[e.target]
	c := t.judgeSlot[e.slot]
	if e.slot == data.LightningSlot {
		p := g.getNextPid(e.target)
		for {
			if g.players[p].judgeSlot[data.LightningSlot] == nil && !g.players[p].hasEffect(guiMeiEffect) {
				g.players[p].judgeSlot[data.LightningSlot] = t.judgeSlot[e.slot]
				t.judgeSlot[data.LightningSlot] = nil
				if c.getID() != 0 {
					g.useCard(e.target, c.getID(), p)
				} else {
					g.useTmpCard(e.target, data.Lightning, c.getDecor(), data.CNum(c.getDecor()), data.VirtualCard, p)
				}
				return
			}
			p = g.getNextPid(p)
			if p == e.target {
				return
			}
		}
	} else {
		t.judgeSlot[e.slot] = nil
		if c.getID() != 0 {
			g.useCard(e.target, c.getID())
			g.dropCards(c.getID())
		} else {
			g.useTmpCard(e.target, c.getName(), c.getDecor(), data.CNum(c.getDecor()), data.VirtualCard)
		}
	}
}
