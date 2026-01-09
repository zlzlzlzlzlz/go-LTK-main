package localclient

import (
	"goltk/data"
	"goltk/front"
	"goltk/sound"
	"image"
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

var cardImgBuf = map[string]*ebiten.Image{}

var cardImgs = struct {
	bg         *ebiten.Image
	nameBg     *ebiten.Image
	tips       *ebiten.Image
	equip      *ebiten.Image
	decor      [7]*ebiten.Image
	numRed     [13]*ebiten.Image
	numblack   [13]*ebiten.Image
	scope      [10]*ebiten.Image
	smallScope [10]*ebiten.Image
	att        *ebiten.Image //-
	def        *ebiten.Image //+
	smallAtt   *ebiten.Image
	smallDef   *ebiten.Image
	decANum    [4][13]*ebiten.Image
	selected   *ebiten.Image
	back       *ebiten.Image
	selLine    *ebiten.Image
	dst        *ebiten.Image
}{
	bg:     loadImg("assets/card/bg.png"),
	nameBg: loadImg("assets/card/nameBg.png"),
	tips:   loadImg("assets/card/tips.png"),
	equip:  loadImg("assets/card/equip.png"),
	decor: [7]*ebiten.Image{
		loadImg("assets/card/dec/spade.png"),
		loadImg("assets/card/dec/heart.png"),
		loadImg("assets/card/dec/club.png"),
		loadImg("assets/card/dec/diamond.png"),
		loadImg("assets/card/dec/nocolor.png"),
		loadImg("assets/card/dec/recolor.png"),
		loadImg("assets/card/dec/blackcolor.png"),
	},
	selected: loadImg("assets/card/selected.png"),
	back:     loadImg("assets/card/back.png"),
	selLine:  loadImg("assets/item/point.png"),
	dst:      loadImg("assets/card/distance.png"),
	att:      loadImg("assets/card/scope/att.png"),
	def:      loadImg("assets/card/scope/def.png"),
	smallAtt: loadImg("assets/card/smallscope/att.png"),
	smallDef: loadImg("assets/card/smallscope/def.png"),
}

// 为cardImg填充数字
func init() {
	for i := 0; i < 13; i++ {
		cardImgs.numRed[i] = loadImg("assets/card/num/red/" + strconv.Itoa(i+1) + ".png")
		cardImgs.numblack[i] = loadImg("assets/card/num/black/" + strconv.Itoa(i+1) + ".png")
	}
	for i := 0; i < 10; i++ {
		cardImgs.scope[i] = loadImg("assets/card/scope/" + strconv.Itoa(i) + ".png")
		cardImgs.smallScope[i] = loadImg("assets/card/smallscope/" + strconv.Itoa(i) + ".png")
	}
	for i := data.Decor(1); i < 5; i++ {
		for j := 0; j < 13; j++ {
			cardImgs.decANum[i-1][j] = loadImg("assets/card/equip/" + i.String() + strconv.Itoa(j+1) + ".png")
		}
	}

}

// 获取卡片图像
func getCardImg(path string) *ebiten.Image {
	img, ok := cardImgBuf[path]
	if ok {
		return img
	}
	i := loadImg(path)
	cardImgBuf[path] = i
	return i
}

// 牌堆坐标
var cardHeapPos = struct {
	x, y float64
}{x: 598, y: 284}

type cardI interface {
	Draw(*ebiten.Image)
	getID() data.CID
	getCardType() data.CardType
	getCardName() data.CardName
	getPos() (float64, float64)
	setPos(x, y float64)
	move2Pos(x, y float64, callback func())
	setFadeOut(t int, callback func()) // 设置卡片在t个tick后淡出s
	resetFadeOut()                     //重置淡化动画
	isOnFade() bool                    //检查卡片是否正在淡出
	setVisibility(v float32)           //设置卡片能见度
	setSelectedAble(selAble bool)
	setShowBack(show bool)
	isClicked() bool
	isSelect() bool
	isSeleable() bool
	setSelect(sel bool) //设置卡牌的isSelected属性，无副作用
	selected(p *player)
	deSelect(p *player)
	pureDeselect(p *player) //无副作用的deselect
	getSelectState() bool
	onUse(g *Games, user data.PID, targets ...data.PID)
	handleUseState(g *Games, p *player)
	handleDropState(g *Games, p *player) bool
	handleDying(g *Games, p *player)
	handleWXKJ(g *Games, p *player)
	handleDodge(g *Games, p *player)
	handleDuel(g *Games, p *player)
	handleNMRQ(g *Games, p *player)
	handleWJQF(g *Games, p *player)
	handleWGFD(g *Games)
	handleBurnShow(g *Games, p *player)
	handleBurnDrop(g *Games, p *player)
	handleSSQY(g *Games)
	handleGHCQ(g *Games)
	handleJDSR(g *Games, p *player)
	handleDropSelfAll(g *Games, p *player) bool
	handleGSF(g *Games, p *player) bool
	handleQLG(g *Games)
	handleQLYYD(g *Games, p *player)
	handleLuanWu(g *Games, p *player)
	handleDropOtherCard(g *Games)
	handleWenJi(g *Games, p *player)
	anime()
	virtualClick()
	setVirtualText(data.CardName)
	setDrawEquipTip(bool)
	setInHandleZoon(bool)
	setPromptText(l1, l2 string)
}

// 无需选择目标的卡
type card struct {
	data.Card
	rect
	g          *Games
	img        *ebiten.Image
	nameImg    *ebiten.Image
	visibility float32 //可见度
	x, y       float64
	selectAble bool
	showBack   bool
	isSelected bool
	moveInf    struct {
		counter  int
		dx, dy   float64
		callback func()
	}
	transparentInf struct {
		counter  int
		dt       float32
		callback func()
	}
	lineAnime struct {
		enable bool
		rido   float64 //遮罩比例,-1,1表示全遮,0表示没有遮罩
		lines  []struct {
			rate  float64 //缩放比例
			theta float64 //旋转角度
		}
		x, y float64 //选择线的起始点
	}
	btnGrup struct {
		confirm buttonI
		cancle  buttonI
		fakebtn buttonI
		recast  buttonI
	}
	virtualtext struct {
		enable bool
		cname  *front.TextItem
	}
	drawEquipTip bool
	inHandleZoon bool
	promptText   struct {
		l1, l2   *front.TextItem2
		dx1, dx2 float64
	}
}

func (c *card) Draw(screen *ebiten.Image) {
	if c.visibility == 0 {
		c.drawLine(screen)
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(c.visibility)
	op.GeoM.Translate(c.x, c.y)
	if c.showBack { //如果显示背面就画背面
		screen.DrawImage(cardImgs.back, op)
		if c.isSelected {
			op.GeoM.Reset()
			op.GeoM.Translate(c.x-10, c.y-14)
			screen.DrawImage(cardImgs.selected, op)
		}
		return
	}
	if !c.selectAble {
		op.ColorScale.Scale(0.5, 0.5, 0.5, 1)
	}
	//绘制背景
	op.GeoM.Reset()
	op.GeoM.Translate(c.x, c.y)
	screen.DrawImage(cardImgs.bg, op)
	//如果被选中则绘制被选中框
	if c.isSelected {
		op.GeoM.Reset()
		op.GeoM.Translate(c.x-10, c.y-14)
		screen.DrawImage(cardImgs.selected, op)
	}
	//计算卡面坐标并绘制
	op.GeoM.Reset()
	op.GeoM.Translate(c.getSubItemPos(66, 113, c.img.Bounds().Size()))
	screen.DrawImage(c.img, op)
	//若不是基本卡则绘制名字背景
	if c.CardType != data.BaseCardType {
		op.GeoM.Reset()
		op.GeoM.Translate(c.getSubItemPos(77, 30, cardImgs.nameBg.Bounds().Size()))
		screen.DrawImage(cardImgs.nameBg, op)
	}
	//绘制名字
	op.GeoM.Reset()
	if c.CardType == data.BaseCardType {
		op.GeoM.Translate(c.getSubItemPos(74, 36, c.nameImg.Bounds().Size()))
	} else {
		op.GeoM.Translate(c.getSubItemPos(77, 30, c.nameImg.Bounds().Size()))
	}
	screen.DrawImage(c.nameImg, op)
	//绘制花色与数字
	if c.Dec != 0 && c.Dec < 5 {
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+5, c.y+30)
		screen.DrawImage(cardImgs.decor[c.Dec-1], op)
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+8, c.y+6)
		if c.Dec.IsRed() {
			screen.DrawImage(cardImgs.numRed[c.Num-1], op)
		} else {
			screen.DrawImage(cardImgs.numblack[c.Num-1], op)
		}
	}
	//绘制颜色
	if c.Dec > 4 {
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+8, c.y+6)
		screen.DrawImage(cardImgs.decor[c.Dec-1], op)
	}
	if c.CardType == data.TipsCardType || c.CardType == data.DealyTipsCardType { //绘制锦囊牌的"锦囊"二字
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+43, c.y+44)
		screen.DrawImage(cardImgs.tips, op)
	} else if c.CardType != data.BaseCardType { //绘制装备牌的"装备"二字
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+43, c.y+44)
		screen.DrawImage(cardImgs.equip, op)
	}
	switch c.CardType {
	case data.WeaponCardType:
		//绘制范围二字
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+50, c.y+144)
		screen.DrawImage(cardImgs.dst, op)
		//绘制武器范围
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+93, c.y+128)
		screen.DrawImage(cardImgs.scope[data.WeaponDstMap[c.Name]], op)
	case data.HorseDownCardType:
		//绘制"-1"
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+77, c.y+118)
		screen.DrawImage(cardImgs.att, op)
		op.GeoM.Translate(10, -2)
		screen.DrawImage(cardImgs.scope[1], op)
	case data.HorseUpCardType:
		//绘制"+1"
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+77, c.y+118)
		screen.DrawImage(cardImgs.def, op)
		op.GeoM.Translate(10, -2)
		screen.DrawImage(cardImgs.scope[1], op)
	}
	//front.DrawText(screen, strconv.Itoa(int(c.ID)), c.x+30, c.y-30, 32, false)
	//如果要转化则画转化目标卡名
	if c.virtualtext.enable {
		op.GeoM.Reset()
		op.GeoM.Translate(c.x+4, c.y+80)
		screen.DrawImage(getImg("assets/virtualCard/cardNameBg.png"), op)
		c.virtualtext.cname.Draw(screen)
	}
	if c.inHandleZoon {
		c.promptText.l1.Draw(screen)
		c.promptText.l2.Draw(screen)
	}
	c.drawBtn(screen)
	c.drawLine(screen)
}

