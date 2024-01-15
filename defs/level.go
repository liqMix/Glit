package defs

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

type Level struct {
	LevelNumber         int
	player              *Player
	things              []Thingi
	levelWidth          int
	levelHeight         int
	ticks               int
	isGlitching         bool
	isRotated           bool
	audioPlayer         *audio.Player
	glitchedAudioPlayer *audio.Player
	loadTick            int
	tick                int
	tickPosition        time.Duration
}

var LOAD_TICKS = 120
var CELL_SCALE = 16

func LevelFromText(levelNumber int, r *Resources, mainAudio *audio.Player) (*Level, error) {
	// split layout by new line
	// each line is a row
	// each character is a tile
	// lookup tile in rune map
	// create cell
	// add cell to row

	layout := r.GetLevel(fmt.Sprintf("%d", levelNumber))
	things := make([]Thingi, 0)
	var player *Player
	x, y := 0, 0
	for _, c := range layout {
		if c == '\n' {
			x = 0
			y++
			continue
		}

		var thing = r.GetThing(c)
		posX := float64(x * CELL_SCALE)
		posY := float64(y * CELL_SCALE)
		if thing != nil {
			if c == '@' {
				player = NewPlayer(posX, posY, thing)
				things = append(things, player)
			} else {
				t := ThingFromThing(thing)
				t.x = posX
				t.y = posY
				things = append(things, t)
			}
		}
		x++
	}
	music := r.GetMusic("main")
	glitchedAudio := audio.Resample(&music, music.Length(), 44100, 48000)
	glitchedAudioPlayer, err := audio.CurrentContext().NewPlayer(glitchedAudio)
	if err != nil {
		return nil, err
	}
	glitchedAudioPlayer.SetVolume(0.4)

	return &Level{
		LevelNumber:         levelNumber,
		levelWidth:          x + 1,
		levelHeight:         y + 1,
		player:              player,
		things:              things,
		audioPlayer:         mainAudio,
		glitchedAudioPlayer: glitchedAudioPlayer,
		tickPosition:        -1,
	}, nil
}

func NewLevel(name string) *Level {
	return &Level{}
}

func (l *Level) Update() int {
	if l.loadTick < LOAD_TICKS {
		l.loadTick++
		return 0
	} else if l.loadTick == LOAD_TICKS {
		l.audioPlayer.SetVolume(0.4)
		l.audioPlayer.Play()
		l.loadTick++
	}
	if !l.audioPlayer.IsPlaying() && !l.glitchedAudioPlayer.IsPlaying() {
		l.audioPlayer.SetPosition(0)
	}
	actions := make([]*Action, 0)
	isGlitching := false
	for _, t := range l.things {
		action := t.Update()
		if action != nil {
			action.Thing = t
			actions = append(actions, action)
		}
	}

	for _, a := range actions {
		if a.Glitch {
			isGlitching = true
			break
		}
	}
	l.isGlitching = isGlitching
	for _, a := range actions {
		actioner := a.Thing

		if a.Move != nil {
			hasCollision := false
			// Check for collision
			for _, t := range l.things {
				if t == nil {
					continue
				}
				if t == a.Thing {
					continue
				}
				// If we are not glitching, check for collisions
				if !actioner.IsGlitching() || t.IsBlockGlitch() {
					// On Collision
					if HasCollision(actioner, t, a.Move) && (l.player == t || l.player == actioner) {
						hasCollision = true
						if t.IsEnemy() && l.player == actioner || t == l.player && actioner.IsEnemy() {
							// Reload level
							l.audioPlayer.Close()
							l.glitchedAudioPlayer.Close()
							return l.LevelNumber
						}
						if t.IsExit() && l.player == actioner {
							// Load next level
							l.audioPlayer.Close()
							l.glitchedAudioPlayer.Close()
							return l.LevelNumber + 1
						}
						if t.IsBlock() {
							a.Move.vX = 0
							a.Move.vY = 0
						} else if l.player == t && (!l.player.isGlitching || t.IsBlockGlitch()) {
							t.ExecuteAction(&Action{
								Move: &Move{
									vX: a.Move.vX,
									vY: a.Move.vY,
								}}, isGlitching)
						}
					}
				}
			}
			if !hasCollision {
				aX, aY := actioner.Position()
				// Check for out of bounds
				if a.Move.vX+aX < 0 {
					a.Move.y = aY
					a.Move.x = 0
				} else if a.Move.vX+aX > float64(l.levelWidth*CELL_SCALE) {
					a.Move.y = aY
					a.Move.x = float64((l.levelWidth - 1) * CELL_SCALE)
				}
				if a.Move.vY+aY < 0 {
					a.Move.x = aX
					a.Move.y = 0
				} else if a.Move.vY+aY > float64(l.levelHeight*CELL_SCALE) {
					a.Move.x = aX
					a.Move.y = float64((l.levelHeight - 1) * CELL_SCALE)
				}
			}
		}
		a.Thing.ExecuteAction(a, isGlitching)
	}
	if l.isGlitching {
		if l.tickPosition <= 0 {
			l.tick = 15
			l.audioPlayer.Pause()
			l.glitchedAudioPlayer.Play()
			l.glitchedAudioPlayer.SetPosition(l.audioPlayer.Position())
			l.tickPosition = l.audioPlayer.Position()
		}
		l.tick--
		if l.tick <= 0 {
			l.tick = 15
			l.glitchedAudioPlayer.SetPosition(l.tickPosition)
		}
	} else {
		l.tick = 15
		l.tickPosition = -1
		l.glitchedAudioPlayer.Pause()
		l.audioPlayer.Play()
	}
	return 0
}

func (l *Level) Draw(screen *ebiten.Image) {

	if l.loadTick < LOAD_TICKS {
		l.loadTick++
		screen.Fill(color.Black)
		return
	}
	// Depending on the ratio of screen and level size set the scale
	// use the min ratio of the two
	width := float64(l.levelWidth * CELL_SCALE)
	height := float64(l.levelHeight * CELL_SCALE)
	widthRatio := float64(screen.Bounds().Dx()) / width
	heightRatio := float64(screen.Bounds().Dy()) / height
	levelScale := math.Min(widthRatio, heightRatio)

	// Depending on the scale, set the offset of the level
	levelOffsetX := (float64(screen.Bounds().Dx()) - width*levelScale) / 2
	levelOffsetY := (float64(screen.Bounds().Dy()) - height*levelScale) / 2

	levelOverlay := ebiten.NewImage(int(width), int(height))
	levelOverlay.Fill(color.Black)

	levelOptions := &ebiten.DrawImageOptions{}
	levelOptions.GeoM.Scale(levelScale, levelScale)
	levelOptions.GeoM.Translate(levelOffsetX, levelOffsetY)

	if l.isGlitching {
		if l.ticks%2 == 0 {
			levelOptions.GeoM.Rotate(0.01)
		}
		levelOptions.ColorScale.Scale(0, 100, 0, 0.85)
		l.ticks++
	} else {
		l.ticks = 0
	}
	for _, t := range l.things {
		if t == nil {
			continue
		}
		// Create overlay
		// Draw cells on overlay
		// Draw overly on screen
		t.Draw(levelOverlay)
		screen.DrawImage(levelOverlay, levelOptions)
	}
}
