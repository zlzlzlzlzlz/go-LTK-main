package localclient

import (
	"goltk/data"
	"goltk/front"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

type skillBoxI interface {
	Draw(x, y float64, screen *ebiten.Image)
}

type skillBox struct {
	name *front.TextItem2
}

func newSkillBox(sid data.SID) skillBox {
	return skillBox{name: front.NewTextItem2("[white]"+sid.Name(), 0, 0, 24, 4, 28)}
}

func (b *skillBox) Draw(x, y float64, screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(playerImgs.skillBoxImg, op)
	b.name.SetPos(x, y+5)
	b.name.Draw(screen)
}

type skillI interface {
	update(*Games, *player)
	Draw(screen *ebiten.Image)
	deselect(g *Games, p *player)
	setActive(active bool)
	getID() data.SID
	isActiveSkill() bool
	setPos(x, y float64)
	getPos() (x, y float64)
}

type skill struct {
	id      data.SID
	nameImg *ebiten.Image
	x, y    float64
}

func newSkill(id data.SID) skill {
	return skill{id: id, nameImg: getSkillNameImg(id.String())}
}

func (s *skill) update(*Games, *player) {}

func (s *skill) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.x, s.y)
	screen.DrawImage(s.nameImg, op)
}

func (s *skill) deselect(g *Games, p *player) {}

func (s *skill) setActive(active bool) {}

func (s *skill) getID() data.SID {
	return s.id
}

func (s *skill) setPos(x, y float64) {
	s.x, s.y = x, y
}

func (s *skill) getPos() (x, y float64) {
	x, y = s.x, s.y
	return
}

func (s *skill) isActiveSkill() bool {
	return false
}

type activeSkill struct {
	skill
	active     bool
	enable     bool
	btn        *skillBtn
	comfirmBtn buttonI
	cancleBtn  buttonI
}

func (s *activeSkill) Draw(screen *ebiten.Image) {
	s.btn.Draw(screen)
	if s.comfirmBtn != nil {
		s.comfirmBtn.Draw(screen)
	}
	if s.cancleBtn != nil {
		s.cancleBtn.Draw(screen)
	}
}

func (s *activeSkill) isActiveSkill() bool {
	return true
}

func (s *activeSkill) setPos(x, y float64) {
	s.skill.setPos(x, y)
	s.btn.setPos(x, y)
}

func (s *activeSkill) setActive(active bool) {
	s.active = active
	if active {
		s.btn.switch2Active()
	} else {
		s.btn.switch2UnActive()
	}
}

func (s *activeSkill) updateBtnGup(g *Games) {
	s.btn.Update(g)
	if s.comfirmBtn != nil {
		s.comfirmBtn.Update(g)
	}
	if s.cancleBtn != nil {
		s.cancleBtn.Update(g)
	}
}

// 丈八蛇矛
type zbsmSkill struct {
	activeSkill
	target []data.PID
}

func newZBSMSkill(g *Games, owner data.PID) *zbsmSkill {
	s := &zbsmSkill{}
	s.id = data.ZBSMSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.ZBSM.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.ZBSMSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.ZBSMSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		for _, cid := range rsp.Cards {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *zbsmSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			break
		}
		c.selected(p)
		if len(p.selCard) > 2 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
		break
	}
	if len(p.selCard) == 2 {
		if g.state != data.UseCardState {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.ZBSMSkill,
						Cards: append([]data.CID{}, p.selCard...), Args: []byte{1}}
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
			return
		}
		for _, t := range g.playList {
			if !t.isClicked() {
				continue
			}
			if t.isSelected {
				t.isSelected = false
				s.target = nil
				continue
			}
			s.target = append(s.target, t.pid)
			t.isSelected = true
			if len(s.target) > 1 {
				g.getPlayer(s.target[0]).isSelected = false
				s.target = s.target[1:]
			}
		}
		if len(s.target) != 0 {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.ZBSMSkill, Cards: append([]data.CID{}, p.selCard...),
						TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
		} else {
			s.comfirmBtn = nil
		}
	} else {
		s.comfirmBtn = nil
		for _, p := range g.playList {
			p.isSelected = false
		}
		s.target = nil
	}
}

func (s *zbsmSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.ZBSMSkill)
	p.selCard = nil
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 仁德技能
type rendeSkill struct {
	activeSkill
	targets []data.PID
}

func newRenDeSkill(g *Games, owner data.PID) *rendeSkill {
	s := &rendeSkill{}
	s.id = data.RenDeSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.RenDeSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.RenDeSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.enableSkill(g, data.RenDeSkill)
		p.btnGrup = nil
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		for _, cid := range rsp.Cards {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *rendeSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
	}
	if len(p.selCard) == 0 {
		s.comfirmBtn = nil
		for _, p1 := range g.playList {
			p1.isSelected = false
		}
		return
	}
	if s.comfirmBtn == nil && len(s.targets) > 0 {
		s.comfirmBtn = newConfirmBtn(func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.RenDeSkill, Cards: append([]data.CID{}, p.selCard...),
				TargetList: []data.PID{s.targets[0]}, Args: []byte{1}}
			s.deselect(g, p)
		})
	}
	for _, t := range g.playList {
		if t.unSelAble {
			continue
		}
		if !t.isClicked() {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.targets = nil
			continue
		}
		t.isSelected = true
		s.targets = append(s.targets, t.pid)
		if len(s.targets) > 1 {
			g.getPlayer(s.targets[0]).isSelected = false
			s.targets = s.targets[1:]
		}
	}
}

func (s *rendeSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.comfirmBtn = nil
	s.cancleBtn = nil
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.RenDeSkill)
	p.selCard = nil
	s.targets = nil
}

// 武圣
type wushengSkill struct {
	activeSkill
	target    []data.PID
	prvLength int
}

