package resources

import (
	//	"log"
	"fmt"
	"path"
	"sync"
)

type GraphicLoader interface {
	// The extensions that the loader can load
	Extensions() []string

	// Loads the given file, returns a mapping from resource name to graphic
	LoadFile(string) (map[string]Graphic, error)
}

var (
	// Map between filename extension and the graphics loaders. Loaders with
	// multiple extensions have multiple mappings
	registeredLoaders = make(map[string]GraphicLoader)

	// And locking
	loadersMutex sync.RWMutex
)

// Registers the loader. Returns an error if any of the loader's extensions
// have already been registered.
func RegisterLoader(loader GraphicLoader) error {
	extensions := loader.Extensions()

	loadersMutex.RLock()
	// Check that all the extensions are unique
	for _, extension := range extensions {
		if _, ok := registeredLoaders[extension]; !ok {
			return fmt.Errorf(
				"Extension %q already has a registered loader",
				extension)
		}
	}
	loadersMutex.RUnlock()

	loadersMutex.Lock()
	for _, extension := range extensions {
		registeredLoaders[extension] = loader
	}
	loadersMutex.Unlock()

	return nil
}

// Returns the loader for the given extension.
func GetLoader(extension string) (GraphicLoader, bool) {
	loadersMutex.RLock()
	defer loadersMutex.RUnlock()
	loader, ok := registeredLoaders[extension]
	return loader, ok
}

// Loads the resources from the file into the resource manager. Uses the
// GraphicLoader associated with the filename's extension via RegisterLoader.
func (rm *ResourceManager) LoadFile(filename string) error {
	return rm.LoadFilePrefix(filename, "")
}

// Loads the file in the same manner as LoadFile, but prepends the prefix to
// the name of all graphics loaded.
func (rm *ResourceManager) LoadFilePrefix(filename, prefix string) error {
	ext := path.Ext(filename)
	loader, ok := GetLoader(ext)
	if !ok {
		return fmt.Errorf("No loader associated with the extension %q", ext)
	}

	graphics, err := loader.LoadFile(filename)
	if err != nil {
		return err
	}

	for name, graphic := range graphics {
		rm.AddGraphic(prefix+name, graphic)
	}
	return nil
}
