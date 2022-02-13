package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/gif"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/metalblueberry/pokedialog/pkg/pokedialog"
	tele "gopkg.in/telebot.v3"
)

type tdlog struct {
	c   tele.Context
	buf bytes.Buffer
}

func (td *tdlog) Write(m []byte) (int, error) {
	return td.buf.Write(m)
}

func (td *tdlog) Flush() error {
	err := td.c.Send(td.buf.String())
	if err != nil {
		return err
	}
	td.buf.Reset()
	return nil
}

func (td *tdlog) Close() error {
	return td.Flush()
}

func main() {
	token, err := os.ReadFile("token.txt")
	if err != nil {
		panic(err)
	}
	pref := tele.Settings{
		Token:  string(token),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnText, func(c tele.Context) error {
		log.Println(c.Text())

		dw, err := pokedialog.NewDrawer()
		if err != nil {
			return c.Send(err.Error())
		}
		logger := &tdlog{c: c}
		defer logger.Close()
		dw.Log = log.New(logger, "", 0)

		buf := &bytes.Buffer{}
		g, err := dw.Gif(c.Text(), pokedialog.GifConfig{})
		if err != nil {
			return c.Send(err.Error())
		}
		err = gif.EncodeAll(buf, g)
		if err != nil {
			return c.Send(err.Error())
		}

		return c.Send(&tele.Document{
			File:     tele.FromReader(buf),
			FileName: "poke.gif",
			// DisableTypeDetection: true,
		})
	})

	b.Handle("/gif", func(c tele.Context) error {
		log.Println(c.Text())

		logger := &tdlog{c: c}
		defer logger.Close()

		args, err := shellwords.Parse(c.Message().Text)
		if err != nil {
			return c.Send(err.Error())
		}

		f := flag.NewFlagSet("pokedialog", flag.ContinueOnError)
		f.SetOutput(logger)

		f.Usage = func() {
			if f.Name() == "" {
				fmt.Fprintf(f.Output(), "Usage:\n")
			} else {
				fmt.Fprintf(f.Output(), "Usage of %s:\nYou can generate gifs that look like old good pokedialogs\n\n/gif [flags] \"text you want to create\"\n", f.Name())
			}
			f.PrintDefaults()
		}

		if len(args) <= 1 {
			return c.Send("You have to specify the text after the /gif command")
		}

		frames := f.Int("frames", 0, "number of frames")
		duration := f.Float64("duration", 0, "duration for the gif in seconds")
		hr := f.Bool("hr", false, "generate a high resolution gif")
		endParagraphFrames := f.Int("endParagraphFrames", 0, "end paragraph frames, this will give you more time to read the paragraph until the end")

		err = f.Parse(args[1:])
		if err != nil {
			return nil
		}

		dw, err := pokedialog.NewDrawer()
		if err != nil {
			panic(err)
		}
		dw.Log = log.New(logger, "", 0)
		text := strings.Join(f.Args(), " ")
		gifs, err := dw.Gif(
			text,
			pokedialog.GifConfig{
				FrameCount:         *frames,
				Duration:           time.Duration(*duration * float64(time.Second)),
				EndParagraphFrames: *endParagraphFrames,
			},
		)

		buf := &bytes.Buffer{}
		err = gif.EncodeAll(buf, gifs)
		if err != nil {
			return c.Send(err.Error())
		}
		return c.Send(&tele.Document{
			File:                 tele.FromReader(buf),
			FileName:             "poke.gif",
			DisableTypeDetection: *hr,
		})

	})

	b.Handle("/start", func(c tele.Context) error {
		return c.Send(`Welcome! Just send me a text and I will create a poke dialog with it. 
If you add multiple lines, each one will be in a different paragraph.
Try /gif -help if you need more control`)
	})

	println("ready")
	b.Start()
}