func newWuShengSkill(g *Games, owner data.PID) *wushengSkill {
	s := &wushengSkill{}
	s.id = data.WuShengSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.WuShengSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.WuShengSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.WuShengSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.equipSlot[i].setDrawEquipTip(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *wushengSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) > 1 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) == 1 {
		if g.state != data.UseCardState {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.WuShengSkill,
						Cards: append([]data.CID{}, p.selCard...), Args: []byte{1}}
					s.deselect(g, p)
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
			return
		}
		for _, t := range g.playList {
			if !t.isClicked() {
				continue
			}
			if t.isSelected {
				t.isSelected = false
				s.target = nil
				continue
			}
			s.target = append(s.target, t.pid)
			t.isSelected = true
			if len(s.target) > 1 {
				g.getPlayer(s.target[0]).isSelected = false
				s.target = s.target[1:]
			}
		}
		if len(s.target) != 0 {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.WuShengSkill, Cards: append([]data.CID{}, p.selCard...),
						TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
					s.deselect(g, p)
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
		} else {
			s.comfirmBtn = nil
		}
	} else {
		s.comfirmBtn = nil
		for _, p := range g.playList {
			p.isSelected = false
		}
		s.target = nil
	}
}

func (s *wushengSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setDrawEquipTip(false)
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	p.disableSkill(data.WuShengSkill)
	p.selCard = nil
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 龙胆
type longDanSkill struct {
	activeSkill
	target []data.PID
}

func newLongDanSkill(g *Games, owner data.PID) *longDanSkill {
	s := &longDanSkill{}
	s.id = data.LongDanSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.LongDanSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.LongDanSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.LongDanSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		for _, cid := range rsp.Cards {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *longDanSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) > 1 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) == 1 {
		if g.state != data.UseCardState {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.LongDanSkill,
						Cards: append([]data.CID{}, p.selCard...), Args: []byte{1}}
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
			return
		}
		for _, t := range g.playList {
			if !t.isClicked() {
				continue
			}
			if t.isSelected {
				t.isSelected = false
				s.target = nil
				continue
			}
			s.target = append(s.target, t.pid)
			t.isSelected = true
			if len(s.target) > 1 {
				g.getPlayer(s.target[0]).isSelected = false
				s.target = s.target[1:]
			}
		}
		if len(s.target) != 0 {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.LongDanSkill, Cards: append([]data.CID{}, p.selCard...),
						TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
		} else {
			s.comfirmBtn = nil
		}
	} else {
		s.comfirmBtn = nil
		for _, p := range g.playList {
			p.isSelected = false
		}
		s.target = nil
	}
}

func (s *longDanSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.LongDanSkill)
	p.selCard = nil
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

type liMuSkill struct {
	activeSkill
	prvLength int
}

func newLiMuSkill(g *Games, owner data.PID) *liMuSkill {
	s := &liMuSkill{}
	s.id = data.LiMuSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.LiMuSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.LiMuSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		p.btnGrup = nil
		p.enableSkill(g, data.LiMuSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.equipSlot[i].setDrawEquipTip(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *liMuSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) > 1 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) == 1 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.LiMuSkill, Cards: append([]data.CID{}, p.selCard...), Args: []byte{1}}
				s.deselect(g, p)
				removeInf := <-g.removeReceiver
				g.removeCard(removeInf)
				iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
			})
		}

	} else {
		s.comfirmBtn = nil
	}
}

func (s *liMuSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setDrawEquipTip(false)
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	p.disableSkill(data.LiMuSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 业炎技能
type yeyanSkill struct {
	activeSkill
	targets []data.PID
	cards   []cardI
}

func newYeYanSkill(g *Games, owner data.PID) *yeyanSkill {
	s := &yeyanSkill{}
	s.id = data.YeYanSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.YeYanSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.YeYanSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.enableSkill(g, data.YeYanSkill)
		p.btnGrup = nil
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		for _, cid := range rsp.Cards {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *yeyanSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			for i, cc := range s.cards {
				if cc == c {
					s.cards = append(s.cards[:i], s.cards[i+1:]...)
				}
			}
			continue
		}
		c.selected(p)
		s.cards = append(s.cards, c)
		if len(s.cards) > 4 {
			s.cards[0].pureDeselect(p)
			s.cards = append([]cardI{}, s.cards[1:]...)
		}
	}
	if len(s.cards) != 4 {
		s.comfirmBtn = nil
		for _, p1 := range g.playList {
			p1.isSelected = false
		}
		return
	}
	if s.comfirmBtn == nil && len(s.targets) > 0 {
		s.comfirmBtn = newConfirmBtn(func(g *Games) {
			list := []data.CID{}
			for _, c := range s.cards {
				list = append(list, c.getID())
			}
			g.useSkillInf <- data.UseSkillInf{ID: data.YeYanSkill, Cards: list,
				TargetList: []data.PID{s.targets[0]}, Args: []byte{1}}
			s.deselect(g, p)
		})
	}
	for _, t := range g.playList {
		if t.unSelAble {
			continue
		}
		if !t.isClicked() {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.targets = nil
			continue
		}
		t.isSelected = true
		s.targets = append(s.targets, t.pid)
		if len(s.targets) > 1 {
			g.getPlayer(s.targets[0]).isSelected = false
			s.targets = s.targets[1:]
		}
	}
}

func (s *yeyanSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.comfirmBtn = nil
	s.cancleBtn = nil
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.YeYanSkill)
	p.selCard = nil
	s.cards = nil
	s.targets = nil
}

// 急救
type jijiuSkill struct {
	activeSkill
	prvLength int
}

func newJiJiuSkill(g *Games, owner data.PID) *jijiuSkill {
	s := &jijiuSkill{}
	s.id = data.JiJiuSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.JiJiuSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.JiJiuSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		p.btnGrup = nil
		p.enableSkill(g, data.JiJiuSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *jijiuSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) > 1 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) == 1 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.JiJiuSkill, Cards: append([]data.CID{}, p.selCard...), Args: []byte{1}}
				s.deselect(g, p)
				removeInf := <-g.removeReceiver
				g.removeCard(removeInf)
				iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
			})
		}

	} else {
		s.comfirmBtn = nil
	}
}

