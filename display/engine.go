package display

import (
	"sync"
	"log"

	"github.com/bluepeppers/allegro"

	"github.com/bluepeppers/danckelmann/resources"
	"github.com/bluepeppers/danckelmann/config"
)

const (
	// Default dimensions of the display. Often not used
	DEFAULT_WIDTH = 600
	DEFAULT_HEIGHT = 400
)

// The interface that a game engine must implement for the display engine to
// be able to display it
type GameEngine interface {
	// Returns the display configuration for the engine. Will be called
	// once at engine startup.
	GetDisplayConfig() DisplayConfig

	// For best results, make sure that the bitmaps all share a common
	// parent. This is done automatically if they are loaded via a
	// resouces.ResourceManager
	GetTile(int, int) []*allegro.Bitmap

	// Passes a fully initialized DisplayEngine to the GameEngine. This
	// allows the GameEngine to inform the DisplayEngine of changes of state
	// without the DisplayEngine having to explicitly poll for them.
	RegisterDisplayEngine(*DisplayEngine)
}

type DisplayEngine struct {
	config     DisplayConfig
	gameEngine *GameEngine

	statusLock sync.RWMutex
	running    bool

	drawLock     sync.RWMutex
	currentFrame int
	viewport     Viewport
	display      *allegro.Display

	resourceManager *resources.ResourceManager
}

func CreateDisplayEngine(conf *allegro.Config, gameEngine *GameEngine) *DisplayEngine {
	var displayEngine DisplayEngine

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		displayEngine.display = createDisp(conf)
		wg.Done()
	}()
	go func() {
		resourceDir := config.GetString(conf, "resources", "directory", ".")
		conf, ok := resources.LoadResourceManagerConfig(resourceDir, "")
		if !ok {
			log.Fatalf("Could not load resource manager config")
		}
		displayEngine.resourceManager = resources.CreateResourceManager(conf)
		wg.Done()
	}()
	wg.Wait()

	displayEngine.running = false

	displayEngine.viewport.W, displayEngine.viewport.H = displayEngine.display.GetDimensions()
	displayEngine.viewport.X = -displayEngine.viewport.W / 2
	displayEngine.viewport.Y = -displayEngine.viewport.H / 2

	displayEngine.gameEngine = gameEngine
	displayEngine.config = (*gameEngine).GetDisplayConfig()
	(*gameEngine).RegisterDisplayEngine(&displayEngine)

	return &displayEngine
}

func createDisp(conf *allegro.Config) *allegro.Display {
	width := config.GetInt(conf, "display", "width", DEFAULT_WIDTH)
	height := config.GetInt(conf, "display", "height", DEFAULT_HEIGHT)

	flags := allegro.RESIZABLE
	switch config.GetString(conf, "display", "windowed", "windowed") {
	case "fullscreen":
		flags |= allegro.FULLSCREEN
	case "fullscreenwindow":
		flags |= allegro.FULLSCREEN_WINDOW
	default:
		log.Printf("display.windowed not one of \"fullscreen\", \"fullscreenwindow\", or \"windowed\"")
		log.Printf("Defaulting to display.windowed=\"windowed\"")
		fallthrough
	case "windowed":
		flags |= allegro.WINDOWED
	}

	disp := allegro.CreateDisplay(width, height, flags)
	if disp == nil {
		log.Fatalf("Could not create display")
	}
	return disp
}

func (d *DisplayEngine) Stop() {
	d.statusLock.Lock()
	d.running = false
	d.statusLock.Unlock()
}

func (d *DisplayEngine) MoveViewport(v Viewport) {
	d.drawLock.Lock()
	d.viewport = v
	d.drawLock.Unlock()
}

func (d *DisplayEngine) Run() {
	running := true
	d.statusLock.Lock()
	d.running = true
	d.statusLock.Unlock()
	lastFrame := 1
	d.drawLock.Lock()
	d.currentFrame = lastFrame
	d.drawLock.Unlock()
	for running {
		d.drawLock.RLock()
		if d.currentFrame != lastFrame {
			lastFrame = d.currentFrame
			go d.drawFrame(lastFrame)
		}
		d.drawLock.RUnlock()

		d.statusLock.RLock()
		running = d.running
		d.statusLock.RUnlock()
	}
}

func (d *DisplayEngine) drawFrame(currFrame int) {
	toDraw := make([][]*allegro.Bitmap, d.config.MapW*d.config.MapH)
	drawPasses := 0
	for x := 0; x < d.config.MapW; x++ {
		for y := 0; y < d.config.MapH; y++ {
			toDraw[x*d.config.MapW+y] = (*d.gameEngine).GetTile(x, y)
			length := len(toDraw[x*d.config.MapW+y])
			if length > drawPasses {
				drawPasses = length
			}
		}
	}

	// Don't want anyone changing the viewport mid frame or any such highjinks
	d.drawLock.RLock()
	viewport := d.viewport
	d.drawLock.RUnlock()
	d.display.SetTargetBackbuffer()
	for p := 0; p < drawPasses; p++ {
		for s := 0; s < d.config.MapW+d.config.MapH; s++ {
			start := 0
			if s > d.config.MapW {
				start = s - d.config.MapW
			}
			for x := start; x < s; x++ {
				y := s - x
				// Trust me, I study maths
				// Coordinates in terms of pixels
				px := (y/2 - x/2) * d.config.TileW
				py := (x/2 - y/2) * d.config.TileH
				bmp := toDraw[x*d.config.MapW+y][p]
				// Coordinates in terms of pixels on screen
				sx := px - viewport.X
				sy := py - viewport.Y
				if sx < -d.config.TileW || sx > viewport.W ||
					sy < d.config.TileH || sy > viewport.H {
					continue
				}
				bmp.Draw(float32(sx), float32(sy), 0)
			}
		}
		d.drawLock.RLock()
		viewport = d.viewport
		d.drawLock.RUnlock()
	}
	allegro.Flip()

	d.drawLock.Lock()
	d.currentFrame = currFrame
	d.drawLock.Unlock()
}
