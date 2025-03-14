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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY is not set.")
	}

	fmt.Print("\033[H\033[2J")
	fmt.Println("Coder v0.0.1")
	fmt.Println("\nType \"/help\" for usage instructions.")

	client := NewOpenAIClient(apiKey)

	defaultModel := Model{
		Name:            "gpt-4o-mini",
		InputTokenCost:  0.15 / 1000000,
		OutputTokenCost: 0.60 / 1000000,
	}

	conversationState := &ConversationState{}

	for {
		responseRequest := ResponseRequest{
			Model: defaultModel.Name,
			// MaxTokens: 100,
			PreviousID: conversationState.PreviousID,
		}

		reader := bufio.NewReader(os.Stdin)

		if conversationState.Elapsed > time.Duration(0) {
			statusStr := fmt.Sprintf("[%v; %v ->; -> %v]", conversationState.Elapsed.Round(time.Millisecond), conversationState.InputTokens, conversationState.OutputTokens)
			fmt.Printf("\n%s > ", statusStr)
		} else {
			fmt.Print("\n> ")
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}

		input = strings.TrimSpace(input)

		if strings.EqualFold(input, "/exit") {
			break
		}

		if strings.EqualFold(input, "/new") {
			conversationState.clearState()
			continue
		}

		if strings.EqualFold(input, "/session") {
			fmt.Printf("\nSession token consumption: [in: %v; out: %v], [in: $%v; out: $%v]", conversationState.InputTokens, conversationState.OutputTokens, float64(conversationState.TotalInputTokens)*defaultModel.InputTokenCost, float64(conversationState.TotalOutputTokens)*defaultModel.OutputTokenCost)

			conversationState.Elapsed = time.Duration(0)
			continue
		}

		if strings.EqualFold(input, "/help") {
			fmt.Println("\nUsage:")
			fmt.Println("- Type your message and press Enter to get a response.")
			fmt.Println("- Type \"/new\" to start a new conversation.")
			fmt.Println("- Type \"/include <file-path> <prompt>\" to include a file in context.")
			fmt.Println("- Type \"/exit\" to exit the program.")

			conversationState.Elapsed = time.Duration(0)
			continue
		}

		if strings.HasPrefix(strings.ToLower(input), "/include") {
			parts := strings.Fields(input)
			if len(parts) < 3 {
				fmt.Println("Usage: /include <file-path> <query>")
				continue
			}

			filePath := parts[1]
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", filePath, err)
				continue
			}

			query := strings.Join(parts[2:], " ")

			query = strings.TrimSpace(query)
			input = fmt.Sprintf("%s\n--- INPUT FILE START ---\n%s\n--- INPUT FILE END ---", query, string(fileContent))
		}

		responseRequest.Input = input

		start := time.Now()

		response, err := client.CreateResponse(responseRequest)
		if err != nil {
			log.Fatalf("Error creating response: %v", err)
		}

		fmt.Print("\033[H\033[2J")

		fmt.Println(response.Output[0].Content[0].Text)

		conversationState.UpdateState(response, time.Since(start))
	}
}
