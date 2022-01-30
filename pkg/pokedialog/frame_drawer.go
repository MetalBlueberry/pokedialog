package pokedialog

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type FrameDrawer struct {
	font    *truetype.Font
	frame   image.Image
	palette []color.Color
	img     *image.Paletted
}

//go:embed "Pokemon GB.ttf"
var pokefont []byte

func NewDrawer(dialogFilePath string) (*FrameDrawer, error) {
	dialogFile, err := os.Open(dialogFilePath)
	if err != nil {
		panic(err)
	}
	defer dialogFile.Close()
	dialog, err := png.Decode(dialogFile)
	if err != nil {
		panic(err)
	}

	var palette = []color.Color{
		color.Alpha{},
		color.RGBA{0x00, 0x00, 0x00, 0xff},
		color.RGBA{0x00, 0x00, 0xff, 0xff},
		color.RGBA{0x00, 0xff, 0x00, 0xff},
		color.RGBA{0x00, 0xff, 0xff, 0xff},
		color.RGBA{0xff, 0x00, 0x00, 0xff},
		color.RGBA{0xff, 0x00, 0xff, 0xff},
		color.RGBA{0xff, 0xff, 0x00, 0xff},
		color.RGBA{0xff, 0xff, 0xff, 0xff},
	}

	img := image.NewPaletted(dialog.Bounds(), palette)
	draw.Draw(img, img.Bounds(), dialog, dialog.Bounds().Min, draw.Src)

	f, err := truetype.Parse(pokefont)
	if err != nil {
		return nil, err
	}
	return &FrameDrawer{
		font:    f,
		palette: palette,
		img:     img,
	}, nil
}

func (fd *FrameDrawer) baseImage() *image.Paletted {
	base := image.NewPaletted(fd.img.Rect, fd.palette)
	draw.Draw(base, base.Bounds(), fd.img, fd.img.Bounds().Min, draw.Src)
	return base
}

func (fd *FrameDrawer) Draw(frameRect image.Rectangle, text string) *image.Paletted {
	base := fd.baseImage()
	frameImage := base.SubImage(frameRect).(*image.Paletted)

	fontColor := color.RGBA{0, 0, 0, 255}

	bounds := frameImage.Bounds()
	fontSize := float64(bounds.Dy()*33) / float64(100)

	face := truetype.NewFace(fd.font, &truetype.Options{
		Size: fontSize,
		DPI:  72,
	})

	d := &font.Drawer{
		Dst:  frameImage,
		Src:  image.NewUniform(fontColor),
		Face: face,
	}

	for l, sentences := range SplitLines(face, text, bounds.Dx()) {
		d.Dot = dotForLine(bounds.Min.X, bounds.Min.Y, fontSize, l)
		d.DrawString(sentences)
		d.DrawString(" ")
	}
	return base
}

func dotForLine(x int, y int, size float64, n int) fixed.Point26_6 {
	linespace := 1.5
	enter := size * linespace * float64(n)
	return fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6((y + int(size) + int(enter)) * 64),
	}
}

func SplitLines(face font.Face, text string, width int) []string {
	i := 0
	lines := []string{}
	words := strings.Split(text, " ")
	for j := range words {
		line := strings.Join(words[i:j], " ")
		length := font.MeasureString(face, line)
		if length.Ceil() > width {
			line := strings.Join(words[i:j-1], " ")
			lines = append(lines, line)
			i = j - 1
		}
	}
	if i != len(words) {
		lines = append(lines, strings.Join(words[i:], " "))
	}
	log.Println(lines)
	return lines
}

func FrameAt(lines []string, charsPerSecond float64, duration time.Duration) []string {
	startPosition := int(charsPerSecond * duration.Seconds())
	offset := 0
	var i int
	var line string
	for i, line = range lines {
		if offset+len(line) >= int(startPosition) {
			break
		}
		offset += len(line)
	}
	lastLine := line[:startPosition-offset]
	if i > 0 {
		return []string{
			lines[i-1],
			lastLine,
		}
	}
	return []string{lastLine}
}
