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
)

type AmmoLimit uint32
const (
	PISTOL_AMMO_LIMIT AmmoLimit = 16
	SHOTGUN_AMMO_LIMIT = 5
	ASSAULT_RIFLE_AMMO_LIMIT = 30
	SMG_AMMO_LIMIT = 30
	ROCKET_LAUNCHER_AMMO_LIMIT = 1
)

type ProjectileSpeed float64

const (
	PISTOL_PROJECTILE_SPEED ProjectileSpeed = 10.0
	SHOTGUN_PROJECTILE_SPEED = 10.0
	ASSAULT_RIFLE_PROJECTILE_SPEED = 10.0
	SMG_PROJECTILE_SPEED = 10.0
	ROCKET_LAUNCHER_PROJECTILE_SPEED = 10.0
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
	pb.Weapon_ROCKET_LAUNCHER: {
		pb.Rarity_COMMON: 70,
		pb.Rarity_UNCOMMON: 85,
		pb.Rarity_RARE: 100,
		pb.Rarity_EPIC: 115,
		pb.Rarity_LEGENDARY: 130,
	},
}