func (s *jijiuSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	p.disableSkill(data.JiJiuSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

type qingnangSkill struct {
	activeSkill
	target []data.PID
}

func newQingNangSkill(g *Games, owner data.PID) *qingnangSkill {
	s := &qingnangSkill{}
	s.id = data.QingNangSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.QingNangSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.QingNangSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		p.btnGrup = nil
		p.enableSkill(g, data.QingNangSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		for _, cid := range rsp.Cards {
			p.findCard(cid).setSelectedAble(true)
		}
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
	})
	return s
}

func (s *qingnangSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) > 1 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) == 0 {
		s.comfirmBtn = nil
		for _, p1 := range g.playList {
			p1.isSelected = false
		}
		return
	}
	if s.comfirmBtn == nil && len(s.target) > 0 {
		s.comfirmBtn = newConfirmBtn(func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.QingNangSkill, Cards: append([]data.CID{}, p.selCard...),
				TargetList: []data.PID{s.target[0]}, Args: []byte{1}}
			s.deselect(g, p)
		})
	}
	for _, t := range g.playList {
		if t.unSelAble {
			continue
		}
		if !t.isClicked() {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.target = nil
			continue
		}
		t.isSelected = true
		s.target = append(s.target, t.pid)
		if len(s.target) > 1 {
			g.getPlayer(s.target[0]).isSelected = false
			s.target = s.target[1:]
		}
	}
}

func (s *qingnangSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.comfirmBtn = nil
	s.cancleBtn = nil
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.QingNangSkill)
	p.selCard = nil
	s.target = nil
}

// 结姻技能
type jieyinSkill struct {
	activeSkill
	targets []data.PID
}

func newJieYinSkill(g *Games, owner data.PID) *jieyinSkill {
	s := &jieyinSkill{}
	s.id = data.JieYinSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.JieYinSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.JieYinSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.enableSkill(g, data.JieYinSkill)
		p.btnGrup = nil
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		for _, cid := range rsp.Cards {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *jieyinSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) == 3 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) != 2 {
		s.comfirmBtn = nil
		for _, p1 := range g.playList {
			p1.isSelected = false
		}
		return
	}
	if s.comfirmBtn == nil && len(s.targets) > 0 {
		s.comfirmBtn = newConfirmBtn(func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.JieYinSkill, Cards: append([]data.CID{}, p.selCard...),
				TargetList: []data.PID{s.targets[0]}, Args: []byte{1}}
			s.deselect(g, p)
		})
	}
	for _, t := range g.playList {
		if t.unSelAble {
			continue
		}
		if !t.isClicked() {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.targets = nil
			continue
		}
		t.isSelected = true
		s.targets = append(s.targets, t.pid)
		if len(s.targets) > 1 {
			g.getPlayer(s.targets[0]).isSelected = false
			s.targets = s.targets[1:]
		}
	}
}

func (s *jieyinSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.comfirmBtn = nil
	s.cancleBtn = nil
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.JieYinSkill)
	p.selCard = nil
	s.targets = nil
}

// 灭吴
type miewuSkill struct {
	activeSkill
	cardBtnList []buttonI
	showCardBtn bool
	srcCard     cardI         //用来转化的卡
	virtualCard data.CardName //当前选择的虚拟卡
	cardButton  []buttonI     //虚拟卡持有的按钮
	targetNum   int
	targets     []data.PID //选择的目标列表
	needTarget  bool
	prvLength   int
	gaint       bool //是否已经获得该技能
}

func newMieWuSkill(g *Games, owner data.PID) *miewuSkill {
	s := &miewuSkill{}
	s.id = data.MieWuSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.MieWuSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		p.enableSkill(g, data.MieWuSkill)
		p.btnGrup = nil
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].setDrawEquipTip(true)
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
		switch g.state {
		case data.UseCardState:
			noTargetCards := []data.CardName{data.WGFD, data.WJQF, data.NMRQ, data.WZSY, data.WXKJ, data.Dodge,
				data.Peach, data.Drunk, data.TYJY, data.Lightning}
			needTargetCards := []data.CardName{data.Attack, data.LightnAttack, data.FireAttack,
				data.TSLH, data.Duel, data.Burn, data.SSQY, data.GHCQ, data.JDSR, data.BLCD, data.LBSS}
			getOnclick := func(cname data.CardName) func(*Games) {
				return func(g *Games) {
					s.virtualCard = cname
					s.srcCard.setVirtualText(cname)
					s.showCardBtn = false
					if isItemInList(noTargetCards, cname) {
						s.cardButton = []buttonI{
							newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) }),
							newConfirmBtn(func(g *Games) {
								g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, Cards: []data.CID{s.srcCard.getID()},
									TargetList: []data.PID{owner}, Args: []byte{1, byte(cname)}}
								s.deselect(g, g.getPlayer(owner))
							}),
						}
						return
					}
					if isItemInList(needTargetCards, cname) {
						s.needTarget = true
						s.cardButton = []buttonI{newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })}
						g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, Cards: []data.CID{s.srcCard.getID()},
							Args: []byte{2, byte(cname)}}
						rsp := <-g.useSkillRspRec
						s.targetNum = int(rsp.Args[0])
						for _, p := range g.playList {
							p.unSelAble = true
							p.isSelected = false
						}
						for _, pid := range rsp.Targets {
							g.getPlayer(pid).unSelAble = false
						}
						return
					}
				}
			}
			s.cardButton = nil
			s.needTarget = false
			s.virtualCard = data.NoName
			x, y := 140., 260.
			//取出arg中可用卡名
			rspList := []data.CardName{}
			for _, cname := range rsp.Args {
				cname := data.CardName(cname)
				rspList = append(rspList, cname)
			}
			//制作卡名总列表
			cnameList := append(needTargetCards, noTargetCards...)
			for i, cnameb := range cnameList {
				cname := data.CardName(cnameb)
				var btn button
				if isItemInList(rspList, cname) {
					btn = newButton(getImg("assets/virtualCard/"+cname.String()+".png"), x, y, getOnclick(cname))
				} else {
					btn = newButton(getImg("assets/virtualCard/"+cname.String()+"1.png"), x, y, nil)
				}
				s.cardBtnList = append(s.cardBtnList, &btn)
				if i%7 == 6 {
					x = 140.
					y += 80
				} else {
					x += 146
				}
			}
		default:
			if s.cardButton == nil {
				s.cardButton = append(s.cardButton, newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) }))
			}
			return
		}
	})
	return s
}