func (c *card) drawBtn(screen *ebiten.Image) {
	draw := func(btn buttonI) {
		if btn == nil {
			return
		}
		btn.Draw(screen)
	}
	draw(c.btnGrup.fakebtn)
	draw(c.btnGrup.cancle)
	draw(c.btnGrup.confirm)
	draw(c.btnGrup.recast)
}

// 绘制选择线
func (c *card) drawLine(screen *ebiten.Image) {
	if !c.lineAnime.enable {
		return
	}
	wide, height := cardImgs.selLine.Bounds().Dx(), cardImgs.selLine.Bounds().Dy()
	var x0, y0, x1, y1 int
	y0, y1 = 0, height
	if c.lineAnime.rido < 0 {
		x0 = 0
		x1 = wide + int(float64(wide)*c.lineAnime.rido)
	} else {
		x0 = int(float64(wide) * c.lineAnime.rido)
		x1 = wide
	}
	op := &ebiten.DrawImageOptions{}
	for _, l := range c.lineAnime.lines {
		op.GeoM.Reset()
		img := cardImgs.selLine.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
		if c.lineAnime.rido > 0 {
			op.GeoM.Translate(c.lineAnime.rido*float64(wide), 0)
		}
		op.GeoM.Scale(l.rate, 1)
		op.GeoM.Rotate(l.theta)
		op.GeoM.Translate(c.lineAnime.x, c.lineAnime.y)
		screen.DrawImage(img, op)
	}
}

