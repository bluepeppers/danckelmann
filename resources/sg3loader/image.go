package sg3loader

import (
	"encoding/binary"
	"io"

	"github.com/go-gl/gl"
)

const (
	IMAGE_SIZE = 72
)

type Image struct {
	Record ImageRecord
	Parent *Bitmap
	Id int
}

type ImageRecord struct {
	Offset uint32
	Length uint32
	UncompressedLength uint32
	_ [4]byte
	InvertOffset int32
	Width int16
	Height int16
	_ [26]byte
	Type uint16
	Flags [4]byte
	BitmapId uint8
	_ [7]byte
	AlphaOffset uint32
	AlphaLength uint32
}

func LoadImage(file io.Reader, id int) (*Image, error) {
	var img Image
	img.Id = id

	err := binary.Read(file, binary.LittleEndian, &img.Record)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (i *Image) LoadImageToTexture(file io.Reader) {
	if i.Record.Width <= 0 || i.Record.Height <= 0 {
		log.Printf("Image has invalid width or height: %vx%v", i.Record.Width, i.Record.Height)
		return
	}

	pixels := i.getPixels()
	gl.CopyTexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, i.Record.Width, i.Record.Height, gl.RGB5_A1, pixels)

	
}