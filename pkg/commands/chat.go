package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arctir/devgraph-cli/pkg/util"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/term"
)

type Chat struct {
	EnvWrapperCommand
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

// extractErrorMessage cleans up error messages to show only relevant details
func extractErrorMessage(fullError string) string {
	// Common patterns to extract meaningful error messages

	// Handle specific HTTP status codes first
	if strings.Contains(fullError, "401") {
		return "Authentication failed. Please check your credentials."
	}
	if strings.Contains(fullError, "403") {
		return "Access denied. Please check your permissions."
	}
	if strings.Contains(fullError, "404") {
		return "Model not found. Please check your model configuration."
	}
	if strings.Contains(fullError, "429") {
		return "Rate limit exceeded. Please try again later."
	}
	if strings.Contains(fullError, "500") {
		return "Server error. Please try again later."
	}

	// Handle OpenAI API errors - extract the actual error message
	if strings.Contains(fullError, "error,") {
		parts := strings.Split(fullError, "error,")
		if len(parts) > 1 {
			// Clean up the error message part
			msg := strings.TrimSpace(parts[1])
			// Remove trailing codes and extra info
			if idx := strings.Index(msg, " (status"); idx != -1 {
				msg = msg[:idx]
			}
			if idx := strings.Index(msg, ", status:"); idx != -1 {
				msg = msg[:idx]
			}
			return msg
		}
	}

	// Handle network errors
	if strings.Contains(fullError, "no such host") {
		return "Unable to connect to the server. Please check your network connection."
	}
	if strings.Contains(fullError, "connection refused") {
		return "Connection refused. Please check if the service is running."
	}
	if strings.Contains(fullError, "timeout") {
		return "Request timed out. Please try again."
	}

	// Handle model/configuration errors
	if strings.Contains(fullError, "model") && strings.Contains(fullError, "does not exist") {
		return "The specified model is not available. Please check your model configuration."
	}

	// If no specific pattern matches, return a cleaned version
	// Remove common prefixes and technical details
	msg := fullError
	prefixes := []string{
		"failed to create chat completion: ",
		"error creating completion: ",
		"API error: ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(msg, prefix) {
			msg = strings.TrimPrefix(msg, prefix)
			break
		}
	}

	// Limit length and clean up
	if len(msg) > 100 {
		msg = msg[:97] + "..."
	}

	return msg
}

func (*Chat) BeforeApply() error {
	return nil
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

	scanner := bufio.NewScanner(os.Stdin)
	for {
		userPrompt(username) // Prompt for user input
		var input string

		if !scanner.Scan() {
			// No more input available (EOF or error)
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
			}
			break
		}
		input = scanner.Text()

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
			Model:     c.Model,
			Messages:  messages,
			MaxTokens: c.MaxTokens,
		})
		if err != nil {
			devgraphPrompt()
			// Extract just the relevant error message without verbose context
			errorMsg := extractErrorMessage(err.Error())
			fmt.Print(red(fmt.Sprintf("Error: %s\n\n", errorMsg)))
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
