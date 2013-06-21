package resources

import (
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"
)

const (
	DEFAULT_GRAPHIC_NAME   = "____DEFAULT____"
	DEFAULT_GRAPHIC_WIDTH  = 128
	DEFAULT_GRAPHIC_HEIGHT = 128
)

var DEFAULT_GRAPHIC_DATA []uint32

// Generate some random pixel data for the default graphic
func init() {
	DEFAULT_GRAPHIC_DATA = make([]uint32, DEFAULT_GRAPHIC_WIDTH*DEFAULT_GRAPHIC_HEIGHT)
	for x := 0; x < DEFAULT_GRAPHIC_WIDTH; x++ {
		for y := 0; y < DEFAULT_GRAPHIC_HEIGHT; y++ {
			fw := float64(DEFAULT_GRAPHIC_WIDTH)
			fh := float64(DEFAULT_GRAPHIC_HEIGHT)
			p := y*DEFAULT_GRAPHIC_WIDTH + x

			// Something funky
			fr := math.Cos(float64(y) / fw)
			fg := math.Sin(float64(x) / fh)
			fb := math.Tan(float64(p) / float64(len(DEFAULT_GRAPHIC_DATA)))

			r := uint32(fr * 255)
			g := uint32(fg * 255)
			b := uint32(fb * 255)
			a := uint32(255)
			DEFAULT_GRAPHIC_DATA[p] =
				(r << 6) |
					(g << 4) |
					(b << 2) |
					(a << 0)
		}
	}
}

// Our custom super special graphic class
type Graphic struct {
	Tex gl.Texture
	// The offset of the graphic when drawing
	OffX, OffY int
	// Dimensions
	Width, Height int
}

type ResourceManager struct {
	graphicsMutex sync.RWMutex
	graphics      map[string]Graphic

	fontMap map[string]*allegro.Font
}

func CreateResourceManager() *ResourceManager {
	var manager ResourceManager

	manager.graphics = make(map[string]Graphic)
	manager.addDefaultGraphic()

	manager.fontMap = make(map[string]*allegro.Font)
	RunInThread(func() {
		manager.fontMap["builtin"] = allegro.CreateBuiltinFont()
	})
	return &manager
}

func (rm *ResourceManager) addDefaultGraphic() {
	var graphic Graphic
	RunInThread(func() {
		graphic.Tex = gl.GenTexture()
		graphic.Tex.Bind(gl.TEXTURE_2D)

		// Maybe REPEAT will be more noticable?
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

		// Why not
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
			DEFAULT_GRAPHIC_WIDTH, DEFAULT_GRAPHIC_HEIGHT,
			0, gl.RGBA, gl.UNSIGNED_INT_8_8_8_8, DEFAULT_GRAPHIC_DATA)
	})
	graphic.Width = DEFAULT_GRAPHIC_WIDTH
	graphic.Height = DEFAULT_GRAPHIC_HEIGHT

	rm.graphics[DEFAULT_GRAPHIC_NAME] = graphic
}

// Adds the graphic to the resource manager with the given name. Overwrites
// existing graphics with the same name.
func (rm *ResourceManager) AddGraphic(name string, graphic Graphic) {
	if name == DEFAULT_GRAPHIC_NAME {
		log.Fatalf("Cannot add graphic with default name %q", name)
	}

	rm.graphicsMutex.Lock()
	defer rm.graphicsMutex.Unlock()
	rm.graphics[name] = graphic
}

func (rm *ResourceManager) GetGraphic(name string) (*Graphic, error) {
	rm.graphicsMutex.RLock()
	defer rm.graphicsMutex.RUnlock()
	graphic, ok := rm.graphics[name]
	if !ok {
		return nil, fmt.Errorf("Could not find graphic %q", name)
	}
	return &graphic, nil
}

// Gets a tile that can be drawn, no matter what. Won't be pretty, but won't crash.
func (rm *ResourceManager) GetDefaultGraphic() *Graphic {
	sub, err := rm.GetGraphic(DEFAULT_GRAPHIC_NAME)
	if err != nil {
		log.Fatalf("Could not find default graphic %q", DEFAULT_GRAPHIC_NAME)
	}
	return sub
}

func (rm *ResourceManager) GetGraphicOrDefault(name string) *Graphic {
	graphic, err := rm.GetGraphic(name)
	if err != nil {
		log.Printf("Could not find graphic named %q. Defaulting to default graphic.", name)
		return rm.GetDefaultGraphic()
	}
	return graphic
}

func (rm *ResourceManager) GetFont(name string) (*allegro.Font, error) {
	font, ok := rm.fontMap[name]
	if !ok {
		return nil, fmt.Errorf("Could not find font %q", name)
	}
	return font, nil
}
