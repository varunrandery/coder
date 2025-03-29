package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

func (cs *ConversationState) UpdateState(response ResponseBody, elapsed time.Duration) {
	cs.PreviousID = response.ID
	cs.InputTokens = response.Usage.InputTokens
	cs.OutputTokens = response.Usage.OutputTokens
	cs.TotalInputTokens += response.Usage.InputTokens
	cs.TotalOutputTokens += response.Usage.OutputTokens
	cs.Elapsed = elapsed
}

func (cs *ConversationState) ClearState() {
	cs.PreviousID = ""
	cs.InputTokens = 0
	cs.OutputTokens = 0
	cs.TotalInputTokens = 0
	cs.TotalOutputTokens = 0
	cs.Elapsed = time.Duration(0)
}

func handleSlashCommand(input *string, cs *ConversationState, selectedModel *Model, validModels map[string]Model) bool {
	parts := strings.Fields(*input)
	if len(parts) == 0 {
		return true
	}

	switch parts[0] {
	case "/exit":
		os.Exit(0)
		return true

	case "/new":
		cs.ClearState()
		return true

	case "/session":
		fmt.Printf("Session token consumption: [in: %v; out: %v], [in: $%.2f; out: $%.2f]\n", cs.TotalInputTokens, cs.TotalOutputTokens, float64(cs.TotalInputTokens)*selectedModel.InputTokenCost, float64(cs.TotalOutputTokens)*selectedModel.OutputTokenCost)
		cs.Elapsed = time.Duration(0)
		return true

	case "/help":
		fmt.Println("Usage:")
		fmt.Println("- Type your message and press Enter to get a response.")
		fmt.Println("- Type \"/new\" to start a new conversation.")
		fmt.Println("- Type \"/include <file-path> <prompt>\" to include a file in context.")
		fmt.Println("- Type \"/session\" to view current-conversation token consumption.")
		fmt.Println("- Type \"/exit\" to exit the program.")
		fmt.Println("- Type \"/model info\" to view the current model.")
		fmt.Println("- Type \"/model switch\" to change the current model.")
		cs.Elapsed = time.Duration(0)
		return true

	case "/include":
		if len(parts) < 3 {
			fmt.Println("Usage: /include <file-path> <query>")
			return true
		}

		filePath := parts[1]
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			return true
		}

		query := strings.Join(parts[2:], " ")
		query = strings.TrimSpace(query)
		*input = fmt.Sprintf("%s\n--- INPUT FILE START ---\n%s\n--- INPUT FILE END ---", query, string(fileContent))
		return false

	case "/model": // "/model info" or "/model switch" commands
		if len(parts) == 2 && parts[1] == "info" {
			fmt.Printf("Current model: %s (input token cost: $%.2f/M, output token cost: $%.2f/M)\n", selectedModel.Name, selectedModel.InputTokenCost*1000000, selectedModel.OutputTokenCost*1000000)
			return true
		} else if len(parts) >= 2 && parts[1] == "switch" {
			keys := slices.Sorted(maps.Keys(validModels))

			fmt.Println("Usage: /model switch <model-name>")
			fmt.Printf("\nAvailable models:\n%s\n", strings.Join(keys, ", "))

			if len(parts) == 3 {
				modelKey := parts[2]
				if model, exists := validModels[modelKey]; exists {
					*selectedModel = model
					fmt.Printf("\nSwitched to model: %s\n", modelKey)
					cs.ClearState()
				} else {
					fmt.Printf("\nInvalid model name: %s\n", modelKey)
				}
			}
			return true
		} else {
			// return /model info
			fmt.Printf("Current model: %s (input token cost: $%.2f/M, output token cost: $%.2f/M)\n", selectedModel.Name, selectedModel.InputTokenCost*1000000, selectedModel.OutputTokenCost*1000000)
			return true
		}

	default:
		fmt.Printf("Unknown command: %s\n", parts[0])
		return true
	}
}

func startSpinner(stop chan struct{}) {
	spinner := []rune{'-', '\\', '|', '/'}
	index := 0

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				fmt.Printf("\r%c", spinner[index])
				index = (index + 1) % len(spinner)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}
