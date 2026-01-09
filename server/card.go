package server

import (
	"goltk/data"
)

type cardI interface {
	use(g *Games, user data.PID, target ...data.PID)
	getID() data.CID
	useAble(*Games, data.PID) bool //检测卡在当前事件中是否可用
	getName() data.CardName
	getDecor() data.Decor
	getNum() data.CNum
	getType() data.CardType
	getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8)
}

type card struct {
	data.Card
}

func (c *card) use(g *Games, user data.PID, target ...data.PID) {}

func (c *card) useAble(g *Games, pid data.PID) bool {
	if isItemInList(g.players[pid].unUseableCol, c.Dec) {
		return false
	}
	if _, ok := g.events.list[g.index].(*useCardEvent); ok {
		return true
	}
	if _, ok := g.events.list[g.index].(*dropCardEvent); ok {
		return true
	}
	return false
}

func (c *card) getAvailableTarget(g *Games, user data.PID) ([]data.PID, uint8) {
	return g.getAllAliveOther(user), 1
}

func (c *card) getID() data.CID {
	return c.ID
}

func (c *card) getName() data.CardName {
	return c.Name
}

func (c *card) getDecor() data.Decor {
	return c.Dec
}

func (c *card) getNum() data.CNum {
	return c.Num
}

func (c *card) getType() data.CardType {
	return c.CardType
}

// 根据卡片名字生成对应的卡片
func newCard(c data.Card) cardI {
	switch c.Name {
	case data.Attack:
		return &atkCard{card: card{Card: c}}
	case data.FireAttack:
		return &atkCard{card: card{Card: c}, dmgType: data.FireDmg}
	case data.LightnAttack:
		return &atkCard{card: card{Card: c}, dmgType: data.LightningDmg}
	case data.Dodge:
		return &dodgeCard{card: card{Card: c}}
	case data.Drunk:
		return &drunkCard{card: card{Card: c}}
	case data.Peach:
		return &peachCard{card: card{Card: c}}
	//锦囊牌
	case data.WXKJ:
		return &wxkjcard{card: card{Card: c}}
	case data.WZSY:
		return &wzsyCard{card: card{Card: c}}
	case data.Duel:
		return &duelCard{card: card{Card: c}}
	case data.NMRQ:
		return &nmrqCard{card: card{Card: c}}
	case data.WJQF:
		return &wjqfCard{card: card{Card: c}}
	case data.TYJY:
		return &tyjyCard{card: card{Card: c}}
	case data.WGFD:
		return &wgfdCard{card: card{Card: c}}
	case data.Burn:
		return &burnCard{card: card{Card: c}}
	case data.TSLH:
		return &tslhCard{card: card{Card: c}}
	case data.SSQY:
		return &ssqyCard{card: card{Card: c}}
	case data.GHCQ:
		return &ghcqCard{card: card{Card: c}}
	case data.JDSR:
		return &jdsrCard{card: card{Card: c}}
	case data.LBSS:
		return &lbssCard{card: card{Card: c}}
	case data.BLCD:
		return &blcdCard{card: card{Card: c}}
	case data.Lightning:
		return &lightnCard{card: card{Card: c}}
	}
	switch c.CardType {
	case data.HorseUpCardType:
		return &horseCard{card: card{Card: c}, isUpHorse: true}
	case data.HorseDownCardType:
		return &horseCard{card: card{Card: c}}
	case data.ArmorCardType:
		return &armorCard{card: card{Card: c}}
	case data.WeaponCardType:
		return newWeaponCard(c)
	}
	panic("名为：" + c.Name.String() + " 的卡片类型不存在")
}

// 生成卡片列表
func newCardList() []cardI {
	rawCardList := data.GetCards()
	cardList := make([]cardI, len(rawCardList)+1)
	for i := 0; i < len(rawCardList); i++ {
		cardList[rawCardList[i].ID] = newCard(rawCardList[i])
	}
	return cardList
}
