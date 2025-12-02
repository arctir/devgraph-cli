package commands

import (
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
)

type EnvWrapperCommand struct {
	config.Config
}

func (e *EnvWrapperCommand) BeforeApply() error {
	// Apply defaults from environment config map
	e.Config.ApplyDefaults()

	// Skip environment check if not authenticated
	// This allows commands to proceed and let main.go handle first-time setup
	if !util.IsAuthenticated() {
		return nil
	}

	ok, err := util.CheckEnvironment(&e.Config)
	if !ok {
		return err
	}
	return nil
}
