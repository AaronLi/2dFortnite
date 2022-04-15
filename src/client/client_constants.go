package main

import (
	pb "2dFortnite/proto"
	"github.com/veandco/go-sdl2/sdl"
	"image/color"
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

var ResourceColours = map[pb.Consumable]color.RGBA{
	pb.Consumable_BANDAGES: color.RGBA{R: 255, G: 100, B: 100, A: 255},
	pb.Consumable_MEDKIT: color.RGBA{R: 100, G: 255, B: 100, A: 255},
	pb.Consumable_SMALL_SHIELD_POTION: color.RGBA{R: 64, G: 159, B: 237, A: 255},
	pb.Consumable_LARGE_SHIELD_POTION: color.RGBA{R: 17, G: 95, B: 240, A: 255},
	pb.Consumable_CHUG_JUG: color.RGBA{R: 255, G: 0, B: 255, A: 255},
}
