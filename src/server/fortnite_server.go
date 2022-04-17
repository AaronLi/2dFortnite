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

	walls map[uint64]*pb.WorldWall

	items map[uint64]*pb.WorldItem

	projectiles map[uint64]*pb.Projectile

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
		Inventory: make([]*pb.InventorySlot, 5),
		Resources: make([]*pb.ResourceStack, 0),
	}

	for i := 0; i < 5; i++ {
		player.Inventory[i] = &pb.InventorySlot{
		}
	}

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_MATERIAL,
		ItemData: &pb.ResourceStack_Material{
			Material: pb.Material_WOOD,
		},
	})

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_MATERIAL,
		ItemData: &pb.ResourceStack_Material{
			Material: pb.Material_BRICK,
		},
	})

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_MATERIAL,
		ItemData: &pb.ResourceStack_Material{
			Material: pb.Material_METAL,
		},
	})

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_AMMO,
		ItemData: &pb.ResourceStack_Ammo{
			Ammo: pb.Ammo_PISTOL_AMMO,
		},
	})

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_AMMO,
		ItemData: &pb.ResourceStack_Ammo{
			Ammo: pb.Ammo_ASSAULT_RIFLE_AMMO,
		},
	})

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_AMMO,
		ItemData: &pb.ResourceStack_Ammo{
			Ammo: pb.Ammo_SHOTGUN_AMMO,
		},
	})

	player.Resources = append(player.Resources, &pb.ResourceStack{
		Item: pb.ItemType_AMMO,
		ItemData: &pb.ResourceStack_Ammo{
			Ammo: pb.Ammo_SNIPER_AMMO,
		},
	})

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
		walls: make(map[uint64]*pb.WorldWall),
		items: make(map[uint64]*pb.WorldItem),
		projectiles: make(map[uint64]*pb.Projectile),
		queuedActions: make(chan *pb.DoActionRequest, 128),
		connections: make(map[uint64]*ClientConnection),
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
	s.populateWorld()
	terminator := make(chan bool)
	defer func() {terminator <- true}()

	go s.serverThread(terminator)

	err = grpcServer.Serve(lis)

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
		wallsList = append(wallsList, wall)
	}

	itemsList := make([]*pb.WorldItem, 0)
	for _, item := range server.items{
		itemsList = append(itemsList, item)
	}

	projectilesList := make([]*pb.Projectile, 0)
	for _, projectile := range server.projectiles{
		projectilesList = append(projectilesList, projectile)
	}

	worldState := pb.WorldStateResponse{
		Players: playersList,
		Walls: wallsList,
		Items: itemsList,
		Projectiles: projectilesList,
	}

	for k, connection := range server.connections{
		worldState.Player = server.players[k]
		err := (*(connection.connection)).Send(&worldState)
		if err != nil {
			delete(server.connections, k)
		}
	}
}

