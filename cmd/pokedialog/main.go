package main

import (
	"flag"
	"image"
	"image/gif"
	"os"
	"time"

	"github.com/metalblueberry/pokedialog/pkg/pokedialog"
)

func main() {

	text := flag.String("text", "hello world", "text to be render")
	frames := flag.Int("frames", 0, "number of frames")
	duration := flag.Float64("duration", 0, "duration for the gif in seconds")
	optimize := flag.Bool("optimize", true, "post process gif to reduce size by using transparency on each frame")
	output := flag.String("output", "pokedialog.gif", "file output")

	flag.Parse()

	// b, err := os.Open("basic.png")
	// if err != nil {
	// 	panic(err)
	// }
	// defer b.Close()
	// dw, err := pokedialog.NewDrawerWithDialog(b, image.Rect(70, 100, 750, 340))
	dw, err := pokedialog.NewDrawer()
	if err != nil {
		panic(err)
	}

	gifs := dw.Gif(
		*text,
		pokedialog.GifConfig{
			FrameCount: *frames,
			Duration:   time.Duration(*duration) * time.Second,
		},
	)

	if *optimize {
		opt := GifFrameOptimizer()
		for i := range gifs.Image {
			opt(gifs.Image[i])
		}
	}

	f, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, gifs)
	if err != nil {
		panic(err)
	}
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
