package server

import (
	"goltk/data"
	"strconv"
)

type player struct {
	g         *Games
	pid       data.PID
	hp        data.HP
	maxHp     data.HP
	isfemale  bool
	side      data.RoleSide
	death     bool                    //似了没
	cards     []data.CID              //手牌
	skills    []skillI                //技能列表
	equipSlot [4]cardI                //装备槽
	judgeSlot [3]cardI                //判定槽
	tsCards   map[data.SID][]data.CID //临时存储的卡牌
	effectMap map[effect]*struct {
		counter        uint8
		skillEffectMap map[data.SID]uint8
	} //已启用的技能表
	hasAttack    bool         //是否用过杀
	isDrunk      bool         //是否有酒效果
	isLinked     bool         //是否被连环
	hasUseDrunk  bool         //回合是否喝过酒
	turnBack     bool         //翻面
	unUseableCol []data.Decor //不可使用花色
	upDropNum    uint8        //手牌上限减少
	originRole   data.Role
}

func newPlayer(g *Games, role data.Role, pid data.PID) player {
	p := player{
		g:        g,
		pid:      pid,
		hp:       role.MaxHP,
		side:     role.Side,
		maxHp:    role.MaxHP,
		isfemale: role.Female,
		tsCards:  map[data.SID][]data.CID{},
		effectMap: map[effect]*struct {
			counter        uint8
			skillEffectMap map[data.SID]uint8
		}{},
		originRole: role,
	}
	return p
}

// 向玩家手牌堆中加卡
func (p *player) addCard(id ...data.CID) {
	//debug 检查要添加的卡片
	testMap := map[data.CID]struct{}{}
	for _, c := range id {
		if c == 0 {
			panic("")
		}
		testMap[c] = struct{}{}
	}
	if len(testMap) != len(id) {
		panic("")
	}
	testMap = map[data.CID]struct{}{}
	for _, c := range p.cards {
		testMap[c] = struct{}{}
	}
	for _, equip := range p.equipSlot {
		if equip == nil || equip.getID() == 0 {
			continue
		}
		testMap[equip.getID()] = struct{}{}
	}
	for _, judge := range p.judgeSlot {
		if judge == nil || judge.getID() == 0 {
			continue
		}
		testMap[judge.getID()] = struct{}{}
	}
	for _, cards := range p.tsCards {
		for _, c := range cards {
			testMap[c] = struct{}{}
		}
	}
	for _, c := range id {
		if _, ok := testMap[c]; ok {
			panic("")
		}
	}
	p.cards = append(p.cards, id...)
	sortList(p.cards)
}

// 添加临时存储的卡牌
func (p *player) addTSCard(sid data.SID, cards ...data.CID) {
	if _, ok := p.tsCards[sid]; ok {
		p.tsCards[sid] = append(p.tsCards[sid], cards...)
		return
	}
	p.tsCards[sid] = cards
}

// 获取并删除临时存储的卡牌最后一张
func (p *player) delTsCardBottom(sid data.SID) data.CID {
	card := p.tsCards[sid][len(p.tsCards[sid])-1]
	p.tsCards[sid] = p.tsCards[sid][:len(p.tsCards[sid])-1]
	return card
}

// 获取临时存储的卡牌(将会删除)
func (p *player) getTSCard(sid data.SID) []data.CID {
	cards := p.tsCards[sid]
	delete(p.tsCards, sid)
	return cards
}

