package fortnite

import pb "2dFortnite/proto"

const (
	// one unit in the wall grid is WALL_GRID_SIZE in terms of player position
	WALL_GRID_SIZE = 10.0
	WALL_GRID_START_X = 0.0
	WALL_GRID_START_Y = 0.0

	MAX_PLAYERS = 16

	MAX_SPEED = 10.0

	MAX_INVENTORY_SIZE = 5

	PICKUP_RANGE = 5.0

	SERVER_TICKRATE = 30.0
)

type AmmoLimit uint32
const (
	PISTOL_AMMO_LIMIT AmmoLimit = 16
	SHOTGUN_AMMO_LIMIT = 5
	ASSAULT_RIFLE_AMMO_LIMIT = 30
	SMG_AMMO_LIMIT = 30
)

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
	pb.Weapon_PISTOL: 2.0,
	pb.Weapon_PUMP_SHOTGUN: 6.0,
	pb.Weapon_ASSAULT_RIFLE: 3.0,
	pb.Weapon_SMG: 4.0,
}

var WeaponCooldowns map[pb.Weapon]uint32 = map[pb.Weapon]uint32{
	pb.Weapon_PISTOL: 3,
	pb.Weapon_PUMP_SHOTGUN: 10,
	pb.Weapon_ASSAULT_RIFLE: 2,
	pb.Weapon_SMG: 1,
}

var WallHealth map[pb.Material]uint32 = map[pb.Material]uint32{
	pb.Material_WOOD: 150,
	pb.Material_BRICK: 300,
	pb.Material_METAL: 500,
}