func (s *miewuSkill) Draw(screen *ebiten.Image) {
	if !s.gaint {
		return
	}
	op := &ebiten.DrawImageOptions{}
	s.activeSkill.Draw(screen)
	if s.showCardBtn {
		op.GeoM.Reset()
		op.GeoM.Translate(0, 180)
		screen.DrawImage(getImg("assets/game/gameSkillBg.png"), op)
		for i := 0; i < len(s.cardBtnList); i++ {
			s.cardBtnList[i].Draw(screen)
		}
	}
	for i := 0; i < len(s.cardButton); i++ {
		s.cardButton[i].Draw(screen)
	}
}

func (s *miewuSkill) update(g *Games, p *player) {
	if !s.gaint {
		return
	}
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	if s.showCardBtn {
		for i := 0; i < len(s.cardBtnList); i++ {
			s.cardBtnList[i].Update(g)
		}
	}
	for i := 0; i < len(s.cardButton); i++ {
		s.cardButton[i].Update(g)
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			s.deselect(g, p)
			s.srcCard = nil
			break
		}
		if s.srcCard == nil {
			c.selected(p)
			s.srcCard = c
		} else {
			s.srcCard = nil
			s.deselect(g, p)
			break
		}
		//只在出牌阶段且第一次选源卡设定
		if g.state == data.UseCardState {
			s.showCardBtn = true
		}
		break
	}
	// 出牌外阶段
	//高危险
	if g.state != data.UseCardState {
		if len(s.cardButton) == 1 && s.srcCard != nil {
			s.cardButton = append(s.cardButton, newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, Cards: []data.CID{s.srcCard.getID()},
					Args: []byte{1, byte(data.NoName)}}
				s.deselect(g, p)
			}))
		}
		return
	}
	//未选虚拟目标结束
	if s.virtualCard == data.NoName {
		return
	}
	switch s.virtualCard {
	default:
		//不需要目标
		if !s.needTarget {
			return
		}
		//借刀杀人
		// /高危险
		if s.virtualCard == data.JDSR {
			if len(s.targets) == 0 {
				for _, p1 := range g.playList {
					if !p1.isClicked() {
						continue
					}
					if p1.equipSlot[data.WeaponSlot] == nil || p1.equipSlot[data.WeaponSlot].getID() == 0 {
						continue
					}
					p1.isSelected = true
					s.targets = append(s.targets, p1.pid)
					g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, TargetList: []data.PID{p1.pid}, Args: []byte{3}}
					rsp := <-g.useSkillRspRec
					for _, p2 := range g.playList {
						if isItemInList(rsp.Targets, p2.pid) {
							p2.unSelAble = false
						} else {
							p2.unSelAble = true
						}
					}
				}
			} else if len(s.targets) == 1 {
				for _, p2 := range g.playList {
					if p2.unSelAble || !p2.isClicked() {
						continue
					}
					p2.isSelected = true
					s.targets = append(s.targets, p2.pid)
					if len(s.cardButton) == 1 {
						s.cardButton = append(s.cardButton, newConfirmBtn(func(g *Games) {
							g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, Cards: []data.CID{s.srcCard.getID()},
								TargetList: s.targets, Args: []byte{1, byte(s.virtualCard)}}
							s.deselect(g, p)
						}))
					}
					return
				}
			} else {
				for _, p2 := range g.playList {
					if p2.unSelAble || !p2.isClicked() {
						continue
					}
					g.getPlayer(s.targets[1]).isSelected = false
					s.targets = s.targets[:1]
					p2.isSelected = true
					s.targets = append(s.targets, p2.pid)
					return
				}
			}
			return
		}
		//需要目标
		for _, p1 := range g.playList {
			if p1.unSelAble || !p1.isClicked() {
				continue
			}
			if p1.isSelected {
				p1.isSelected = false
				for i := 0; i < len(s.targets); i++ {
					if s.targets[i] == p1.pid {
						s.targets = append(s.targets[:i], s.targets[i+1:]...)
						break
					}
				}
				continue
			}
			p1.isSelected = true
			s.targets = append(s.targets, p1.pid)
			if len(s.targets) > s.targetNum {
				g.getPlayer(s.targets[0]).isSelected = false
				s.targets = s.targets[1:]
			}
			if len(s.targets) > 0 {
				if len(s.cardButton) == 1 {
					s.cardButton = append(s.cardButton, newConfirmBtn(func(g *Games) {
						g.useSkillInf <- data.UseSkillInf{ID: data.MieWuSkill, Cards: []data.CID{s.srcCard.getID()},
							TargetList: s.targets, Args: []byte{1, byte(s.virtualCard)}}
						s.deselect(g, p)
					}))
				}
			} else {
				s.cardButton = s.cardButton[1:]
			}
		}
	}
}

