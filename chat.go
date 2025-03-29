package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
)

type Chat struct {
	Config
}

var cyan = color.New(color.FgCyan).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

const header = "      dP                                                       dP      \n" +
	"      88                                                       88      \n" +
	".d888b88 .d8888b. dP   .dP .d8888b. 88d888b. .d8888b. 88d888b. 88d888b.\n" +
	"88'  `88 88ooood8 88   d8' 88'  `88 88'  `88 88'  `88 88'  `88 88'  `88\n" +
	"88.  .88 88.  ... 88 .88'  88.  .88 88       88.  .88 88.  .88 88    88\n" +
	"`88888P8 `88888P' 8888P'   `8888P88 dP       `88888P8 88Y888P' dP    dP\n" +
	"                                .88                   88               \n" +
	"                            d8888P                    dP               \n"

func (c *Chat) Run() error {

	creds, err := loadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %v", err)
	}

	clientConfig := openai.DefaultConfig(creds.IDToken)
	clientConfig.BaseURL = "http://localhost:8000/devgraph/api" // Set your API base URL here
	clientConfig.OrgID = "1234"
	client := openai.NewClientWithConfig(clientConfig)
	ctx := context.Background()

	var messages []openai.ChatCompletionMessage
	username, ok := (*creds.Claims)["preferred_username"].(string)
	if !ok {
		username = "you"
	}

	fmt.Print(cyan(header))
	fmt.Print("\n\nWelcome to devgraph! Type 'exit' to quit.\n\n")

	for {
		fmt.Printf(green("@%s: "), username)
		var input string
		fmt.Scanln(&input)

		if strings.ToLower(input) == "exit" {
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
			return fmt.Errorf("failed to get response: %v", err)
		}

		if len(resp.Choices) == 0 {
			fmt.Println("AI: No response generated")
			continue
		}

		aiResponse := resp.Choices[0].Message.Content
		fmt.Printf("@devgraph: %s\n", aiResponse)

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: aiResponse,
		})
	}

	return nil
}
