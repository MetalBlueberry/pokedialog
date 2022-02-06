# Pokedialog

![pokedialog](welcome.gif)

```bash
go run cmd/pokedialog/main.go --duration 10 --output welcome.gif --text 'Welcome!!
     This project helps you to build pokedialogs!
     Just follow the instructions below'
```

## Usage

```sh
  -duration float
        duration for the gif in seconds
  -frames int
        number of frames
  -optimize
        post process gif to reduce size by using transparency on each frame (default true)
  -output string
        file output (default "pokedialog.gif")
  -text string
        text to be render (default "hello world")
```
