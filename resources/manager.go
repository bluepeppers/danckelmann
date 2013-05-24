package resources

import (
	"log"
	_ "sync"
	"path"
	"path/filepath"

	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/bluepeppers/allegro"
)

const (
	DEFAULT_TILE_NAME   = "____DEFAULT____"
	DEFAULT_TILE_WIDTH  = 128
	DEFAULT_TILE_HEIGHT = 128
)

// Our custom super special bitmap class
type Bitmap struct {
	Tex gl.Texture
	// The offset of the bitmap when drawing
	OffX, OffY int
	// Dimensions
	W, H int
}

// Infomation about a tile in the atlas
type tileMetadata struct {
	x, y, w, h int
	offx, offy int
	name       string
}

type ResourceManager struct {
	tileMetadatas map[string]tileMetadata
	tileBmps      map[string]Bitmap

	fontMap map[string]*allegro.Font
}

func CreateResourceManager(config *ResourceManagerConfig) *ResourceManager {
/*	defaultTileConfig := TileConfig{Name: DEFAULT_TILE_NAME, Filename: DEFAULT_TILE_NAME}
	config.TileConfigs = append(config.TileConfigs, defaultTileConfig)

	tileConfs := make([]TileConfig, 0)
	for _, v := range config.TileConfigs {
		tileConfs = append(tileConfs, v)
	}
	
	var manager ResourceManager

	nBmps := len(tileConfs)
	log.Printf("Loading %v bitmaps", nBmps)

	loadedBmps := make([]*allegro.Bitmap, nBmps)
	manager.tileMetadatas = make([]tileMetadata, nBmps)
	var wg sync.WaitGroup
	wg.Add(nBmps)
	for j := 0; j < nBmps; j++ {
		go func (i int) {
			cfg := tileConfs[i]

			// Load the bitmap
			var bmp *allegro.Bitmap
			if cfg.Name == DEFAULT_TILE_NAME {
				bmp = allegro.NewBitmap(DEFAULT_TILE_WIDTH, DEFAULT_TILE_HEIGHT)
			} else {
				bmp = allegro.LoadBitmap(cfg.Filename)
			}
			if bmp == nil {
				log.Fatalf("Failed to load bmp %v", cfg)
			}
			
			loadedBmps[i] = bmp
			manager.tileMetadatas[i] = generateMetadata(bmp, cfg)

			wg.Done()
		} (j)
	}
	wg.Wait()
	log.Printf("Loaded bitmaps")
	
	manager.tileBmps = make([]*Bitmap, nBmps)
	for i, bmp := range loadedBmps {
		metadata := manager.tileMetadatas[i]
		fBmp := &Bitmap{gl.Texture(bmp.GetGLTexture()), bmp,
			metadata.offx, metadata.offy, metadata.w, metadata.h}
		manager.tileBmps[i] = fBmp
	}

	manager.tilePos = make(map[string]int, nBmps)
	for i, metadata := range manager.tileMetadatas {
		manager.tilePos[metadata.name] = i
	}*/

	var manager ResourceManager
	manager.tileMetadatas = make(map[string]tileMetadata)
	manager.tileBmps = make(map[string]Bitmap)
	
	// Load the fonts
	manager.fontMap = make(map[string]*allegro.Font)
	for _, v := range config.FontConfigs {
		var font *allegro.Font
		if v.Filename == "builtin" {
			font = allegro.CreateBuiltinFont()
		} else {
			font = allegro.LoadFont(v.Filename, v.Size, 0)
		}
		manager.fontMap[v.Name] = font
	}

	return &manager
}

func (rm *ResourceManager) GetTile(name string) (*Bitmap, bool) {
	bmp, ok := rm.tileBmps[name]
	if !ok {
		return rm.loadTile(name)
	}
	return &bmp, true
}

// Gets a tile that can be drawn, no matter what. Won't be pretty, but won't crash.
func (rm *ResourceManager) GetDefaultTile() *Bitmap {
	sub, ok := rm.GetTile(DEFAULT_TILE_NAME)
	if !ok {
		log.Panicf("Could not find default key in atlas: %v", DEFAULT_TILE_NAME)
	}
	return sub
}

func (rm *ResourceManager) GetTileOrDefault(name string) *Bitmap {
	tile, ok := rm.GetTile(name)
	if !ok {
		log.Printf("Could not find tile named %q. Defaulting to default tile.", name)
		return rm.GetDefaultTile()
	}
	return tile
}

func (rm *ResourceManager) GetFont(name string) (*allegro.Font, bool) {
	font, ok := rm.fontMap[name]
	return font, ok && font != nil
}

func (rm *ResourceManager) loadTile(name string) (*Bitmap, bool) {
	fname, _ := filepath.Abs(path.Join("resources", name))
	var tex gl.Texture
	allegro.RunInThread(func() {
		tex = gl.GenTexture()
		tex.Bind(gl.TEXTURE_2D)
		glfw.LoadTexture2D(fname, 0)
	})
	bmp := Bitmap{tex, 0, 0, DEFAULT_TILE_WIDTH, DEFAULT_TILE_HEIGHT}
	rm.tileBmps[name] = bmp
	return &bmp, true
}

func generateMetadata(bmp *allegro.Bitmap, cfg TileConfig) tileMetadata {
	// Load the metadata, and then sanitize it
	x := cfg.X
	y := cfg.Y
	w := cfg.W
	h := cfg.H
	ox := cfg.OffX
	oy := cfg.OffY
	bmpw, bmph := bmp.GetDimensions()
	if bmpw < x {
		x = 0
		w = bmpw
	} else if bmpw < x+cfg.W {
		w = bmpw - x
	} else {
		w = cfg.W
	}
	if bmph < y {
		y = 0
		h = bmph
	} else if bmph < y+cfg.H {
		h = bmph - y
	} else {
		h = cfg.H
	}

	if w == 0 {
		w = bmpw - x
	}
	if h == 0 {
		h = bmph - y
	}

	if ox > w {
		ox = 0
	}
	if oy > h {
		oy = 0
	}

	return tileMetadata{x, y, w, h, ox, oy, cfg.Name}
}