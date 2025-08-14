package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/term"
)

type Chat struct {
	config.Config
}

var cyan = color.New(color.FgCyan).SprintFunc()
var boldCyan = color.New(color.FgCyan, color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

const smallHeader = "devgraph.ai"

const largeHeader = "      dP                                                       dP      \n" +
	"      88                                                       88      \n" +
	".d888b88 .d8888b. dP   .dP .d8888b. 88d888b. .d8888b. 88d888b. 88d888b.\n" +
	"88'  `88 88ooood8 88   d8' 88'  `88 88'  `88 88'  `88 88'  `88 88'  `88\n" +
	"88.  .88 88.  ... 88 .88'  88.  .88 88       88.  .88 88.  .88 88    88\n" +
	"`88888P8 `88888P' 8888P'   `8888P88 dP       `88888P8 88Y888P' dP    dP\n" +
	"                                .88                   88               \n" +
	"                            d8888P                    dP               \n"

func typeWriter(text string) {
	for _, char := range text {
		fmt.Print(string(char))
		// Flush the output to ensure immediate display
		time.Sleep(10 * time.Millisecond) // Adjust delay as needed
	}
	fmt.Println()
	fmt.Println()
}

func userPrompt(username string) {
	fmt.Printf(green("@%s:\n"), username)
}

func devgraphPrompt() {
	fmt.Print(cyan("\n@%devgraph:\n"))
}

func (c *Chat) Run() error {
	authHttpClient, err := util.GetAuthenticatedHTTPClient(c.Config)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	clientConfig := openai.DefaultConfig("")
	clientConfig.BaseURL = c.ApiURL + "/api/v1/model"
	clientConfig.HTTPClient = authHttpClient
	client := openai.NewClientWithConfig(clientConfig)

	ctx := context.Background()

	username, err := util.GetUsername()
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}
	var messages []openai.ChatCompletionMessage

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && width > 75 {
		fmt.Print(boldCyan(largeHeader))
	} else {
		fmt.Print(boldCyan(smallHeader))
	}
	fmt.Print("\n\nWelcome to devgraph! Type '/exit' to quit.\n\n")

	for {
		userPrompt(username) // Prompt for user input
		var input string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input = scanner.Text()
			break
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}

		if input == "" {
			continue // Skip empty input
		}

		if strings.ToLower(input) == "/exit" {
			fmt.Println("Goodbye!")
			break
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:     "gpt-4o-mini", //c.Model,
			Messages:  messages,
			MaxTokens: c.MaxTokens,
		})
		if err != nil {
			devgraphPrompt()
			fmt.Print(red(fmt.Sprintf("Oops! It seems like there was an error while generating a response.\n%s\n", err.Error())))
			continue
		}

		if len(resp.Choices) == 0 {
			devgraphPrompt()
			typeWriter("No response generated")
			continue
		}

		for _, choice := range resp.Choices {
			aiResponse := choice.Message.Content
			devgraphPrompt()
			typeWriter(aiResponse)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: aiResponse,
			})
		}
	}

	return nil
}
