package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"os/signal"
	"syscall"

	"github.com/chzyer/readline"
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
	fmt.Println("Coder v0.0.2")
	fmt.Println("\nType \"/help\" for usage instructions.")

	client := NewOpenAIClient(apiKey)

	validModels := map[string]Model{
		"gpt-4o-mini": {
			Name:            "gpt-4o-mini",
			InputTokenCost:  0.15 / 1000000,
			OutputTokenCost: 0.60 / 1000000,
		},
		"gpt-4o": {
			Name:            "gpt-4o",
			InputTokenCost:  2.50 / 1000000,
			OutputTokenCost: 10.0 / 1000000,
		},
		"o3-mini": {
			Name:            "o3-mini",
			InputTokenCost:  1.10 / 1000000,
			OutputTokenCost: 4.40 / 1000000,
		},
	}

	selectedModel := validModels["gpt-4o-mini"]
	cs := &ConversationState{}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var completer = readline.NewPrefixCompleter(
		readline.PcItem("/include",
			readline.PcItemDynamic(listFiles("./")),
		),
		readline.PcItem("/write",
			readline.PcItemDynamic(listFiles("./")),
			readline.PcItem("-code",
				readline.PcItemDynamic(listFiles("./")),
			),
		),
		readline.PcItem("/model",
			readline.PcItem("info"),
			readline.PcItem("switch",
				readline.PcItemDynamic(func(line string) []string {
					var models []string
					for modelName := range validModels {
						models = append(models, modelName)
					}
					return models
				}),
			),
		),
		readline.PcItem("/new"),
		readline.PcItem("/session"),
		readline.PcItem("/help"),
		readline.PcItem("/exit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		HistoryFile:  "/tmp/coder_history",
		AutoComplete: completer,
	})
	if err != nil {
		log.Fatalf("Error creating readline: %v", err)
	}
	defer rl.Close()

	go func() {
		<-signalChan
		rl.Close()
		os.Exit(0)
	}()

	for {
		responseRequest := ResponseRequest{
			Model: selectedModel.Name,
			/* MaxTokens: func() *int {
				i := 100
				return &i
			}(), */
			PreviousID: cs.PreviousID,
		}

		if cs.Elapsed > time.Duration(0) {
			statusStr := fmt.Sprintf("[%v; %v ->; -> %v]", cs.Elapsed.Round(time.Millisecond), cs.InputTokens, cs.OutputTokens)
			fmt.Printf("\n%s\n", statusStr)
		} else {
			fmt.Print("\n")
		}

		input, err := rl.Readline()
		if err != nil {
			if err = readline.ErrInterrupt; err != nil {
				os.Exit(0)
			}
			log.Fatalf("Error reading input: %v", err)
		}

		input = strings.TrimSpace(input)

		if strings.HasPrefix(strings.ToLower(input), "/") {
			if handleSlashCommand(&input, cs, &selectedModel, validModels) {
				cs.Elapsed = time.Duration(0)
				continue
			}
		}

		responseRequest.Input = input

		start := time.Now()

		stopSpinner := make(chan struct{})
		startSpinner(stopSpinner)

		response, err := client.CreateResponse(responseRequest)

		close(stopSpinner)
		fmt.Print("\r")

		if err != nil {
			log.Fatalf("Error creating response: %v", err)
		}

		// fmt.Print("\033[H\033[2J")

		var responseText string
		for _, output := range response.Output {
			if output.Role == "assistant" {
				if len(output.Content) > 0 {
					responseText = output.Content[0].Text
					break
				}
			}
		}

		if responseText != "" {
			fmt.Println(responseText)
			language, code := extractCodeBlock(responseText)
			if language != "" && code != "" {
				lines := len(strings.Split(code, "\n"))
				fmt.Printf("\n\u2728 Code block detected: Language=%s, Lines=%d. Use /write -code <file-path> to write to file.\n", language, lines)
				// fmt.Printf("\n%s\n", code)
			}
			cs.UpdateState(response, responseText, time.Since(start))
		} else {
			fmt.Println("Error: no \"role\": \"assistant\" response object found.")
		}
	}

	// do cleanup if necessary
}
