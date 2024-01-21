package main

import (
	"embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/liqmix/glit/defs"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
)

//go:embed assets/*
var embedFS embed.FS

func main() {
	ebiten.SetWindowTitle("Glit")
	ebiten.SetWindowDecorated(true)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// ebiten.SetCursorMode(ebiten.CursorModeHidden)
	// ebiten.SetFullscreen(true)

	var resources = defs.InitResources(embedFS)

	audio.NewContext(44100)
	if err := ebiten.RunGame(NewGame(resources)); err != nil {
		panic(err)
	}
}
