package sg3loader

import (
	"encoding/binary"
	"log"
	"os"
)

const (
	HEADER_SIZE = 680
)

type File struct {
	Bitmaps  []*Bitmap
	Images   []*Image
	Filename string

	Header Header
}

type Header struct {
	Filesize         uint32
	Version          uint32
	_                uint32
	MaxImageRecords  int32
	NumImageRecords  int32
	NumBitmapRecords int32
	_                int32
	TotalFilesize    uint32
	Filesize555      uint32
	FilesizeExternal uint32
}

func LoadFile(filename string) *File {
	var sgfile File
	sgfile.Filename = filename

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Could not open file %q: %v", filename, err)
		return nil
	}

	err = binary.Read(file, binary.LittleEndian, &sgfile.Header)
	if err != nil {
		log.Printf("Could not read header from %q: %v", filename, err)
		return nil
	}
	log.Printf("%v", sgfile)

	sgfile.CheckHeader()

	if !sgfile.LoadBitmaps(file) {
		return nil
	}
	if !sgfile.LoadImages(file) {
		return nil
	}

	if len(sgfile.Bitmaps) > 1 && len(sgfile.Images) == len(sgfile.Bitmaps[0].Images) {
		log.Printf("SG file %q has %v bitmaps but only the first is in use",
			sgfile.Filename, len(sgfile.Images))
		sgfile.Bitmaps = []*Bitmap{sgfile.Bitmaps[0]}
	}
	
	return &sgfile
}

// Checks things such as the version numbers and issues warnings where needed
func (f *File) CheckHeader() {
	h := f.Header
	if h.Version != 0xd5 && h.Version != 0xd6 {
		log.Printf("SG file's version is unsupported: %q.version = %v", f.Filename, h.Version)
	}
	if h.Filesize555+h.FilesizeExternal != h.TotalFilesize {
		log.Printf("SG files' filesize's are incorrect: %q, %v + %v != %v",
			f.Filename, h.Filesize555, h.FilesizeExternal, h.TotalFilesize)
	}
}

func (f *File) MaxBitmapRecords() int {
	if f.Header.Version == 0xd3 {
		return 100
	} else {
		return 200
	}
}

func (f *File) LoadBitmaps(file *os.File) bool {
	f.Bitmaps = make([]*Bitmap, f.Header.NumBitmapRecords)
	for i := 0; int32(i) < f.Header.NumBitmapRecords; i++ {
		
		file.Seek(int64(HEADER_SIZE + BITMAP_SIZE * i), 0)
		bmp, err := LoadBitmap(file, i)
		f.Bitmaps[i] = bmp
		if err != nil {
			log.Printf("Could not read bitmap %v from %q: %v", i, f.Filename, err)
			return false
		}
	}
	
	return true
}

func (f *File) LoadImages(file *os.File) bool {
	f.Images = make([]*Image, f.Header.NumImageRecords + 1)
	img, err := LoadImage(file, 0)
	f.Images[0] = img
	if err != nil {
		log.Printf("Could not load image 0 from %q: %v", f.Filename, err)
		return false
	}
	for i := 1; int32(i) < f.Header.NumImageRecords + 1; i++ {
		//file.Seek(int64(HEADER_SIZE + BITMAP_SIZE * f.MaxBitmapRecords() + IMAGE_SIZE * i), 0)
		img, err = LoadImage(file, 1)
		f.Images[i] = img
		if err != nil {
			log.Printf("Could not load image %v from %q: %v", i, f.Filename, err)
			return false
		}
		bmpId := f.Images[i].Record.BitmapId
		if bmpId < 0 || bmpId > uint8(len(f.Bitmaps)) {
			log.Printf("Image %v from %q has invalid parent: %v", i, f.Filename, bmpId)
		} else {
			f.Bitmaps[bmpId].AddImage(f.Images[i])
		}
	}
	
	return true
}