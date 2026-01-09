package localclient

import (
	"goltk/data"
	"goltk/front"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

var btnImgList = [...]*ebiten.Image{
	loadImg("assets/button/confirm.png"),
	loadImg("assets/button/cancle.png"),
	loadImg("assets/button/end.png"),
	loadImg("assets/button/drop.png"),
	loadImg("assets/button/recast.png"),
	loadImg("assets/button/drop1.png"),
	loadImg("assets/button/confirm1.png"),
	loadImg("assets/button/recast.png"),
	loadImg("assets/button/longbtn.png"),
	loadImg("assets/button/skill.png"),
	loadImg("assets/button/skill1.png"),
	loadImg("assets/button/skill2.png"),
	loadImg("assets/button/selconfirm.png"),
	loadImg("assets/button/skillConfirm.png"),
	loadImg("assets/button/fakeSkillConfirm.png"),
	loadImg("assets/menu/button/kaizhan.png"),
	loadImg("assets/button/fakeLongbtn.png"),
}

type buttonI interface {
	Draw(screen *ebiten.Image)  //处理绘制
	Update(g *Games)            //处理更新
	setOnClick(fn func(*Games)) //设置被点击回调
	setPos(x, y float64)
}

type button struct {
	rect
	img       *ebiten.Image
	highLight bool //是否高亮
	x, y      float64
	callback  func(*Games)
}

func newButton(img *ebiten.Image, x, y float64, onClick func(*Games)) button {
	return button{rect: rect{wide: img.Bounds().Dx(), height: img.Bounds().Dy(), x: int(x), y: int(y)},
		img: img, x: x, y: y, callback: onClick}
}

func (b *button) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	if b.highLight {
		op.ColorScale.Scale(1.3, 1.3, 1.3, 1)
	}
	screen.DrawImage(b.img, op)
}

func (b *button) Update(g *Games) {
	if b.isPointInRect(ebiten.CursorPosition()) {
		b.highLight = true
	} else {
		b.highLight = false
	}
	if (b.callback != nil) && b.isClicked() {
		b.callback(g)
	}
}

func (b *button) setOnClick(fn func(*Games)) {
	b.callback = fn
}

func (b *button) setPos(x, y float64) {
	b.x, b.y = x, y
	b.rect.x, b.rect.y = int(x), int(y)
}

// 新建确定按钮
func newConfirmBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[0], 480, 490, onClick)
	return &btn
}

type fakeConfirmBtn struct {
	button
}

func (b *fakeConfirmBtn) Update(g *Games) {}

// 新建虚假确定按钮
func newFakeConfirmBtn(onClick func(*Games)) *fakeConfirmBtn {
	btn := newButton(btnImgList[6], 480, 490, onClick)
	return &fakeConfirmBtn{button: btn}
}

// 新建取消按钮
func newCancleBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[1], 820, 498, onClick)
	return &btn
}

// 新建结束按钮
func newEndStateBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[2], 820, 498, onClick)
	return &btn
}

type dropBtn struct {
	button
	unActiveImg *ebiten.Image
	isActive    bool
}

func (b *dropBtn) Update(g *Games) {
	if g.curPlayer != g.pid {
		return
	}
	b.isActive = len(g.getPlayer(g.pid).selCard) == int(g.dropAbleInf.dropNum)
	if !b.isActive {
		b.highLight = false
		return
	}
	if b.isPointInRect(ebiten.CursorPosition()) {
		b.highLight = true
	} else {
		b.highLight = false
	}
	if (b.callback != nil) && b.isClicked() {
		b.callback(g)
	}
}

func (b *dropBtn) Draw(screen *ebiten.Image) {
	if b.isActive {
		b.button.Draw(screen)
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.unActiveImg, op)
}

// 新建弃牌按钮
func newDropBtn(onClick func(*Games)) *dropBtn {
	btn := newButton(btnImgList[3], 480, 490, onClick)
	return &dropBtn{button: btn, unActiveImg: btnImgList[5]}
}

// 新建重铸按钮
func newRecastBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[7], 280, 498, onClick)
	return &btn
}

