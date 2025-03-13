package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultModel = "gpt-4o-mini"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set.")
	}

	client := NewOpenAIClient(apiKey)

	conversationState := ConversationState{
		PreviousID: "",
	}

	for {
		responseRequest := ResponseRequest{
			Model: defaultModel,
			// MaxTokens: 100,
			PreviousID: conversationState.PreviousID,
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}

		input = strings.TrimSpace(input)

		if strings.EqualFold(input, "/exit") {
			break
		}

		if strings.EqualFold(input, "/new") {
			conversationState.PreviousID = ""
			continue
		}

		if strings.HasPrefix(strings.ToLower(input), "/attach") {
			parts := strings.Fields(input)
			if len(parts) < 2 {
				fmt.Println("Usage: /attach <file-path>")
				continue
			}

			filePath := parts[1]
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				continue
			}

			fmt.Print("\n! File attached\n> ")
			query, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Error reading input: %v", err)
			}

			query = strings.TrimSpace(query)

			input = fmt.Sprintf("%s\n--- INPUT FILE START ---\n%s\n--- INPUT FILE END ---", query, string(fileContent))
		}

		responseRequest.Input = input

		start := time.Now()

		response, err := client.CreateResponse(responseRequest)
		if err != nil {
			log.Fatalf("Error creating response: %v", err)
		}

		elapsed := time.Since(start)

		fmt.Println(response.Output[0].Content[0].Text)
		conversationState.PreviousID = response.ID

		fmt.Printf("\n! Server response time: %v\n", elapsed)
	}
}
