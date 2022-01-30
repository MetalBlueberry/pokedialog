package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func main() {
	dialogFile, err := os.Open("dialog.png")
	if err != nil {
		panic(err)
	}
	defer dialogFile.Close()
	dialog, err := png.Decode(dialogFile)
	if err != nil {
		panic(err)
	}

	img := image.NewRGBA(dialog.Bounds())
	draw.Draw(img, img.Bounds(), dialog, dialog.Bounds().Min, draw.Src)

	dw, err := InitDrawer()
	if err != nil {
		panic(err)
	}

	window := img.SubImage(image.Rect(185, 145, 1530, 435)).(*image.RGBA)

	dw.Draw(
		window,
		"Hello Typefont test length for the provided text",
	)

	f, err := os.Create("hello-go.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

type FrameDrawer struct {
	font *truetype.Font
}

func InitDrawer() (*FrameDrawer, error) {
	b, err := ioutil.ReadFile("Pokemon GB.ttf")
	if err != nil {
		return nil, err
	}
	f, err := truetype.Parse(b)
	if err != nil {
		return nil, err
	}

	return &FrameDrawer{
		font: f,
	}, nil
}

func (fd *FrameDrawer) Draw(img draw.Image, text string) error {
	col := color.RGBA{0, 0, 0, 255}

	bounds := img.Bounds()
	size := float64(bounds.Dy()*33) / float64(100)

	face := truetype.NewFace(fd.font, &truetype.Options{
		Size: size,
		DPI:  72,
	})

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	log.Print(bounds)
	for l, sentences := range SplitLines(face, text, bounds.Dx()) {
		d.Dot = dotForLine(bounds.Min.X, bounds.Min.Y, size, l)
		d.DrawString(sentences)
		d.DrawString(" ")
	}
	return nil
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

func addLabel(img *image.RGBA, x, y int, size float64, label string) {
	log.Println("drawing at %d %d", x, y)

	col := color.RGBA{0, 0, 0, 255}

	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	b, err := ioutil.ReadFile("Pokemon GB.ttf")
	if err != nil {
		log.Println(err)
		return
	}
	f, err := truetype.Parse(b)
	if err != nil {
		log.Println(err)
		return
	}

	d := &font.Drawer{
		Dst: img,
		Src: image.NewUniform(col),
		Face: truetype.NewFace(f, &truetype.Options{
			Size: size,
			DPI:  72,
		}),
		Dot: point,
	}
	d.DrawString(label)
}
