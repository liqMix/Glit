package main

import (
	"holidayebijam23/defs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Game struct {
	ebiten.Game
	level      *defs.Level
	resources  *defs.Resources
	isGlitched bool
	audio      *audio.Player
}

func NewGame(r *defs.Resources) ebiten.Game {
	music := r.GetMusic("main")
	audioPlayer, err := audio.CurrentContext().NewPlayer(&music)
	audioPlayer.SetVolume(0.5)
	audioPlayer.Play()

	level, err := defs.LevelFromText(1, r, audioPlayer)
	if err != nil {
		panic(err)
	}
	return &Game{
		level:     level,
		resources: r,
		audio:     audioPlayer,
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	nextLevel := g.level.Update()
	if nextLevel != 0 {
		if nextLevel == 17 {
			nextLevel = 1
		}
		g.audio.Pause()
		level, err := defs.LevelFromText(nextLevel, g.resources, g.audio)
		if err != nil {
			panic(err)
		}
		g.level = level
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.level.Draw(screen)
}
