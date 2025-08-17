package commands

import (
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/util"
	devgraphv1 "github.com/arctir/go-devgraph/pkg/apis/devgraph/v1"
)

type EnvironmentListCommand struct {
	EnvWrapperCommand
}

type EnvironmentCommand struct {
	List EnvironmentListCommand `cmd:"list" help:"List all environments for Devgraph"`
}

func (e *EnvironmentListCommand) Run() error {
	envs, err := util.GetEnvironments(e.Config)
	if err != nil {
		return err
	}

	if envs == nil || len(*envs) == 0 {
		fmt.Println("No environments found.")
		return nil
	}

	displayEnvironments(envs)
	return nil
}

func displayEnvironments(envs *[]devgraphv1.EnvironmentResponse) {
	if envs == nil || len(*envs) == 0 {
		fmt.Println("No environments found.")
		return
	}

	headers := []string{"ID", "Name", "Slug"}
	data := make([]map[string]interface{}, len(*envs))
	for i, env := range *envs {
		data[i] = map[string]interface{}{
			"Name": env.Name,
			"ID":   env.Id,
			"Slug": env.Slug,
		}
	}
	util.DisplaySimpleTable(data, headers)
}
