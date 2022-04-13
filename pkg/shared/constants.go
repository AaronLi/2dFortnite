package fortnite

import (
	pb "2dFortnite/proto"
	"math"
	"image/color"
)

const (
	// one unit in the wall grid is WALL_GRID_SIZE in terms of player position
	WALL_GRID_SIZE = 50.0
	WALL_GRID_START_X = 0.0
	WALL_GRID_START_Y = 0.0

	MAX_PLAYERS = 16

	MAX_SPEED = 100.0

	PLAYER_RADIUS = 10.0

	MAX_INVENTORY_SIZE = 5

	PICKUP_RANGE = 50.0

	SERVER_TICKRATE = 30.0

	MIN_WORLD_X = -2000.0
	MAX_WORLD_X = 2000.0
	MIN_WORLD_Y = -2000.0
	MAX_WORLD_Y = 2000.0
)


var WeaponAmmoLimits = map[pb.Weapon]uint32{
	pb.Weapon_PISTOL: 16,
	pb.Weapon_PUMP_SHOTGUN: 5,
	pb.Weapon_ASSAULT_RIFLE: 30,
	pb.Weapon_SMG: 30,
}

type ProjectileSpeed float64

const (
	PISTOL_PROJECTILE_SPEED ProjectileSpeed = 10.0
	SHOTGUN_PROJECTILE_SPEED = 15.0
	ASSAULT_RIFLE_PROJECTILE_SPEED = 18.0
	SMG_PROJECTILE_SPEED = 15.0
)


var WeaponDamage map[pb.Weapon]map[pb.Rarity]uint32 = map[pb.Weapon]map[pb.Rarity]uint32{
	pb.Weapon_PISTOL: {
		pb.Rarity_COMMON: 24,
		pb.Rarity_UNCOMMON: 25,
		pb.Rarity_RARE: 26,
		pb.Rarity_EPIC: 28,
		pb.Rarity_LEGENDARY: 29,
	},
	pb.Weapon_PUMP_SHOTGUN: {
		pb.Rarity_COMMON: 84,
		pb.Rarity_UNCOMMON: 92,
		pb.Rarity_RARE: 100,
		pb.Rarity_EPIC: 108,
		pb.Rarity_LEGENDARY: 116,
	},
	pb.Weapon_ASSAULT_RIFLE: {
		pb.Rarity_COMMON: 30,
		pb.Rarity_UNCOMMON: 31,
		pb.Rarity_RARE: 33,
		pb.Rarity_EPIC: 35,
		pb.Rarity_LEGENDARY: 36,
	},
	pb.Weapon_SMG: {
		pb.Rarity_COMMON: 16,
		pb.Rarity_UNCOMMON: 17,
		pb.Rarity_RARE: 18,
		pb.Rarity_EPIC: 19,
		pb.Rarity_LEGENDARY: 20,
	},
}

var WeaponInaccuracy map[pb.Weapon]float64 = map[pb.Weapon]float64{
	pb.Weapon_PISTOL: 4.0,
	pb.Weapon_PUMP_SHOTGUN: 10.0,
	pb.Weapon_ASSAULT_RIFLE: 5.0,
	pb.Weapon_SMG: 8.0,
}

var WeaponCooldowns map[pb.Weapon]uint32 = map[pb.Weapon]uint32{
	pb.Weapon_PISTOL: 8,
	pb.Weapon_PUMP_SHOTGUN: 30,
	pb.Weapon_ASSAULT_RIFLE: 4,
	pb.Weapon_SMG: 1,
}

var WallHealth map[pb.Material]uint32 = map[pb.Material]uint32{
	pb.Material_WOOD: 150,
	pb.Material_BRICK: 300,
	pb.Material_METAL: 500,
}

var WeaponAmmoUsage = map[pb.Weapon]pb.Ammo {
	pb.Weapon_PISTOL: pb.Ammo_PISTOL_AMMO,
	pb.Weapon_PUMP_SHOTGUN: pb.Ammo_SHOTGUN_AMMO,
	pb.Weapon_ASSAULT_RIFLE: pb.Ammo_ASSAULT_RIFLE_AMMO,
	pb.Weapon_SMG: pb.Ammo_SMG_AMMO,	
}

var WeaponReloadTime map[pb.Weapon]uint32 = map[pb.Weapon]uint32{
	pb.Weapon_PISTOL: uint32(math.Round(1.5 * SERVER_TICKRATE)),
	pb.Weapon_PUMP_SHOTGUN: uint32(math.Round(5.0 * SERVER_TICKRATE)),
	pb.Weapon_ASSAULT_RIFLE: uint32(math.Round(2.5 * SERVER_TICKRATE)),
	pb.Weapon_SMG: uint32(math.Round(2.3 * SERVER_TICKRATE)),
}

var RarityColours map[pb.Rarity]color.RGBA = map[pb.Rarity]color.RGBA{
	pb.Rarity_COMMON: color.RGBA{
		R: 200,
		G: 200,
		B: 200,
	},

	pb.Rarity_UNCOMMON: color.RGBA{
		R: 100,
		G: 255,
		B: 100,
	},

	pb.Rarity_RARE: color.RGBA{
		R: 80,
		G: 80,
		B: 255,
	},

	pb.Rarity_EPIC: color.RGBA{
		R: 106,
		G: 39,
		B: 214,
	},

	pb.Rarity_LEGENDARY: color.RGBA{
		R: 242,
		G: 170,
		B: 31,
	},
}

var WeaponDisplayNames = map[pb.Weapon]string{
	pb.Weapon_PISTOL: "Pistol",
	pb.Weapon_PUMP_SHOTGUN: "Shotgun",
	pb.Weapon_ASSAULT_RIFLE: "Assault Rifle",
	pb.Weapon_SMG: "SMG",
}