// 从玩家区域中删除卡
func (p *player) delCard(IDs ...data.CID) {
	for _, id := range IDs {
		//检查手牌堆
		for i := 0; i < len(p.cards); i++ {
			if p.cards[i] == id {
				p.cards = append(p.cards[:i], p.cards[i+1:]...)
				//殇逝
				if p.hasEffect(shangShiEffect) && len(p.cards) < int(p.maxHp-p.hp) {
					p.g.sendCard2Player(p.pid, p.g.getCardsFromTop(int(p.maxHp-p.hp)-len(p.cards))...)
				}
				//决策
				if p.g.players[p.g.turnOwner].hasEffect(jueceEffect) && len(p.cards) == 0 && p.g.turnOwner != p.pid {
					p.g.events.insert(p.g.index, newSkillSelectEvent(data.JueCeSkill, p.g.turnOwner, []data.PID{p.pid}))
				}
				//克己
				if p.hasEffect(kejiEffect) && p.g.turnOwner != p.pid {
					if p.findSkill(data.KeJiSkill).(*keJiSkill).count < 2 {
						p.g.events.insert(p.g.index, newSkillSelectEvent(data.KeJiSkill, p.pid, nil))
					}
				}
				goto loop
			}
		}
		//检查判定区
		for i := 0; i < len(p.judgeSlot); i++ {
			if p.judgeSlot[i] != nil && p.judgeSlot[i].getID() == id {
				p.judgeSlot[i] = nil
				goto loop
			}
		}
		//检查装备区
		for i := 0; i < len(p.equipSlot); i++ {
			if p.equipSlot[i] == nil || p.equipSlot[i].getID() == 0 {
				continue
			}
			if p.equipSlot[i].getID() == id {
				if p.hasEffect(byszEffect) && p.equipSlot[i].getName() == data.BYSZ {
					p.g.recover(p.pid, p.pid, 1)
				}
				p.disableEffect(equipEffectMap[p.g.cards[id].getName()])
				if skill, ok := equipSkillMap[p.equipSlot[i].getName()]; ok {
					p.removeSkill(skill)
				}
				p.equipSlot[i] = nil
				//枭姬
				if p.hasEffect(xiaojiEffect) {
					p.g.sendCard2Player(p.pid, p.g.getCardsFromTop(2)...)
					p.g.useSkill(p.pid, data.XiaoJiSkill)
				}
				//视为拥有装备
				if p.hasEffect(zhangbaEffect) {
					p.findSkill(data.ZhangBaSkill).(*zhangbaSkill).check(p.g, p)
				}
				if p.hasEffect(qingGangEffect) {
					p.findSkill(data.QingGangSkill).(*qinggangSkill).check(p.g, p)
				}
				if p.hasEffect(yanyueEffect) {
					p.findSkill(data.YanYueSkill).(*yanyueSkill).check(p.g, p)
				}
				if p.hasEffect(qilinEffect) {
					p.findSkill(data.QiLinSkill).(*qilinSkill).check(p.g, p)
				}
				if p.hasEffect(zhuQueEffect) {
					p.findSkill(data.ZhuQueSkill).(*zhuQueSkill).check(p.g, p)
				}
				if p.hasEffect(guDingEffect) {
					p.findSkill(data.GuDingSkill).(*gudingSkill).check(p.g, p)
				}
				if p.hasEffect(fangTianEffect) {
					p.findSkill(data.FangTianSkill).(*fangTianSkill).check(p.g, p)
				}
				if p.hasEffect(baZhenEffect) {
					p.findSkill(data.BaZhenSkill).(*bazhenSkill).check(p.g, p)
				}
				if p.hasEffect(manjiaEffect) {
					p.findSkill(data.ManJiaSkill).(*manjiaSkill).check(p.g, p)
				}
				goto loop
			}
		}
		panic("玩家区域没有id=" + strconv.Itoa(int(id)) + "的卡牌")
	loop:
	}
	//魔炎
	if p.hasEffect(moYanEffect) && p.g.turnOwner != p.pid {
		p.g.addPriorityEvent(newSkillSelectEvent(data.MoYanSkill, p.pid, nil))
	}
	//丹术
	if p.hasEffect(danshueffect) && p.g.turnOwner != p.pid {
		p.g.addPriorityEvent(newSkillSelectEvent(data.DanShuSkill, p.pid, nil))
	}
	sortList(p.cards)
}

// 检测玩家区域是否有牌并弃置
func (p *player) cleanCards(g *Games) {
	cards := []data.CID{}
	cards = append(cards, p.cards...)
	p.cards = nil
	//清空装备槽
	for i := 0; i < len(p.equipSlot); i++ {
		if p.equipSlot[i] == nil || p.equipSlot[i].getID() == 0 {
			continue
		}
		cards = append(cards, p.equipSlot[i].getID())
		p.disableEffect(equipEffectMap[p.equipSlot[i].getName()], equipSkillMap[p.equipSlot[i].getName()])
		p.equipSlot[i] = nil
	}
	//清空判定区
	for i := 0; i < len(p.judgeSlot); i++ {
		if p.judgeSlot[i] == nil || p.judgeSlot[i].getID() == 0 {
			continue
		}
		cards = append(cards, p.judgeSlot[i].getID())
		p.judgeSlot[i] = nil
	}
	//清空暂存区
	for _, card := range p.tsCards {
		cards = append(cards, card...)
	}
	p.tsCards = map[data.SID][]data.CID{}
	g.dropCards(cards...)
	iterators(g.clients, func(c clientI) { c.RemoveCard(p.pid, cards...) })
}

// 检测玩家区域是否有牌并弃置(不弃置暂存区)
func (p *player) dropAllCards(g *Games) {
	cards := []data.CID{}
	cards = append(cards, p.cards...)
	p.cards = nil
	//清空装备槽
	for i := 0; i < len(p.equipSlot); i++ {
		if p.equipSlot[i] == nil || p.equipSlot[i].getID() == 0 {
			continue
		}
		cards = append(cards, p.equipSlot[i].getID())
		//枭姬
		if p.hasEffect(xiaojiEffect) {
			g.sendCard2Player(p.pid, g.getCardsFromTop(2)...)
			g.useSkill(p.pid, data.XiaoJiSkill)
		}
		p.disableEffect(equipEffectMap[p.equipSlot[i].getName()], equipSkillMap[p.equipSlot[i].getName()])
		p.equipSlot[i] = nil
	}
	//清空判定区
	for i := 0; i < len(p.judgeSlot); i++ {
		if p.judgeSlot[i] == nil || p.judgeSlot[i].getID() == 0 {
			continue
		}
		cards = append(cards, p.judgeSlot[i].getID())
		p.judgeSlot[i] = nil
	}
	g.dropCards(cards...)
	iterators(g.clients, func(c clientI) { c.RemoveCard(p.pid, cards...) })
}