func (s *miewuSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setDrawEquipTip(false)
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	for _, p1 := range g.playList {
		p1.unSelAble = false
		p1.isSelected = false
	}
	if s.srcCard != nil {
		s.srcCard.setVirtualText(data.NoName)
	}
	p.disableSkill(data.MieWuSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
	s.cardBtnList = nil
	s.showCardBtn = false
	s.cardButton = nil
	s.needTarget = false
	s.virtualCard = data.NoName
	s.srcCard = nil
	s.targets = nil
}

// 成略
type chengLueSkill struct {
	activeSkill
}

func newChenghLueSkill(g *Games, owner data.PID) *chengLueSkill {
	s := &chengLueSkill{}
	s.id = data.ChengLueSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.ChengLueSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		s.comfirmBtn = newConfirmBtn(func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.ChengLueSkill}
		})
		p.btnGrup = nil
		p.enableSkill(g, data.ChengLueSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
	})
	return s
}

func (s *chengLueSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
}

func (s *chengLueSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.ChengLueSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 乱武
type luanWuSkill struct {
	activeSkill
}

func newLuanWuSkill(g *Games, owner data.PID) *luanWuSkill {
	s := &luanWuSkill{}
	s.id = data.LuanWuSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.LuanWuSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		s.comfirmBtn = newConfirmBtn(func(g *Games) {
			g.useSkillInf <- data.UseSkillInf{ID: data.LuanWuSkill}
		})
		p.btnGrup = nil
		p.enableSkill(g, data.LuanWuSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
	})
	return s
}

func (s *luanWuSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
}

func (s *luanWuSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.LuanWuSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 雪恨
type xuehenSkill struct {
	activeSkill
	target    []data.PID
	prvLength int
}

func newXueHenSkill(g *Games, owner data.PID) *xuehenSkill {
	s := &xuehenSkill{}
	s.id = data.XueHenSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.XueHenSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.XueHenSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		g.selNum = int(rsp.Args[0])
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.XueHenSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *xuehenSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		c.selected(p)
		if len(p.selCard) > 1 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
	}
	if len(p.selCard) == 1 {
		for _, t := range g.playList {
			if !t.isClicked() {
				continue
			}
			if t.isSelected {
				t.isSelected = false
				s.target = nil
				continue
			}
			s.target = append(s.target, t.pid)
			t.isSelected = true
			if len(s.target) > g.selNum {
				g.getPlayer(s.target[0]).isSelected = false
				s.target = s.target[1:]
			}
		}
		if len(s.target) == g.selNum {
			if s.comfirmBtn == nil {
				s.comfirmBtn = newConfirmBtn(func(g *Games) {
					g.useSkillInf <- data.UseSkillInf{ID: data.XueHenSkill, Cards: append([]data.CID{}, p.selCard...),
						TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
					s.deselect(g, p)
					removeInf := <-g.removeReceiver
					g.removeCard(removeInf)
					iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
				})
			}
		} else {
			s.comfirmBtn = nil
		}
	} else {
		s.comfirmBtn = nil
		for _, p := range g.playList {
			p.isSelected = false
		}
		s.target = nil
	}
}

func (s *xuehenSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	p.disableSkill(data.XueHenSkill)
	p.selCard = nil
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 排异
type paiYiSkill struct {
	activeSkill
	target []data.PID
}

func newPaiYiSkill(g *Games, owner data.PID) *paiYiSkill {
	s := &paiYiSkill{}
	s.id = data.PaiYiSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.PaiYiSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.PaiYiSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.PaiYiSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
	})
	return s
}

func (s *paiYiSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, t := range g.playList {
		if !t.isClicked() || t.unSelAble {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.target = nil
			continue
		}
		s.target = append(s.target, t.pid)
		t.isSelected = true
		if len(s.target) > 1 {
			g.getPlayer(s.target[0]).isSelected = false
			s.target = s.target[1:]
		}
	}
	if len(s.target) != 0 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.PaiYiSkill,
					TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
				s.deselect(g, p)
			})
		}
	} else {
		s.comfirmBtn = nil
	}
}

func (s *paiYiSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	p.disableSkill(data.PaiYiSkill)
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

type xionHuoSkill struct {
	activeSkill
	target []data.PID
}

func newXionHuoSkill(g *Games, owner data.PID) *xionHuoSkill {
	s := &xionHuoSkill{}
	s.id = data.XionHuoSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.XionHuoSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.XionHuoSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.XionHuoSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
	})
	return s
}

func (s *xionHuoSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, t := range g.playList {
		if !t.isClicked() || t.unSelAble {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.target = nil
			continue
		}
		s.target = append(s.target, t.pid)
		t.isSelected = true
		if len(s.target) > 1 {
			g.getPlayer(s.target[0]).isSelected = false
			s.target = s.target[1:]
		}
	}
	if len(s.target) != 0 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.XionHuoSkill,
					TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
				s.deselect(g, p)
			})
		}
	} else {
		s.comfirmBtn = nil
	}
}

func (s *xionHuoSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	p.disableSkill(data.XionHuoSkill)
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 椎锋
type zhuiFengSkill struct {
	activeSkill
	target []data.PID
}

func newZhuiFeng(g *Games, owner data.PID) *zhuiFengSkill {
	s := &zhuiFengSkill{}
	s.id = data.ZhuiFengSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.ZhuiFengSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.ZhuiFengSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = !isItemInList(rsp.Targets, t.pid)
		}
		p.btnGrup = nil
		p.enableSkill(g, data.ZhuiFengSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
	})
	return s
}

func (s *zhuiFengSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, t := range g.playList {
		if !t.isClicked() || t.unSelAble {
			continue
		}
		if t.isSelected {
			t.isSelected = false
			s.target = nil
			continue
		}
		s.target = append(s.target, t.pid)
		t.isSelected = true
		if len(s.target) > 1 {
			g.getPlayer(s.target[0]).isSelected = false
			s.target = s.target[1:]
		}
	}
	if len(s.target) != 0 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.ZhuiFengSkill,
					TargetList: append([]data.PID{}, s.target...), Args: []byte{1}}
				s.deselect(g, p)
			})
		}
	} else {
		s.comfirmBtn = nil
	}
}

