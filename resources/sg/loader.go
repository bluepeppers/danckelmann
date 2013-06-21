package sg

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	libsg "github.com/bluepeppers/libsg/golibsg"
	"github.com/go-gl/gl"

	"github.com/bluepeppers/danckelmann/resources"
)

func init() {
	err := resources.RegisterLoader(SGLoader{})
	if err != nil {
		log.Printf("Could not register SG loader: %v", err)
	}
}

type SGLoader struct{}

func (l SGLoader) Extensions() []string {
	return []string{"sg3", "sg2"}
}

func (l SGLoader) LoadFile(filenameSG string) (map[string]resources.Graphic, error) {
	// Can I get a monad in here?
	if !isFile(filenameSG) {
		return nil, fmt.Errorf("File %q does not exist")
	}

	sgFile, err := libsg.ReadFile(filenameSG)
	if err != nil {
		return nil, err
	}

	// Check if all the images are internal to see if we need to find the 555
	// file.
	imgs, err := sgFile.Images()
	if err != nil {
		return nil, err
	}
	allInternal := true
	for _, image := range imgs {
		allInternal = allInternal && !image.IsExtern()
	}

	filename555 := ""
	if !allInternal {
		filename555, err = find555File(filenameSG)
		if err != nil {
			return nil, err
		}
	}

	graphics := make(map[string]resources.Graphic, len(imgs))
	for _, img := range imgs {
		parent := img.Parent()
		name := fmt.Sprintf("%v.%v", parent.Filename(), img.ID())

		resourceFilename := filenameSG
		if img.IsExtern() {
			resourceFilename = filename555
		}
		graphic, err := loadSGImage(img, resourceFilename)
		if err != nil {
			return nil, err
		}
		graphics[name] = graphic
	}

	return graphics, nil
}

func loadSGImage(img *libsg.SgImage, dataFile string) (resources.Graphic, error) {
	var graphic resources.Graphic
	imgData, err := img.LoadData(dataFile)
	if err != nil {
		return graphic, err
	}
	tex, err := loadTexture(imgData)
	if err != nil {
		return graphic, err
	}
	graphic.Tex = tex
	graphic.Width = imgData.Width
	graphic.Height = imgData.Height

	return graphic, nil
}

func loadTexture(imgData *libsg.SgImageData) (gl.Texture, error) {
	// We only support ARGB32 from libsg (convert to RGBA later)
	if imgData.BMask != 0xff ||
		imgData.GMask != 0xff00 ||
		imgData.RMask != 0xff0000 ||
		imgData.AMask != 0xff000000 {
		return gl.Texture(0), fmt.Errorf("Unsupport image format: %v", imgData)
	}

	// Convert the argb to rgba
	data := imgData.Data
	for i := 0; i < imgData.Width*imgData.Height; i++ {
		p := i * 4
		a := data[p]
		b := data[p+1]
		c := data[p+2]
		d := data[p+3]
		data[p] = d
		data[p+1] = a
		data[p+2] = b
		data[p+3] = c
	}

	var tex gl.Texture
	resources.RunInThread(func() {
		tex = gl.GenTexture()
		tex.Bind(gl.TEXTURE_2D)

		// Why not
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		// To keep pixely art, GL_NEAREST
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
			imgData.Width, imgData.Height,
			0, gl.RGBA, gl.UNSIGNED_INT_8_8_8_8, imgData.Data)
	})
	return tex, nil
}

func isFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	fi, err := file.Stat()
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func find555File(fileSG3 string) (string, error) {
	// I.e. foo.sg3 -> foo.555
	sameDir555 := filepath.Join(filepath.Dir(fileSG3), filepath.Base(fileSG3)+".555")
	if isFile(sameDir555) {
		return sameDir555, nil
	}
	return "", fmt.Errorf("Cannot find 555 file for %q", fileSG3)
}
