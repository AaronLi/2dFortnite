package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"log"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "2dFortnite/proto"
)

func main() {
	var exitcode int

	// connect to fortnite server over GRPC
	
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	client := pb.NewFortniteServiceClient(conn)

	request := pb.RegisterPlayerRequest{
		Name: "test",
		Skin: 0,
	}

	response, err := client.RegisterPlayer(context.Background(), &request)

	if err != nil {
		panic(err)
	}

	log.Printf("Response: ID: %d", response.Id)

	sdl.Main(func() {
		exitcode = run(&request, response.Id)
	})

	os.Exit(exitcode)
}