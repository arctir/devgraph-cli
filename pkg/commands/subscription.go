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
	headers := []string{"ID", "Status", "Plan", "Period Start", "Period End", "Environments"}
	data := make([]map[string]interface{}, len(subscriptions))

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

		data[i] = map[string]interface{}{
			"ID":           sub.ID.String(),
			"Status":       sub.Status,
			"Plan":         plan,
			"Period Start": periodStart,
			"Period End":   periodEnd,
			"Environments": environments,
		}
	}

	displayEntityTable(data, headers)

	return nil
}
