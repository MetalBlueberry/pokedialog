package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image/gif"
	"log"
	"os"
	"runtime"
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
	token, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		panic("BOT_TOKEN variable not found")
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

	logFile, err := os.Create(fmt.Sprintf("logs-%s.json", time.Now().Format(time.RFC3339)))
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	cl := NewConcurrentLimit(runtime.NumCPU() - 1)

	b.Handle(tele.OnText, func(c tele.Context) error {
		if c.Chat().Type != tele.ChatPrivate {
			log.Println("it is not a private chat")
			return nil
		}

		if err := c.Notify(tele.Typing); err != nil {
			log.Println(err)
		}

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

		if err := c.Notify(tele.UploadingDocument); err != nil {
			log.Println(err)
		}

		return c.Send(&tele.Document{
			File:     tele.FromReader(buf),
			FileName: "poke.gif",
			// DisableTypeDetection: true,
		})
	}, Monitor, cl.Handler)

	b.Handle("/pokedialog", func(c tele.Context) error {
		logger := &tdlog{c: c}
		defer logger.Close()

		if err := c.Notify(tele.Typing); err != nil {
			log.Println(err)
		}

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
				fmt.Fprintf(f.Output(), "Usage of %s:\nYou can generate gifs that look like old good pokedialogs\n\n/pokedialog [flags] \"text you want to create\"\n", f.Name())
			}
			f.PrintDefaults()
		}

		frames := f.Int("frames", 0, "number of frames")
		duration := f.String("duration", "0", "duration for the gif. 1s 10s 1m...")
		hr := f.Bool("hr", false, "generate a high resolution gif")
		endParagraphFrames := f.Int("endParagraphFrames", 0, "end paragraph frames, this will give you more time to read the paragraph until the end")

		err = f.Parse(args[1:])
		if err != nil {
			return nil
		}

		dw, err := pokedialog.NewDrawer()
		if err != nil {
			return err
		}
		dw.Log = log.New(logger, "", 0)
		text := strings.Join(f.Args(), " ")

		if len(text) <= 1 {
			f.Usage()
			return nil
		}

		parsedDuration, err := time.ParseDuration(*duration)
		if err != nil {
			return c.Send(fmt.Sprintf("Invalid duration %s, %s", *duration, err))
		}

		gifs, err := dw.Gif(
			text,
			pokedialog.GifConfig{
				FrameCount:         *frames,
				Duration:           parsedDuration,
				EndParagraphFrames: *endParagraphFrames,
			},
		)

		buf := &bytes.Buffer{}
		err = gif.EncodeAll(buf, gifs)
		if err != nil {
			return c.Send(err.Error())
		}

		if err := c.Notify(tele.UploadingDocument); err != nil {
			log.Println(err)
		}

		return c.Send(&tele.Document{
			File:                 tele.FromReader(buf),
			FileName:             "poke.gif",
			DisableTypeDetection: *hr,
		})

	}, Monitor, cl.Handler)

	b.Handle("/start", func(c tele.Context) error {
		return c.Send(`Welcome! Just send me a text and I will create a poke dialog with it. 
If you add multiple lines, each one will be in a different paragraph.
Try /pokedialog -help if you need more control`)
	}, Monitor)

	log.Println(cl)
	log.Println("ready")
	b.Start()
}

type LogEntry struct {
	Message LogEntryMessage `json:"message,omitempty"`
}

type LogEntryMessage struct {
	MessageID int         `json:"message_id,omitempty"`
	From      *tele.User  `json:"from,omitempty"`
	Chat      *tele.Chat  `json:"chat,omitempty"`
	Date      time.Time   `json:"date,omitempty"`
	Text      string      `json:"text,omitempty"`
	Error     error       `json:"error,omitempty"`
	Panic     interface{} `json:"panic,omitempty"`
}

func Monitor(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) (err error) {
		defer func() {
			pan := recover()
			if pan != nil {
				c.Send("ouch! don't do that!!")
			}
			entry := LogEntry{
				Message: LogEntryMessage{
					MessageID: c.Message().ID,
					From:      c.Message().Sender,
					Chat:      c.Message().Chat,
					Date:      c.Message().Time(),
					Text:      c.Message().Text,
					Error:     err,
					Panic:     pan,
				},
			}
			text, err := json.Marshal(entry)
			if err != nil {
				log.Println(err)
			} else {
				log.Println(string(text))
			}
		}()
		return next(c)
	}
}

func NewConcurrentLimit(max int) *ConcurrentLimit {
	cl := &ConcurrentLimit{
		workers: make(chan *Worker, max),
	}
	for i := 0; i < max; i++ {
		cl.workers <- &Worker{ID: i}
	}
	return cl
}

type ConcurrentLimit struct {
	workers chan *Worker
}

type Worker struct {
	ID int
}

func (cl *ConcurrentLimit) String() string {
	return fmt.Sprintf("Concurrency set to %d", len(cl.workers))
}

func (cl *ConcurrentLimit) Adquire(ctx context.Context) (*Worker, error) {
	select {
	case w := <-cl.workers:
		return w, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (cl *ConcurrentLimit) Return(w *Worker) {
	cl.workers <- w
}

func (cl *ConcurrentLimit) Handler(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		w, err := cl.Adquire(ctx)
		if err != nil {
			return c.Send(fmt.Sprintf("I'm too busy, try again later\n%s", err.Error()))
		}
		defer cl.Return(w)

		return next(c)
	}
}