// 获取可用的卡列表
func (p *player) getUseAbleCards(g *Games) (list []data.CID) {
	for i := 0; i < len(p.cards); i++ {
		if g.cards[p.cards[i]].useAble(g, p.pid) {
			list = append(list, p.cards[i])
		}
	}
	return
}

func (p *player) addSkill(id data.SID) {
	if eff, ok := skillEffectMap[id]; ok {
		p.enableEffect(eff, id)
	}
	if id >= data.SkillListEndPos {
		return
	}
	s := newSkill(id)
	p.skills = append(p.skills, s)
	s.init(p.g, p)
}

func (p *player) findSkill(id data.SID) skillI {
	for _, s := range p.skills {
		if s.getID() == id {
			return s
		}
	}
	panic("玩家没有id=" + strconv.Itoa(int(id)) + "的技能")
}

func (p *player) hasSkill(id data.SID) bool {
	for _, s := range p.skills {
		if s.getID() == id {
			return true
		}
	}
	return false
}

func (p *player) removeSkill(id data.SID) {
	if eff, ok := skillEffectMap[id]; ok {
		p.disableEffect(eff, id)
	}
	for i := 0; i < len(p.skills); i++ {
		if p.skills[i].getID() == id {
			p.skills = append(p.skills[:i], p.skills[i+1:]...)
			return
		}
	}
}

func (p *player) getUseAbleSkill(g *Games) (list []data.SID) {
	for i := 0; i < len(p.skills); i++ {
		if p.skills[i].isUseAble(g, p) {
			list = append(list, p.skills[i].getID())
		}
	}
	return
}

// 启用效果
func (p *player) enableEffect(effect effect, skill ...data.SID) {
	if effect == noEffect {
		panic("没有这个效果")
	}
	if inf, ok := p.effectMap[effect]; ok {
		inf.counter += 1
		for _, s := range skill {
			if _, ok := inf.skillEffectMap[s]; ok {
				inf.skillEffectMap[s]++
			} else {
				inf.skillEffectMap[s] = 1
			}
		}
		return
	}
	inf := &struct {
		counter        uint8
		skillEffectMap map[data.SID]uint8
	}{
		counter:        1,
		skillEffectMap: map[data.SID]uint8{},
	}
	for _, s := range skill {
		inf.skillEffectMap[s] = 1
	}
	p.effectMap[effect] = inf
}

// 关闭效果
func (p *player) disableEffect(effect effect, skill ...data.SID) {
	if _, ok := p.effectMap[effect]; !ok {
		return
	}
	inf := p.effectMap[effect]
	if inf.counter > 1 {
		p.effectMap[effect].counter--
		for _, s := range skill {
			if _, ok := inf.skillEffectMap[s]; ok {
				if inf.skillEffectMap[s] > 1 {
					inf.skillEffectMap[s]--
				} else {
					delete(inf.skillEffectMap, s)
				}
			}
		}
		return
	}
	delete(p.effectMap, effect)
}

// 是否至少拥有效果中的一个
func (p *player) hasEffect(effects ...effect) bool {
	for i := 0; i < len(effects); i++ {
		_, ok := p.effectMap[effects[i]]
		if ok {
			return true
		}
	}
	return false
}

// 获取效果数量
func (p *player) getEffectCount(effect effect) uint8 {
	if inf, ok := p.effectMap[effect]; ok {
		return inf.counter
	}
	return 0
}

// 返回添加改效果的技能
func (p *player) getEffectSkill(effect effect) data.SID {
	if inf, ok := p.effectMap[effect]; ok {
		if len(inf.skillEffectMap) == 0 {
			panic("效果" + strconv.Itoa(int(effect)) + "没有对应的技能")
		}
		for sid := range inf.skillEffectMap {
			return sid
		}
	}
	panic("效果" + strconv.Itoa(int(effect)) + "未启用")
}

func (p *player) getEquipCount() (count int) {
	for _, equip := range p.equipSlot {
		if equip != nil && equip.getID() != 0 {
			count++
		}
	}
	return
}

func (p *player) getAtkDst() (dst distence) {
	if p.equipSlot[data.WeaponSlot] != nil {
		dst = p.equipSlot[data.WeaponSlot].(*weaponCard).dst
		return
	}
	return 1
}
