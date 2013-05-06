package display

import (
	"github.com/bluepeppers/allegro"
)

// Configuration for the display engine. To be created by the game engine and
// passed back via GetDisplayConfig
type DisplayConfig struct {
	// The dimensions of the map in tiles
	MapW, MapH int
	// The size of a tile
	TileW, TileH int
	// The color of the "void" (i.e. where no tiles are drawn)
	BGColor allegro.Color
}
