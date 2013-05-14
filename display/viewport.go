package display

import (
//	"log"
	"math"
	
	"github.com/bluepeppers/allegro"
)

var ISOMETRIC_ROTATION = float32(3 * math.Pi / 8)

type Viewport struct {
	x, y, w, h int
	xZoom, yZoom float64

	trans allegro.Transform
}

func CreateViewport(x, y, w, h int, xZoom, yZoom float64) Viewport {
	var v Viewport
	v.x, v.y, v.w, v.h = x, y, w, h
	v.xZoom, v.yZoom = xZoom, yZoom
	v.buildTrans()
	return v
}

func (v *Viewport) ResizeViewport(w, h int) {
	v.w = w
	v.h = h
	v.buildTrans()
}

func (v *Viewport) Move(dx, dy int) {
	v.x += dx
	v.y += dy
}

func (v *Viewport) GetTransform() *allegro.Transform {
	return &v.trans
}

func (v *Viewport) buildTrans() {
	v.trans.Identity()
	v.trans.Build(float32(-v.x), float32(-v.y), float32(v.xZoom), float32(v.yZoom), 0)
}

func (v *Viewport) OnScreen(x, y, w, h int) bool {
	return !(
        // Off left side
        x + w < v.x  ||
        // Off right side
        x > v.x + int(float64(v.w) * v.xZoom) ||
        // Off top
        y + h < v.y ||
        // Off bottom
        y > v.y + int(float64(v.h) * v.yZoom))
}

func (v *Viewport) TileCoordinatesToScreen(tx, ty float64, config DisplayConfig) (float64, float64) {
	var trans allegro.Transform
	trans.Build(float32(-v.x), float32(-v.y), float32(v.xZoom), float32(v.yZoom),
		ISOMETRIC_ROTATION)
	x, y := trans.Apply(float32(tx), float32(ty))
	return float64(x), float64(y)
}

func (v *Viewport) ScreenCoordinatesToTile(sx, sy int, config DisplayConfig) (float64, float64) {
	w, h := float64(config.TileW), float64(config.TileH)
	fx, fy := float32(sx), float32(sy)
	var trans allegro.Transform
	trans.Identity()
	trans.Build(float32(-v.x), float32(-v.y), float32(v.xZoom), float32(v.yZoom),
		0)
	trans.Invert()
	trans.Translate(-float32(w/2), -float32(h/2))

	x, y := trans.Apply(fx, fy)
	tx := float64(float64(y) * w - float64(x) * h) / (w * h)
	ty := float64(float64(y) * w + float64(x) * h) / (w * h)
	return tx, ty
}