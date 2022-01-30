package main

import (
	"image"
	"image/png"
	"os"

	"github.com/metalblueberry/pokedialog/pkg/pokedialog"
)

func main() {

	dw, err := pokedialog.NewDrawer("dialog.png")
	if err != nil {
		panic(err)
	}

	img := dw.Draw(
		image.Rect(185, 145, 1530, 435),
		// window,
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
