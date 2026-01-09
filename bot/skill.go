package bot

import (
	"goltk/data"
	"sort"
)

type skillI interface {
	getID() data.SID        //返回技能id
	handleSelect(*Bot)      //处理技能选择阶段
	handleActive(*Bot) bool //思考主动技,返回值代表是否结束当前阶段
}

type skill struct {
	id data.SID
}

func newSkill(id data.SID) *skill {
	return &skill{id: id}
}

func (s *skill) getID() data.SID {
	return s.id
}

func (s *skill) handleSelect(*Bot) {}

func (s *skill) handleActive(*Bot) bool {
	return false
}

// 武圣
type wushengSkill struct {
	skill
}

// 获得最适合用来转化的卡
func (s *wushengSkill) getBestCard(cards []data.CID) data.CID {
	sort.Slice(cards, func(i, j int) bool { return cards[i] < cards[j] })
	return cards[len(cards)-1]
}

func (s *wushengSkill) handleActive(b *Bot) bool {
	b.useSkillInf <- data.UseSkillInf{ID: data.WuShengSkill, Args: []byte{0}}
	rsp := <-b.useSkillRspRec
	//如果没有可用牌则不使用技能
	if compareArray(rsp.Cards, []data.CID{0, 0, 0, 0}) {
		return false
	}
	//如果身上有杀也不用技能
	for _, c := range b.players[b.pid].cards {
		if isItemInList([]data.CardName{data.Attack, data.LightnAttack, data.FireAttack}, b.cards[c].Name) {
			return false
		}
	}
	srcCard := s.getBestCard(rsp.Cards)
	switch b.state {
	case data.UseCardState:
		enemys := b.getEnemy(rsp.Targets...)
		//检查藤甲
		for i := len(enemys) - 1; i >= 0; i-- {
			if b.players[enemys[i]].equips[data.ArmorSlot] != nil {
				if b.players[enemys[i]].equips[data.ArmorSlot].Name == data.TengJia {
					enemys = append(enemys[:i], enemys[i+1:]...)
				}
			}
		}
		if len(enemys) == 0 {
			return false
		}
		bestTarget := enemys[0]
		for _, p := range enemys {
			if b.players[bestTarget].hp > b.players[p].hp {
				bestTarget = p
			}
		}
		b.useSkillInf <- data.UseSkillInf{ID: data.WuShengSkill, TargetList: []data.PID{bestTarget},
			Cards: []data.CID{srcCard}, Args: []byte{1}}
		return true
	}
	return false
}

func newSkillI(id data.SID) skillI {
	switch id {
	case data.WuShengSkill:
		return &wushengSkill{skill: *newSkill(data.WuShengSkill)}
	default:
		return newSkill(id)
	}
}