func (s *zhuiFengSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	p.disableSkill(data.ZhuiFengSkill)
	s.target = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 制衡
type zhihengSkill struct {
	activeSkill
	prvLength int
}

func newZhiHengSkill(g *Games, owner data.PID) *zhihengSkill {
	s := &zhihengSkill{}
	s.id = data.ZhiHengSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.ZhiHengSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.ZhiHengSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		p.btnGrup = nil
		p.enableSkill(g, data.ZhiHengSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.equipSlot[i].setDrawEquipTip(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
	})
	return s
}

func (s *zhihengSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for i := len(p.cards) - 1; i >= 0; i-- {
		c := p.cards[i]
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			break
		}
		c.selected(p)
		break
	}
	if len(p.selCard) != 0 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.ZhiHengSkill,
					Cards: append([]data.CID{}, p.selCard...), Args: []byte{1}}
				s.deselect(g, p)
				removeInf := <-g.removeReceiver
				g.removeCard(removeInf)
				iterators(removeInf.cards, func(c data.CID) { g.moveCard2Drop(c) })
			})
		}
		return
	} else {
		s.comfirmBtn = nil
	}
}

func (s *zhihengSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, p1 := range g.playList {
		p1.isSelected = false
		p1.unSelAble = false
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setDrawEquipTip(false)
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	p.disableSkill(data.ZhiHengSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 燕语
type yanYuSkill struct {
	activeSkill
}

func newYanYuSkill(g *Games, owner data.PID) *yanYuSkill {
	s := &yanYuSkill{}
	s.id = data.YanYuSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.YanYuSkill.String()), func(g *Games) {
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		p := g.getPlayer(owner)
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.YanYuSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		for _, t := range g.playList {
			t.unSelAble = false
		}
		p.btnGrup = nil
		p.enableSkill(g, data.YanYuSkill)
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(isItemInList(rsp.Cards, c.getID()))
		}
	})
	return s
}

func (s *yanYuSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			c.pureDeselect(p)
			continue
		}
		if len(p.selCard) > 0 {
			p.findCard(p.selCard[0]).pureDeselect(p)
		}
		c.selected(p)
		break
	}
	if len(p.selCard) == 1 {
		if s.comfirmBtn == nil {
			s.comfirmBtn = newConfirmBtn(func(g *Games) {
				g.useSkillInf <- data.UseSkillInf{ID: data.YanYuSkill,
					Cards: p.selCard, Args: []byte{1}}
				s.deselect(g, p)
			})
		}
	} else {
		s.comfirmBtn = nil
	}
}

func (s *yanYuSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.disableSkill(data.YanYuSkill)
	s.comfirmBtn = nil
	s.cancleBtn = nil
}

// 渐营
type jianyingSkill struct {
	activeSkill
	cardBtnList []buttonI
	showCardBtn bool
	srcCard     cardI         //用来转化的卡
	virtualCard data.CardName //当前选择的虚拟卡
	cardButton  []buttonI     //虚拟卡持有的按钮
	targets     []data.PID    //选择的目标列表
	needTarget  bool
	prvLength   int
	atkNum      int
}

func newjianYingSkill(g *Games, owner data.PID) *jianyingSkill {
	s := &jianyingSkill{}
	s.id = data.JianYingSkill
	s.btn = newSkillBtn(0, 0, getSkillNameImg(data.JianYingSkill.String()), func(g *Games) {
		p := g.getPlayer(owner)
		if !s.active {
			return
		}
		if s.enable {
			s.deselect(g, g.getPlayer(owner))
			return
		}
		//先关闭所有技能
		for _, s := range p.skills {
			s.deselect(g, p)
		}
		s.enable = true
		s.btn.switch2Selected()
		s.cancleBtn = newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })
		g.useSkillInf <- data.UseSkillInf{ID: data.JianYingSkill, Args: []byte{0}}
		rsp := <-g.useSkillRspRec
		p.enableSkill(g, data.JianYingSkill)
		p.btnGrup = nil
		for _, cid := range p.selCard {
			p.findCard(cid).deSelect(p)
		}
		for _, c := range p.cards {
			c.setSelectedAble(false)
		}
		s.prvLength = len(p.cards)
		for i, cid := range rsp.Cards[:4] {
			if cid == 0 {
				continue
			}
			p.equipSlot[i].setDrawEquipTip(true)
			p.equipSlot[i].resetFadeOut()
			p.equipSlot[i].setVisibility(1)
			p.equipSlot[i].setSelectedAble(true)
			p.cards = append(p.cards, p.equipSlot[i])
		}
		p.calculatePos()
		for _, cid := range rsp.Cards[4:] {
			p.findCard(cid).setSelectedAble(true)
		}
		getOnclick := func(cname data.CardName) func(*Games) {
			return func(g *Games) {
				s.virtualCard = cname
				s.srcCard.setVirtualText(cname)
				s.showCardBtn = false
				if isItemInList([]data.CardName{data.Drunk, data.Peach}, cname) {
					s.cardButton = []buttonI{
						newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) }),
						newConfirmBtn(func(g *Games) {
							g.useSkillInf <- data.UseSkillInf{ID: data.JianYingSkill, Cards: []data.CID{s.srcCard.getID()},
								TargetList: []data.PID{owner}, Args: []byte{1, byte(cname)}}
							s.deselect(g, g.getPlayer(owner))
						}),
					}
					return
				}
				if isItemInList([]data.CardName{data.FireAttack, data.Attack, data.LightnAttack}, cname) {
					s.needTarget = true
					s.cardButton = []buttonI{newCancleBtn(func(g *Games) { s.deselect(g, g.getPlayer(owner)) })}
					g.useSkillInf <- data.UseSkillInf{ID: data.JianYingSkill, Cards: []data.CID{s.srcCard.getID()},
						Args: []byte{2, byte(cname)}}
					rsp := <-g.useSkillRspRec
					s.atkNum = int(rsp.Args[0])
					for _, p := range g.playList {
						p.unSelAble = true
						p.isSelected = false
					}
					for _, pid := range rsp.Targets {
						g.getPlayer(pid).unSelAble = false
					}
					return
				}
			}
		}
		s.cardButton = nil
		s.needTarget = false
		s.virtualCard = data.NoName
		x, y := 140., 260.
		//取出arg中可用卡名
		rspList := []data.CardName{}
		for _, cname := range rsp.Args {
			cname := data.CardName(cname)
			rspList = append(rspList, cname)
		}
		//制作卡名总列表
		cnameList := []data.CardName{data.Attack, data.FireAttack, data.LightnAttack, data.Peach, data.Drunk, data.Dodge}
		for i, cnameb := range cnameList {
			cname := data.CardName(cnameb)
			var btn button
			if isItemInList(rspList, cname) {
				btn = newButton(getImg("assets/virtualCard/"+cname.String()+".png"), x, y, getOnclick(cname))
			} else {
				btn = newButton(getImg("assets/virtualCard/"+cname.String()+"1.png"), x, y, nil)
			}
			s.cardBtnList = append(s.cardBtnList, &btn)
			if i%7 == 6 {
				x = 140.
				y += 80
			} else {
				x += 146
			}
		}
	})
	return s
}

