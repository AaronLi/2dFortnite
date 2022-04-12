package main
import (
	empty "github.com/golang/protobuf/ptypes/empty"
	pb "2dFortnite/proto"
	"2dFortnite/pkg/shared"
	"context"
	"math/rand"
	"math"
	"net"
	"log"
	"time"
	"google.golang.org/grpc"
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
		Name: in.Name,
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

	log.Printf("Player %s(%d) registered", player.Name, player.Id)

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

func (s *FortniteServer) DoAction(ctx context.Context, request *pb.DoActionRequest) (*empty.Empty, error){
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
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}	

	grpcServer := grpc.NewServer()

	pb.RegisterFortniteServiceServer(grpcServer, s)

	log.Printf("Starting server on port 50051")
	
	terminator := make(chan bool)

	go s.serverThread(terminator)

	err = grpcServer.Serve(lis)
	terminator <- true

	if err != nil {	
		log.Fatalf("failed to serve: %v", err)
	}
}

func (server *FortniteServer) serverThread(terminator chan bool){
	ticker := time.NewTicker(time.Second * 1.0 / fortnite.SERVER_TICKRATE)

	for {
		select {
		case <- terminator:
			return
		case <- ticker.C:
			server.updateWorld()
			server.transmitState()
		}
	}
}

func (server *FortniteServer) transmitState(){
	playersList := make([]*pb.Player, 0)
	for _, player := range server.players{
		playersList = append(playersList, player)
	}

	wallsList := make([]*pb.WorldWall, 0)
	for _, wall := range server.walls{
		wallsList = append(wallsList, &wall)
	}

	itemsList := make([]*pb.WorldItem, 0)
	for _, item := range server.items{
		itemsList = append(itemsList, &item)
	}

	projectilesList := make([]*pb.Projectile, 0)
	for _, projectile := range server.projectiles{
		projectilesList = append(projectilesList, &projectile)
	}

	for _, connection := range server.connections{
		err := (*(connection.connection)).Send(&pb.WorldStateResponse{
			Players: playersList,
			Walls: wallsList,
			Items: itemsList,
			Projectiles: projectilesList,
		})
		if err != nil {
			connection.errorChannel <- err
		}
	}
}

