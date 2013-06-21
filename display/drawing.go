package display

import (
	"fmt"
	"time"

	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"

	"github.com/bluepeppers/danckelmann/resources"
)

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
	// Get all of the graphics to be drawn, and put them into a big array
	toDraw := make([][]*resources.Graphic, d.config.MapW*d.config.MapH)
	drawPasses := 0
	for x := 0; x < d.config.MapW; x++ {
		for y := 0; y < d.config.MapH; y++ {
			toDraw[x*d.config.MapW+y] = d.gameEngine.GetTile(x, y)
			length := len(toDraw[x*d.config.MapW+y])
			if length > drawPasses {
				drawPasses = length
			}
		}
	}

	viewport := d.viewport

	resources.RunInThread(func() {
		// Don't want anyone changing the viewport mid frame or any such highjinks
		d.Display.SetTargetBackbuffer()

		r, g, b, a := d.config.BGColor.GetRGBA()
		gl.ClearColor(
			gl.GLclampf(r)/255.0,
			gl.GLclampf(g)/255.0,
			gl.GLclampf(b)/255.0,
			gl.GLclampf(a)/255.0)

		gl.Clear(gl.COLOR_BUFFER_BIT)

		viewport.SetupTransform()

		for pass := 0; pass < drawPasses; pass++ {
			mapW := d.config.MapW
			mapH := d.config.MapH
			// Iterate through the map coordinates diagonally
			for s := 0; s < mapW+mapH; s++ {
				for x := 0; x < s; x++ {
					y := s - x - 1
					if x >= mapW || y < 0 || y >= mapH {
						continue
					}
					// If this tile doesn't have any graphics on this pass, cont.
					if len(toDraw[x*d.config.MapW+y]) < p {
						continue
					}

					// Coordinates in terms of pixels
					px := (y - x) * d.config.TileW / 2
					py := (x + y) * d.config.TileH / 2
					graphic := toDraw[x*d.config.MapW+y][p]
					/*					ox := bmp.OffX
										oy := bmp.OffY*/
					gw, gh := graphic.Width, graphic.Height
					if viewport.OnScreen(px, py, gw, gh) {
						gl.Begin(gl.QUADS)
						graphic.Tex.Bind(gl.TEXTURE_2D)
						gl.TexCoord2f(0, 0)
						gl.Vertex3i(px, py, 0)
						gl.TexCoord2f(0, 1)
						gl.Vertex3i(px, py+gw, 0)
						gl.TexCoord2f(1, 1)
						gl.Vertex3i(px+gh, py+gw, 0)
						gl.TexCoord2f(1, 0)
						gl.Vertex3i(px+gh, py, 0)
						gl.End()
					}
				}
			}
		}

		gl.Flush()

		var trans allegro.Transform
		trans.Identity()
		trans.Use()

		font, err := d.resourceManager.GetFont("builtin")
		if err == nil {
			font.Draw(allegro.CreateColor(0, 255, 0, 255), 0, 0, 0, fmt.Sprint(int(d.fps)))
		}

		allegro.Flip()
	})

	d.frameDrawing.Unlock()
}
