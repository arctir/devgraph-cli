package commands

import (
	"context"
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
	"github.com/google/uuid"
)

// SuggestionCommand manages chat suggestions
type SuggestionCommand struct {
	List   SuggestionListCommand   `cmd:"list" help:"List chat suggestions"`
	Create SuggestionCreateCommand `cmd:"create" help:"Create a chat suggestion"`
	Delete SuggestionDeleteCommand `cmd:"delete" help:"Delete a chat suggestion"`
}

type SuggestionListCommand struct {
	EnvWrapperCommand
}

type SuggestionCreateCommand struct {
	EnvWrapperCommand
	Title  string `arg:"" required:"" help:"Suggestion title"`
	Label  string `short:"l" required:"" help:"Button label for the suggestion"`
	Action string `short:"a" required:"" help:"Action/prompt to execute when selected"`
}

type SuggestionDeleteCommand struct {
	EnvWrapperCommand
	ID string `arg:"" required:"" help:"Suggestion ID to delete"`
}

func (s *SuggestionListCommand) Run() error {
	client, err := util.GetAuthenticatedClient(s.Config)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	params := api.ListChatSuggestionsParams{}
	resp, err := client.ListChatSuggestions(ctx, params)
	if err != nil {
		return err
	}

	switch r := resp.(type) {
	case *api.ListChatSuggestionsOKApplicationJSON:
		suggestions := []api.ChatSuggestionResponse(*r)
		if len(suggestions) == 0 {
			fmt.Println("No chat suggestions found.")
			return nil
		}
		displaySuggestions(&suggestions)
	default:
		return fmt.Errorf("failed to list chat suggestions")
	}
	return nil
}

func (s *SuggestionCreateCommand) Run() error {
	client, err := util.GetAuthenticatedClient(s.Config)
	if err != nil {
		return err
	}

	ctx := context.TODO()

	suggestion := api.ChatSuggestionCreate{
		Title:  s.Title,
		Label:  s.Label,
		Action: s.Action,
	}

	resp, err := client.CreateChatSuggestion(ctx, &suggestion)
	if err != nil {
		return err
	}

	switch r := resp.(type) {
	case *api.ChatSuggestionResponse:
		fmt.Printf("✅ Created chat suggestion: %s\n", r.ID)
	default:
		return fmt.Errorf("failed to create chat suggestion")
	}
	return nil
}

func (s *SuggestionDeleteCommand) Run() error {
	client, err := util.GetAuthenticatedClient(s.Config)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	suggestionUUID, err := uuid.Parse(s.ID)
	if err != nil {
		return fmt.Errorf("invalid suggestion UUID: %w", err)
	}

	params := api.DeleteChatSuggestionParams{
		SuggestionID: suggestionUUID,
	}
	resp, err := client.DeleteChatSuggestion(ctx, params)
	if err != nil {
		return err
	}

	switch resp.(type) {
	case *api.DeleteChatSuggestionNoContent:
		fmt.Printf("✅ Deleted chat suggestion: %s\n", s.ID)
	default:
		return fmt.Errorf("failed to delete chat suggestion")
	}
	return nil
}

func displaySuggestions(suggestions *[]api.ChatSuggestionResponse) {
	if suggestions == nil || len(*suggestions) == 0 {
		fmt.Println("No chat suggestions found.")
		return
	}

	headers := []string{"ID", "Title", "Label", "Action"}
	data := make([]map[string]any, len(*suggestions))
	for i, s := range *suggestions {
		data[i] = map[string]any{
			"ID":     s.ID,
			"Title":  s.Title,
			"Label":  s.Label,
			"Action": s.Action,
		}
	}
	util.DisplaySimpleTable(data, headers)
}