func (c *card) getID() data.CID {
	return c.ID
}

func (c *card) getCardType() data.CardType {
	return c.CardType
}

func (c *card) getCardName() data.CardName {
	return c.Name
}

func (c *card) getPos() (float64, float64) {
	return c.x, c.y
}

func (c *card) isSelect() bool {
	return c.isSelected
}

func (c *card) setSelect(sel bool) {
	c.isSelected = sel
}

func (c *card) isSeleable() bool {
	return c.selectAble
}

func (c *card) setVirtualText(cname data.CardName) {
	if cname == data.NoName {
		c.virtualtext.enable = false
		return
	}
	c.virtualtext.enable = true
	c.virtualtext.cname.SetPos(c.x, c.y+60)
	c.virtualtext.cname.SetStr(cname.ChnName())
}

func (c *card) setPromptText(l1, l2 string) {
	c.promptText.l1.SetText(l1)
	c.promptText.l2.SetText(l2)
	c.promptText.dx1 = (127-c.promptText.l1.Width)/2 + 4
	c.promptText.dx2 = (127-c.promptText.l2.Width)/2 + 4
}

// 获取子元素坐标,输入子元素的中心坐标与尺寸，返回子元素的左上角坐标
func (c *card) getSubItemPos(x, y float64, size image.Point) (float64, float64) {
	return x - float64(size.X/2) + c.x, y - float64(size.Y/2) + c.y
}

// 完成卡片的动画运算
func (c *card) anime() {
	if c.moveInf.counter > 0 {
		c.setPos(c.x+c.moveInf.dx, c.y+c.moveInf.dy)
		c.moveInf.counter--
	} else if c.moveInf.callback != nil {
		c.moveInf.callback()
		c.moveInf.callback = nil
	}
	if c.transparentInf.counter > 0 {
		c.visibility -= c.transparentInf.dt
		c.transparentInf.counter--
	} else if c.transparentInf.callback != nil {
		c.transparentInf.callback()
		c.transparentInf.callback = nil
	}
	if c.lineAnime.enable {
		c.lineAnime.rido += 0.042
		if c.lineAnime.rido >= 0.99 {
			c.lineAnime.lines = nil
			c.lineAnime.enable = false
			c.lineAnime.rido = -1
		}
	}
	if c.inHandleZoon {
		c.promptText.l1.SetPos(c.x+c.promptText.dx1, c.y+90)
		c.promptText.l1.SetVisibility(c.visibility)
		c.promptText.l2.SetPos(c.x+c.promptText.dx2, c.y+116)
		c.promptText.l2.SetVisibility(c.visibility)
	}
}

// 将坐标设为x,y(瞬移)
func (c *card) setPos(x, y float64) {
	c.x, c.y = x, y
	c.rect.x, c.rect.y = int(x), int(y)
}

// 将卡片移动至x,y
func (c *card) move2Pos(x, y float64, callback func()) {
	c.moveInf.callback = callback
	const speed = 16 //每tick移动像素
	dx, dy := x-c.x, y-c.y
	distence := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2)) //起点到终点的距离
	if distence == 0 {
		return
	}
	c.moveInf.counter = int(distence / speed)
	c.moveInf.dx = dx * speed / distence
	c.moveInf.dy = dy * speed / distence
	remain := int(distence) % speed
	c.setPos(c.x+dx*float64(remain)/distence, c.y+dy*float64(remain)/distence)
}

// 设置卡片在t个tick后淡出
func (c *card) setFadeOut(t int, callback func()) {
	c.transparentInf.callback = callback
	c.transparentInf.counter = t
	c.transparentInf.dt = 1.0 / float32(t)
}

func (c *card) resetFadeOut() {
	c.transparentInf.callback = nil
	c.transparentInf.counter = 0
	c.transparentInf.dt = 0
}

// 检查是否正在淡出
func (c *card) isOnFade() bool {
	return c.transparentInf.counter > 0
}

// 设置卡片可见性
func (c *card) setVisibility(v float32) {
	c.visibility = v
}

func (c *card) setSelectedAble(selAble bool) {
	c.selectAble = selAble
}

func (c *card) setShowBack(show bool) {
	c.showBack = show
}

