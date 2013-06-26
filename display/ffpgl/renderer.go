package ffpgl

import (
	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"

	"github.com/bluepeppers/danckelmann/display"
	"github.com/bluepeppers/danckelmann/resources"
)

type Renderer display.RendererConfig

func CreateRenderer() display.RenderingBackend {
	// We can let this be null equivilent, as the config is set every SetupFrame
	// call.
	var conf display.RendererConfig
	return display.RenderingBackend((*Renderer)(&conf))
}

func (r Renderer) Name() string {
	return "Fixed Function Pipeline OpenGL Render"
}

func (r *Renderer) SetupFrame(config display.RendererConfig) {
	*r = Renderer(config)

	v := r.Viewport
	resources.RunInThread(func() {
		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, 0, float64(v.W), float64(v.H), -100, 100)
		gl.Translatef(float32(-v.X), float32(-v.Y), 0)
	})
}

func (r *Renderer) Draw(drawable display.Drawable) {
	x, y := drawable.Position()

	// Pixel coordinates
	screenX := (y - x) * r.TileWidth / 2
	screenY := (y + x) * r.TileHeight / 2
	for _, graphic := range drawable.Graphics() {
		resources.RunInThread(func() {
			graphic.Tex.Bind(gl.TEXTURE_2D)
			gl.Begin(gl.QUADS)
			gl.TexCoord2f(0, 0)
			gl.Vertex3i(screenX, screenY, 0)
			gl.TexCoord2f(0, 1)
			gl.Vertex3i(screenX, screenY+graphic.Width, 0)
			gl.TexCoord2f(1, 1)
			gl.Vertex3i(screenX+graphic.Width,
				screenY+graphic.Height, 0)
			gl.TexCoord2f(1, 0)
			gl.Vertex3i(screenX+graphic.Height, screenY, 0)
			gl.End()
		})
	}
}

func (r Renderer) DrawText(text string, x, y int) {
	var trans allegro.Transform
	trans.Identity()
	resources.RunInThread(func() {
		trans.Use()

		r.Font.Draw(r.TextColor, float32(x), float32(y), 0, text)
	})
}

func (r Renderer) ClearBackground() {
	resources.RunInThread(func() {
		r, g, b, a := r.BackgroundColor.GetRGBA()
		gl.ClearColor(
			gl.GLclampf(r)/255.0,
			gl.GLclampf(g)/255.0,
			gl.GLclampf(b)/255.0,
			gl.GLclampf(a)/255.0)

		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
}

func (r Renderer) EndFrame() {
	resources.RunInThread(func() {
		gl.Flush()
		allegro.Flip()
	})
}
