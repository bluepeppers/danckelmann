package display

import (
	"log"
	"sync"
	"time"

	"github.com/bluepeppers/allegro"

	"github.com/bluepeppers/danckelmann/config"
	"github.com/bluepeppers/danckelmann/resources"
)

const (
	// Default dimensions of the display. Often not used
	DEFAULT_WIDTH  = 600
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

	// Called when the display receivs a DisplayCloseEvent or the game is
	// otherwise terminated
	GameFinished()
}

func InitializeAllegro() {
	allegro.Init()
	allegro.InitFont()
	allegro.InitImage()
	allegro.InitTTF()
	allegro.InstallKeyboard()
	allegro.InstallMouse()
}

type DisplayEngine struct {
	config     DisplayConfig
	gameEngine *GameEngine

	statusLock sync.RWMutex
	running    bool

	drawLock     sync.RWMutex
	frameDrawing sync.RWMutex // Locked -> Frame drawing atm
	currentFrame int
	viewport     Viewport
	Display      *allegro.Display

	resourceManager *resources.ResourceManager
}

func CreateDisplayEngine(resourceDir string, conf *allegro.Config, gameEngine GameEngine) *DisplayEngine {
	var displayEngine DisplayEngine

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		displayEngine.Display = createDisp(conf)
		wg.Done()
	}()
	go func() {
		conf, ok := resources.LoadResourceManagerConfig(resourceDir, "")
		if !ok {
			log.Fatalf("Could not load resource manager config from %q", resourceDir)
		}
		displayEngine.resourceManager = resources.CreateResourceManager(conf)
		wg.Done()
	}()
	wg.Wait()

	displayEngine.running = false

	displayEngine.viewport.W, displayEngine.viewport.H = displayEngine.Display.GetDimensions()
	displayEngine.viewport.X = -displayEngine.viewport.W / 2
	displayEngine.viewport.Y = -displayEngine.viewport.H / 2

	displayEngine.gameEngine = &gameEngine
	displayEngine.config = (*displayEngine.gameEngine).GetDisplayConfig()
	(*displayEngine.gameEngine).RegisterDisplayEngine(&displayEngine)

	return &displayEngine
}

func createDisp(conf *allegro.Config) *allegro.Display {
	/*	width := config.GetInt(conf, "display", "width", DEFAULT_WIDTH)
		height := config.GetInt(conf, "display", "height", DEFAULT_HEIGHT)*/

	flags := allegro.RESIZABLE
	switch config.GetString(conf, "display", "windowed", "fullscreenwindow") {
	case "fullscreen":
		flags |= allegro.FULLSCREEN
	case "windowed":
		flags |= allegro.WINDOWED
	default:
		log.Printf("display.windowed not one of \"fullscreen\", \"fullscreenwindow\", or \"windowed\"")
		log.Printf("Defaulting to display.windowed=\"fullscreenwindow\"")
		fallthrough
	case "fullscreenwindow":
		flags |= allegro.FULLSCREEN_WINDOW
	}

	disp := allegro.CreateDisplay(1, 1, flags)
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

func (d *DisplayEngine) GetViewport() Viewport {
	d.drawLock.RLock()
	defer d.drawLock.RUnlock()
	return d.viewport
}

func (d *DisplayEngine) SetViewport(v Viewport) {
	d.drawLock.Lock()
	d.viewport = v
	d.drawLock.Unlock()
}

func (d *DisplayEngine) GetResourceManager() *resources.ResourceManager {
	return d.resourceManager
}

func (d *DisplayEngine) Run() {
	running := true
	d.statusLock.Lock()
	d.running = true
	d.statusLock.Unlock()

	go d.eventHandler()

	start := time.Now()
	frames := 0
	for running {
		d.frameDrawing.Lock()
		go d.drawFrame()
		frames++
		if frames >= 30 {
			fps := float64(frames) / time.Since(start).Seconds()
			log.Printf("FPS: %v", fps)
			start = time.Now()
			frames = 0
		}

		d.statusLock.RLock()
		running = d.running
		d.statusLock.RUnlock()
	}
}

func (d *DisplayEngine) drawFrame() {
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
	d.viewport.W, d.viewport.H = d.Display.GetDimensions()
	viewport := d.viewport
	d.drawLock.RUnlock()

	d.Display.SetTargetBackbuffer()
	d.config.BGColor.Clear()

	viewport.GetTransform(d.Display).Use()

	allegro.HoldBitmapDrawing(true)
	for p := 0; p < drawPasses; p++ {
		m := d.config.MapW
		n := d.config.MapH
		for s := 0; s < m+n; s++ {
			for x := 0; x < s; x++ {
				y := s - x - 1

				if x >= m || y < 0 || y >= n {
					continue
				}
				if len(toDraw[x*d.config.MapW+y]) < p {
					continue
				}

				// Trust me, I study maths
				// Coordinates in terms of pixels
				px := (y-x)*d.config.TileW/2 - d.config.TileW/2
				py := (x+y)*d.config.TileH/2 - d.config.TileH/2
				bmp := toDraw[x*d.config.MapW+y][p]
				bw, bh := bmp.GetDimensions()
				if viewport.OnScreen(px, py, bw, bh) {
					bmp.Draw(float32(px), float32(py), 0)
				}
			}
		}
		d.drawLock.RLock()
		viewport = d.viewport
		d.drawLock.RUnlock()
	}

	allegro.HoldBitmapDrawing(false)
	allegro.Flip()

	d.frameDrawing.Unlock()
}
