package pokedialog

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type FrameDrawer struct {
	Log       *log.Logger
	font      *truetype.Font
	palette   []color.Color
	img       *image.Paletted
	frameRect image.Rectangle
}

//go:embed "Pokemon GB.ttf"
var Pokefont []byte

//go:embed "dialog.png"
var DefaultDialog []byte
var DefaultDialogFrameRect = image.Rect(70, 70, 1425, 395)

func NewDrawer() (*FrameDrawer, error) {
	return NewDrawerWithDialog(bytes.NewReader(DefaultDialog), DefaultDialogFrameRect)
}

func NewDrawerWithDialog(dialogFile io.Reader, frameRect image.Rectangle) (*FrameDrawer, error) {
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

	f, err := truetype.Parse(Pokefont)
	if err != nil {
		return nil, err
	}
	return &FrameDrawer{
		font:      f,
		palette:   palette,
		img:       img,
		frameRect: frameRect,
		Log:       log.Default(),
	}, nil
}

func (fd *FrameDrawer) baseImage() *image.Paletted {
	pix := make([]uint8, len(fd.img.Pix))
	copy(pix, fd.img.Pix)
	return &image.Paletted{
		Pix:     pix,
		Stride:  fd.img.Stride,
		Rect:    fd.img.Rect,
		Palette: fd.img.Palette,
	}
}

type GifConfig struct {
	FrameCount         int
	Duration           time.Duration
	EndParagraphFrames int
}

func (fd *FrameDrawer) Gif(text string, conf GifConfig) (*gif.GIF, error) {
	maxFrames := len(text)

	frameCount := conf.FrameCount
	if conf.FrameCount == 0 {
		frameCount = maxFrames
	}
	if conf.FrameCount > maxFrames {
		fd.Log.Printf("WARNING!! too many frames, maximum is %d", maxFrames)
		frameCount = maxFrames
	}

	duration := time.Millisecond * 250 * time.Duration(len(text))
	if conf.Duration != 0 {
		duration = conf.Duration
	}

	paragraphs := splitParagraphs(text)
	frames := []*image.Paletted{}
	delays := []int{}

	endParagraphFrames := 5
	if conf.EndParagraphFrames != 0 {
		endParagraphFrames = conf.EndParagraphFrames
	}
	time := duration / time.Duration(frameCount+len(paragraphs)*endParagraphFrames)

	// fd.Log.Println(frameCount)
	for _, paragraph := range paragraphs {
		paragraphFrameCount := int(math.Ceil(float64(frameCount*len(paragraph)) / float64(len(text))))
		paragraphFrames := fd.DrawFrames(paragraph, paragraphFrameCount)
		frames = append(frames, paragraphFrames...)
		// fd.Log.Println(paragraph)
		// fd.Log.Println(len(paragraphFrames))
		delays = append(delays, makeWithValue(len(paragraphFrames), int(time.Seconds()*100))...)

		for i := 0; i < endParagraphFrames; i++ {
			frames = append(frames, copyImage(frames[len(frames)-1]))
			delays = append(delays, delays[len(delays)-1])
		}
	}

	// fd.Log.Println(len(delays))
	// fd.Log.Println(delays)

	opt := GifFrameOptimizer()
	for i := range frames {
		opt(frames[i])
	}

	return &gif.GIF{
		Image: frames,
		Delay: delays,
	}, nil

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
		line := strings.Join(words[i:j+1], " ")
		length := font.MeasureString(face, line)
		if length.Ceil() > width {
			line := strings.Join(words[i:j], " ")
			lines = append(lines, line)
			i = j
		}
	}
	if i != len(words) {
		lines = append(lines, strings.Join(words[i:], " "))
	}
	return lines
}

func LinesAt(lines []string, position int) []string {

	offset := 0
	var (
		i         int
		line      string
		completed bool
	)
	for i, line = range lines {
		if offset+len(line) >= int(position) {
			completed = true
			break
		}
		offset += len(line)
	}

	result := []string{}

	if i > 0 {
		result = append(result, lines[i-1])
	}
	if completed {
		result = append(result, line[:position-offset])
	} else {
		result = append(result, line)
	}

	return result
}

func splitParagraphs(text string) []string {
	return strings.Split(text, "\n")
}

func makeWithValue(n int, value int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = value
	}
	return s
}

func constantDelay(n int, duration time.Duration) []int {
	d := duration.Seconds() * 100
	ints := make([]int, n)
	for i := range ints {
		ints[i] = int(d)
	}
	return ints
}

func sum(s []int) (r int) {
	for i := range s {
		r += s[i]
	}
	return
}

// GifFrameOptimizer turns repeated pixels to transparent to the final gif size is minimal.
func GifFrameOptimizer() func(img *image.Paletted) {
	var currentImage *image.Paletted

	return func(img *image.Paletted) {
		if currentImage == nil {
			currentImage = &image.Paletted{}
			currentImage.Palette = img.Palette
			currentImage.Rect = img.Rect
			currentImage.Stride = img.Stride
			currentImage.Pix = make([]uint8, len(img.Pix))
			copy(currentImage.Pix, img.Pix)
			return
		}

		for i := range img.Pix {
			if img.Pix[i] == currentImage.Pix[i] {
				img.Pix[i] = 0
			} else {
				currentImage.Pix[i] = img.Pix[i]
			}
		}
	}
}

func copyImage(origin *image.Paletted) *image.Paletted {
	pix := make([]uint8, len(origin.Pix))
	copy(pix, origin.Pix)
	return &image.Paletted{
		Pix:     pix,
		Stride:  origin.Stride,
		Rect:    origin.Rect,
		Palette: origin.Palette,
	}
}