func (s *jianyingSkill) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	s.activeSkill.Draw(screen)
	if s.showCardBtn {
		op.GeoM.Reset()
		op.GeoM.Translate(0, 180)
		screen.DrawImage(getImg("assets/game/gameSkillBg.png"), op)
		for i := 0; i < len(s.cardBtnList); i++ {
			s.cardBtnList[i].Draw(screen)
		}
	}
	for i := 0; i < len(s.cardButton); i++ {
		s.cardButton[i].Draw(screen)
	}
}

func (s *jianyingSkill) update(g *Games, p *player) {
	s.updateBtnGup(g)
	if !s.enable {
		return
	}
	if s.showCardBtn {
		for i := 0; i < len(s.cardBtnList); i++ {
			s.cardBtnList[i].Update(g)
		}
	}
	for i := 0; i < len(s.cardButton); i++ {
		s.cardButton[i].Update(g)
	}
	for _, c := range p.cards {
		if !c.isSeleable() {
			continue
		}
		if !c.isClicked() {
			continue
		}
		if c.isSelect() {
			s.deselect(g, p)
			s.srcCard = nil
			break
		}
		if s.srcCard == nil {
			c.selected(p)
			s.srcCard = c
		} else {
			s.srcCard = nil
			s.deselect(g, p)
			break
		}
		//只在出牌阶段且第一次选源卡设定
		if g.state == data.UseCardState {
			s.showCardBtn = true
		}
		break
	}
	//未选虚拟目标结束
	if s.virtualCard == data.NoName {
		return
	}
	switch s.virtualCard {
	default:
		//不需要目标
		if !s.needTarget {
			return
		}
		//需要目标
		for _, p1 := range g.playList {
			if p1.unSelAble || !p1.isClicked() {
				continue
			}
			if p1.isSelected {
				p1.isSelected = false
				s.targets = []data.PID{}
				continue
			}
			p1.isSelected = true
			s.targets = append(s.targets, p1.pid)
			if len(s.targets) > s.atkNum {
				g.getPlayer(s.targets[0]).isSelected = false
				s.targets = s.targets[1:]
			}
			if len(s.targets) > 0 {
				if len(s.cardButton) == 1 {
					s.cardButton = append(s.cardButton, newConfirmBtn(func(g *Games) {
						g.useSkillInf <- data.UseSkillInf{ID: data.JianYingSkill, Cards: []data.CID{s.srcCard.getID()},
							TargetList: s.targets, Args: []byte{1, byte(s.virtualCard)}}
						s.deselect(g, p)
					}))
				}
			} else {
				s.cardButton = s.cardButton[1:]
			}
		}
	}
}

func (s *jianyingSkill) deselect(g *Games, p *player) {
	if !s.enable {
		return
	}
	s.enable = false
	s.btn.switch2Active()
	for _, card := range p.cards {
		card.setSelectedAble(isItemInList(g.useAbleCards, card.getID()))
	}
	for _, cid := range p.selCard {
		p.findCard(cid).pureDeselect(p)
	}
	p.cards = p.cards[:s.prvLength]
	for _, c := range p.equipSlot {
		if c == nil {
			continue
		}
		c.setDrawEquipTip(false)
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
	}
	for _, p1 := range g.playList {
		p1.unSelAble = false
		p1.isSelected = false
	}
	if s.srcCard != nil {
		s.srcCard.setVirtualText(data.NoName)
	}
	p.disableSkill(data.JianYingSkill)
	p.selCard = nil
	s.comfirmBtn = nil
	s.cancleBtn = nil
	s.cardBtnList = nil
	s.showCardBtn = false
	s.cardButton = nil
	s.needTarget = false
	s.virtualCard = data.NoName
	s.srcCard = nil
	s.targets = nil
}

