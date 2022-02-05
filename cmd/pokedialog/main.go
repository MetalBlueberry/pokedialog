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
	frames := flag.Int("frames", 5, "number of frames")
	flag.Parse()

	dw, err := pokedialog.NewDrawer("dialog.png", 3, image.Rect(185, 145, 1530, 435))
	if err != nil {
		panic(err)
	}

	gifs := dw.Gif(
		*text,
		*frames,
		time.Second*5,
	)

	f, err := os.Create("hello-go.gif")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, gifs)
	if err != nil {
		panic(err)
	}
}
