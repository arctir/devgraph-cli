package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/oauth2"
	"golang.org/x/term"
)

type Chat struct {
	Config
}

var cyan = color.New(color.FgCyan).SprintFunc()
var boldCyan = color.New(color.FgCyan, color.Bold).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

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

	environment := os.Getenv("DEVGRAPH_ENVIRONMENT")
	if environment == "" {
		return fmt.Errorf("DEVGRAPH_ENVIRONMENT is not set")
	}

	creds, err := loadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %v", err)
	}

	endpoints, err := getWellKnownEndpoints(c.IssuerURL)
	if err != nil {
		return fmt.Errorf("failed to get well-known endpoints: %v", err)
	}
	oauth2Config := oauth2.Config{
		ClientID:    c.ClientID,
		RedirectURL: c.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  endpoints.AuthorizationEndpoint,
			TokenURL: endpoints.TokenEndpoint,
		},
		Scopes: []string{"openid", "profile", "email"},
	}

	var exp float64
	exp, _ = (*creds.Claims)["exp"].(float64)

	expTime := time.Unix(int64(exp), 0)
	if exp > 0 {
		expTime = expTime.Add(-30 * time.Second)
	}

	token := &oauth2.Token{
		AccessToken:  creds.IDToken,
		RefreshToken: creds.RefreshToken,
		TokenType:    "Bearer",
		Expiry:       expTime,
	}

	provider, err := oidc.NewProvider(context.Background(), c.IssuerURL)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %v", err)
	}

	tokenManager := NewOIDCTokenManager(
		oauth2Config,
		token,
		provider,
		environment,
	)
	httpClient := tokenManager.HTTPClient()

	clientConfig := openai.DefaultConfig(creds.IDToken)
	clientConfig.BaseURL = c.Config.ApiURL
	clientConfig.HTTPClient = httpClient // Use the HTTP client that auto-refreshes the token
	client := openai.NewClientWithConfig(clientConfig)
	ctx := context.Background()

	var messages []openai.ChatCompletionMessage
	username, ok := (*creds.Claims)["preferred_username"].(string)
	if !ok {
		username, ok = (*creds.Claims)["email"].(string)
		if !ok {
			username = "localuser"
		}
	}

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
			Model:     c.Model,
			Messages:  messages,
			MaxTokens: c.MaxTokens,
		})
		if err != nil {
			devgraphPrompt()
			typeWriter("Oops! It seems like there was an error while generating a response. " + err.Error() + "\n")
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
