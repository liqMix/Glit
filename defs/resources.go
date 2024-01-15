package defs

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kettek/go-multipath/v2"
)

type Resources struct {
	files     multipath.FS
	spriteMap map[string]*image.Image  // Indexed by name
	levelMap  map[string]string        // Indexed by name
	thingMap  map[string]*Thing        // Indexed by symbol
	musicMap  map[string]vorbis.Stream // musicMap
	// soundMap
}

var (
	ErrNoSuchCategory   = errors.New("no such category")
	ErrMissingDirectory = errors.New("missing directory")
)

func InitResources(embedFS embed.FS) *Resources {
	resources := &Resources{
		spriteMap: make(map[string]*image.Image),
		levelMap:  make(map[string]string),
		thingMap:  make(map[string]*Thing),
		musicMap:  make(map[string]vorbis.Stream),
	}
	// Allow loading from filesystem.
	resources.files.InsertFS(os.DirFS("assets"), multipath.FirstPriority)

	// Also allow loading from embedded filesystem.
	sub, err := fs.Sub(embedFS, "assets")
	if err != nil {
		panic(err)
	}
	resources.files.InsertFS(sub, multipath.LastPriority)

	resources.Load()
	return resources
}

func (r *Resources) GetSprite(name string) *ebiten.Image {
	return ebiten.NewImageFromImage(*r.spriteMap[name+".png"])
}

func (r *Resources) GetLevel(name string) string {
	return r.levelMap[name]
}

func (r *Resources) GetThing(roon rune) *Thing {
	return r.thingMap[string(roon)]
}

func (r *Resources) GetMusic(name string) vorbis.Stream {
	return r.musicMap[name+".ogg"]
}

func (r *Resources) Load() error {
	// Load sprites from assets/sprites.
	err := r.files.Walk("sprites", func(path string, entry fs.DirEntry, err error) error {
		if entry == nil {
			return ErrMissingDirectory
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		_, img, err := ebitenutil.NewImageFromFileSystem(r.files, fmt.Sprintf("sprites/%s", name))
		if err != nil {
			return err
		}

		r.spriteMap[name] = &img
		fmt.Printf("Loaded image %s\n", name)
		return nil
	})
	if err != nil {
		return err
	}

	// Load music from assets/music.
	err = r.files.Walk("music", func(path string, entry fs.DirEntry, err error) error {
		if entry == nil {
			return ErrMissingDirectory
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		b, err := r.files.ReadFile(fmt.Sprintf("music/%s", name))
		if err != nil {
			return err
		}
		song, err := vorbis.DecodeWithSampleRate(44100, bytes.NewReader(b))
		if err != nil {
			return err
		}
		r.musicMap[name] = *song
		fmt.Printf("Loaded music %s\n", name)
		return nil
	})
	if err != nil {
		return err
	}

	// Load levels from assets/levels.
	err = r.files.Walk("levels", func(path string, entry fs.DirEntry, err error) error {
		if entry == nil {
			return ErrMissingDirectory
		}
		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		bytes, err := r.files.ReadFile(fmt.Sprintf("levels/%s", name))
		if err != nil {
			return err
		}
		// Bytes to text
		level := string(bytes)
		r.levelMap[name] = level

		// print message
		fmt.Printf("Loaded level %s\n", name)
		return nil
	})
	if err != nil {
		return err
	}

	// Load things from assets/things.
	err = r.files.Walk("things", func(path string, entry fs.DirEntry, err error) error {
		if entry == nil {
			return ErrMissingDirectory
		}
		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		bytes, err := r.files.ReadFile(fmt.Sprintf("things/%s", name))
		if err != nil {
			return err
		}

		// Bytes to text
		thing, err := ThingFromFile(bytes)
		if err != nil {
			return err
		}
		r.thingMap[thing.Symbol] = thing

		// Associate sprite with thing
		thing.sprite = r.GetSprite(thing.Name)
		thing.width = thing.sprite.Bounds().Dx()
		thing.height = thing.sprite.Bounds().Dy()

		// print message
		fmt.Printf("Loaded thing %s\n", thing.Name)
		return nil
	})

	return err
}