func (c *card) selected(p *player) {
	c.isSelected = true
	c.setPos(c.x, c.y-8)
	p.selCard = append(p.selCard, c.ID)
}

func (c *card) deSelect(p *player) {
	if !c.isSelected {
		return
	}
	c.isSelected = false
	c.setPos(c.x, c.y+8)
	c.btnGrup.cancle = nil
	c.btnGrup.confirm = nil
	c.btnGrup.fakebtn = nil
	c.btnGrup.recast = nil
	for i := 0; i < len(p.selCard); i++ {
		if p.selCard[i] == c.getID() {
			p.selCard = append(p.selCard[:i], p.selCard[i+1:]...)
		}
	}
}

// 无任何副作用的deselect
func (c *card) pureDeselect(p *player) {
	c.deSelect(p)
}

func (c *card) getSelectState() bool {
	return c.isSelected
}

func (c *card) isClicked() bool {
	if c.rect.isClicked() {
		sound.PlayClickCard()
		return true
	}
	return false
}

func (c *card) onUse(g *Games, user data.PID, targets ...data.PID) {
	sound.PlayCardSound(c.Name, !g.getPlayer(user).male)
	if c.getCardType() == data.BaseCardType {
		g.getPlayer(user).basicCardAnime.name = c.Name
		g.getPlayer(user).basicCardAnime.enable = true
	}
	if len(targets) > 0 {
		c.lineAnime.enable = true
		c.lineAnime.rido = -1
		x0, y0 := c.getLinePos(user, g)
		c.lineAnime.x, c.lineAnime.y = x0, y0
		for _, target := range targets {
			x1, y1 := c.getLinePos(target, g)
			rate, theta := c.getLineDirection(x0, y0, x1, y1)
			c.lineAnime.lines = append(c.lineAnime.lines, struct {
				rate  float64
				theta float64
			}{rate: rate, theta: theta})
		}
	}
	//卡片提示文本
	l1 := "[white]" + g.getPlayer(user).dspName
	var l2 string
	if !isItemInList([]data.CardName{data.Drunk, data.Dodge, data.WGFD, data.WXKJ, data.WJQF}, c.Name) && len(targets) != 0 {
		l2 += "[yellow]对 "
		for _, t := range targets {
			l2 += "[white]" + g.getPlayer(t).dspName
		}
	}
	if c.CardType == data.BaseCardType || c.CardType == data.TipsCardType {
		c.setPromptText(l1, l2)
	} else {
		//因为装备牌和延时锦囊牌是复制了一张送进处理区，所以我们要设置的是处理区最后一张牌的文本
		g.handleZone[len(g.handleZone)-1].setPromptText(l1, l2)
	}
}

// 获取选择线指向的坐标
func (c *card) getLinePos(pid data.PID, g *Games) (x, y float64) {
	if pid != g.pid {
		x = g.getPlayer(pid).x + 87
		y = g.getPlayer(pid).y + 116
		return
	}
	x, y = 667, 505
	return
}

// 获取选择线的指向
func (c *card) getLineDirection(x0, y0, x1, y1 float64) (rate, theat float64) {
	dx, dy := x1-x0, y1-y0
	rate = math.Sqrt(dx*dx+dy*dy) / float64(cardImgs.selLine.Bounds().Dx())
	theat = math.Atan(dy / dx)
	if dx < 0 {
		theat += math.Pi
	}
	return
}

func (c *card) setDrawEquipTip(a bool) {
	c.drawEquipTip = a
}

func (c *card) setInHandleZoon(b bool) {
	c.inHandleZoon = b
}

func (c *card) updateBtn(g *Games) {
	update := func(b buttonI) {
		if b == nil {
			return
		}
		b.Update(g)
	}
	update(c.btnGrup.confirm)
	update(c.btnGrup.cancle)
	update(c.btnGrup.fakebtn)
	update(c.btnGrup.recast)
}

