package main

import (
	"context"
	"fmt"
	pokemonpc "github.com/jayansmart/GoLang-Tutorial/go-grpc-pokemon/pokemon"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
)

const defaultHost = "localhost"
const defaultPort = "4041"

var collection *mongo.Collection

type server struct {
	pokemonpc.PokemonServiceServer
}

type PokemonItem struct {
	ID          primitive.ObjectID `bson:"_id, omitempty"`
	Pid         string             `bson:"pid"`
	Name        string             `bson:"name"`
	Power       string             `bson:"power"`
	Description string             `bson:"description"`
}

func getPokemonData(data *PokemonItem) *pokemonpc.Pokemon {
	return &pokemonpc.Pokemon{
		Pid:         data.Pid,
		Name:        data.Name,
		Power:       data.Power,
		Description: data.Description,
	}
}

func (*server) CreatePokemon(ctx context.Context, req *pokemonpc.CreatePokemonRequest) (*pokemonpc.CreatePokemonResponse, error) {
	fmt.Println("Create Pokemon")
	pokemon := req.GetPokemon()

	data := PokemonItem{
		Pid:         pokemon.GetPid(),
		Name:        pokemon.GetName(),
		Power:       pokemon.GetPower(),
		Description: pokemon.GetDescription(),
	}

	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Error: %v", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID"))
	}

	return &pokemonpc.CreatePokemonResponse{
		Pokemon: &pokemonpc.Pokemon{
			Id:          oid.Hex(),
			Pid:         pokemon.GetPid(),
			Name:        pokemon.GetName(),
			Power:       pokemon.GetPower(),
			Description: pokemon.GetDescription(),
		},
	}, nil
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("POKEMON_HOST")
	if host == "" {
		host = defaultHost
	}

	port := os.Getenv("POKEMON_PORT")
	if port == "" {
		port = defaultPort
	}

	mongoUrl := os.Getenv("MONGODB_URL")

	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Pokemon Service Started at " + host + ":" + port)
	collection = client.Database("pokemondb").Collection("pokemon")

	lis, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	pokemonpc.RegisterPokemonServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	// First we close the connection with MongoDB:
	fmt.Println("Closing MongoDB Connection")
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatalf("Error on disconnection with MongoDB : %v", err)
	}

	// Finally, we stop the server
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("End of Program")
}
