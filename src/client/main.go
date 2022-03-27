package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

func main() {
	var exitcode int

	sdl.Main(func() {
		exitcode = run()
	})

	os.Exit(exitcode)
}