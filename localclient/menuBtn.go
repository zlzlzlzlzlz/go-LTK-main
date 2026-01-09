package localclient

import (
	"goltk/app"
	"goltk/data"
	"goltk/front"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type menuButtonI interface {
	Draw(screen *ebiten.Image)
	Update(a *app.App)
}

type menuButton struct {
	rect
	x, y      float64
	img       *ebiten.Image
	highLight bool
	onclick   func(*app.App)
}

func newMenuButton(x, y float64, img *ebiten.Image, onclick func(*app.App)) *menuButton {
	return &menuButton{x: x, y: y, img: img, onclick: onclick,
		rect: rect{x: int(x), y: int(y), wide: img.Bounds().Dx(), height: img.Bounds().Dy()}}
}

func (b *menuButton) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	if b.highLight {
		op.ColorScale.Scale(1.5, 1.5, 1.5, 1)
	}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(b.img, op)
}

func (b *menuButton) Update(a *app.App) {
	if b.isPointInRect(ebiten.CursorPosition()) {
		b.highLight = true
	} else {
		b.highLight = false
	}
	if (b.onclick != nil) && b.isClicked() {
		b.onclick(a)
	}
}

func newMenuConfirmBtn(x, y float64, onclick func(*app.App)) *menuButton {
	return newMenuButton(x, y, getImg("assets/menu/button/find.png"), onclick)
}

type menuSearchBtn struct {
	menuButton
	promptText *front.TextItem2
}

func newMenuSearchBtn(x, y float64, onclick func(*app.App)) *menuSearchBtn {
	return &menuSearchBtn{menuButton: *newMenuButton(x, y, getImg("assets/menu/button/find.png"), onclick),
		promptText: front.NewTextItem2("", 0, 0, 24, 1, 24)}
}

func (b *menuSearchBtn) Draw(screen *ebiten.Image) {
	b.menuButton.Draw(screen)
	b.promptText.Draw(screen)
}

type inputBox struct {
	menuButton
	runes []rune
	text  front.TextItem
}

func newInputBox(x, y float64) *inputBox {
	box := &inputBox{
		menuButton: *newMenuButton(x, y, getImg("assets/menu/button/inputBox.png"), nil),
		text:       *front.NewTextItem("", x+15, y+5, 38, false, 0, color.White),
	}
	return box
}

func (b *inputBox) Draw(screen *ebiten.Image) {
	b.menuButton.Draw(screen)
	b.text.Draw(screen)
}

func (b *inputBox) Update(a *app.App) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if b.highLight {
			b.highLight = false
			return
		}
		if b.isClicked() {
			b.highLight = true
			return
		}
	}
	if len(b.text.Str) > 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			b.text.SetStr(b.text.Str[:len(b.text.Str)-1])
			return
		}
	}
	b.runes = ebiten.AppendInputChars(b.runes[:0])
	b.text.SetStr(b.text.Str + string(b.runes))
}

type selectButton struct {
	menuButton
	text front.TextItem
}

func newSelectButton(x, y float64, str string, onclick func(a *app.App)) *selectButton {
	return &selectButton{menuButton: *newMenuButton(x, y, getImg("assets/button/skill1.png"), onclick),
		text: *front.NewTextItem(str, x, y, 32, false, 0, color.White)}
}

func (b *selectButton) Draw(screen *ebiten.Image) {
	b.menuButton.Draw(screen)
	b.text.Draw(screen)
}

func (b *selectButton) Update(a *app.App) {
	if b.isClicked() && b.onclick != nil {
		b.onclick(a)
	}
}

func (b *selectButton) setSelect(sel bool) {
	b.highLight = sel
}

func newBackBtn(x, y float64, onClick func(a *app.App)) *menuButton {
	b := newMenuButton(x, y, getImg("assets/menu/button/back.png"), onClick)
	return b
}

type menuTextBtn struct {
	menuButton
	text *front.TextItem2
}

func newMenuTextBtn(x, y float64, str string, onclick func(*app.App)) *menuTextBtn {
	btn := &menuTextBtn{menuButton: *newMenuButton(x, y, btnImgList[8], onclick),
		text: front.NewTextItem2("[yellow]"+str, 0, 0, 28, 4, 28)}
	btn.text.SetPos(x+(230-btn.text.Width)/2, y+2+(55-btn.text.Heigh)/2)
	return btn
}

func (b *menuTextBtn) Draw(screen *ebiten.Image) {
	b.menuButton.Draw(screen)
	b.text.Draw(screen)
}

type modeSelectBtn struct {
	menuButton
	selected bool
}

func newModeSelectBtn(x, y float64, img *ebiten.Image, onclick func(*app.App)) *modeSelectBtn {
	return &modeSelectBtn{menuButton: *newMenuButton(x, y, img, onclick)}
}

func (b *modeSelectBtn) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	if b.selected {
		screen.DrawImage(b.img, op)
		return
	}
	if b.highLight {
		op.ColorScale.Scale(1.3, 1.3, 1.3, 1)
	} else {
		op.ColorScale.Scale(0.4, 0.4, 0.4, 1)
	}
	screen.DrawImage(b.img, op)
}

type roleDetailBtn struct {
	menuButton
	role data.Role
}

var detailTopImg = loadImg("assets/game/select/littleBoxTop.png")

func newRoleDetailBtn(x, y float64, role data.Role, onClick func(*app.App)) *roleDetailBtn {
	img := getPlayerImg("assets/role/smallChar/" + role.Name + ".png")
	return &roleDetailBtn{menuButton: *newMenuButton(x, y, img, nil), role: role}
}

func (b *roleDetailBtn) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.x, b.y)
	screen.DrawImage(roleSelectBox[0], op)
	if b.highLight {
		op.ColorScale.Scale(1.5, 1.5, 1.5, 1)
	}
	op.GeoM.Translate(19, 17)
	screen.DrawImage(b.img, op)
	op.ColorScale.Reset()
	op.GeoM.Translate(-19, -17)
	screen.DrawImage(detailTopImg, op)
}