func newSkillI(g *Games, owner data.PID, id data.SID) skillI {
	switch id {
	case data.ZBSMSkill:
		return newZBSMSkill(g, owner)
	case data.RenDeSkill:
		return newRenDeSkill(g, owner)
	case data.WuShengSkill:
		return newWuShengSkill(g, owner)
	case data.JiJiuSkill:
		return newJiJiuSkill(g, owner)
	case data.LongDanSkill:
		return newLongDanSkill(g, owner)
	case data.LiMuSkill:
		return newLiMuSkill(g, owner)
	case data.YeYanSkill:
		return newYeYanSkill(g, owner)
	case data.QingNangSkill:
		return newQingNangSkill(g, owner)
	case data.JieYinSkill:
		return newJieYinSkill(g, owner)
	case data.ChengLueSkill:
		return newChenghLueSkill(g, owner)
	case data.XueHenSkill:
		return newXueHenSkill(g, owner)
	case data.MieWuSkill:
		return newMieWuSkill(g, owner)
	case data.PaiYiSkill:
		return newPaiYiSkill(g, owner)
	case data.XionHuoSkill:
		return newXionHuoSkill(g, owner)
	case data.ZhuiFengSkill:
		return newZhuiFeng(g, owner)
	case data.LuanWuSkill:
		return newLuanWuSkill(g, owner)
	case data.ZhiHengSkill:
		return newZhiHengSkill(g, owner)
	case data.YanYuSkill:
		return newYanYuSkill(g, owner)
	case data.JianYingSkill:
		return newjianYingSkill(g, owner)
	default:
		s := newSkill(id)
		return &s
	}
}

type chenglueSkillBox struct {
	skillBox
	decs []data.Decor
}

func (b *chenglueSkillBox) Draw(x, y float64, screen *ebiten.Image) {
	b.skillBox.Draw(x, y, screen)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x+55, y+8)
	for i := 0; i < len(b.decs); i++ {
		screen.DrawImage(cardImgs.decor[b.decs[i]-1], op)
		op.GeoM.Translate(22, 0)
	}
}

type shiCaiSkillBox struct {
	skillBox
}

func (b *shiCaiSkillBox) addType(ctype data.CardType) {
	var str string
	switch ctype {
	case data.BaseCardType:
		str = " 基"
	case data.TipsCardType:
		str = " 锦"
	case data.WeaponCardType:
		str = " 装"
	case data.ArmorCardType:
		str = " 装"
	case data.HorseDownCardType:
		str = " 装"
	case data.HorseUpCardType:
		str = " 装"
	}
	b.name.SetText(b.name.GetStr() + str)
}

type fengYinSkillBox struct {
	skillBox
}

func (b *fengYinSkillBox) setColor(dec data.Decor) {
	var str string
	if dec.ISBlack() {
		str = " 黑"
	} else if dec.IsRed() {
		str = " 红"
	}
	b.name.SetText("奋音" + str)
}

type jiLiSkillBox struct {
	skillBox
}

func (b *jiLiSkillBox) setNum(num int) {
	b.name.SetText("蒺藜 " + strconv.Itoa(num))
}

type wenJiSkillBox struct {
	skillBox
}

func (b *wenJiSkillBox) setNum(cardName data.CardName) {
	b.name.SetText("问计 " + cardName.ChnName())
}

type qiZhiSkillBox struct {
	skillBox
}

func (b *qiZhiSkillBox) setNum(num int) {
	b.name.SetText("奇制 " + strconv.Itoa(num))
}

type liYuSkillBox struct {
	skillBox
}

func (b *liYuSkillBox) setNum(num int) {
	b.name.SetText("利驭 " + strconv.Itoa(num))
}

type qianxiSkillbox struct {
	skillBox
}

func (b *qianxiSkillbox) setCol(col data.Decor) {
	if col == data.RedCol {
		b.name.SetText("潜袭 红")
	} else {
		b.name.SetText("潜袭 黑")
	}
}

type wukuSkillBox struct {
	skillBox
}

func (b *wukuSkillBox) setNum(num int) {
	b.name.SetText("武库 " + strconv.Itoa(num))
}

type quanJiSkillBox struct {
	skillBox
}

func (b *quanJiSkillBox) setNum(num int) {
	b.name.SetText("权计 " + strconv.Itoa(num))
}

type zhenGuSkillBox struct {
	skillBox
}

type baoliSkillBox struct {
	skillBox
}

func (b *baoliSkillBox) setNum(num int) {
	b.name.SetText("暴戾 " + strconv.Itoa(num))
}

type jiQiaoSkillBox struct {
	skillBox
	count int
}

func (b *jiQiaoSkillBox) setNum(num int) {
	b.name.SetText("藏机 " + strconv.Itoa(num))
}

type jianYingSkillBox struct {
	skillBox
	dec data.Decor
	num data.CNum
}

func (b *jianYingSkillBox) Draw(x, y float64, screen *ebiten.Image) {
	b.skillBox.Draw(x, y, screen)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x+55, y+8)
	screen.DrawImage(cardImgs.decor[b.dec-1], op)
	op.GeoM.Translate(22, 0)
	if b.dec.IsRed() {
		screen.DrawImage(cardImgs.numRed[b.num-1], op)
	} else {
		screen.DrawImage(cardImgs.numblack[b.num-1], op)
	}
}

type yanYuSkillBox struct {
	skillBox
}

func (b *yanYuSkillBox) setNum(num int) {
	b.name.SetText("燕语 " + strconv.Itoa(num))
}

type lueYingSkillBox struct {
	skillBox
}

func (b *lueYingSkillBox) setNum(num int) {
	b.name.SetText("椎 " + strconv.Itoa(num))
}

type shouXiSkillBox struct {
	skillBox
	count int
	sid   data.SID
}

func (b *shouXiSkillBox) setNum(num int) {
	b.count = num
	if b.sid == 0 {
		b.name.SetText("兽 " + strconv.Itoa(num))
	} else {
		b.name.SetText("兽 " + strconv.Itoa(num) + " " + b.sid.Name())
	}
}

func (b *shouXiSkillBox) setAddSkill(sid data.SID) {
	if sid == 0 {
		b.sid = 0
		b.name.SetText("兽 " + strconv.Itoa(b.count))
	} else {
		b.sid = sid
		b.name.SetText("兽 " + strconv.Itoa(b.count) + " " + sid.Name())
	}
}
