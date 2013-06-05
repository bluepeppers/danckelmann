package resources

import (
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/bluepeppers/allegro"
)

const (
	// Default value for the size field of font resources
	DEFAULT_SIZE = "12"
)

var (
	positionRegexp   = regexp.MustCompile(`\d+,\d+`)
	dimensionsRegexp = regexp.MustCompile(`\d+,\d+`)
	sizeRegexp       = regexp.MustCompile(`\d+`)
	offsetRegexp     = regexp.MustCompile(`\d+,\d+`)
)

// Information on how to load a tile resouce.
type TileConfig struct {
	Name     string
	Filename string

	// Set any of these to 0 to use the default values
	X, Y, W, H int
	OffX, OffY int
}

// Information on how to load a font resource.
type FontConfig struct {
	Name string
	// If filename is `builtin`, will not check file exists
	Filename string
	Size     int
}

type ResourceManagerConfig struct {
	TileConfigs []TileConfig
	FontConfigs []FontConfig
}

func LoadResourceManagerConfig(directory string, prefix string) (*ResourceManagerConfig, bool) {
	configFilename := path.Join(directory, "resources.ini")
	_, err := os.Open(configFilename)
	if os.IsNotExist(err) {
		return nil, false
	}

	var rmConfig ResourceManagerConfig
	rawConfig := allegro.LoadConfig(configFilename)
	for sectionName := range rawConfig.IterSections() {
		resourceType, ok := rawConfig.Get(sectionName, "type")
		if !ok {
			log.Printf("Section %v of %v resource file has no type field",
				sectionName, directory)
			log.Printf("Skipping section")
			continue
		}

		switch resourceType {
		case "tile":
			tileConfig, ok := loadTileConfig(rawConfig, sectionName, prefix, directory)
			if ok {
				rmConfig.TileConfigs = append(rmConfig.TileConfigs, tileConfig)
			}
		case "font":
			fontConfig, ok := loadFontConfig(rawConfig, sectionName, prefix, directory)
			if ok {
				rmConfig.FontConfigs = append(rmConfig.FontConfigs, fontConfig)
			}
		case "subdirectory":
			fname, ok := rawConfig.Get(sectionName, "filename")
			if !ok {
				log.Printf("Subdir %v had no filename field", sectionName)
				log.Printf("Skipping directory")
				continue
			}
			dirname := path.Join(directory, fname)
			stat, err := os.Stat(dirname)

			if err != nil || !stat.Mode().IsDir() {
				log.Printf("Subdir %v's filename field is not a directory: %v",
					sectionName, dirname)
				log.Printf("Skipping directory")
				continue
			}

			var subPrefix string
			if prefix != "" {
				subPrefix = prefix + "." + sectionName
			} else {
				subPrefix = sectionName
			}
			subConfig, ok := LoadResourceManagerConfig(dirname, subPrefix)
			if ok {
				rmConfig.Merge(subConfig)
			}
		default:
			log.Printf("Resource %q was of type %q that was not recognised",
				sectionName, resourceType)
			log.Printf("Skipping resource")
		}
	}

	return &rmConfig, true
}

func loadTileConfig(rawConfig *allegro.Config, name, prefix, directory string) (TileConfig, bool) {
	var tileConf TileConfig

	if prefix != "" {
		tileConf.Name = prefix + "." + name
	} else {
		tileConf.Name = name
	}

	fname, ok := rawConfig.Get(name, "filename")
	if !ok {
		log.Printf("Resource %v has no filename field")
		log.Printf("Skipping resource")
		return tileConf, false
	}
	filename := path.Join(directory, fname)
	_, err := os.Open(filename)
	if os.IsNotExist(err) {
		log.Printf("Resource %v's assigned file did not exist: %v",
			name, filename)
		log.Printf("Skipping resource")
		return tileConf, false
	}
	tileConf.Filename = filename

	position, ok := rawConfig.Get(name, "position")
	if !ok {
		position = "0,0"
	} else if !positionRegexp.MatchString(position) {
		log.Printf("Resource %v's position field was not valid: %v",
			name, position)
		log.Printf("Using default of 0,0")
		position = "0,0"
	}
	// Already checked format with regexp, so don't need to check error codes
	// and the like
	split := strings.Split(position, ",")
	x, _ := strconv.Atoi(split[0])
	y, _ := strconv.Atoi(split[1])
	tileConf.X = x
	tileConf.Y = y

	dimensions, ok := rawConfig.Get(name, "dimensions")
	if !ok {
		dimensions = "0,0"
	} else if !dimensionsRegexp.MatchString(dimensions) {
		log.Printf("Resource %v's dimensions filed was not valid: %v",
			name, dimensions)
		log.Printf("Using default of 0,0")
		dimensions = "0,0"
	}
	split = strings.Split(dimensions, ",")
	w, _ := strconv.Atoi(split[0])
	h, _ := strconv.Atoi(split[1])
	tileConf.W = w
	tileConf.H = h

	offset, ok := rawConfig.Get(name, "offset")
	if !ok {
		offset = "0,0"
	} else if !offsetRegexp.MatchString(dimensions) {
		log.Printf("Resource %v's offset filed was not valid: %v",
			name, offset)
		log.Printf("Using default of 0,0")
		offset = "0,0"
	}
	split = strings.Split(offset, ",")
	ox, _ := strconv.Atoi(split[0])
	oy, _ := strconv.Atoi(split[1])
	tileConf.OffX = ox
	tileConf.OffY = oy

	return tileConf, true
}

func loadFontConfig(rawConfig *allegro.Config, name, prefix, directory string) (FontConfig, bool) {
	var fontConf FontConfig

	if prefix != "" {
		fontConf.Name = prefix + "." + name
	} else {
		fontConf.Name = name
	}

	fname, ok := rawConfig.Get(name, "filename")
	if !ok {
		log.Printf("Resource %v has no filename field")
		log.Printf("Skipping resource")
		return fontConf, false
	}
	var filename string
	if fname != "builtin" {
		filename = path.Join(directory, fname)
		_, err := os.Open(filename)
		if os.IsNotExist(err) {
			log.Printf("Resource %v's assigned file did not exist: %v",
				name, filename)
			log.Printf("Skipping resource")
			return fontConf, false
		}
	} else {
		filename = fname
	}
	fontConf.Filename = filename

	strSize, ok := rawConfig.Get(name, "size")
	if !ok {
		strSize = DEFAULT_SIZE
	}
	if !sizeRegexp.MatchString(strSize) {
		log.Printf("Resource %v's size field is invalid: %v",
			name, strSize)
		log.Printf("Defaulting to %v", DEFAULT_SIZE)
		strSize = DEFAULT_SIZE
	}
	size, _ := strconv.Atoi(strSize)
	fontConf.Size = size

	return fontConf, true
}

func (rm *ResourceManagerConfig) Merge(sub *ResourceManagerConfig) {
	rm.TileConfigs = append(rm.TileConfigs, sub.TileConfigs...)
	rm.FontConfigs = append(rm.FontConfigs, sub.FontConfigs...)
}