func (server *FortniteServer) updateWorld(){
	// step the world forward

	// place players and walls in a super grid that holds all entities in a WALL_GRID_SIZExWALL_GRID_SIZE of the game world in each tile
	// makes searching through much easier since we only have to look at adjacent grid squares
	playerGridPositions := make(map[int64]map[int64]map[uint64]*pb.Player)
	wallGridPositions := make(map[int64]map[int64]uint64)


	for k, player := range server.players{
		gridY := int64(player.Position.Y/fortnite.WALL_GRID_SIZE)
		gridX := int64(player.Position.X/fortnite.WALL_GRID_SIZE)

		if _, exists := playerGridPositions[gridY]; !exists {
			playerGridPositions[gridY] = make(map[int64]map[uint64]*pb.Player)
		}
		if _, exists := playerGridPositions[gridY][gridX]; !exists{
			playerGridPositions[gridY][gridX] = make(map[uint64]*pb.Player)
		}

		playerGridPositions[gridY][gridX][k] = player
	}

	for k, wall := range server.walls {
		wallGridPositions[wall.Y][wall.X] = k
	}

	// process actions
	for i := len(server.queuedActions); i > 0; i-- {
		action := <- server.queuedActions

		switch action.ActionType {
			case pb.ActionType_PICKUP_ITEM:
				pickupItemRequest := action.GetPickupItem()
				// verify item exists and is in range
				item, exists := server.items[pickupItemRequest.ItemId]
				// verify player has inventory space
				player := server.players[action.PlayerId.Id]
				
				if exists {
					dx := item.Pos.X - player.Position.X
					dy := item.Pos.Y - player.Position.Y
					distance := math.Hypot(dx, dy)
					if distance < fortnite.PICKUP_RANGE {
						switch item.ItemType {
						case pb.ItemType_MATERIAL: // materials and ammo interact with the inventory in the same way
						fallthrough	
						case pb.ItemType_AMMO:
						for _, slot := range player.Resources {
							if slot.Item == item.ItemType {
								slot.StackSize += item.StackSize
								break
							}
						}
						case pb.ItemType_CONSUMABLE: // both consumables and weapons interact with the inventory the same way
						fallthrough
						case pb.ItemType_WEAPON:
							pickedUp := false
							// put item in first empty slot
							for _, slot := range player.Inventory {
								if slot.Item == pb.ItemType_NONE {
									slot.Item = item.ItemType
									slot.Rarity = item.ItemRarity
									slot.StackSize = item.StackSize
									slot.Cooldown = 0
									slot.Reload =  0
									pickedUp = true
									break
								}
							}
							if !pickedUp {
								// swap with item at currently selected slot

							}
						}
					}
				}
			case pb.ActionType_DROP_ITEM:
				dropItemRequest := action.GetDropItem() 

				server.dropItemInventory(dropItemRequest.SlotNumber, action.PlayerId.Id)
			case pb.ActionType_DROP_RESOURCE:
				dropResourceRequest := action.GetDropResource()
				server.dropItemResource(dropResourceRequest.SlotNumber, action.PlayerId.Id)
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
				user := server.players[action.PlayerId.Id]

				user.Rotation = shootWeaponInfo.Facing

				equippedItem := user.Inventory[user.EquippedSlot]
				if equippedItem.Item == pb.ItemType_WEAPON {
					if equippedItem.Cooldown == 0 && equippedItem.StackSize > 0{
						weaponInfo := equippedItem.GetWeapon()
						switch weaponInfo {
						case pb.Weapon_PISTOL:
							server.spawnBullet(weaponInfo, equippedItem, user)
						case pb.Weapon_PUMP_SHOTGUN:
							for i := 0; i < 10; i++ {
								server.spawnBullet(weaponInfo, equippedItem, user)
							}
						case pb.Weapon_SMG:
							server.spawnBullet(weaponInfo, equippedItem, user)
						case pb.Weapon_ASSAULT_RIFLE:
							server.spawnBullet(weaponInfo, equippedItem, user)
						}
						equippedItem.StackSize -= 1
					}
				}
			case pb.ActionType_BUILD_WALL:
				buildWallInfo := action.GetBuildWall()
				user := server.players[action.PlayerId.Id]
				// check if player isn't too far from the desired position
				// check if player has enough resources
				for _, resource := range user.Resources {
					if resource.Item == pb.ItemType_MATERIAL{
						materialType := resource.GetMaterial()

						if materialType == buildWallInfo.Material {
							if resource.StackSize >= 10 {
								wall_uuid, ok := wallGridPositions[buildWallInfo.Y][buildWallInfo.X]
								if !ok{
									// wall doesn't exist, create it
									wallGridPositions[buildWallInfo.Y][buildWallInfo.X] = server.buildWall(buildWallInfo, resource)
									break
								}else if server.walls[wall_uuid].Health < fortnite.WallHealth[server.walls[wall_uuid].Material]{
									// wall exists but does not have full health
									wallGridPositions[buildWallInfo.Y][buildWallInfo.X] = server.buildWall(buildWallInfo, resource)
									break
								}
							}
						}
					}
				}
			case pb.ActionType_USE_ITEM:
				user := server.players[action.PlayerId.Id]
				item := user.Inventory[user.EquippedSlot]
				if item.Item == pb.ItemType_CONSUMABLE {
					consumableInfo := item.GetConsumable()
					switch consumableInfo {
					case pb.Consumable_BANDAGES:
						if user.Health < 75 {
							user.Health += 15
							if user.Health > 75 {
								user.Health = 75
							}
						}
					case pb.Consumable_MEDKIT:
						user.Health += 100
						if user.Health > 100 {
							user.Health = 100
						}
					case pb.Consumable_SMALL_SHIELD_POTION:
						if user.Shields < 50 {
							user.Shields += 25
							if user.Shields > 50 {
								user.Shields = 50
							}
						}
					case pb.Consumable_LARGE_SHIELD_POTION:
						if user.Shields < 100 {
							user.Shields += 50
							if user.Shields > 100 {
								user.Shields = 100
							}
						}
					case pb.Consumable_CHUG_JUG:
						if user.Health < 100 || user.Shields < 100{
							user.Health = 100
							user.Shields = 100
						}
					}
				}
			case pb.ActionType_SWAP_ITEM:
				swapItemRequest := action.GetSwapItem()
				user := server.players[action.PlayerId.Id]
				toSlot := user.Inventory[swapItemRequest.SlotNumber]
				user.Inventory[swapItemRequest.SlotNumber] = user.Inventory[swapItemRequest.SlotNumber2]
				user.Inventory[swapItemRequest.SlotNumber2] = toSlot
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
				if wall_uuid, ok := wallGridPositions[check_y][check_x]; ok {
					if server.collidesWall(projectileUUID, wall_uuid) {
						wall := server.walls[wall_uuid]
						if wall.Health > projectile.Damage {
							wall.Health -= projectile.Damage
						}else{
							wall.Health = 0
							walls_broken = append(walls_broken, wall_uuid)
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

func (server *FortniteServer) dropItemInventory(inventoryIndex int32, player_uuid uint64){

	itemToDrop := server.players[player_uuid].Inventory[inventoryIndex]

	var worldItem pb.WorldItem

	if itemToDrop.Item == pb.ItemType_CONSUMABLE {
		worldItem = pb.WorldItem{
			Id: rand.Uint64(),
			Pos: &pb.NetworkPosition{
				X: server.players[player_uuid].Position.X,
				Y: server.players[player_uuid].Position.Y,
			},
			ItemType: itemToDrop.Item,
			ItemData: &pb.WorldItem_Consumable{Consumable: itemToDrop.GetConsumable()},
			ItemRarity: itemToDrop.Rarity,
			StackSize: itemToDrop.StackSize,
		}
	}else if itemToDrop.Item == pb.ItemType_WEAPON {
		worldItem = pb.WorldItem{
			Id: rand.Uint64(),
			Pos: &pb.NetworkPosition{
				X: server.players[player_uuid].Position.X,
				Y: server.players[player_uuid].Position.Y,
			},
			ItemType: itemToDrop.Item,
			ItemData: &pb.WorldItem_Weapon{Weapon: itemToDrop.GetWeapon()},
			ItemRarity: itemToDrop.Rarity,
			StackSize: itemToDrop.StackSize,
		}
	}
	server.items[worldItem.Id] = worldItem
	server.players[player_uuid].Inventory[inventoryIndex] = &pb.InventorySlot{}
}

func (server *FortniteServer) dropItemResource(resourceIndex int32, player_uuid uint64){
	itemToDrop := server.players[player_uuid].Resources[resourceIndex]
	var worldItem pb.WorldItem

	if itemToDrop.Item == pb.ItemType_AMMO {
		worldItem = pb.WorldItem{
			Id: rand.Uint64(),
			Pos: &pb.NetworkPosition{
				X: server.players[player_uuid].Position.X,
				Y: server.players[player_uuid].Position.Y,
			},
			ItemType: itemToDrop.Item,
			ItemData: &pb.WorldItem_Ammo{Ammo: itemToDrop.GetAmmo()},
			StackSize: itemToDrop.StackSize,
		}
	}else if itemToDrop.Item == pb.ItemType_MATERIAL {
		worldItem = pb.WorldItem{
			Id: rand.Uint64(),
			Pos: &pb.NetworkPosition{
				X: server.players[player_uuid].Position.X,
				Y: server.players[player_uuid].Position.Y,
			},
			ItemType: itemToDrop.Item,
			ItemData: &pb.WorldItem_Material{Material: itemToDrop.GetMaterial()},
			StackSize: itemToDrop.StackSize,
		}
	}
	server.items[worldItem.Id] = worldItem
	server.players[player_uuid].Resources[resourceIndex] = &pb.ResourceStack{
		Item: itemToDrop.Item,
	}
}

func (server *FortniteServer) buildWall(buildWallInfo *pb.BuildWallRequest, resource *pb.ResourceStack) uint64{
	wallId := rand.Uint64()
	wall := pb.WorldWall{
		X: buildWallInfo.X,
		Y: buildWallInfo.Y,
		Material: buildWallInfo.Material,
	}
	server.walls[wallId] = wall
	resource.StackSize -= 10
	return wallId
}

func (server *FortniteServer) spawnBullet(weaponInfo pb.Weapon, equippedItem *pb.InventorySlot, user *pb.Player) {
	angle_offset := rand.Float64() * fortnite.WeaponInaccuracy[weaponInfo]

	projectile := pb.Projectile{
		Id: rand.Uint64(),
		Position: &pb.NetworkPosition{
			X: user.Position.X,
			Y: user.Position.Y,
			VX: math.Cos((user.Rotation + angle_offset) * math.Pi / 180) * float64(fortnite.PISTOL_PROJECTILE_SPEED),
			VY: math.Sin((user.Rotation + angle_offset) * math.Pi / 180) * float64(fortnite.PISTOL_PROJECTILE_SPEED),
		},
		Damage: fortnite.WeaponDamage[weaponInfo][equippedItem.Rarity],
		Life: 300,
	}
	server.projectiles[projectile.Id] = projectile

	equippedItem.Cooldown = fortnite.WeaponCooldowns[weaponInfo]
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