func (server *FortniteServer) updateWorld(){
	// step the world forward

	// place players and walls in a super grid that holds all entities in a WALL_GRID_SIZExWALL_GRID_SIZE of the game world in each tile
	// makes searching through much easier since we only have to look at adjacent grid squares
	playerGridPositions := make(map[int64]map[int64]map[uint64]*pb.Player)
	wallGridPositions := make(map[int64]map[int64]map[pb.WallOrientation]uint64)

	//log.Println("An item position", server.items[0].Pos)

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

		// player inventory cooldown
		for _, item := range player.Inventory{
			if item.Cooldown > 0 {
				item.Cooldown--
			}
			if item.Reload > 0 {
				item.Reload--
			}
		}
	}

	for k, wall := range server.walls {
		if _, present := wallGridPositions[wall.Y]; !present {
			wallGridPositions[wall.Y] = make(map[int64]map[pb.WallOrientation]uint64)
		}
		if _, present := wallGridPositions[wall.Y][wall.X]; !present {
			wallGridPositions[wall.Y][wall.X] = make(map[pb.WallOrientation]uint64, 2)
		}
		wallGridPositions[wall.Y][wall.X][wall.Orientation] = k
	}

	// process actions
	for i := len(server.queuedActions); i > 0; i-- {
		action := <- server.queuedActions
		player, present := server.players[action.PlayerId.Id]
		if !present {
			continue
		}
		switch action.ActionType {
			case pb.ActionType_PICKUP_ITEM:
				pickupItemRequest := action.GetPickupItem()

				// verify item exists and is in range
				item, exists := server.items[pickupItemRequest.ItemId]
				
				if exists {
					dx := item.Pos.X - player.Position.X
					dy := item.Pos.Y - player.Position.Y
					distance := math.Hypot(dx, dy)
					if distance < fortnite.PICKUP_RANGE {
						switch item.ItemType {
						case pb.ItemType_MATERIAL: // materials and ammo interact with the inventory in the same way
						for _, slot := range player.Resources {
							if slot.Item == pb.ItemType_MATERIAL && slot.GetMaterial() == item.GetMaterial() {
								slot.StackSize += item.StackSize
								break
							}
						}
						case pb.ItemType_AMMO:
						for _, slot := range player.Resources {
							if slot.Item == pb.ItemType_AMMO && slot.GetAmmo() == item.GetAmmo() {
								slot.StackSize += item.StackSize
								break
							}
						}
						case pb.ItemType_CONSUMABLE: // both consumables and weapons interact with the inventory the same way
						pickedUp := false
							// put item in first empty slot
							for _, slot := range player.Inventory {
								if slot.Item == pb.ItemType_NONE {
									slot.Item = item.ItemType
									slot.Rarity = item.ItemRarity
									slot.ItemData = &pb.InventorySlot_Consumable{
										Consumable: item.GetConsumable(),
									}
									slot.StackSize = item.StackSize
									slot.Cooldown = 0
									slot.Reload =  0
									pickedUp = true
									break
								}
							}
							if !pickedUp {
								// swap with item at currently selected slot
								server.dropItemInventory(player.EquippedSlot, player.Id)
								
								player.Inventory[player.EquippedSlot].Item = item.ItemType
								player.Inventory[player.EquippedSlot].Rarity = item.ItemRarity
								player.Inventory[player.EquippedSlot].StackSize = item.StackSize
								player.Inventory[player.EquippedSlot].ItemData = &pb.InventorySlot_Weapon{
									Weapon: item.GetWeapon(),
								}
								player.Inventory[player.EquippedSlot].Cooldown = 0
								player.Inventory[player.EquippedSlot].Reload =  0
							}
						case pb.ItemType_WEAPON:
							pickedUp := false
							// put item in first empty slot
							for _, slot := range player.Inventory {
								if slot.Item == pb.ItemType_NONE {
									slot.Item = item.ItemType
									slot.Rarity = item.ItemRarity
									slot.ItemData = &pb.InventorySlot_Weapon{
										Weapon: item.GetWeapon(),
									}
									slot.StackSize = item.StackSize
									slot.Cooldown = 0
									slot.Reload =  0
									pickedUp = true
									break
								}
							}
							if !pickedUp {
								// swap with item at currently selected slot
								server.dropItemInventory(player.EquippedSlot, player.Id)
								
								player.Inventory[player.EquippedSlot].Item = item.ItemType
								player.Inventory[player.EquippedSlot].Rarity = item.ItemRarity
								player.Inventory[player.EquippedSlot].StackSize = item.StackSize
								player.Inventory[player.EquippedSlot].ItemData = &pb.InventorySlot_Weapon{
									Weapon: item.GetWeapon(),
								}
								player.Inventory[player.EquippedSlot].Cooldown = 0
								player.Inventory[player.EquippedSlot].Reload =  0
							}
						}
						delete(server.items, pickupItemRequest.ItemId)
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

				moveMagnitude := math.Min(requestMagnitude, 1.0)
				if moveMagnitude > 0 {
					server.players[action.PlayerId.Id].Position.VX = fortnite.MAX_SPEED * moveMagnitude * moveRequest.Vx / requestMagnitude
					server.players[action.PlayerId.Id].Position.VY = fortnite.MAX_SPEED * moveMagnitude * moveRequest.Vy / requestMagnitude
				} else{
					server.players[action.PlayerId.Id].Position.VX = 0
					server.players[action.PlayerId.Id].Position.VY = 0
				}
				server.players[action.PlayerId.Id].Rotation = moveRequest.Facing
			case pb.ActionType_SHOOT_PROJECTILE:

				shootWeaponInfo := action.GetShootProjectile()
				// check if player has weapon equipped
				// check if player has ammo

				player.Rotation = shootWeaponInfo.Facing

				equippedItem := player.Inventory[player.EquippedSlot]
				if equippedItem.Item == pb.ItemType_WEAPON {
					if equippedItem.Cooldown == 0 && equippedItem.Reload == 0 && equippedItem.StackSize > 0{
						weaponInfo := equippedItem.GetWeapon()
						switch weaponInfo {
						case pb.Weapon_PUMP_SHOTGUN:
							for i := 0; i < 10; i++ {
								server.spawnBullet(weaponInfo, equippedItem, player)
							}
						default:
							server.spawnBullet(weaponInfo, equippedItem, player)
						}
						equippedItem.StackSize -= 1
						equippedItem.Cooldown = fortnite.WeaponCooldowns[weaponInfo]
					}
				}
			case pb.ActionType_BUILD_WALL:
				buildWallInfo := action.GetBuildWall()
				// check if player isn't too far from the desired position
				// check if player has enough resources
				for _, resource := range player.Resources {
					if resource.Item == pb.ItemType_MATERIAL{
						materialType := resource.GetMaterial()

						if materialType == buildWallInfo.Material {
							if resource.StackSize >= 10 {
								if _, present := wallGridPositions[buildWallInfo.Y]; !present {
									wallGridPositions[buildWallInfo.Y] = make(map[int64]map[pb.WallOrientation]uint64)
								}
								if _, present := wallGridPositions[buildWallInfo.Y][buildWallInfo.X]; !present {
									wallGridPositions[buildWallInfo.Y][buildWallInfo.X] = make(map[pb.WallOrientation]uint64, 2)
								}
								_, ok := wallGridPositions[buildWallInfo.Y][buildWallInfo.X][buildWallInfo.Facing]
								if !ok{
									// wall doesn't exist, create it
									log.Println("Build")
									wallGridPositions[buildWallInfo.Y][buildWallInfo.X][buildWallInfo.Facing] = server.buildWall(buildWallInfo, resource)
									break
								}
							}
						}
					}
				}
			case pb.ActionType_USE_ITEM:
				user := server.players[action.PlayerId.Id]
				item := user.Inventory[user.EquippedSlot]
				if item.Item == pb.ItemType_CONSUMABLE && item.Cooldown == 0 && item.StackSize > 0 {
					consumableInfo := item.GetConsumable()
					log.Println("Attempting to use", consumableInfo)
					switch consumableInfo {
					case pb.Consumable_BANDAGES:
						if user.Health < 75 {
							user.Health += 15
							if user.Health > 75 {
								user.Health = 75
							}
							item.StackSize -= 1
							item.Cooldown = 30
						}
					case pb.Consumable_MEDKIT:
						user.Health += 100
						if user.Health > 100 {
							user.Health = 100
						}
						item.StackSize -= 1
						item.Cooldown = 120
					case pb.Consumable_SMALL_SHIELD_POTION:
						if user.Shields < 50 {
							user.Shields += 25
							if user.Shields > 50 {
								user.Shields = 50
							}
							item.StackSize -= 1
							item.Cooldown = 30
						}
					case pb.Consumable_LARGE_SHIELD_POTION:
						if user.Shields < 100 {
							user.Shields += 50
							if user.Shields > 100 {
								user.Shields = 100
							}
							item.StackSize -= 1
							item.Cooldown = 120
						}
					case pb.Consumable_CHUG_JUG:
						if user.Health < 100 || user.Shields < 100{
							user.Health = 100
							user.Shields = 100
							item.StackSize -= 1
							item.Cooldown = 300
						}
					}
					if item.StackSize == 0{
						user.Inventory[user.EquippedSlot].Item = pb.ItemType_NONE
					}
				} else if item.Item == pb.ItemType_WEAPON {
					// check if weapon is fully loaded
					// check if reload timer is finished
					// check if player has enough ammo
					if item.StackSize < fortnite.WeaponAmmoLimits[item.GetWeapon()] {
						// weapon not fully loaded
						if item.Reload <= 0 {
							// weapon not being reloaded
							for _, resource := range user.Resources {
								if resource.Item == pb.ItemType_AMMO {
									// item is ammo
									if resource.GetAmmo() == fortnite.WeaponAmmoUsage[item.GetWeapon()] {
										// correct ammo type for this weapon
										if resource.StackSize > 0 {
											reloadAmount := fortnite.WeaponAmmoLimits[item.GetWeapon()] - item.StackSize
											if resource.StackSize < reloadAmount {
												// player doesn't have enough ammo for full reload
												reloadAmount = resource.StackSize
											}
											item.StackSize += reloadAmount
											resource.StackSize -= reloadAmount
											item.Reload = fortnite.WeaponReloadTime[item.GetWeapon()]
										}
										break
									}
								}
							}
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
		// check walls around player
		player_center_x := int64(player.Position.X/fortnite.WALL_GRID_SIZE)
		player_center_y := int64(player.Position.Y/fortnite.WALL_GRID_SIZE)
		hitSomething := false

		for y_off := int64(-1); y_off <= 1; y_off++ {
			for x_off := int64(-1); x_off <= 1; x_off++ {
				check_y := player_center_y + y_off
				check_x := player_center_x + x_off
				for orientation := 0; orientation < 2; orientation++ {
					if wall_uuid, present := wallGridPositions[check_y][check_x][pb.WallOrientation(orientation)]; present {
						wall := server.walls[wall_uuid]
						wallStartX := float64(wall.X*fortnite.WALL_GRID_SIZE + fortnite.WALL_GRID_START_X)
						wallStartY := float64(wall.Y*fortnite.WALL_GRID_SIZE + fortnite.WALL_GRID_START_Y)
						wallEndX := float64(wallStartX)
						wallEndY := float64(wallStartY)
						
						if orientation == int(pb.WallOrientation_HORIZONTAL) {
							wallEndX += fortnite.WALL_GRID_SIZE
						} else {
							wallEndY += fortnite.WALL_GRID_SIZE
						}

						if server.rayCollidesLine(
							player.Position.X,
							player.Position.Y,
							player.Position.VX * (1.0 / fortnite.SERVER_TICKRATE),
							player.Position.VY * (1.0 / fortnite.SERVER_TICKRATE),
							wallStartX,
							wallStartY,
							wallEndX,
							wallEndY){
							hitSomething = true
							break
						}
					}
				}
				if hitSomething {
					break
				}
			}
			if hitSomething {
				break
			}
		}
		if !hitSomething{
			player.Position.X += player.Position.VX * (1.0 / fortnite.SERVER_TICKRATE)
			player.Position.Y += player.Position.VY * (1.0 / fortnite.SERVER_TICKRATE)
		}
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
		hitSomething := false
		// check 3x3 centered on projectile for collisions
		for y_off := int64(-1); y_off <= 1; y_off++{
			for x_off := int64(-1); x_off <= 1; x_off++{
				check_y := projectile_center_y + y_off
				check_x := projectile_center_x+x_off
				hitPlayer := false
				hitWall := false
				for uuid, player := range playerGridPositions[check_y][check_x] {
					if server.collidesPlayer(projectileUUID, uuid) {
						if player.Shields > projectile.Damage {
							player.Shields -= projectile.Damage
						}else{
							projectile.Damage -= player.Shields
							player.Shields = 0

						}
						if player.Shields == 0 {
							if player.Health > projectile.Damage{
								player.Health -= projectile.Damage
							}else{
								player.Health = 0
								deaths = append(deaths, uuid)
							}
						}
					
						hitPlayer = true
						break
					}
				}
				for orientation := 0; orientation < 2; orientation++ {
					if wall_uuid, ok := wallGridPositions[check_y][check_x][pb.WallOrientation(orientation)]; ok && !hitPlayer {
						if server.collidesWall(projectileUUID, wall_uuid) {
							wall := server.walls[wall_uuid]
							if wall.Health > projectile.Damage {
								wall.Health -= projectile.Damage
							}else{
								wall.Health = 0
								walls_broken = append(walls_broken, wall_uuid)
							}
							hitWall = true
						}
					}
				}
				if hitPlayer || hitWall {
					hitSomething = true
					break
				}
			}
			if hitSomething {
				break
			}
		}
		if hitSomething {
			log.Printf("Deleting projectile at %.3f,%.3f moving at %.3f %.3f", projectile.Position.X, projectile.Position.Y, projectile.Position.VX, projectile.Position.VY)
			delete(server.projectiles, projectileUUID)
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
	server.items[worldItem.Id] = &worldItem
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
	server.items[worldItem.Id] = &worldItem
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
		Orientation: buildWallInfo.Facing,
		Health: fortnite.WallHealth[buildWallInfo.Material],
	}
	server.walls[wallId] = &wall
	resource.StackSize -= 10
	return wallId
}

func (server *FortniteServer) spawnBullet(weaponInfo pb.Weapon, equippedItem *pb.InventorySlot, user *pb.Player) {
	angle_offset := (rand.Float64() - 0.5) * fortnite.WeaponInaccuracy[weaponInfo]

	projectile := pb.Projectile{
		Id: rand.Uint64(),
		Position: &pb.NetworkPosition{
			X: user.Position.X,
			Y: user.Position.Y,
			VX: math.Cos((user.Rotation + angle_offset) * math.Pi / 180) * float64(fortnite.WeaponProjectileSpeed[weaponInfo]),
			VY: math.Sin((user.Rotation + angle_offset) * math.Pi / 180) * float64(fortnite.WeaponProjectileSpeed[weaponInfo]),
		},
		Damage: fortnite.WeaponDamage[weaponInfo][equippedItem.Rarity],
		Life: 300,
	}
	server.projectiles[projectile.Id] = &projectile

	equippedItem.Cooldown = fortnite.WeaponCooldowns[weaponInfo]
}

func (server *FortniteServer) collidesPlayer(projectileUUID uint64, playerUUID uint64) bool {
	projectile := server.projectiles[projectileUUID]
	player := server.players[playerUUID]

	if projectile.Owner == playerUUID {
		return false
	}

	// line (projectile +velocity) intersects circle (player)
	
	Ux := player.Position.X - projectile.Position.X
	Uy := player.Position.Y - projectile.Position.Y

	VelocityMagnitude := math.Hypot(projectile.Position.VX, projectile.Position.VY)

	Vx := projectile.Position.VX / VelocityMagnitude
	Vy := projectile.Position.VY / VelocityMagnitude

	U1intermediate := Ux * Vx + Uy * Vy

	U1x := U1intermediate * Vx
	U1y := U1intermediate * Vy

	U1dotDirection := U1x * projectile.Position.VX + U1y * projectile.Position.VY

	if U1dotDirection < 0 {
		return false
	}
	
	U2x := Ux - U1x
	U2y := Uy - U1y

	d := math.Hypot(U2x, U2y)
	
	targetRadius := float64(fortnite.PLAYER_RADIUS)
	if d <= targetRadius {
		m := math.Sqrt(targetRadius * targetRadius - d * d)
		mVx := m * Vx
		mVy := m * Vy
		p1x := projectile.Position.X + U1x + mVx
		p1y := projectile.Position.Y + U1y + mVy

		// check if projectile to p1 is less than projectile to velocity
		p1DistanceX := projectile.Position.X - p1x
		p1DistanceY := projectile.Position.Y - p1y

		if math.Hypot(p1DistanceX, p1DistanceY) < math.Hypot(projectile.Position.VX, projectile.Position.VY) {
			return true
		}
		if d < targetRadius{
			p2x := projectile.Position.X + U1x - mVx
			p2y := projectile.Position.Y + U1y - mVy

			p2DistanceX := projectile.Position.X - p2x
			p2DistanceY := projectile.Position.Y - p2y
			if math.Hypot(p2DistanceX, p2DistanceY) < math.Hypot(projectile.Position.VX, projectile.Position.VY) {
				return true
			}
		}
	}
	return false

}

func (server *FortniteServer) collidesWall(projectileUUID uint64, wallUUID uint64) bool {
	//wall := server.walls[wallUUID]
	//projectile := server.projectiles[projectileUUID]

	// line (projectile + velocity) intersects line (wall)
	projectile := server.projectiles[projectileUUID]
	wall := server.walls[wallUUID]

	// v1 = self.pos - L.p1
	wallX := (float64(wall.X) * fortnite.WALL_GRID_SIZE + fortnite.WALL_GRID_START_X)
	wallY := (float64(wall.Y) * fortnite.WALL_GRID_SIZE + fortnite.WALL_GRID_START_Y)

	wallEndX := wallX
	wallEndY := wallY

	if wall.Orientation == pb.WallOrientation_HORIZONTAL {
		wallEndX += fortnite.WALL_GRID_SIZE
	}else{
		wallEndY += fortnite.WALL_GRID_SIZE
	}
	return server.rayCollidesLine(projectile.Position.X, projectile.Position.Y, projectile.Position.VX, projectile.Position.VY, wallX, wallY, wallEndX, wallEndY)
}

func (server *FortniteServer) rayCollidesLine(rX, rY, rDx, rDy, l1x, l1y, l2x, l2y float64) bool{
//wall := server.walls[wallUUID]
	//projectile := server.projectiles[projectileUUID]

	// line (projectile + velocity) intersects line (wall)

	// v1 = self.pos - L.p1
	v1x := rX - l1x
	v1y := rY - l1y

	// v2 = L.p2 - L.p1
	v2x := l2x - l1x
	v2y := l2y - l1y

	// v3 = -self.direction[1], self.direction[0]
	v30 := -rDy
	v31 := rDx

	v2dotv3 := v2x * v30 + v2y * v31

	// t1 = cross(v2, v1) / dot(v2, v3)
	t1 := (v2x * v1y - v2y * v1x) / v2dotv3

	// t2 = dot(v1, v3) / dot(v2, v3)
	t2 := (v1x * v30 + v1y * v31) / v2dotv3

	if t1 >= 0.0 && t2 >= 0.0 && t2 <= 1.0{
		intersectX := rX + t1 * rDx
		intersectY := rY + t1 * rDy

		intersectXDirection := intersectX - rX
		intersectYDirection := intersectY - rY
		intersectDistance := math.Hypot(intersectXDirection, intersectYDirection)

		if intersectDistance < math.Hypot(rDx, rDy) {
			return true
		}
		
	}
	return false
}

func (server *FortniteServer) populateWorld(){
	for i := 0; i < 100; i++ {
		// spawn weapons
		var worldItem pb.WorldItem
		worldItem.Id = rand.Uint64()
		worldItem.Pos = &pb.NetworkPosition{
			X: fortnite.MIN_WORLD_X + rand.Float64() * float64(fortnite.MAX_WORLD_X - fortnite.MIN_WORLD_X),
			Y: fortnite.MIN_WORLD_Y + rand.Float64() * float64(fortnite.MAX_WORLD_Y - fortnite.MIN_WORLD_Y),
		}

		worldItem.ItemType = pb.ItemType_WEAPON

		worldItem.ItemData = &pb.WorldItem_Weapon{
			Weapon: pb.Weapon(rand.Intn(5)),
		}
		worldItem.ItemRarity = server.intToRarity(rand.Intn(4))

		worldItem.StackSize = fortnite.WeaponAmmoLimits[worldItem.GetWeapon()]
		server.items[worldItem.Id] = &worldItem

		// spawn appropriate ammo near weapon
		var resource pb.WorldItem
		resource.Id = rand.Uint64()
		resource.Pos = &pb.NetworkPosition{
			X: worldItem.Pos.X + (rand.Float64()-0.5) * 30.0,
			Y: worldItem.Pos.Y + (rand.Float64()-0.5) * 30.0,
		}

		resource.ItemType = pb.ItemType_AMMO

		resource.ItemData = &pb.WorldItem_Ammo{
			Ammo: fortnite.WeaponAmmoUsage[worldItem.GetWeapon()],
		}
		resource.StackSize = uint32(rand.Intn(20) + 10)
		server.items[resource.Id] = &resource
	}

	for i := 0; i < 200; i++ {
		// spawn materials
		var worldItem pb.WorldItem
		worldItem.Id = rand.Uint64()
		worldItem.Pos = &pb.NetworkPosition{
			X: fortnite.MIN_WORLD_X + rand.Float64() * float64(fortnite.MAX_WORLD_X - fortnite.MIN_WORLD_X),
			Y: fortnite.MIN_WORLD_Y + rand.Float64() * float64(fortnite.MAX_WORLD_Y - fortnite.MIN_WORLD_Y),
		}

		worldItem.ItemType = pb.ItemType_MATERIAL

		worldItem.ItemData = &pb.WorldItem_Material{
			Material: server.intToMaterial(rand.Intn(3)),
		}

		worldItem.StackSize = uint32(rand.Intn(20) + 10)
		server.items[worldItem.Id] = &worldItem
	}

	for i := 0; i< 100; i++ {
		// spawn consumables
		var worldItem pb.WorldItem
		worldItem.Id = rand.Uint64()
		worldItem.Pos = &pb.NetworkPosition{
			X: fortnite.MIN_WORLD_X + rand.Float64() * float64(fortnite.MAX_WORLD_X - fortnite.MIN_WORLD_X),
			Y: fortnite.MIN_WORLD_Y + rand.Float64() * float64(fortnite.MAX_WORLD_Y - fortnite.MIN_WORLD_Y),
		}

		worldItem.ItemType = pb.ItemType_CONSUMABLE

		worldItem.ItemData = &pb.WorldItem_Consumable{
			Consumable: pb.Consumable(rand.Intn(5)),
		}

		worldItem.StackSize = uint32(rand.Intn(2)+1)
		server.items[worldItem.Id] = &worldItem
	}
}

func (server *FortniteServer) intToMaterial(m int) pb.Material {
	switch m {
	case 0:
		return pb.Material_WOOD
	case 1:
		return pb.Material_BRICK
	case 2:
		return pb.Material_METAL
	}
	return pb.Material_WOOD
}

func (server *FortniteServer) intToRarity(r int) pb.Rarity {
	switch r {
	case 0:
		return pb.Rarity_COMMON
	case 1:
		return pb.Rarity_UNCOMMON
	case 2:
		return pb.Rarity_RARE
	case 3:
		return pb.Rarity_EPIC
	}
	return pb.Rarity_COMMON
}