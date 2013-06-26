package display

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/bluepeppers/allegro"

	"github.com/bluepeppers/danckelmann/resources"
	"github.com/bluepeppers/danckelmann/utils/priorityqueue"
)

type RendererConfig struct {
	TileWidth, TileHeight      int
	Viewport                   Viewport
	TextColor, BackgroundColor allegro.Color
	Font                       *allegro.Font
}

// The rendering backend.
type RenderingBackend interface {
	// The name of the backend. For logging purposes and jazz
	Name() string

	// Set up various matricies etc
	SetupFrame(RendererConfig)

	// Draws a drawable
	Draw(Drawable)

	// Draws the text to the given position
	DrawText(string, int, int)

	// Sets the entire screen to the background color
	ClearBackground()

	// Flip the backbuffer etc
	EndFrame()
}

type Drawable interface {
	// Lower layers are drawn before higher ones
	Layer() int

	// The lowest position of the tiles occupied by the drawable. Tiles lower
	// down will be drawn after tiles closer to the top of the screen
	Position() (int, int)

	// The graphics to be drawn
	Graphics() []resources.Graphic
}

func (d *DisplayEngine) Run(backend RenderingBackend) {
	running := true
	d.statusLock.Lock()
	d.running = true
	d.statusLock.Unlock()

	go d.eventHandler()

	start := time.Now()
	frames := 0

	for running {
		d.frameDrawing.Lock()
		go d.drawFrame(backend)
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

func (d *DisplayEngine) drawFrame(renderer RenderingBackend) {
	// Get all of the graphics to be drawn, and put them into a PriorityQueue
	// for drawing
	toDraw := (*priorityqueue.PriorityQueue)(&[]*priorityqueue.Item{})
	heap.Init(toDraw)

	for i := 0; ; i++ {
		drawable, ok := d.gameEngine.GetDrawable(i)
		if !ok {
			break
		}
		dx, dy := drawable.Position()
		dLayer := drawable.Layer()

		// Layer > Y coord > X coord
		priority := dLayer*d.config.MapW*d.config.MapH +
			dy*d.config.MapW +
			dx

		heap.Push(toDraw, &priorityqueue.Item{Value: drawable, Priority: priority})
	}

	config := RendererConfig{
		d.config.TileW, d.config.TileH,
		d.viewport,
		allegro.CreateColor(255, 0, 0, 0),
		allegro.CreateColor(0, 0, 255, 0),
		allegro.CreateBuiltinFont()}

	renderer.SetupFrame(config)

	renderer.ClearBackground()

	for toDraw.Len() != 0 {
		drawable := heap.Pop(toDraw).(priorityqueue.Item).Value.(Drawable)
		// The coordinates in the tile system
		tileX, tileY := drawable.Position()

		// The coordinates in the pixel system
		pixelX := (tileY - tileX) * config.TileWidth / 2
		pixelY := (tileY + tileX) * config.TileHeight / 2
		// Width & Height is the largest of the graphic properties
		pixelWidth := 0
		pixelHeight := 0
		for _, graphic := range drawable.Graphics() {
			if graphic.Width > pixelWidth {
				pixelWidth = graphic.Width
			}
			if graphic.Height > pixelHeight {
				pixelHeight = graphic.Height
			}
		}

		if config.Viewport.OnScreen(pixelX, pixelY, pixelWidth, pixelHeight) {
			renderer.Draw(drawable)
		}
	}

	renderer.DrawText(fmt.Sprint(int(d.fps)), 0, 0)

	renderer.EndFrame()

	d.frameDrawing.Unlock()
}
