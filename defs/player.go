package defs

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	Thing
}

func NewPlayer(x, y float64, t *Thing) *Player {
	t.energy = t.MaxEnergy
	t.x = x
	t.y = y
	return &Player{
		Thing: *t,
	}
}

func (p *Player) Update() *Action {
	p.Thing.Update()
	return p.GetAction()
}

func (p *Player) Draw(screen *ebiten.Image) {
	p.Thing.Draw(screen)
}

func (p *Player) GetAction() *Action {
	action := &Action{
		Move: &Move{},
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		action.Move.vY = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		action.Move.vY = 1
	} else {
		action.Move.vY = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		action.Move.vX = -1
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		action.Move.vX = 1
	} else {
		action.Move.vX = 0
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) && !p.cooldown {
		action.Glitch = true
		if p.energy <= 0 {
			p.cooldown = true
		}
	} else {
		if p.energy < p.MaxEnergy {
			p.cooldown = true
		}
	}
	return action
}
