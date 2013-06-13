package display

import (
	"fmt"
	"log"
	"sync"
	"time"
	"path/filepath"

	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"

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
	GetTile(int, int) []*resources.Bitmap

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
	var displayEngine DisplayEngine

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		displayEngine.Display = createDisp(conf)
		wg.Done()
	}()
	go func() {
		displayEngine.resourceManager, err = createResourceManager(conf)
		if err != nil {
			fmt.Fatalf(err)
		}
		wg.Done()
	}()
	wg.Wait()

	displayEngine.running = false

	w, h := displayEngine.Display.GetDimensions()
	displayEngine.viewport = CreateViewport(-w/2, -h/w, w, h, 1.0, 1.0)

	displayEngine.gameEngine = &gameEngine
	displayEngine.config = (*displayEngine.gameEngine).GetDisplayConfig()
	(*displayEngine.gameEngine).RegisterDisplayEngine(&displayEngine)

	return &displayEngine
}

func createResourceManager(conf *allegro.Config) (*resources.ResourceManager, error) {
	rm := resources.CreateResourceManager()
	if rm == nil {
		return nil, fmt.Errorf("Could not create new resource manager")
	}
	dataDir := config.GetString(conf, "data", "asset_dir", "DATA")
	sg3Files, err := filepath.Glob(filepath.Join(dataDir, "*.(sg|SG)3")
	if err != nil {
		return nil, err
	}
	for _, fileSG3 := range sg3Files {
		prefix := strings.TrimSuffix(strings.TrimSuffix(fileSG3, ".sg3"), ".SG3")
		file555 := prefix + ".555"
		// Images are already loaded in parallel here, not sure it's worth
		// doing these calls in a seperate routine
		err = rm.LoadSG3File(fileSG3, file555, prefix)
		if err != nil {
			return nil, err
		}
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
			d.fps = float64(frames) / time.Since(start).Seconds()
			start = time.Now()
			frames = 0
		}

		d.statusLock.RLock()
		running = d.running
		d.statusLock.RUnlock()
	}
}

func (d *DisplayEngine) drawFrame() {
	toDraw := make([][]*resources.Bitmap, d.config.MapW*d.config.MapH)
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

	viewport := d.viewport
	font := allegro.CreateBuiltinFont()

	// Don't want anyone changing the viewport mid frame or any such highjinks
	d.Display.SetTargetBackbuffer()

	allegro.RunInThread(func() {
		r, g, b, a := d.config.BGColor.GetRGBA()
		gl.ClearColor(
			gl.GLclampf(r)/255.0,
			gl.GLclampf(g)/255.0,
			gl.GLclampf(b)/255.0,
			gl.GLclampf(a)/255.0)

		gl.Clear(gl.COLOR_BUFFER_BIT)

		viewport.SetupTransform()

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

					// Coordinates in terms of pixels
					px := (y - x) * d.config.TileW / 2
					py := (x + y) * d.config.TileH / 2
					bmp := toDraw[x*d.config.MapW+y][p]
					/*					ox := bmp.OffX
										oy := bmp.OffY*/
					bw, bh := bmp.W, bmp.H
					if viewport.OnScreen(px, py, bw, bh) {
						gl.Begin(gl.QUADS)
						bmp.Tex.Bind(gl.TEXTURE_2D)
						gl.TexCoord2f(0, 0)
						gl.Vertex3i(px, py, 0)
						gl.TexCoord2f(0, 1)
						gl.Vertex3i(px, py+bw, 0)
						gl.TexCoord2f(1, 1)
						gl.Vertex3i(px+bh, py+bw, 0)
						gl.TexCoord2f(1, 0)
						gl.Vertex3i(px+bh, py, 0)
						gl.End()
					}
				}
			}
		}

		gl.Flush()
	})

	var trans allegro.Transform
	trans.Identity()
	trans.Use()

	font.Draw(allegro.CreateColor(0, 255, 0, 255), 0, 0, 0, fmt.Sprint(int(d.fps)))

	allegro.Flip()

	d.frameDrawing.Unlock()
}
