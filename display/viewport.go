package display

import (
	"github.com/bluepeppers/allegro"
)

type Viewport struct {
	X, Y, W, H int
}

func (v *Viewport) GetTransform(d *allegro.Display) *allegro.Transform {
	var final allegro.Transform
	width, height := d.GetDimensions()
	final.Build(float32(-v.X), float32(-v.Y),
		float32(v.W)/float32(width), float32(v.H)/float32(height), 0)
	return &final
}
