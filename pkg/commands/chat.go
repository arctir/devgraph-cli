package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/arctir/devgraph-cli/pkg/util"
	"github.com/charmbracelet/glamour"
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
var yellow = color.New(color.FgYellow).SprintFunc()
var magenta = color.New(color.FgMagenta).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var gray = color.New(color.FgHiBlack).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()

const smallHeader = "devgraph.ai"

const largeHeader = "      dP                                                       dP      \n" +
	"      88                                                       88      \n" +
	".d888b88 .d8888b. dP   .dP .d8888b. 88d888b. .d8888b. 88d888b. 88d888b.\n" +
	"88'  `88 88ooood8 88   d8' 88'  `88 88'  `88 88'  `88 88'  `88 88'  `88\n" +
	"88.  .88 88.  ... 88 .88'  88.  .88 88       88.  .88 88.  .88 88    88\n" +
	"`88888P8 `88888P' 8888P'   `8888P88 dP       `88888P8 88Y888P' dP    dP\n" +
	"                                .88                   88               \n" +
	"                            d8888P                    dP               \n"

// Enhanced response formatter with markdown and syntax highlighting
func formatResponse(text string) {
	// First try to render as markdown
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0), // 0 = no word wrap limit, use full terminal width
	)
	
	if err == nil {
		if formatted, err := renderer.Render(text); err == nil {
			typeWriter(formatted)
			return
		}
	}
	
	// Fallback to manual formatting if glamour fails
	typeWriter(formatTextWithSyntaxHighlighting(text))
}

// Manual text formatting with syntax highlighting for code blocks
func formatTextWithSyntaxHighlighting(text string) string {
	// Regex to find code blocks
	codeBlockRegex := regexp.MustCompile("```(\\w+)?\\s*\\n([\\s\\S]*?)\\n```")
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	
	// Handle code blocks
	result := codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		if len(parts) >= 3 {
			language := parts[1]
			code := parts[2]
			
			// Try syntax highlighting
			var highlighted strings.Builder
			if err := quick.Highlight(&highlighted, code, language, "terminal", "monokai"); err == nil {
				return "\n" + gray("┌─ Code ("+language+")") + "\n" + highlighted.String() + gray("└─") + "\n"
			}
			
			// Fallback to simple formatting
			return "\n" + gray("┌─ Code") + "\n" + code + "\n" + gray("└─") + "\n"
		}
		return match
	})
	
	// Handle inline code
	result = inlineCodeRegex.ReplaceAllStringFunc(result, func(match string) string {
		code := strings.Trim(match, "`")
		return yellow("`" + code + "`")
	})
	
	return result
}

// Enhanced typewriter with word-by-word printing
func typeWriter(text string) {
	// Show typing indicator briefly
	fmt.Print(gray("● "))
	time.Sleep(200 * time.Millisecond)
	fmt.Print("\r                    \r") // Clear the line completely
	
	// Clean up the text to avoid extra trailing newlines
	text = strings.TrimRight(text, "\n")
	
	// Type out the text word by word for more natural feel
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			fmt.Println()
			continue
		}
		
		// Different delays for different content types
		delay := 40 * time.Millisecond  // Default delay between words (readable but not too slow)
		if strings.Contains(line, "```") || strings.HasPrefix(strings.TrimSpace(line), "┌─") || 
		   strings.HasPrefix(strings.TrimSpace(line), "└─") {
			// Faster for code block delimiters
			delay = 15 * time.Millisecond
		} else if strings.HasPrefix(strings.TrimSpace(line), "    ") || 
				  strings.HasPrefix(strings.TrimSpace(line), "\t") {
			// Faster for code content (indented lines)
			delay = 25 * time.Millisecond
		}
		
		// Split line into words while preserving spacing
		words := splitLineIntoWords(line)
		
		// If no words (empty line after trimming), treat as empty line
		if len(words) == 0 {
			// This shouldn't happen since we check for empty lines above,
			// but handle it gracefully just in case
			continue
		}
		
		for j, word := range words {
			fmt.Print(word)
			
			// Add space and delay between words (but not after the last word of a line)
			if j < len(words)-1 {
				fmt.Print(" ")
				time.Sleep(delay)
			}
		}
		
		if i < len(lines)-1 {
			fmt.Println()
		}
	}
	fmt.Println() // Single newline at the end of content
	fmt.Println() // Extra newline before next prompt
}

// splitLineIntoWords splits a line into words while preserving leading spaces but not trailing
func splitLineIntoWords(line string) []string {
	// For lines that are empty or only whitespace, return empty slice
	// The main loop will handle these as empty lines with just newlines
	if strings.TrimSpace(line) == "" {
		return []string{}
	}
	
	// Trim trailing spaces to avoid printing spaces until newline
	line = strings.TrimRight(line, " \t")
	
	// Simple approach: split by spaces and preserve leading spaces
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return []string{}
	}
	
	// Find leading spaces to preserve indentation
	leadingSpaces := ""
	for _, char := range line {
		if char == ' ' || char == '\t' {
			leadingSpaces += string(char)
		} else {
			break
		}
	}
	
	// Add leading spaces to first word if any
	if leadingSpaces != "" && len(fields) > 0 {
		fields[0] = leadingSpaces + fields[0]
	}
	
	return fields
}

