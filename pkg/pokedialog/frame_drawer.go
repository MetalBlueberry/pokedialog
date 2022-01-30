package pokedialog

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type FrameDrawer struct {
	font           *truetype.Font
	palette        []color.Color
	img            *image.Paletted
	charsPerSecond float64
	frameRect      image.Rectangle
}

//go:embed "Pokemon GB.ttf"
var pokefont []byte

func NewDrawer(dialogFilePath string, charsPerSecond float64, frameRect image.Rectangle) (*FrameDrawer, error) {
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
		font:           f,
		palette:        palette,
		img:            img,
		charsPerSecond: charsPerSecond,
		frameRect:      frameRect,
	}, nil
}

func (fd *FrameDrawer) baseImage() *image.Paletted {
	base := image.NewPaletted(fd.img.Rect, fd.palette)
	draw.Draw(base, base.Bounds(), fd.img, fd.img.Bounds().Min, draw.Src)
	return base
}

func (fd *FrameDrawer) Gif(text string, frameCount int) *gif.GIF {
	frames := fd.DrawFrames(text, frameCount)

	return &gif.GIF{
		Image: frames,
	}

}

func (fd *FrameDrawer) DrawFrames(text string, frameCount int) []*image.Paletted {
	maxFrames := len(text)
	if frameCount > maxFrames {
		frameCount = maxFrames
	}

	frames := make([]*image.Paletted, frameCount)

	for i := 1; i <= frameCount; i++ {
		frames[i-1] = fd.DrawFrameAt(text, maxFrames*i/frameCount)
	}
	return frames
}

func (fd *FrameDrawer) DrawFrameAt(text string, frame int) *image.Paletted {
	base := fd.baseImage()
	frameImage := base.SubImage(fd.frameRect).(*image.Paletted)

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

	lines := SplitLines(face, text, bounds.Dx())
	linesAt := LinesAt(lines, frame)
	for l, sentences := range linesAt {
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
	return lines
}

func LinesAt(lines []string, position int) []string {
	offset := 0
	var i int
	var line string
	for i, line = range lines {
		if offset+len(line) >= int(position) {
			break
		}
		offset += len(line)
	}
	lastLine := line[:position-offset]
	if i > 0 {
		return []string{
			lines[i-1],
			lastLine,
		}
	}
	return []string{lastLine}
}
