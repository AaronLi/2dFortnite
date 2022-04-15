package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"log"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"flag"
	pb "2dFortnite/proto"
)

func main() {
	var exitcode int

	var serverAddress = flag.String("server", "localhost:50051", "The server address in the format of host:port")

	flag.Parse()
	// connect to fortnite server over GRPC

	log.Println("Connecting to", *serverAddress)
	
	conn, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))

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
		exitcode = run(&request, response.Id, &client)
	})

	os.Exit(exitcode)
}