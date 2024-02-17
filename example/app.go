package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type App struct {
	img *ebiten.Image
}

func (app *App) Update() error { return nil }

func (app *App) Draw(screen *ebiten.Image) {
	if app.img != nil {
		screen.DrawImage(app.img, nil)
	}
}

func (app *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
