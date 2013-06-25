package ffpgl

import (
	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"

	"github.com/bluepeppers/danckelmann/resources"
	"github.com/bluepeppers/danckelmann/display"
)

type Renderer display.RendererConfig

func (r Renderer) Name() string {
	return "Fixed Function Pipeline OpenGL Render"
}

func (r *Renderer) SetupFrame(config display.RendererConfig) {
	*r = &config

	v := r.Viewport
	resources.RunInThread(func(){
		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, 0, float64(v.w), float64(v.h), -100, 100)
		gl.Translatef(float32(-v.x), float32(-v.y), 0)
		config.Viewport.SetupTransform()
	})
}

func (r *Renderer) Draw(drawable Drawable) {
	x, y := drawable.Position()

	graphics := drawable.Graphics()

	// Pixel coordinates
	screenX := (y - x) * r.TileWidth / 2
	screenY := (y + x) * r.TileHeight / 2
	resources.RunInThread(func() {
		graphic.Tex.Bind(gl.TEXTURE_2D)
		gl.Begin(gl.QUADS)
		gl.TexCoord2f(0, 0)
		gl.Vertex3i(screenX, screenY, 0)
		gl.TexCoord2f(0, 1)
		gl.Vertex3i(screenX, screenY + graphic.Width, 0)
		gl.TexCoord2f(1, 1)
		gl.Vertex3i(screenX + graphic.Width,
			screenY + graphic.Height, 0)
		gl.TexCoord2f(1, 0)
		gl.Vertex3i(screenX + graphic.Height, screenY, 0)
		gl.End()
	})
}

func (r Renderer) DrawText(text string, x, y int) {
	var trans allegro.Transform
	trans.Identity()
	resources.RunInThread(func() {
		trans.Use()

		config.Font.Draw(r.TextColor, x, y, 0, text)
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