type textButton struct {
	button
	text *front.TextItem2
}

func (b *textButton) Draw(screen *ebiten.Image) {
	b.button.Draw(screen)
	b.text.Draw(screen)
}

func newTextButton(x, y float64, str string, onClick func(*Games)) *textButton {
	btn := &textButton{
		button: newButton(btnImgList[8], x, y, onClick),
		text:   front.NewTextItem2("[yellow]"+str, 0, 0, 28, 4, 28),
	}
	btn.text.SetPos(x+(230-btn.text.Width)/2, y+2+(55-btn.text.Heigh)/2)
	return btn
}

func newFakeTextButton(x, y float64, str string, onClick func(*Games)) *textButton {
	btn := &textButton{
		button: newButton(btnImgList[16], x, y, onClick),
		text:   front.NewTextItem2("[grey]"+str, 0, 0, 28, 4, 28),
	}
	btn.text.SetPos(x+(230-btn.text.Width)/2, y+2+(55-btn.text.Heigh)/2)
	return btn
}

// 新建左选择按钮
func newChooseLeftBtn(str string, onClick func(*Games)) *textButton {
	return newTextButton(350, 480, str, onClick)
}

// 新建右选择按钮
func newChooseRightBtn(str string, onClick func(*Games)) *textButton {
	return newTextButton(650, 480, str, onClick)
}

// 新建灰色选择按钮
func newFakeChooseBtn(str string, x, y float64, onClick func(*Games)) *textButton {
	return newTextButton(x, y, str, onClick)
}

// 背水按钮
func newBeiShuiBtn(str string, useAble bool, onClick func(*Games)) *textButton {
	var img *ebiten.Image
	if useAble {
		img = btnImgList[10]
	} else {
		img = btnImgList[9]
	}
	btn := &textButton{
		button: newButton(img, 200, 484, onClick),
		text:   front.NewTextItem2("[yellow]"+str, 0, 0, 28, 4, 28),
	}
	btn.text.SetPos(200+(94-btn.text.Width)/2, 480+2+(55-btn.text.Heigh)/2)
	return btn
}

// 选将按钮
type roleSelBtn struct {
	button
	isSelected bool
	role       data.Role
	img        *ebiten.Image
	skillText  *front.TextItem2
}

func newRoleSelBtn(x, y float64, img *ebiten.Image, role data.Role, onclick func(*Games)) *roleSelBtn {
	text := ""
	for _, skill := range role.SkillList {
		const maxLen = 30 //每行的最大长度
		str := []rune(skill.Name() + "：" + skill.Text())
		for i := maxLen; i < len(str); i += maxLen {
			str = append(str[:i], append([]rune("\n"), str[i:]...)...)
		}
		text += "[yellow]" + skill.Name() + "：[white]" + string(str[len([]rune(skill.Name()))+1:])
		if len(str) != 0 && str[len(str)-1] != rune('\n') {
			text += "\n\n"
		}
	}
	btn := &roleSelBtn{role: role, img: getPlayerImg("assets/role/charImg/" + role.Name + ".png"),
		skillText: front.NewTextItem2(text, 400, 100, 24, 1, 24), button: newButton(img, x, y, onclick)}
	return btn
}

func (b *roleSelBtn) resetOthers(g *Games) {
	for i := 0; i < len(g.btnList); i++ {
		if b1, ok := g.btnList[i].(*roleSelBtn); ok {
			b1.isSelected = false
		}
	}
}

var roleSelectBox = [...]*ebiten.Image{
	loadImg("assets/game/select/littleBoxBottom.png"),
	loadImg("assets/game/select/select.png"),
}

