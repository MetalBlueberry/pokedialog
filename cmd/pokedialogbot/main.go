package main

import (
	"bytes"
	"image/gif"
	"log"
	"os"
	"time"

	"github.com/metalblueberry/pokedialog/pkg/pokedialog"
	tele "gopkg.in/telebot.v3"
)

type tdlog struct {
	c tele.Context
}

func (td *tdlog) Write(m []byte) (int, error) {
	err := td.c.Send(string(m))
	return len(m), err
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
		dw, err := pokedialog.NewDrawer()
		if err != nil {
			return c.Send(err.Error())
		}
		dw.Log = log.New(&tdlog{c: c}, "", 0)

		buf := &bytes.Buffer{}
		g, err := dw.Gif(c.Text(), pokedialog.GifConfig{
			Duration:           time.Millisecond * 250 * time.Duration(len(c.Text())),
			EndParagraphFrames: 5,
		})
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

	println("ready")
	b.Start()
}
