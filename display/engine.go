package display

import (
	"fmt"
	"log"
	"sync"

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
	GetDrawable(int) (Drawable, bool)

	// Passes a fully initialized DisplayEngine to the GameEngine. This
	// allows the GameEngine to inform the DisplayEngine of changes of state
	// without the DisplayEngine having to explicitly poll for them.
	RegisterDisplayEngine(DisplayEngine)

	// Called when the display receivs a DisplayCloseEvent or the game is
	// otherwise terminated
	GameFinished()
}

func initializeAllegro() {
	allegro.Init()
	allegro.InitFont()
	allegro.InitImage()
	allegro.InitTTF()
	allegro.InstallKeyboard()
	allegro.InstallMouse()
}

type DisplayEngine struct {
	config     DisplayConfig
	gameEngine GameEngine

	statusLock sync.RWMutex
	running    bool

	drawLock         sync.RWMutex
	frameDrawing     sync.RWMutex // Locked -> Frame drawing atm
	currentFrame     int
	viewport         Viewport
	Display          *allegro.Display
	fps              float64
	cursorX, cursorY float64

	resourceManager *resources.ResourceManager
}

func CreateDisplayEngine(resourceDir string, conf *allegro.Config, gameEngine GameEngine) *DisplayEngine {
	resources.RunInThread(initializeAllegro)

	var displayEngine DisplayEngine

	displayEngine.Display = createDisp(conf)
	var err error
	displayEngine.resourceManager, err = createResourceManager(conf)
	if err != nil {
		log.Fatalf("%v", err)
	}

	displayEngine.running = false

	w, h := displayEngine.Display.GetDimensions()
	displayEngine.viewport = CreateViewport(-w/2, -h/w, w, h, 1.0, 1.0)

	displayEngine.gameEngine = gameEngine
	displayEngine.config = displayEngine.gameEngine.GetDisplayConfig()
	displayEngine.gameEngine.RegisterDisplayEngine(displayEngine)

	return &displayEngine
}

func createResourceManager(conf *allegro.Config) (*resources.ResourceManager, error) {
	rm := resources.CreateResourceManager()
	if rm == nil {
		return nil, fmt.Errorf("Could not create new resource manager")
	}
	return rm, nil
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

	var disp *allegro.Display
	resources.RunInThread(func() {
		disp = allegro.CreateDisplay(1, 1, flags)
	})
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

func (d *DisplayEngine) GetViewport() *Viewport {
	d.drawLock.RLock()
	defer d.drawLock.RUnlock()
	return &d.viewport
}

func (d *DisplayEngine) SetViewport(v *Viewport) {
	d.drawLock.Lock()
	d.viewport = *v
	d.drawLock.Unlock()
}

func (d *DisplayEngine) GetResourceManager() *resources.ResourceManager {
	return d.resourceManager
}
