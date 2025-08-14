package admincommands

import (
	"fmt"
	"io"
	"net/http"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
)

type DevEnvironment struct {
	Create DevEnvironmentCreateCommand `cmd:"create" help:"Create a new ModelProvider resource."`
}

type DevEnvironmentCreateCommand struct {
	config.Config
}

func (e *DevEnvironmentCreateCommand) Run() error {
	client, err := util.GetAuthenticatedHTTPClient(e.Config)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/dev/environments", e.Config.ApiURL)

	// Create new request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned non-2xx status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
