package sg3loader

import (
	"encoding/binary"
	"io"
	"errors"

	"github.com/go-gl/gl"
)

const (
	IMAGE_SIZE = 72

	ISOMETRIC_TILE_WIDTH = 58
	ISOMETRIC_TILE_HEIGHT = 30
	ISOMETRIC_TILE_BYTES = 1800
	ISOMETRIC_LARGE_TILE_WIDTH = 78
	ISOMETRIC_LARGE_TILE_HEIGHT = 40
	ISOMETRIC_LARGE_TILE_BYTES = 3200
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

func (i *Image) LoadImageToTexture(file io.ReadSeeker) bool {
	if i.Record.Width <= 0 || i.Record.Height <= 0 {
		log.Printf("Image has invalid width or height: %vx%v", i.Record.Width, i.Record.Height)
		return false
	}
	buffer, err := i.GetImageBuffer(file)
	if err != nil {
		log.Printf("Could not open image %v buffer: %v", i.Id, err)
		return false
	}
	
	var pixels []float32
	switch i.Record.Type {
	case 0, 1, 10, 12, 13:
		pixels, err = i.loadPlainImage(buffer)
	case 30
		pixels, err = i.loadIsometricImage(buffer)
	case 256, 257, 276:
		pixels, err = i.loadSpriteImage(buffer)
	default:
		log.Printf("Image %v has unknown type: %v", i.Id, i.Record.Type)
		return false
	}

	if err != nil {
		log.Printf("Could not load image %v: %v", i.Id, err)
	}
	
	if i.Record.AlphaLength != 0 {
		err = i.LoadAlpha(buffer, pixels)
		if err != nil {
			log.Printf("Could not load image %v's alpha: %v", i.Id, err)
		}
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, i.Record.Width, i.Record.Type, 0, gl.RGBA, gl.FLOAT, pixels)
}

func (i *Image) GetImageBuffer(file io.ReadSeeker) ([]byte, error) {
	seekPos := i.Record.Offset
	if i.Record.Flags[0] == 1 {
		seekPos--
	}
	_, err := file.Seek(seekPos, 0)
	if err != nil {
		return []byte{}, err
	}

	dataLength := i.Record.Length + i.Record.AlphaLength
	buffer := make([]byte, dataLength)
	nRead, err := file.Read(buffer)
	if err != nil {
		return []byte{}, err
	}
	if nRead != dataLength {
		if nRead + 4 == dataLength {
			buffer[n] = buffer[n+1] = buffer[n+2] = buffer[n+3] = 0
		}  else {
			return []byte{}, errors.New("Could not read all image data into buffer")
		}
	}
	return buffer, nil
}

func (i *Image) loadPlainImage(buffer []byte) ([]float32, error) {
	if (i.Record.Height * i.Record.Width * 2 != len(buffer)) {
		return []float32{}, errors.New("Image data was of invalid length")
	}
	pixels := make([]float32, len(buffer) * 2) // 4 floats for every pixel
	max := 0x1f
	for i, j := 0, 0; i < len(buffer); i += 2, j += 4 {
		val := uint16(buffer[i]) + uint16(buffer[i+1] << 8)

		write555Pixel(pixels[j:], val)
	}
	return pixels, nil
}

func write555Pixel(pixels []float32, val uint16) {
	max := 0x1f

	pixels[0] = float32((val & (max << 10)) >> 10) / float32(max)
	pixels[1] = float32((val & (max << 5)) >> 5) / float32(max)
	pixels[2] = float32(val & max) / float32(max)
	pixels[3] = 0.0
}

func (i *Image) loadIsometricImage(buffer []byte) ([]float32, error) {
	width := i.Record.Width
	height := (width + 2) / 2
	heightOffset := i.Record.Height - height
	size := i.Record.Flags[3]

	if size == 0 {
		if height % ISOMETRIC_TILE_HEIGHT == 0 {
			size = height / ISOMETRIC_TILE_HEIGHT
		} else if height % ISOMETRIC_LARGE_TILE_HEIGHT == 0 {
			size = height / ISOMETRIC_LARGE_TILE_HEIGHT
		}
	}

	var tileBytes, tileHeight, tileWidth int
	if ISOMETRIC_TILE_HEIGHT * size == height {
		tileBytes = ISOMETRIC_TILE_BYTES
		tileHeight = ISOMETRIC_TILE_HEIGHT
		tileWidth = ISOMETRIC_TILE_WIDTH
	} else ISOMETRIC_LARGE_TYPE_BITES * size == height {
		tileBytes = ISOMETRIC_LARGE_TILE_BYTES
		tileHeight = ISOMETRIC_LARGE_TILE_HEIGHT
		tileWidth = ISOMETRIC_LARGE_TILE_WIDTH
	}
}