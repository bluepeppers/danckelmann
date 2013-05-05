package resources

import (
	"github.com/bluepeppers/allegro"
)

// Infomation about a tile in the atlas
type tileMetadata struct {
	index      int
	x, y, w, h int
	name       string
}

type ResourceManager struct {
	tileAtlas     *allegro.Bitmap
	tileMetadatas []tileMetadata
	tilePositions map[string]int

	fontMap map[string]*allegro.Font
}

func CreateResourceManager(config *ResourceManagerConfig) (*ResourceManager, bool) {
	tileConfs := make([]TileConfig, 0)
	for _, v := range config.TileConfigs {
		tileConfs = append(tileConfs, v)
	}
	tileBmps := make([]*allegro.Bitmap, len(tileConfs))
	tileMetadatas := make([]tileMetadata, len(tileConfs))
	maxHeight := 0
	totalWidth := 0
	for i := 0; i < len(tileBmps); i++ {
		cfg := tileConfs[i]
		bmp := allegro.LoadBitmap(cfg.Filename)
		tileBmps[i] = bmp
		var x, y, w, h int
		x = cfg.X
		y = cfg.Y
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

		if h > maxHeight {
			maxHeight = h
		}
		totalWidth += w

		tileMetadatas[i] = tileMetadata{i, x, y, w, h, cfg.Name}
	}

	atlas := allegro.NewBitmap(totalWidth, maxHeight)

	atlas.SetTargetBitmap()
	allegro.HoldBitmapDrawing(true)

	currentPos := 0
	for i := 0; i < len(tileBmps); i++ {
		bmp := tileBmps[i]
		metadata := tileMetadatas[i]
		bmp.DrawRegion(float32(metadata.x), float32(metadata.y),
			float32(metadata.w), float32(metadata.h),
			float32(currentPos), float32(0), 0)
		metadata.x = currentPos
		metadata.y = 0
		currentPos += metadata.w
	}

	allegro.HoldBitmapDrawing(false)

	var manager ResourceManager
	manager.tileAtlas = atlas
	manager.tileMetadatas = tileMetadatas
	manager.tilePositions = make(map[string]int, len(tileMetadatas))
	for _, v := range tileMetadatas {
		manager.tilePositions[v.name] = v.index
	}

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

	return &manager, true
}

func (rm *ResourceManager) GetTile(name string) (*allegro.Bitmap, bool) {
	pos, ok := rm.tilePositions[name]
	if !ok {
		return nil, false
	}
	metadata := rm.tileMetadatas[pos]
	sub := rm.tileAtlas.CreateSubBitmap(metadata.x, metadata.y,
		metadata.w, metadata.h)
	return sub, sub != nil
}

func (rm *ResourceManager) GetFont(name string) (*allegro.Font, bool) {
	font, ok := rm.fontMap[name]
	return font, font != nil
}
