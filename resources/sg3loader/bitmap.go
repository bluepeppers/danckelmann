package sg3loader

import (
	"encoding/binary"
	"io"
	//"log"
	//"errors"
)

const (
	BITMAP_SIZE = 200
)

type Bitmap struct {
	Images []*Image
	Record BitmapRecord
	Id     int
}

type BitmapRecord struct {
	Filename   [65]byte
	Comment    [51]byte
	Width      uint32
	Height     uint32
	NumImages  uint32
	StartIndex uint32
	EndIndex   uint32
}

func LoadBitmap(file io.Reader, id int) (*Bitmap, error) {
	var bmp Bitmap

	err := binary.Read(file, binary.LittleEndian, &bmp.Record)
	if err != nil {
		return nil, err
	}

	bmp.Id = id
	return &bmp, nil
}

func (b *Bitmap) AddImage(img *Image) {
	img.Parent = b
	b.Images = append(b.Images, img)
}
