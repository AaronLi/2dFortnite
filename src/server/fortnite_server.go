package main
import (
	pb "2dFortnite/proto"
	"2dFortnite/pkg/shared"
)

type FortniteServer struct{
	pb.UnimplementedFortniteServiceServer

	players map[uint64]pb.Player

	walls map[uint64]pb.WorldWall

	items map[uint64]pb.WorldItem

	projectiles map[uint64]pb.Projectile
}

func NewFortniteServer() *FortniteServer{
	return &FortniteServer{
		players: make(map[uint64]pb.Player),
		walls: make(map[uint64]pb.WorldWall),
		items: make(map[uint64]pb.WorldItem),
		projectiles: make(map[uint64]pb.Projectile),
	}
}

func (s *FortniteServer) StartServer(){

}

func (server *FortniteServer) updateWorld(){
	// step the world forward
	// move players
	for _, player := range server.players {
		player.Position.X += player.Position.VX
		player.Position.Y += player.Position.VY
	}
	// process player actions
	//TODO
	// move projectiles
	for k, projectile := range server.projectiles {
		projectile.Life -= 1
		if projectile.Life == 0{
			// delete projectile if it's at eol
			delete(server.projectiles, k)
			continue
		}
		projectile.Position.X += projectile.Position.VX
		projectile.Position.Y += projectile.Position.VY
	}
	// check hits

	// place players and walls in a super grid that holds all entities in a WALL_GRID_SIZExWALL_GRID_SIZE of the game world in each tile
	// makes searching through much easier since we only have to look at adjacent grid squares
	playerGridPositions := make(map[int64]map[int64]map[uint64]*pb.Player)
	wallGridPositions := make(map[int64]map[int64]map[uint64]*pb.WorldWall)


	for k, player := range server.players{
		playerGridPositions[int64(player.Position.Y/fortnite.WALL_GRID_SIZE)][int64(player.Position.X/fortnite.WALL_GRID_SIZE)][k] = &player
	}

	for k, wall := range server.walls {
		wallGridPositions[wall.Y][wall.X][k] = &wall
	}

	for projectileUUID, projectile := range server.projectiles{
		projectile_center_x := int(projectile.Position.X/fortnite.WALL_GRID_SIZE)
		projectile_center_y := int(projectile.Position.Y/fortnite.WALL_GRID_SIZE)

		// check 3x3 centered on projectile for collisions
		for y_off := -1; y_off <= 1; y_off++{
			for x_off := -1; x_off <= 1; x_off++{
				for uuid, player := range playerGridPositions[int64(projectile_center_y + y_off)][int64(projectile_center_x+x_off)] {
					if server.collidesPlayer(projectileUUID, uuid) {
						if player.Health > projectile.Damage {
							player.Health -= projectile.Damage
						}else{
							player.Health = 0
						}
					}
				}
			}
		}
	}
}

func (server *FortniteServer) collidesPlayer(projectileUUID uint64, playerUUID uint64) bool {
	//projectile := server.projectiles[projectileUUID]
	//player := server.players[playerUUID]

	// line (projectile +velocity) intersects circle (player)
	
	panic("Not implemented")
	
}

func (server *FortniteServer) collidesWall(projectileUUID uint64, wallUUID uint64) bool {
	//wall := server.walls[wallUUID]
	//projectile := server.projectiles[projectileUUID]

	// line (projectile + velocity) intersects line (wall)
	panic("Not implemented")
}