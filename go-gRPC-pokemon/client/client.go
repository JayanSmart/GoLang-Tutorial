package main

import (
	"context"
	"fmt"
	pokemonpc "github.com/jayansmart/GoLang-Tutorial/go-grpc-pokemon/pokemon"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
)

const defaultPort = "4041"
const defaultHost = "localhost"

func main() {

	fmt.Println("Starting Pokemon Client...")

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

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())

	cc, err := grpc.Dial(host+":"+port, opts)
	if err != nil {
		log.Fatal("Failed to connect to server at: " + host + ":" + port)
	}
	defer cc.Close() // Consider handling this in a separate function later.

	c := pokemonpc.NewPokemonServiceClient(cc)

	// Create a pokemon
	fmt.Println("Creating a new Pokemon")
	pokemon := &pokemonpc.Pokemon{
		Pid:         "Poke0001",
		Name:        "Pikachu",
		Power:       "Electric",
		Description: "Mouse Pokemon",
	}

	createPokemonRes, err := c.CreatePokemon(context.Background(), &pokemonpc.CreatePokemonRequest{Pokemon: pokemon})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}

	fmt.Printf("Pokemon has been created: %v \n", createPokemonRes)
	pokemonID := createPokemonRes.GetPokemon().GetId()

	// Read Pokemon
	fmt.Printf("Reading Pokemon: %v \n", pokemonID)
	readPokemonReq := &pokemonpc.ReadPokemonRequest{Pid: pokemonID}
	readPokemonRes, err := c.ReadPokemon(context.Background(), readPokemonReq)
	if err != nil {
		fmt.Printf("Error happened while reading: %v \n", err)
	}

	fmt.Printf("Pokemon was read: %v \n", readPokemonRes)

	// update Pokemon
	newPokemon := &pokemonpc.Pokemon{
		Id:          pokemonID,
		Pid:         "Poke01",
		Name:        "Pikachu",
		Power:       "Fire Fire Fire",
		Description: "Fluffy",
	}
	updateRes, updateErr := c.UpdatePokemon(context.Background(), &pokemonpc.UpdatePokemonRequest{Pokemon: newPokemon})
	if updateErr != nil {
		fmt.Printf("Error happened while updating: %v \n", updateErr)
	}
	fmt.Printf("Pokemon was updated: %v\n", updateRes)

	// delete Pokemon
	deleteRes, deleteErr := c.DeletePokemon(context.Background(), &pokemonpc.DeletePokemonRequest{Pid: pokemonID})

	if deleteErr != nil {
		fmt.Printf("Error happened while deleting: %v \n", deleteErr)
	}
	fmt.Printf("Pokemon was deleted: %v \n", deleteRes)

	// list Pokemons

	stream, err := c.ListPokemon(context.Background(), &pokemonpc.ListPokemonRequest{})
	if err != nil {
		log.Fatalf("error while calling ListPokemon RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(res.GetPokemon())
	}
}
