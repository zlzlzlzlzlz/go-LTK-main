package app

import (
	"goltk/sound"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type serverI interface {
	Run()
	Close()
}

type menuI interface {
	Draw(screen *ebiten.Image)
	Update(a *App)
}

type App struct {
	CurMenu menuI
	Bgm     *audio.Player
	Server  serverI
}

func (a *App) Update() error {
	a.CurMenu.Update(a)
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	a.CurMenu.Draw(screen)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1334, 750
}

func (a *App) init() {
	a.Bgm = sound.NewBgm("assets/audio/outgame.mp3")
	a.Bgm.Play()
}
