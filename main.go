package main

import (
	"goltk/app"
	"goltk/localclient"
	"goltk/sound"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	app := &app.App{}
	app.Bgm = sound.NewBgm("assets/audio/outgame.mp3")
	app.Bgm.Play()
	app.CurMenu = localclient.NewMianMenu(app)
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetTPS(60)
	ebiten.RunGame(app)
}
