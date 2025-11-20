package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	api "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type SubscriptionListCommand struct {
	config.Config
	Output string `short:"o" help:"Output format: table, json, yaml" default:"table"`
}

type SubscriptionCommand struct {
	List SubscriptionListCommand `cmd:"list" help:"List subscriptions"`
}

func (s *SubscriptionListCommand) Run() error {
	s.Config.ApplyDefaults()

	client, err := util.GetAuthenticatedClient(s.Config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx := context.Background()

	// Fetch all environments first to create a lookup map
	envResponse, err := client.GetEnvironments(ctx)
	if err != nil {
		return fmt.Errorf("failed to get environments: %w", err)
	}

	envMap := make(map[string]api.EnvironmentResponse)
	if envOkResp, ok := envResponse.(*api.GetEnvironmentsOKApplicationJSON); ok {
		environments := []api.EnvironmentResponse(*envOkResp)
		for _, env := range environments {
			envMap[env.ID.String()] = env
		}
	}

	response, err := client.GetSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	// Type assert to the success response
	okResp, ok := response.(*api.GetSubscriptionsOKApplicationJSON)
	if !ok {
		return fmt.Errorf("unexpected response type")
	}

	subscriptions := []api.SubscriptionResponse(*okResp)
	if len(subscriptions) == 0 {
		fmt.Println("No subscriptions found.")
		return nil
	}

	// Build table data
	type subOutput struct {
		ID           string `json:"id" yaml:"id"`
		Status       string `json:"status" yaml:"status"`
		Plan         string `json:"plan,omitempty" yaml:"plan,omitempty"`
		PeriodStart  string `json:"period_start,omitempty" yaml:"period_start,omitempty"`
		PeriodEnd    string `json:"period_end,omitempty" yaml:"period_end,omitempty"`
		Environments string `json:"environments" yaml:"environments"`
	}

	structured := make([]subOutput, len(subscriptions))
	tableData := make([]map[string]any, len(subscriptions))

	for i, sub := range subscriptions {
		plan := ""
		if sub.PlanName.Set {
			plan = sub.PlanName.Value
		}

		periodStart := ""
		periodEnd := ""
		if sub.CurrentPeriodStart.Set {
			startTime := time.Unix(int64(sub.CurrentPeriodStart.Value), 0)
			periodStart = startTime.Format("2006-01-02")
		}
		if sub.CurrentPeriodEnd.Set {
			endTime := time.Unix(int64(sub.CurrentPeriodEnd.Value), 0)
			periodEnd = endTime.Format("2006-01-02")
		}

		// Build environment names list
		envNames := []string{}
		for _, envId := range sub.EnvironmentIds {
			if env, ok := envMap[envId.String()]; ok {
				envNames = append(envNames, env.Name)
			}
		}
		environments := fmt.Sprintf("%d", len(sub.EnvironmentIds))
		if len(envNames) > 0 {
			environments = fmt.Sprintf("%d (%s)", len(envNames), envNames[0])
		}

		structured[i] = subOutput{
			ID:           sub.ID.String(),
			Status:       string(sub.Status),
			Plan:         plan,
			PeriodStart:  periodStart,
			PeriodEnd:    periodEnd,
			Environments: environments,
		}
		tableData[i] = map[string]any{
			"ID":           sub.ID.String(),
			"Status":       sub.Status,
			"Plan":         plan,
			"Period Start": periodStart,
			"Period End":   periodEnd,
			"Environments": environments,
		}
	}

	headers := []string{"ID", "Status", "Plan", "Period Start", "Period End", "Environments"}
	return util.FormatOutput(s.Output, structured, headers, tableData)
}
