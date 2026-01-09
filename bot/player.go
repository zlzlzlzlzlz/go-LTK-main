package bot

import "goltk/data"

type player struct {
	pid    data.PID
	hp     data.HP
	maxHp  data.HP
	female bool
	side   data.RoleSide
	cards  []data.CID
	judges [3]data.CID
	equips [4]*data.Card
	death  bool
}

func newPlayer(role data.Role, pid data.PID) player {
	return player{
		pid:    pid,
		hp:     role.MaxHP,
		maxHp:  role.MaxHP,
		female: role.Female,
		side:   role.Side,
	}
}

func (p *player) addCard(cards ...data.CID) {
	p.cards = append(p.cards, cards...)
}

//从玩家区域中移除卡牌
func (p *player) removeCard(cards ...data.CID) {
	//检查手牌堆
	for i := len(p.cards) - 1; i >= 0; i-- {
		if isItemInList(cards, p.cards[i]) {
			p.cards = append(p.cards[:i], p.cards[i+1:]...)
		}
	}
	//检查装备区
	for i := 0; i < len(p.equips); i++ {
		if p.equips[i] == nil {
			continue
		}
		if isItemInList(cards, p.equips[i].ID) {
			p.equips[i] = nil
		}
	}
	//检查判定区
	for i := 0; i < len(p.judges); i++ {
		if isItemInList(cards, p.judges[i]) {
			p.judges[i] = 0
		}
	}
}
