package shapeshifter

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"

	"code.google.com/p/draw2d/draw2d"
)

var reverseBits = [256]byte{
	0, 128, 64, 192, 32, 160, 96, 224,
	16, 144, 80, 208, 48, 176, 112, 240,
	8, 136, 72, 200, 40, 168, 104, 232,
	24, 152, 88, 216, 56, 184, 120, 248,
	4, 132, 68, 196, 36, 164, 100, 228,
	20, 148, 84, 212, 52, 180, 116, 244,
	12, 140, 76, 204, 44, 172, 108, 236,
	28, 156, 92, 220, 60, 188, 124, 252,
	2, 130, 66, 194, 34, 162, 98, 226,
	18, 146, 82, 210, 50, 178, 114, 242,
	10, 138, 74, 202, 42, 170, 106, 234,
	26, 154, 90, 218, 58, 186, 122, 250,
	6, 134, 70, 198, 38, 166, 102, 230,
	22, 150, 86, 214, 54, 182, 118, 246,
	14, 142, 78, 206, 46, 174, 110, 238,
	30, 158, 94, 222, 62, 190, 126, 254,
	1, 129, 65, 193, 33, 161, 97, 225,
	17, 145, 81, 209, 49, 177, 113, 241,
	9, 137, 73, 201, 41, 169, 105, 233,
	25, 153, 89, 217, 57, 185, 121, 249,
	5, 133, 69, 197, 37, 165, 101, 229,
	21, 149, 85, 213, 53, 181, 117, 245,
	13, 141, 77, 205, 45, 173, 109, 237,
	29, 157, 93, 221, 61, 189, 125, 253,
	3, 131, 67, 195, 35, 163, 99, 227,
	19, 147, 83, 211, 51, 179, 115, 243,
	11, 139, 75, 203, 43, 171, 107, 235,
	27, 155, 91, 219, 59, 187, 123, 251,
	7, 135, 71, 199, 39, 167, 103, 231,
	23, 151, 87, 215, 55, 183, 119, 247,
	15, 143, 79, 207, 47, 175, 111, 239,
	31, 159, 95, 223, 63, 191, 127, 255,
}

const (
	namesOffset    = 0x0F00AB
	wavesOffset    = 0x1000AB
	numBanks       = 128
	wavesPerBank   = 8
	samplesPerWave = 512
	nameLength     = 8
)

type Config [128]Bank

type Bank struct {
	Name  string
	Waves [wavesPerBank]Wave
}

type Wave [samplesPerWave]int16

type bitReversingReader struct {
	io.ReadSeeker
}

func (r bitReversingReader) Read(b []byte) (int, error) {
	n, err := r.ReadSeeker.Read(b)
	for i := 0; i < n; i++ {
		b[i] = reverseBits[b[i]]
	}
	return n, err
}

type bitReversingWriter struct {
	io.WriteSeeker
}

func (w bitReversingWriter) Write(b []byte) (int, error) {
	b2 := make([]byte, len(b))
	for i, c := range b {
		b2[i] = reverseBits[c]
	}
	return w.WriteSeeker.Write(b2)
}

func Read(r io.ReadSeeker) (*Config, error) {
	_, err := r.Seek(namesOffset, 0)
	if err != nil {
		return nil, err
	}

	r = bitReversingReader{r}

	var config Config
	name := make([]byte, nameLength)
	for i := range config {
		_, err := io.ReadFull(r, name)
		if err != nil {
			return nil, err
		}
		config[i].Name = string(name)
	}

	_, err = r.Seek(wavesOffset, 0)
	if err != nil {
		return nil, err
	}

	for i := range config {
		for j := 0; j < wavesPerBank; j++ {
			err := binary.Read(r, binary.LittleEndian, &config[i].Waves[j])
			if err != nil {
				return nil, err
			}
		}
	}

	return &config, nil
}

func Write(w io.WriteSeeker, config *Config) error {
	_, err := w.Seek(namesOffset, 0)
	if err != nil {
		return err
	}

	w = bitReversingWriter{w}

	for _, c := range config {
		_, err := w.Write([]byte(c.Name))
		if err != nil {
			return err
		}
	}

	_, err = w.Seek(wavesOffset, 0)
	if err != nil {
		return err
	}

	for _, c := range config {
		for _, wave := range c.Waves {
			err := binary.Write(w, binary.LittleEndian, wave)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

const (
	imageHeight = 256
	imageWidth  = samplesPerWave
)

func DrawWave(w Wave) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	gc := draw2d.NewGraphicContext(img)
	for x, sample := range w {
		y := -int(sample)*imageHeight/int(math.MaxInt16)/2 + imageHeight/2
		if x == 0 {
			gc.MoveTo(float64(x), float64(y))
		}
		gc.LineTo(float64(x), float64(y))
	}
	gc.Stroke()
	return img
}