func (b *roleSelBtn) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x-20, b.y-16)
	screen.DrawImage(roleSelectBox[0], op)
	if b.isSelected {
		op.GeoM.Translate(6, 2)
		screen.DrawImage(roleSelectBox[1], op)
		//画选将信息
		//绘制阴影
		op.GeoM.Reset()
		op.GeoM.Translate(97, 97)
		screen.DrawImage(playerImgs.frame, op)
		//绘制背景
		op.GeoM.Translate(3, 3)
		screen.DrawImage(playerImgs.bgFaction[b.role.Side], op)
		//计算卡面坐标并绘制
		op.GeoM.Reset()
		op.GeoM.Translate(200-float64(b.img.Bounds().Size().X/2), 202-float64(b.img.Bounds().Size().Y/2))
		screen.DrawImage(b.img, op)
		//画龙
		if b.role.Dragon != 0 {
			op.GeoM.Reset()
			op.GeoM.Translate(93, 60)
			screen.DrawImage(playerImgs.dragonImg[b.role.Dragon-1], op)
		}
		//绘制血上限
		if b.role.MaxHP < 7 {
			op.GeoM.Reset()
			op.GeoM.Translate(100+8, 100+219)
			for i := 0; i < int(b.role.MaxHP); i++ {
				op.GeoM.Translate(0, float64(-20))
				screen.DrawImage(playerImgs.hpmaxImg, op)
			}
		}
		//绘制血
		if b.role.MaxHP < 7 {
			op.GeoM.Reset()
			op.GeoM.Translate(100+8, 100+220)
			var hpimg = playerImgs.hpImg[2]
			for i := 0; i < int(b.role.MaxHP); i++ {
				op.GeoM.Translate(0, float64(-20))
				screen.DrawImage(hpimg, op)
			}
		} else {
			hpimg := playerImgs.hpImg[2]
			op.GeoM.Reset()
			op.GeoM.Translate(100+8, 100+150)
			screen.DrawImage(hpimg, op)
			op.GeoM.Translate(-2, 18)
			screen.DrawImage(playerImgs.hpxxImg, op)
			front.DrawText(screen, strconv.Itoa(int(b.role.MaxHP)), 100+6, 100+190, 24, false)
		}
		//绘制名字
		front.DrawText(screen, b.role.DspName, 100+17, 100+34, 20, true)
		b.skillText.Draw(screen)
	}
	b.button.Draw(screen)
}

// 新建选将确定按钮
func newSelConfirmBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[12], 116, 200, onClick)
	return &btn
}

// 新建技能确定按钮
func newSkillConfirmBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[13], 580, 550, onClick)
	return &btn
}

// 新建技能确定按钮
func newFakeSkillConfirmBtn(onClick func(*Games)) *button {
	btn := newButton(btnImgList[14], 580, 550, onClick)
	return &btn
}

type skillBtn struct {
	button
	active  bool
	nameImg *ebiten.Image
}

func newSkillBtn(x, y float64, name *ebiten.Image, onclick func(*Games)) *skillBtn {
	return &skillBtn{button: newButton(btnImgList[9], x, y, onclick), nameImg: name}
}

func (b *skillBtn) Update(g *Games) {
	if !b.active {
		return
	}
	if b.isPointInRect(ebiten.CursorPosition()) {
		b.highLight = true
	} else {
		b.highLight = false
	}
	if (b.callback != nil) && b.isClicked() {
		b.callback(g)
	}
}

func (b *skillBtn) switch2UnActive() {
	b.active = false
	b.highLight = false
	b.img = btnImgList[9]
}

func (b *skillBtn) switch2Active() {
	b.active = true
	b.img = btnImgList[10]
}

func (b *skillBtn) switch2Selected() {
	b.active = true
	b.img = btnImgList[11]
}

func (b *skillBtn) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	if b.highLight {
		op.ColorScale.Scale(1.5, 1.5, 1.5, 1)
	}
	screen.DrawImage(b.img, op)
	dx := float64(94-b.nameImg.Bounds().Dx()) / 2
	op.GeoM.Translate(dx, 4)
	screen.DrawImage(b.nameImg, op)
}

func newQuitGameBtn() *button {
	b := newButton(getImg("assets/menu/button/back.png"), 1260, 20, func(g *Games) {
		g.app.Server.Close()
	})
	return &b
}

func newVirtualcardBtn(x, y float64, cname data.CardName, result *data.CardName) *button {
	b := newButton(getImg("assets/virtualCard/"+cname.String()+".png"), x, y, func(g *Games) {
		*result = cname
	})
	return &b
}
