package main
import (
	empty "github.com/golang/protobuf/ptypes/empty"
	pb "2dFortnite/proto"
	"2dFortnite/pkg/shared"
	"context"
	"math/rand"
	"math"
	"errors"
)

type FortniteServer struct{
	pb.UnimplementedFortniteServiceServer

	players map[uint64]*pb.Player

	walls map[uint64]pb.WorldWall

	items map[uint64]pb.WorldItem

	projectiles map[uint64]pb.Projectile

	connections map[uint64]*ClientConnection

	queuedActions chan *pb.DoActionRequest
}

type ClientConnection struct{
	connection *pb.FortniteService_WorldStateServer
	errorChannel chan error
}

func (s *FortniteServer) RegisterPlayer(ctx context.Context, in *pb.RegisterPlayerRequest) (*pb.RegisterPlayerResponse, error){
	if len(s.players) >= fortnite.MAX_PLAYERS {
		return nil, errors.New("Server is full")
	}

	player := pb.Player{
		Id: rand.Uint64(),
		Skin: in.Skin,
		Position: &pb.NetworkPosition{
			X: 0,
			Y: 0,
			VX: 0,
			VY: 0,
		},
		Health: 100,
	}

	s.players[player.Id] = &player
	return &pb.RegisterPlayerResponse{
		Id: player.Id,
	}, nil
}

func (s *FortniteServer) WorldState(player *pb.PlayerId, response_channel pb.FortniteService_WorldStateServer) error{
	client_connection := &ClientConnection{
		connection: &response_channel,
		errorChannel: make(chan error),
	}
	s.connections[player.Id] = client_connection
	return <- client_connection.errorChannel
}

func (s *FortniteServer) DoAction(request *pb.DoActionRequest) (*empty.Empty, error){
	// add action to queue
	s.queuedActions <- request
	return &empty.Empty{}, nil
}

func NewFortniteServer() *FortniteServer{
	return &FortniteServer{
		players: make(map[uint64]*pb.Player),
		walls: make(map[uint64]pb.WorldWall),
		items: make(map[uint64]pb.WorldItem),
		projectiles: make(map[uint64]pb.Projectile),
		queuedActions: make(chan *pb.DoActionRequest, 128),
	}
}

func (s *FortniteServer) StartServer(){

}

func (server *FortniteServer) updateWorld(){
	// step the world forward

	// process actions
	for i := len(server.queuedActions); i > 0; i-- {
		action := <- server.queuedActions

		switch action.ActionType {
			case pb.ActionType_PICKUP_ITEM:

			case pb.ActionType_DROP_ITEM:      

			case pb.ActionType_MOVE_PLAYER:
				// Update player's velocity
				// clamp requested velocity to max speed
				moveRequest := action.GetMovePlayer()

				requestMagnitude := math.Hypot(moveRequest.Vx, moveRequest.Vy)

				moveMagnitude := math.Min(requestMagnitude, fortnite.MAX_SPEED)

				server.players[action.PlayerId.Id].Position.VX = moveMagnitude * moveRequest.Vx/ requestMagnitude
				server.players[action.PlayerId.Id].Position.VY = moveMagnitude * moveRequest.Vy / requestMagnitude
				server.players[action.PlayerId.Id].Rotation = moveRequest.Facing
			case pb.ActionType_SHOOT_PROJECTILE:

				shootWeaponInfo := action.GetShootProjectile()
				// check if player has weapon equipped
				// check if player has ammo
				user := server.players[shootWeaponInfo.PlayerId.Id]

				equippedItem := user.Inventory[user.EquippedSlot]
				if equippedItem.Item == pb.ItemType_WEAPON {
					weaponInfo := equippedItem.GetWeapon()

					switch weaponInfo {
					case pb.Weapon_PISTOL:
						// create projectile
						projectile := pb.Projectile{
							Id: rand.Uint64(),
							Position: &pb.NetworkPosition{
								X: user.Position.X,
								Y: user.Position.Y,
								VX: math.Cos(user.Rotation) * float64(fortnite.PISTOL_PROJECTILE_SPEED),
								VY: math.Sin(user.Rotation) * float64(fortnite.PISTOL_PROJECTILE_SPEED),
							},
							Damage: fortnite.WeaponDamage[weaponInfo][equippedItem.Rarity],
							Life: 300,
						}
						server.projectiles[projectile.Id] = projectile
					case pb.Weapon_PUMP_SHOTGUN:
					case pb.Weapon_SMG:
					case pb.Weapon_ASSAULT_RIFLE:
					case pb.Weapon_ROCKET_LAUNCHER:
					}
				}
			case pb.ActionType_BUILD_WALL:

			case pb.ActionType_USE_ITEM:
			
			case pb.ActionType_SWAP_ITEM:
			case pb.ActionType_SELECT_ITEM:
				// select item if possible
				selectItemRequest := action.GetSelectItem()

				if selectItemRequest.SlotNumber < fortnite.MAX_INVENTORY_SIZE {
					server.players[action.PlayerId.Id].EquippedSlot = selectItemRequest.SlotNumber
				}
		}
	}

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
		playerGridPositions[int64(player.Position.Y/fortnite.WALL_GRID_SIZE)][int64(player.Position.X/fortnite.WALL_GRID_SIZE)][k] = player
	}

	for k, wall := range server.walls {
		wallGridPositions[wall.Y][wall.X][k] = &wall
	}

	deaths := make([]uint64, 0)
	walls_broken := make([]uint64, 0)

	for projectileUUID, projectile := range server.projectiles{
		projectile_center_x := int64(projectile.Position.X/fortnite.WALL_GRID_SIZE)
		projectile_center_y := int64(projectile.Position.Y/fortnite.WALL_GRID_SIZE)

		// check 3x3 centered on projectile for collisions
		for y_off := int64(-1); y_off <= 1; y_off++{
			for x_off := int64(-1); x_off <= 1; x_off++{
				check_y := projectile_center_y + y_off
				check_x := projectile_center_x+x_off
				for uuid, player := range playerGridPositions[check_y][check_x] {
					if server.collidesPlayer(projectileUUID, uuid) {
						if player.Health > projectile.Damage {
							player.Health -= projectile.Damage
						}else{
							player.Health = 0
							deaths = append(deaths, uuid)
						}
					}
				}

				for uuid, wall := range wallGridPositions[check_y][check_x] {
					if server.collidesWall(projectileUUID, uuid) {
						if wall.Health > projectile.Damage {
							wall.Health -= projectile.Damage
						}else{
							wall.Health = 0
							walls_broken = append(walls_broken, uuid)
						}
					}
				}
			}
		}
	}

	// removed destroyed walls and dead players
	for _, uuid := range deaths {
		delete(server.players, uuid)
	}

	for _, uuid := range walls_broken {
		delete(server.walls, uuid)
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