func (c *card) handleUseState(g *Games, p *player) {
	if !c.selectAble {
		return
	}
	c.updateBtn(g)
	if !c.isClicked() {
		return
	}
	if c.isSelected {
		p.selCard = nil
		c.deSelect(p)
		return
	}
	if len(p.selCard) > 0 {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	c.btnGrup.confirm = newConfirmBtn(func(g *Games) { g.useCardInf <- data.UseCardInf{Skip: false, ID: c.ID}; c.deSelect(p) })
	c.btnGrup.cancle = newCancleBtn(func(g *Games) { c.deSelect(p) })
	c.btnGrup.fakebtn = newFakeConfirmBtn(nil)
}

func (c *card) handleDropState(g *Games, p *player) bool {
	if !c.selectAble {
		return false
	}
	if !c.isClicked() {
		return false
	}
	if c.isSelected {
		c.deSelect(p)
		return true
	}
	if int(g.dropAbleInf.dropNum) == len(p.selCard) && len(p.selCard) > 0 {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	return true
}

func (c *card) handleDying(g *Games, p *player) {
	if !c.selectAble {
		return
	}
	c.updateBtn(g)
	if !c.isClicked() {
		return
	}
	if c.isSelected {
		p.selCard = nil
		c.deSelect(p)
		return
	}
	if len(p.selCard) > 0 {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	c.btnGrup.confirm = newConfirmBtn(func(g *Games) { g.useCardInf <- data.UseCardInf{Skip: false, ID: c.ID}; c.deSelect(p) })
	c.btnGrup.cancle = newCancleBtn(func(g *Games) { c.deSelect(p) })
}

func (c *card) handleWXKJ(g *Games, p *player) {}

func (c *card) handleDodge(g *Games, p *player) {}

func (c *card) handleDuel(g *Games, p *player) {}

func (c *card) handleNMRQ(g *Games, p *player) {}

func (c *card) handleWJQF(g *Games, p *player) {}

func (c *card) handleWGFD(g *Games) {
	if !c.selectAble {
		return
	}
	if !c.isClicked() {
		return
	}
	c.selectAble = false
	g.useCardInf <- data.UseCardInf{ID: c.ID}
}

func (c *card) handleBurnShow(g *Games, p *player) {
	c.handleUseState(g, p)
}

func (c *card) handleBurnDrop(g *Games, p *player) {
	c.handleUseState(g, p)
}

func (c *card) handleSSQY(g *Games) {
	c.handleWGFD(g)
}

func (c *card) handleGHCQ(g *Games) {
	c.handleWGFD(g)
}

func (c *card) handleQLG(g *Games) {
	c.handleWGFD(g)
}

func (c *card) handleDropOtherCard(g *Games) {
	c.handleWGFD(g)
}

func (c *card) handleJDSR(g *Games, p *player) {}

func (c *card) handleQLYYD(g *Games, p *player) {}

func (c *card) handleLuanWu(g *Games, p *player) {}

func (c *card) handleDropSelfAll(g *Games, p *player) bool {
	if !c.isClicked() {
		return false
	}
	if !c.selectAble {
		return false
	}
	if c.isSelected {
		c.deSelect(p)
		return true
	}
	if len(p.selCard) == g.selNum {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	return true
}

func (c *card) handleGSF(g *Games, p *player) bool {
	if !c.isClicked() {
		return false
	}
	if !c.selectAble {
		return false
	}
	if c.isSelected {
		c.deSelect(p)
		return true
	}
	if len(p.selCard) == g.selNum {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	return true
}

func (c *card) handleWenJi(g *Games, p *player) {
	c.handleUseState(g, p)
}

// 需要目标的卡
type needTargetCard struct {
	card
	targetlist []data.PID
	maxTarget  int
}

func (c *needTargetCard) handleUseState(g *Games, p *player) {
	if !c.selectAble {
		return
	}
	c.updateBtn(g)
	if c.isSelected {
		//当卡被再次点击时取消选中卡片
		if c.isClicked() {
			p.selCard = nil
			c.deSelect(p)
			return
		}
		//当目标还未选定时检测所有玩家
		for i := 0; i < len(g.playList); i++ {
			if g.playList[i].unSelAble {
				continue
			}
			if g.playList[i].isClicked() {
				if g.playList[i].isSelected {
					g.playList[i].isSelected = false
					c.targetlist = delFromList(c.targetlist, func(pid data.PID) bool { return pid == g.playList[i].pid })
				} else {
					c.targetlist = append(c.targetlist, g.playList[i].pid)
					g.playList[i].isSelected = true
				}
				if len(c.targetlist) > c.maxTarget {
					g.getPlayer(c.targetlist[0]).isSelected = false
					c.targetlist = c.targetlist[1:]
				}
				if len(c.targetlist) > 0 {
					c.btnGrup.confirm = newConfirmBtn(func(g *Games) {
						g.useCardInf <- data.UseCardInf{ID: c.ID, TargetList: c.targetlist}
						c.deSelect(p)
						c.restorePlayerList()
					})
				} else {
					c.btnGrup.confirm = nil
				}
				return
			}
		}
		return
	}
	if !c.isClicked() {
		return
	}
	//当卡被点起来时,向服务端请求可攻击的玩家列表
	if len(p.selCard) > 0 {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	g.targetQuest <- c.ID
	targetInf := <-g.availableTargetrec
	//将目标列表中的玩家设为可选中
	for i := 0; i < len(g.playList); i++ {
		g.playList[i].unSelAble = true
		for j := 0; j < len(targetInf.TargetList); j++ {
			if g.playList[i].pid == targetInf.TargetList[j] {
				g.playList[i].unSelAble = false
				break
			}
		}
	}
	c.maxTarget = int(targetInf.TargetNum)
	c.btnGrup.cancle = newCancleBtn(func(g *Games) { c.deSelect(p) })
	c.btnGrup.fakebtn = newFakeConfirmBtn(nil)
}

func (c *needTargetCard) deSelect(p *player) {
	if !c.isSelected {
		return
	}
	c.card.deSelect(p)
	c.targetlist = nil
	c.restorePlayerList()
}

// 恢复玩家列表的状态
func (c *needTargetCard) restorePlayerList() {
	for i := 0; i < len(c.g.playList); i++ {
		c.g.playList[i].unSelAble = false
		c.g.playList[i].isSelected = false
	}
}

type wxkjCard struct {
	card
}

func (c *wxkjCard) handleWXKJ(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

type dodgeCard struct {
	card
}

func (c *dodgeCard) handleDodge(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

func (c *dodgeCard) handleWJQF(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

type attackCard struct {
	needTargetCard
}

func (c *attackCard) handleDuel(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

func (c *attackCard) handleNMRQ(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

func (c *attackCard) handleJDSR(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

func (c *attackCard) handleQLYYD(g *Games, p *player) {
	c.card.handleUseState(g, p)
}

func (c *attackCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	c.card.onUse(g, user, targets...)
	g.getPlayer(user).isDrunk = false
}

func (c *attackCard) handleLuanWu(g *Games, p *player) {
	c.needTargetCard.handleUseState(g, p)
}

type fireAttackCard struct {
	attackCard
}

type lightnAttackCard struct {
	attackCard
}

type drunkCard struct {
	card
}

func (c *drunkCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	c.card.onUse(g, user, targets...)
	if g.state == data.UseCardState {
		g.getPlayer(user).isDrunk = true
	}
}

type tslhCard struct {
	needTargetCard
}

func (c *tslhCard) handleUseState(g *Games, p *player) {
	if !c.selectAble {
		return
	}
	c.updateBtn(g)
	if c.isSelected {
		//当卡被再次点击时取消选中卡片
		if c.isClicked() {
			p.selCard = nil
			c.deSelect(p)
			return
		}
		//当目标还未选定时检测所有玩家
		for i := 0; i < len(g.playList); i++ {
			if g.playList[i].unSelAble {
				continue
			}
			if g.playList[i].isClicked() {
				if g.playList[i].isSelected {
					g.playList[i].isSelected = false
					c.targetlist = delFromList(c.targetlist, func(pid data.PID) bool { return pid == g.playList[i].pid })
				} else {
					c.targetlist = append(c.targetlist, g.playList[i].pid)
					g.playList[i].isSelected = true
				}
				if len(c.targetlist) > c.maxTarget {
					g.getPlayer(c.targetlist[0]).isSelected = false
					c.targetlist = c.targetlist[1:]
				}
				if len(c.targetlist) > 0 {
					c.btnGrup.confirm = newConfirmBtn(func(g *Games) {
						g.useCardInf <- data.UseCardInf{ID: c.ID, TargetList: c.targetlist}
						c.deSelect(p)
						c.restorePlayerList()
					})
				} else {
					c.btnGrup.confirm = nil
				}
				return
			}
		}
		return
	}
	if !c.isClicked() {
		return
	}
	//当卡被点起来时,向服务端请求可攻击的玩家列表
	if len(p.selCard) > 0 {
		p.findCard(p.selCard[0]).deSelect(p)
	}
	c.selected(p)
	g.targetQuest <- c.ID
	targetInf := <-g.availableTargetrec
	//将目标列表中的玩家设为可选中
	for i := 0; i < len(g.playList); i++ {
		for j := 0; j < len(targetInf.TargetList); j++ {
			if g.playList[i].pid == targetInf.TargetList[j] {
				g.playList[i].unSelAble = false
				break
			}
			g.playList[i].unSelAble = true
		}
	}
	c.maxTarget = int(targetInf.TargetNum)
	c.btnGrup.cancle = newCancleBtn(func(g *Games) { c.deSelect(p) })
	c.btnGrup.fakebtn = newFakeConfirmBtn(nil)
	c.btnGrup.recast = newRecastBtn(func(g *Games) {
		g.useCardInf <- data.UseCardInf{ID: c.ID, TargetList: []data.PID{}}
		c.deSelect(p)
	})
}

func (c *tslhCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	if len(targets) != 0 {
		c.card.onUse(g, user, targets...)
	} else {
		//如果是重铸
		c.setPromptText("[white]"+g.getPlayer(user).dspName, " [yellow]重铸")
	}
}

type jdsrCard struct {
	needTargetCard
	targetInf []struct {
		killer  data.PID
		targets []data.PID
	}
}

func (c *jdsrCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	c.card.onUse(g, user, targets[0])

}

func (c *jdsrCard) handleUseState(g *Games, p *player) {
	if !c.selectAble {
		return
	}
	c.updateBtn(g)
	if !c.isSelected {
		if !c.isClicked() {
			return
		}
		//当卡被点起来时,向服务端请求可攻击的玩家列表
		if len(p.selCard) > 0 {
			p.findCard(p.selCard[0]).deSelect(p)
		}
		c.selected(p)
		g.targetQuest <- c.ID
		targetInf := <-g.availableTargetrec
		c.targetInf = nil
		isKillerFlat := true
		for _, pid := range targetInf.TargetList {
			if isKillerFlat {
				c.targetInf = append(c.targetInf, struct {
					killer  data.PID
					targets []data.PID
				}{killer: pid})
				isKillerFlat = false
				continue
			}
			if pid == -1 {
				isKillerFlat = true
				continue
			}
			c.targetInf[len(c.targetInf)-1].targets = append(c.targetInf[len(c.targetInf)-1].targets, pid)
		}
		c.showKiller(g)
		return
	}
	if c.isClicked() {
		c.deSelect(p)
		return
	}
	for _, t := range g.playList {
		if t.unSelAble || !t.isClicked() {
			continue
		}
		//如果t已被选择
		if t.isSelected {
			if c.targetlist[0] == t.pid {
				//如果t是killer
				c.showKiller(g)
				t.isSelected = false
				if len(c.targetlist) == 2 {
					g.getPlayer(c.targetlist[1]).isSelected = false
				}
				c.targetlist = nil
			} else {
				//如果t是target
				c.showTargets(g, c.targetlist[0])
				t.isSelected = false
				c.targetlist = c.targetlist[:1]
			}
			continue
		}
		//如果t未被选中
		if len(c.targetlist) == 0 {
			//如果t是killer
			c.showTargets(g, t.pid)
			c.targetlist = append(c.targetlist, t.pid)
			t.isSelected = true
		} else {
			//如果t是target
			if len(c.targetlist) > 2 {
				//如果已有选中的target
				g.getPlayer(c.targetlist[1]).isSelected = false
				c.targetlist = c.targetlist[:2]
			}
			c.targetlist = append(c.targetlist, t.pid)
			t.isSelected = true
		}
	}
	if c.btnGrup.cancle == nil {
		c.btnGrup.cancle = newCancleBtn(func(g *Games) { c.deSelect(p); g.hasSkip = true })
	}
	if len(c.targetlist) == 2 {
		if c.btnGrup.confirm == nil {
			c.btnGrup.confirm = newConfirmBtn(func(g *Games) {
				g.useCardInf <- data.UseCardInf{ID: c.ID, TargetList: c.targetlist}
				c.deSelect(p)
				g.hasSkip = true
			})
		}
	} else {
		c.btnGrup.confirm = nil
	}
}

// 将所有killer设为可选
func (c *jdsrCard) showKiller(g *Games) {
	for i := 0; i < len(g.playList); i++ {
		g.playList[i].unSelAble = true
	}
	for i := 0; i < len(c.targetInf); i++ {
		g.getPlayer(c.targetInf[i].killer).unSelAble = false
	}
}

// 将对应killer的target设为可选
func (c *jdsrCard) showTargets(g *Games, killer data.PID) {
	for i := 0; i < len(c.targetInf); i++ {
		if c.targetInf[i].killer != killer {
			continue
		}
		for j := 0; j < len(g.playList); j++ {
			g.playList[j].unSelAble = true
		}
		for _, pid := range c.targetInf[i].targets {
			g.getPlayer(pid).unSelAble = false
		}
		g.getPlayer(killer).unSelAble = false
		return
	}
}

type delayTipsCard struct {
	needTargetCard
}

func (c *delayTipsCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	var slot *cardI
	var t *player
	if len(targets) != 0 {
		t = g.getPlayer(targets[0])
	} else {
		t = g.getPlayer(user)
	}
	switch c.Name {
	case data.LBSS:
		slot = &t.judgeSlot[data.LBSSSlot]
	case data.BLCD:
		slot = &t.judgeSlot[data.BLCDSlot]
	}
	if len(targets) != 0 {
		*slot = c
		//sound.PlayCardSound(c.Name, !g.getPlayer(user).male)
		c.card.onUse(g, user, targets...)
		c.setVisibility(0)
		c.setPos(t.x+10, t.y+20)
	} else {
		//以下是弃牌
		if c.getID() != 0 {
			g.moveCard2Handle(c)
			*slot = nil
		} else {
			g.moveCard2Handle(*slot)
			*slot = nil
		}
		g.moveCard2Drop(c.getID())
	}
}

type lightnCard struct {
	card
}

func (c *lightnCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	if len(targets) == 0 {
		//以下是弃牌
		slot := &g.getPlayer(user).judgeSlot[data.LightningSlot]
		if c.getID() != 0 {
			g.moveCard2Handle(c)
			*slot = nil
		} else {
			g.moveCard2Handle(*slot)
			*slot = nil
		}
		g.moveCard2Drop(c.getID())
		return
	}
	if user == targets[0] {
		//sound.PlayCardSound(c.Name, !g.getPlayer(user).male)
		c.card.onUse(g, user)
	}
	t := g.getPlayer(targets[0])
	c.setVisibility(0)
	c.setPos(t.x+10, t.y+20)
	t.judgeSlot[data.LightningSlot] = c
}

type equipCard struct {
	card
	isEquip      bool
	px, py       float64
	item         *ebiten.Image
	smallNameImg *ebiten.Image
}

func (c *equipCard) onUse(g *Games, user data.PID, targets ...data.PID) {
	var slot **equipCard
	p := g.getPlayer(user)
	switch c.CardType {
	case data.WeaponCardType:
		slot = &p.equipSlot[data.WeaponSlot]
	case data.ArmorCardType:
		slot = &p.equipSlot[data.ArmorSlot]
	case data.HorseUpCardType:
		slot = &p.equipSlot[data.HorseUpSlot]
	case data.HorseDownCardType:
		slot = &p.equipSlot[data.HorseDownSlot]
	}
	if *slot != nil {
		if (*slot).getCardName() == data.ZBSM && user == g.pid {
			for i := 0; i < len(p.skills); i++ {
				if p.skills[i].getID() == data.ZBSMSkill {
					p.skills = append(p.skills[:i], p.skills[i+1:]...)
					break
				}
			}
		}
		if (*slot).getID() != 0 {
			(*slot).resetFadeOut()
			(*slot).isEquip = false
			if c.getID() != 0 {
				g.moveCard2Handle(*slot)
			} else {
				p.cards = append(p.cards, *slot)
			}
		}
	}
	if c.getCardName() == data.ZBSM && user == g.pid {
		y := 660.
		if len(p.skills) != 0 {
			lastSkill := p.skills[len(p.skills)-1]
			_, y = lastSkill.getPos()
			y -= 50
		}
		p.skills = append(p.skills, newZBSMSkill(g, g.pid))
		p.skills[len(p.skills)-1].setPos(1020, y)
	}
	*slot = c
	c.isEquip = true
	c.px, c.py = p.x+36, p.y
	switch c.CardType {
	case data.WeaponCardType:
		c.py += 158
	case data.ArmorCardType:
		c.py += 180
	case data.HorseUpCardType:
		c.py += 205
	case data.HorseDownCardType:
		c.px += 64
		c.py += 205
	}
	c.setPos(cardHeapPos.x, cardHeapPos.y)
	c.setVisibility(1)
	c.setShowBack(false)
	c.selectAble = true
	if c.getID() != 0 {
		//sound.PlayCardSound(c.Name, !g.getPlayer(user).male)
		c.card.onUse(g, user)
		c.move2Pos(p.x+10, p.y+20, func() { c.setFadeOut(1*60, func() { c.setVisibility(0) }) })
	} else {
		c.setPos(p.x+10, p.y+20)
		c.setVisibility(0)
		c.selectAble = false
	}

}

func (c *equipCard) Draw(screen *ebiten.Image) {
	if c.isEquip {
		c.drawEquiped(screen)
	}
	c.card.Draw(screen)
	if c.drawEquipTip {
		c.drawEquipTips(screen)
	}
}

func (c *equipCard) drawEquiped(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(c.px, c.py)
	screen.DrawImage(c.item, op)
	switch c.CardType {
	case data.HorseUpCardType:
		op.GeoM.Translate(15, -3)
		screen.DrawImage(cardImgs.smallDef, op)
		op.GeoM.Translate(8, 0)
		screen.DrawImage(cardImgs.smallScope[1], op)
		op.GeoM.Translate(12, -2)
	case data.HorseDownCardType:
		op.GeoM.Translate(15, -3)
		screen.DrawImage(cardImgs.smallAtt, op)
		op.GeoM.Translate(8, 0)
		screen.DrawImage(cardImgs.smallScope[1], op)
		op.GeoM.Translate(12, -2)
	default:
		op.GeoM.Translate(44, 0)
		screen.DrawImage(c.smallNameImg, op)
		op.GeoM.Translate(-26, -2)
	}
	if c.Dec != data.NoDec {
		screen.DrawImage(cardImgs.decANum[c.Dec-1][c.Num-1], op)
	}
	if c.CardType == data.WeaponCardType {
		op.GeoM.Translate(88, 2)
		screen.DrawImage(cardImgs.smallScope[data.WeaponDstMap[c.Name]], op)
	}
}

func (c *equipCard) drawEquipTips(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(c.x-10, c.y+55)
	screen.DrawImage(getImg("assets/card/equipTip.png"), op)
}

var noTargetCardList = []data.CardName{data.Peach, data.WZSY, data.NMRQ, data.WJQF, data.TYJY, data.WGFD}

var needTargetCardList = []data.CardName{data.Duel, data.Burn, data.SSQY, data.GHCQ}

func newCard(c data.Card, g *Games) cardI {
	card := card{Card: c, img: getCardImg("assets/card/img/" + c.Name.String() + ".png"),
		nameImg: getCardImg("assets/card/name/" + c.Name.String() + ".png"), visibility: 1, rect: rect{wide: 136, height: 182}, g: g}
	card.virtualtext.cname = front.NewTextItem("", 0, 0, 24, false, 24, color.Black)
	card.promptText.l1 = front.NewTextItem2("", 0, 0, 24, 2, 24)
	card.promptText.l2 = front.NewTextItem2("", 0, 0, 24, 2, 24)
	card.setPos(cardHeapPos.x, cardHeapPos.y)
	if isItemInList(noTargetCardList, c.Name) {
		return &card
	}
	if isItemInList(needTargetCardList, c.Name) {
		return &needTargetCard{card: card}
	}
	//对于不在以上两个列表中的卡
	switch c.Name {
	//基本卡
	case data.Dodge:
		return &dodgeCard{card: card}
	case data.Attack:
		return &attackCard{needTargetCard: needTargetCard{card: card}}
	case data.FireAttack:
		return &fireAttackCard{attackCard: attackCard{needTargetCard: needTargetCard{card: card}}}
	case data.LightnAttack:
		return &lightnAttackCard{attackCard: attackCard{needTargetCard: needTargetCard{card: card}}}
	case data.Drunk:
		return &drunkCard{card: card}
	//锦囊卡
	case data.WXKJ:
		return &wxkjCard{card: card}
	case data.TSLH:
		return &tslhCard{needTargetCard: needTargetCard{card: card}}
	case data.Lightning:
		return &lightnCard{card: card}
	case data.JDSR:
		return &jdsrCard{needTargetCard: needTargetCard{card: card}}
	}
	//对于延时锦囊牌
	if c.CardType == data.DealyTipsCardType {
		return &delayTipsCard{needTargetCard: needTargetCard{card: card}}
	}
	//对于装备牌
	if c.CardType == data.WeaponCardType || c.CardType == data.ArmorCardType {
		return &equipCard{card: card, item: getCardImg("assets/card/item/" + card.Name.String() + ".png"),
			smallNameImg: getCardImg("assets/card/equip/" + card.Name.String() + ".png")}
	}
	//对于马
	if c.CardType == data.HorseUpCardType || c.CardType == data.HorseDownCardType {
		return &equipCard{card: card, item: getCardImg("assets/card/item/horse.png")}
	}
	panic("卡类型" + c.Name.String() + "不存在")
}

func newCardList(g *Games) []cardI {
	l := data.GetCards()
	cardList := make([]cardI, len(l)+1)
	for i := 0; i < len(l); i++ {
		cardList[l[i].ID] = newCard(l[i], g)
	}
	return cardList
}

// 通过id复制卡片(有id)
func copyCardByID(id data.CID, g *Games) cardI {
	l := data.GetCards()
	return newCard(l[id-1], g)
}

// 复制一张无id卡
func copyNoIDCard(id data.CID, g *Games) cardI {
	l := data.GetCards()
	c := l[id-1]
	c.ID = 0
	return newCard(c, g)
}
