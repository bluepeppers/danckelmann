package resources

import (
	"log"
	"fmt"
	"math"
	"sync"

	libsg "github.com/bluepeppers/libsg/golibsg"
	"github.com/bluepeppers/allegro"
	"github.com/go-gl/gl"
)

const (
	DEFAULT_GRAPHIC_NAME   = "____DEFAULT____"
	DEFAULT_GRAPHIC_WIDTH  = 128
	DEFAULT_GRAPHIC_HEIGHT = 128
)

var DEFAULT_GRAPHIC_DATA uint32[];

func init() {
	DEFAULT_GRAPHIC_DATA = make(uint32[], DEFAULT_GRAPHIC_WIDTH * DEFAULT_GRAPHIC_HEIGHT)
	for x := 0; x < DEFAULT_GRAPHIC_WIDTH; x++ {
		for y := 0; y < DEFAULT_GRAPHIC_HEIGHT; y++ {
			p := y * DEFAULT_GRAPHIC_WIDTH + x

			// Something funky
			fr := math.Cos(y/DEFAULT_GRAPHIC_WIDTH)
			fg := math.Sin(x/DEFAULT_GRAPHIC_WIDTH)
			fb := math.Tan(p/len(DEFAULT_GRAPHIC_DATA))

			r := byte(fr * 255)
			g := byte(fg * 255)
			b := byte(fb * 255)
			a := 255
			DEFAULT_GRAPHIC_DATA[p] =
				(r << 6) |
				(g << 4) |
				(b << 2) |
				(a << 0)
		}
	}
}

// Our custom super special graphic class
type Graphic struct {
	Tex gl.Texture
	// The offset of the graphic when drawing
	OffX, OffY int
	// Dimensions
	W, H int
}

type ResourceManager struct {
	tileGraphics map[string]Graphic

	fontMap map[string]*allegro.Font
}

func CreateResourceManager() *ResourceManager {
	var manager ResourceManager

	manager.tileGraphics = make(map[string]Graphic)
	manager.addDefaultGraphic()

	manager.fontMap = make(map[string]*allegro.Font)
	manager.fontMap["builtin"] = allegro.CreateBuiltinFont()

	return &manager
}

func (rm *ResourceManager) addDefaultGraphic() {
	var graphic Graphic
	allegro.RunInThread(func() {
		graphic.Tex = gl.GenTexture()
		graphic.Tex.Bind(gl.TEXTURE_2D)

		// Maybe REPEAT will be more noticable?
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		
		// Why not
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32,
			DEFAULT_GRAPHIC_WIDTH, DEFAULT_GRAPHIC_HEIGHT,
			0, gl.RGBA32, DEFAULT_GRAPHIC_DATA)
	})
	graphic.Width = DEFAULT_GRAPHIC_WIDTH
	graphic.Height = DEFAULT_GRAPHIC_HEIGHT
	
	rm.tileGraphics[DEFAULT_GRAPHIC_NAME] = graphic
}

func (rm *ResourceManager) LoadSG3File(filenameSG3, filename555, prefix string) error {
	file, err := libsg.ReadFile(filename)
	if err {
		return err
	}
	imgs, err := file.Images()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(imgs) + 1)

	errChan := make(chan error)

	go func() {
		err = nil
		err <- errChan
		wg.Done()
	} ()

	for _, img := range imgs {
		dataFile := filenameSG3
		if img.IsExtern() {
			dataFile = filename555
		}
		go func () {
			err := rm.loadImage(img, name, dataFile)
			if err != nil {
				// Non blocking send
				select {
				case errChan <- err:
				default:
				}
			}
			wg.Done()
		} ()
	}
	wg.Wait()
	return err
}

func (rm *ResourceManager) loadImage(img *libsg.Image, name, dataFile string) error {
	bmpName := strings.TrimSuffix(bmp.Filename(), ".bmp")
	name := fmt.Sprintf("%v.%v.%v", prefix, bmpName, img.ID())
	
	imgData := img.LoadData(dataFile)
	tex, err := img.loadTexture(imgData)
	if err != nil {
		return err
	}
	var graphic Graphic
	graphic.Tex = tex
	graphic.W = imgData.Width
	graphic.H = imgData.Height

	rm.tileGraphics[name] = graphic
	return nil
}

func (rm *ResourceManager) loadTexture(imgData *libsg.SgImageData) (gl.Texture, error) {
	// We only support ARGB32 from libsg (convert to RGBA later)
	if imgData.BMask != 0xff ||
		imgData.GMask != 0xff00 ||
		imgData.RMask != 0xff0000 ||
		imgData.AMask != 0xff000000 {
		return gl.Texture(0), fmt.Errorf("Unsupport image format for %q", name)
	}

	// Convert the argb to rgba
	for i := range imgData {
		imgData[i] = (imgData[i] & (~imgData.AMask)) << 2 | (imgData[i] & imgData.AMask) >> 6
	}

	var tex gl.Texture
	allegro.RunInThread(func() {
		tex = gl.GenTexture()
		tex.Bind(gl.TEXTURE_2D)

		// Why not
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		
		// To keep pixely art, GL_NEAREST
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32,
			imgData.Width, imgData.Height,
			0, gl.RGBA32, imgData.Data)
	})
	return tex, nil
}

func (rm *ResourceManager) GetGraphic(name string) (Graphic, error) {
	graphic, ok := rm.tileGraphics[name]
	if !ok {
		return Graphic{}, fmt.Errorf("Could not find graphic %q", name)
	}
	return &bmp, nil
}

// Gets a tile that can be drawn, no matter what. Won't be pretty, but won't crash.
func (rm *ResourceManager) GetDefaultGraphic() *Bitmap {
	sub, ok := rm.GetGraphic(DEFAULT_GRAPHIC_NAME)
	if !ok {
		log.Panicf("Could not find default graphic %q", DEFAULT_TILE_NAME)
	}
	return sub
}

func (rm *ResourceManager) GetGraphicOrDefault(name string) *Bitmap {
	graphic, ok := rm.GetGraphic(name)
	if !ok {
		log.Printf("Could not find graphic named %q. Defaulting to default graphic.", name)
		return rm.GetDefaultTile()
	}
	return graphic
}

func (rm *ResourceManager) GetFont(name string) (*allegro.Font, error) {
	font, ok := rm.fontMap[name]
	if !ok {
		return nil, fmt.Errorf("Could not find font %q", name)
	}
	return font, nil
}