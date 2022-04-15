package main

import (
	pb "2dFortnite/proto"
	"github.com/veandco/go-sdl2/sdl"
)

var MaterialDrawPositions = map[pb.Material]sdl.Rect{
	pb.Material_WOOD: sdl.Rect{
		X: int32(WindowWidth - (55 * 3) - 5),
		Y: int32(WindowHeight- 120),
		W: 50,
		H: 50,
	},
	pb.Material_BRICK: sdl.Rect{
		X: int32(WindowWidth - (55 * 2) - 5),
		Y: int32(WindowHeight- 120),
		W: 50,
		H: 50,
	},
	pb.Material_METAL: sdl.Rect{
		X: int32(WindowWidth - 55 - 5 ),
		Y: int32(WindowHeight- 120),
		W: 50,
		H: 50,
	},
}