package main

import (
	"flag"
	"image/gif"
	"os"
	"time"

	"github.com/metalblueberry/pokedialog/pkg/pokedialog"
)

func main() {

	text := flag.String("text", "hello world", "text to be render")
	frames := flag.Int("frames", 0, "number of frames")
	duration := flag.Float64("duration", 0, "duration for the gif in seconds")
	output := flag.String("output", "pokedialog.gif", "file output")
	endParagraphDuration := flag.Float64("endParagraphDuration", 0, "end paragraph duration in seconds")

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
			FrameCount:           *frames,
			Duration:             time.Duration(*duration * float64(time.Second)),
			EndParagraphDuration: time.Duration(*endParagraphDuration * float64(time.Second)),
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
