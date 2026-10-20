package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
)

func main() {

	client, err := api.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}
	prompts := []string{
		"what is Go? just 1 sentence",
		"what is go? explain it to a backend developer. just 1 sentence",
		"what is Go? Explain it to a 10-year-old child in just 2 sentences",
	}

	ctx := context.Background()

	for i, prompt := range prompts {
		fmt.Printf("\n--- Test %d ---\n", i+1)
		fmt.Println("Prompt:", prompt)
		fmt.Println("Answer:")

		req := &api.GenerateRequest{
			Model:  "llama3.2:latest",
			Prompt: prompt,
		}

		err := client.Generate(ctx, req, func(resp api.GenerateResponse) error {
			fmt.Print(resp.Response)
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("\n")
	}

}
