package main

import "fmt"

type Message struct {
	Role    string
	Content string
}

type Conversation struct {
	ID       int
	Messages []Message
}

func main() {
	conversations := map[int]*Conversation{
		1: {
			ID: 1,
			Messages: []Message{
				{Role: "user", Content: "What is a goroutine?"},
				{Role: "assistant", Content: "It is a lightweight unit of execution in Go."},
			},
		},
		2: {
			ID: 2,
			Messages: []Message{
				{Role: "user", Content: "What is a PostgreSQL index?"},
				{Role: "assistant", Content: "It is a data structure that can speed up queries."},
			},
		},

		3: {
			ID: 3,
			Messages: []Message{
				{Role: "user", Content: "What is Docker?"},
				// {Role: "assistant", Content: "CREATE INDEX index_name ON table_name (column_name);"},
			},
		},
	}

	// Add a new message to the first conversation.
	conversation := conversations[1]

	conversation.Messages = append(
		conversation.Messages,
		Message{
			Role:    "user",
			Content: "How do I start it?",
		},
	)

	conversations[2].Messages = append(
		conversations[2].Messages,
		Message{
			Role:    "user",
			Content: "What is the SQL command to create an index?",
		},
	)
	// Show the history of the first conversation.
	fmt.Println("--- CONVERSATION 1 ---")

	for _, msg := range conversations[1].Messages {
		fmt.Printf("%s: %s\n", msg.Role, msg.Content)
	}

	fmt.Println("\n--- CONVERSATION 2 ---")

	for _, msg := range conversations[2].Messages {
		fmt.Printf("%s: %s\n", msg.Role, msg.Content)
	}

	fmt.Println("\n--- CONVERSATION 3 ---")

	for _, msg := range conversations[3].Messages {
		fmt.Printf("%s: %s\n", msg.Role, msg.Content)
	}
}