func userPrompt(username string) {
	fmt.Printf("%s %s\n", blue("❯"), bold(green(username)))
}

func devgraphPrompt() {
	fmt.Printf("\n%s %s\n", magenta("◉"), bold(cyan("devgraph")))
}

func showThinkingIndicator() {
	indicators := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for i := 0; i < 10; i++ {
		fmt.Printf("\r%s %s", gray(indicators[i%len(indicators)]), gray("thinking..."))
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print("\r                    \r") // Clear the line
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
	// Validate that a model is configured
	if c.Model == "" {
		return fmt.Errorf("no model configured. Run 'devgraph setup' to configure a model, or use the -m flag to specify one")
	}

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
	fmt.Printf("\n%s Welcome to %s! \n", cyan("✨"), bold(cyan("devgraph")))
	fmt.Printf("%s Type %s to quit, %s to change model, or %s for commands.\n\n", 
		gray("   "), yellow("'/exit'"), yellow("'/model'"), yellow("'/help'"))

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

		// Handle slash commands
		if strings.HasPrefix(input, "/") {
			if err := c.handleSlashCommand(input); err != nil {
				fmt.Printf("%s Error: %s\n\n", red("⚠️"), err)
			}
			continue
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		// Show thinking indicator while making API call
		go showThinkingIndicator()
		
		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:     c.Model,
			Messages:  messages,
			MaxTokens: c.MaxTokens,
		})
		
		if err != nil {
			devgraphPrompt()
			// Extract just the relevant error message without verbose context
			errorMsg := extractErrorMessage(err.Error())
			fmt.Printf("%s %s\n\n", red("✖"), red(fmt.Sprintf("Error: %s", errorMsg)))
			continue
		}

		if len(resp.Choices) == 0 {
			devgraphPrompt()
			fmt.Printf("%s %s\n\n", yellow("⚠"), yellow("No response generated"))
			continue
		}

		for _, choice := range resp.Choices {
			aiResponse := choice.Message.Content
			devgraphPrompt()
			
			// Use enhanced formatting for the response
			formatResponse(aiResponse)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: aiResponse,
			})
		}
	}

	return nil
}

// handleSlashCommand processes slash commands during chat
func (c *Chat) handleSlashCommand(input string) error {
	command := strings.ToLower(strings.TrimSpace(input))
	
	switch command {
	case "/exit":
		fmt.Printf("%s %s\n", cyan("👋"), "Goodbye!")
		os.Exit(0)
		return nil
		
	case "/help":
		fmt.Printf("\n%s %s\n", blue("ℹ"), bold("Available commands:"))
		fmt.Printf("  %s   - Exit the chat\n", yellow("/exit"))
		fmt.Printf("  %s  - Change the current model\n", yellow("/model"))
		fmt.Printf("  %s   - Show this help message\n", yellow("/help"))
		fmt.Println()
		return nil
		
	case "/model":
		return c.changeModel()
		
	default:
		return fmt.Errorf("unknown command: %s. Type '/help' for available commands", input)
	}
}

// changeModel allows the user to select a different model during chat
func (c *Chat) changeModel() error {
	fmt.Printf("\n%s %s\n\n", magenta("🤖"), bold("Available models:"))
	
	// Get available models from API
	models, err := util.GetModels(c.Config)
	if err != nil {
		return fmt.Errorf("failed to fetch available models: %w", err)
	}
	
	if models == nil || len(*models) == 0 {
		return fmt.Errorf("no models are available from the API")
	}
	
	// Display models with current one highlighted
	for i, model := range *models {
		if model.Name == c.Model {
			fmt.Printf("  %s %s %s\n", green("✓"), blue(fmt.Sprintf("%d.", i+1)), 
				boldCyan(model.Name+" "+gray("(current)")))
		} else {
			fmt.Printf("    %s %s\n", blue(fmt.Sprintf("%d.", i+1)), model.Name)
		}
	}
	
	// Get user selection
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\n%s ", cyan("❯"))
		fmt.Print("Select a model (enter number, or 'c' to cancel): ")
		if !scanner.Scan() {
			return fmt.Errorf("failed to read input")
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "c" || input == "cancel" {
			fmt.Printf("%s %s\n", yellow("⚠"), "Model change cancelled.")
			return nil
		}
		
		// Try to parse as number
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(*models) {
			fmt.Printf("%s Invalid choice. Please enter a number between %s and %s, or %s to cancel.\n", 
				red("✖"), blue("1"), blue(fmt.Sprintf("%d", len(*models))), yellow("'c'"))
			continue
		}
		
		selectedModel := (*models)[choice-1]
		if selectedModel.Name == c.Model {
			fmt.Printf("%s Already using model: %s\n", blue("ℹ"), cyan(selectedModel.Name))
			return nil
		}
		
		// Update the current model
		c.Model = selectedModel.Name
		fmt.Printf("%s Switched to model: %s\n\n", green("✅"), bold(green(selectedModel.Name)))
		return nil
	}
}

