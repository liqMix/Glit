package defs

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"gopkg.in/yaml.v2"
)

type Thingi interface {
	Position() (float64, float64)
	Update() *Action
	Draw(screen *ebiten.Image)
	ExecuteAction(a *Action, isGlitching bool)
	IsGlitching() bool
	IsBlock() bool
	IsBlockGlitch() bool
	IsEnemy() bool
	IsExit() bool
	HasCollision(x, y float64) bool
	Girth() (float64, float64)
}

type Thing struct {
	Name         string  `yaml:"name"`
	Symbol       string  `yaml:"symbol"`
	Enemy        bool    `yaml:"enemy"`
	Block        bool    `yaml:"block"`
	BlockGlitch  bool    `yaml:"blockGlitch"`
	Speed        float64 `yaml:"speed"`
	MaxEnergy    int     `yaml:"energy"`
	Exit         bool    `yaml:"exit"`
	Behavior     string  `yaml:"behavior"`
	x            float64
	y            float64
	width        int
	height       int
	sprite       *ebiten.Image
	history      []*Action
	historyIndex int
	energy       int
	isGlitching  bool
	cooldown     bool
}

var MAX_HISTORY = 10

func ThingFromFile(bytes []byte) (*Thing, error) {
	// Unmarshall yaml
	var t *Thing
	if err := yaml.Unmarshal(bytes, &t); err != nil {
		return nil, err
	}
	return t, nil
}

func ThingFromThing(t *Thing) *Thing {
	// Get width, height from sprite
	width, height := t.sprite.Bounds().Dx(), t.sprite.Bounds().Dy()
	return &Thing{
		Name:         t.Name,
		Symbol:       t.Symbol,
		Enemy:        t.Enemy,
		Behavior:     t.Behavior,
		Block:        t.Block,
		BlockGlitch:  t.BlockGlitch,
		Speed:        t.Speed,
		MaxEnergy:    t.MaxEnergy,
		Exit:         t.Exit,
		x:            t.x,
		y:            t.y,
		width:        width,
		height:       height,
		sprite:       t.sprite,
		history:      make([]*Action, 0),
		historyIndex: t.historyIndex,
		energy:       t.energy,
	}
}

func (t *Thing) Girth() (float64, float64) {
	return float64(t.width), float64(t.height)
}

func (t *Thing) Position() (float64, float64) {
	return t.x, t.y
}

func (t *Thing) IsGlitching() bool {
	return t.isGlitching
}

func (t *Thing) IsBlock() bool {
	return t.Block
}

func (t *Thing) IsBlockGlitch() bool {
	return t.BlockGlitch
}

func (t *Thing) IsEnemy() bool {
	return t.Enemy
}

func (t *Thing) IsExit() bool {
	return t.Exit
}

func (t *Thing) Move(move *Move) {
	if move == nil {
		return
	}
	if move.x != 0 || move.y != 0 {
		t.x = move.x
		t.y = move.y
		return
	}

	t.x += move.vX * t.Speed
	t.y += move.vY * t.Speed
}

func (t *Thing) ExecuteAction(a *Action, isGlitching bool) {
	if a == nil {
		return
	}
	if a.Glitch {
		t.isGlitching = true
	} else {
		t.isGlitching = false
	}

	// If we are glitching, and we aren't a glitcher, play history
	playHistory := false
	if isGlitching {
		if t.isGlitching == false {
			playHistory = true
		}
	}

	if playHistory {
		if len(t.history) != 0 {
			if t.historyIndex >= len(t.history) {
				t.historyIndex = 0
			}
			a = t.history[t.historyIndex]
			t.historyIndex++
		}
	} else {
		// Create action from current state
		// to append to history.
		// There are some actions we don't want to replay (like enabling glitch)
		action := &Action{
			Move: &Move{
				x: t.x,
				y: t.y,
			},
		}
		t.history = append(t.history, action)
	}

	// Execute action
	if a.Move != nil {
		t.Move(a.Move)
	}
}

func (t *Thing) HasCollision(x, y float64) bool {
	if t.Block == false {
		return false
	}
	// Check if x, y is within bounds of thing's position + width/height
	if x >= t.x && x <= t.x+float64(t.width) && y >= t.y && y <= t.y+float64(t.height) {
		return true
	}

	return false
}

func (t *Thing) Update() *Action {
	if t.cooldown && t.energy >= t.MaxEnergy {
		t.cooldown = false
	}
	// Truncate history if over max
	if len(t.history) > MAX_HISTORY {
		t.history = t.history[len(t.history)-MAX_HISTORY:]
	}
	if t.isGlitching && t.energy <= 0 {
		t.isGlitching = false
		t.history = make([]*Action, 0)
	} else if t.isGlitching {
		t.energy--
	} else if t.energy < t.MaxEnergy {
		t.energy++
	}
	return t.GetAction()
}

func (t *Thing) GetAction() *Action {
	switch t.Behavior {
	case "left":
		return &Action{
			Move: &Move{
				vX: -1,
			},
		}
	case "right":
		return &Action{
			Move: &Move{
				vX: 1,
			},
		}
	case "down":
		return &Action{
			Move: &Move{
				vY: 1,
			},
		}
	}
	return nil
}

func (t *Thing) Draw(s *ebiten.Image) {
	if t.sprite == nil {
		return
	}

	ops := &ebiten.DrawImageOptions{}
	ops.GeoM.Translate(float64(t.x), float64(t.y))
	// ops.GeoM.Scale(2, 2)

	if t.isGlitching {
		ops.ColorScale.Scale(0, 100, 0, 1)
	}
	s.DrawImage(t.sprite, ops)

	// Draw energy bar
	// Draw energy bar background
	if t.MaxEnergy != 0 && t.energy != t.MaxEnergy {
		energyBarX := t.x - float64(t.width/2)

		vector.DrawFilledRect(
			s,
			float32(energyBarX-0.5),
			float32(t.y-11),
			float32(17),
			6,
			color.RGBA{255, 255, 255, 255},
			false,
		)
		vector.DrawFilledRect(
			s,
			float32(energyBarX),
			float32(t.y-10),
			float32(15),
			4.0,
			color.RGBA{0, 0, 0, 255},
			false,
		)
		// Draw energy bar
		// White energy bar when full
		// Green energy bar when using
		if t.isGlitching {
			vector.DrawFilledRect(
				s,
				float32(energyBarX),
				float32(t.y-10),
				float32(t.energy)/float32(t.MaxEnergy)*16,
				4.0,
				color.RGBA{0, 255, 0, 255},
				false,
			)
		} else if t.cooldown {
			vector.DrawFilledRect(
				s,
				float32(energyBarX),
				float32(t.y-10),
				float32(t.energy)/float32(t.MaxEnergy)*16,
				4.0,
				color.RGBA{255, 0, 0, 255},
				false,
			)
		} else {
			vector.DrawFilledRect(
				s,
				float32(energyBarX),
				float32(t.y-10),
				float32(t.energy)/float32(t.MaxEnergy)*16,
				4.0,
				color.White,
				false,
			)
		}
	}

}
