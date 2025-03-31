package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

func (cs *ConversationState) UpdateState(response ResponseBody, responseText string, elapsed time.Duration) {
	cs.PreviousID = response.ID
	cs.PreviousResponse = responseText
	cs.InputTokens = response.Usage.InputTokens
	cs.OutputTokens = response.Usage.OutputTokens
	cs.TotalInputTokens += response.Usage.InputTokens
	cs.TotalOutputTokens += response.Usage.OutputTokens
	cs.Elapsed = elapsed
}

func (cs *ConversationState) ClearState() {
	cs.PreviousID = ""
	cs.PreviousResponse = ""
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
		return true

	case "/help":
		fmt.Println("Usage:")
		fmt.Println("- Type your message and press Enter to get a response.")
		fmt.Println("- \"/new\": start a new conversation.")
		fmt.Println("- \"/include <file-path> <prompt>\": include a file in context.")
		fmt.Println("- \"/session\": view current-conversation token consumption.")
		fmt.Println("- \"/model info\": view current model info.")
		fmt.Println("- \"/model switch <model-name>\": change current model.")
		fmt.Println("- \"/write [-code] <file-path>\": write the previous response to a file. Use -code to write the first code block only.")
		fmt.Println("- \"/exit\": exit the program.")
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

	case "/write":
		if len(parts) < 2 {
			fmt.Println("Usage: /write [-code] <file-path>")
			return true
		}

		var filePath string
		writeCodeBlock := false

		var language string
		var code string

		if len(parts) > 2 && parts[1] == "-code" {
			writeCodeBlock = true
			filePath = parts[2]
		} else if len(parts) == 2 {
			filePath = parts[1]
		} else {
			fmt.Println("Usage: /write [-code] <file-path>")
			return true
		}

		if cs.PreviousID == "" {
			fmt.Println("No previous response to write.")
			return true
		}

		var contentToWrite string
		if writeCodeBlock {
			language, code = extractCodeBlock(cs.PreviousResponse)
			if language == "" || code == "" {
				fmt.Println("No code block found in the previous response.")
				return true
			}
			contentToWrite = code
		} else {
			contentToWrite = cs.PreviousResponse
		}

		err := os.WriteFile(filePath, []byte(contentToWrite), 0644)
		if err != nil {
			fmt.Printf("Error writing to file %s: %v\n", filePath, err)
			return true
		}

		if writeCodeBlock {
			fmt.Printf("Code block written to %s\n\n", filePath)
			fmt.Printf("Language: %s\n", language)
			fmt.Printf("Lines: %d\n", len(strings.Split(code, "\n")))
		} else {
			fmt.Printf("\nResponse written to %s\n", filePath)
		}

		return true

	default:
		fmt.Printf("Unknown command: %s\n", parts[0])
		return true
	}
}

func extractCodeBlock(responseText string) (string, string) {
	parts := strings.Split(responseText, "```")
	if len(parts) < 3 {
		return "", ""
	}

	language := strings.Split(parts[1], "\n")[0]
	lines := strings.Split(parts[1], "\n")

	code := strings.TrimSpace(strings.Join(lines[1:], "\n"))

	return language, code
}

func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := os.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
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
