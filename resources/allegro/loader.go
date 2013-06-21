package allegro

import (
	"fmt"
	"log"
	"path"

	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"

	"github.com/bluepeppers/danckelmann/resources"
)

func init() {
	resources.RunInThread(func() {
		allegro.InitImage()
	})

	err := resources.RegisterLoader(AllegroLoader{})
	if err != nil {
		log.Printf("Could not register Allegro loader: %v", err)
	}
}

type AllegroLoader struct{}

func (l AllegroLoader) Extensions() []string {
	// While the implementation may support more, these are the ones we know
	// are always supported
	return []string{"bmp", "pcx", "tga", "jpeg", "png"}
}

func (l AllegroLoader) LoadFile(filename string) (map[string]resources.Graphic, error) {
	var bmp *allegro.Bitmap
	resources.RunInThread(func() {
		bmp = allegro.LoadBitmap(filename)
	})
	if bmp == nil {
		return nil, fmt.Errorf("Could not load graphic %q", filename)
	}
	var graphic resources.Graphic
	resources.RunInThread(func() {
		graphic.Tex = gl.Texture(bmp.GetGLTexture())
		graphic.Width, graphic.Height = bmp.GetDimensions()
	})
	if int(graphic.Tex) == 0 {
		return nil, fmt.Errorf("Could not graphic texture for %q", filename)
	}

	name := path.Base(filename)

	return map[string]resources.Graphic{name: graphic}, nil
}
