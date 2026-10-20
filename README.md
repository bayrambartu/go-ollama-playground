# Go Ollama Playground

A hands-on roadmap for learning LLM fundamentals and building LLM applications with Go and Ollama.

Each folder contains a small experiment for learning a specific concept. Differences between examples, such as keeping or clearing conversation history, are intentional learning exercises.

## Learning roadmap

### 01 - LLM fundamentals and Ollama

- Prompt basics
- System, user, and assistant roles
- Tokens and conversation context
- Model options and small experiments

### 02 - Go and LLM integration

- Calling the Ollama HTTP API
- Comparing HTTP requests with the Go SDK
- Building a terminal chatbot with conversation history
- Streaming responses
- Handling errors, timeouts, and cancellation
- Building a REST API with Fiber
- Organizing a mini project into handler, service, and LLM layers

## Running the examples

Install the Go version specified in `go.mod`, install Ollama, and start the Ollama server. Download the model used in the examples:

```sh
ollama pull llama3.2:latest
```

From the repository root, run an example:

```sh
go run ./01-llm-fundamentals-and-ollama/1.1-prompt-and-llm-basics
```

Run the mini project API:

```sh
go run ./02-go-llm-integration/2.7-mini-project/cmd/server
```

The API listens on port 8080 and provides `GET /health` and `POST /chat`. Send a JSON body such as `{"message":"What is a goroutine?"}` to `/chat`.

Source text is written in English. Some prompts intentionally request answers in Turkish.
