package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/gfx"
	"os"
	"sync"
	"fmt"
	"math/rand"
)

const (
	WindowTitle = "2d Fortnite"
	WindowWidth = 1280
	WindowHeight = 720
	FrameRate = 60

	RectWidth = 20
	RectHeight = 40
	NumRects = WindowHeight / RectHeight
)

var rects [NumRects]sdl.Rect
var runningMutex sync.Mutex

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var fpsManager gfx.FPSmanager
	var err error

	sdl.Do(func() {
		window, err = sdl.CreateWindow(WindowTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, WindowWidth, WindowHeight, sdl.WINDOW_OPENGL)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer func() {
		sdl.Do(func() {
			window.Destroy()
		})
	}()

	sdl.Do(func() {
		gfx.InitFramerate(&fpsManager)
		gfx.SetFramerate(&fpsManager, FrameRate)
	})

	sdl.Do(func() {
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	})
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer func() {
		sdl.Do(func() {
			renderer.Destroy()
		})
	}()

	sdl.Do(func() {
		renderer.Clear()
	})

	for i := range rects {
		rects[i] = sdl.Rect{
			X: int32(rand.Int() % WindowWidth),
			Y: int32(i * WindowHeight / len(rects)),
			W: RectWidth,
			H: RectHeight,
		}
	}

	running := true
	for running {
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.QuitEvent:
					runningMutex.Lock()
					running = false
					runningMutex.Unlock()
				}
			}

			renderer.Clear()
			renderer.SetDrawColor(18, 151, 204, 0x20)
			renderer.FillRect(&sdl.Rect{0, 0, WindowWidth, WindowHeight})
		})

		// Do expensive stuff using goroutines
		wg := sync.WaitGroup{}
		for i := range rects {
			wg.Add(1)
			go func(i int) {
				rects[i].X = (rects[i].X + 10) % WindowWidth
				sdl.Do(func() {
					renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff)
					renderer.DrawRect(&rects[i])
				})
				wg.Done()
			}(i)
		}
		wg.Wait()

		sdl.Do(func() {
			renderer.Present()
			gfx.FramerateDelay(&fpsManager)
		})
	}

	return 0
}
