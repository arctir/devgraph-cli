package auth

import (
	"github.com/arctir/devgraph-cli/pkg/config"
)

// Re-export types and functions from config package for backward compatibility
type Credentials = config.Credentials

var LoadCredentials = config.LoadCredentials
var SaveCredentials = config.SaveCredentials
