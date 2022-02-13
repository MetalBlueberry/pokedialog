package main

import (
	"flag"
	"image/gif"
	"os"
	"time"

	"github.com/metalblueberry/pokedialog/pkg/pokedialog"
)

func main() {

	text := flag.String("text", "hello world", "text to be rendered")
	frames := flag.Int("frames", 0, "number of frames")
	duration := flag.Float64("duration", 0, "duration for the gif in seconds")
	output := flag.String("output", "pokedialog.gif", "file output")
	endParagraphFrames := flag.Int("endParagraphFrames", 0, "end paragraph frames, this will give you more time to read the paragraph until the end")

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

	gifs, err := dw.Gif(
		*text,
		pokedialog.GifConfig{
			FrameCount:         *frames,
			Duration:           time.Duration(*duration * float64(time.Second)),
			EndParagraphFrames: *endParagraphFrames,
		},
	)

	if err != nil {
		panic(err)
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
