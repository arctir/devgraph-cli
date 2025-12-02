package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	"github.com/fatih/color"
	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

type ConfigCommand struct {
	CurrentContext CurrentContextCommand `kong:"cmd,name='current-context',help='Display the current context'"`
	CurrentEnv     CurrentEnvCommand     `kong:"cmd,name='current-env',help='Display the current environment ID'"`
	DeleteCluster  DeleteClusterCommand  `kong:"cmd,name='delete-cluster',help='Delete a cluster'"`
	DeleteContext  DeleteContextCommand  `kong:"cmd,name='delete-context',help='Delete a context'"`
	DeleteUser     DeleteUserCommand     `kong:"cmd,name='delete-user',help='Delete a user'"`
	GetClusters    GetClustersCommand    `kong:"cmd,name='get-clusters',help='List all clusters'"`
	GetContexts    GetContextsCommand    `kong:"cmd,aliases='get-contexts',help='List all contexts'"`
	GetUsers       GetUsersCommand       `kong:"cmd,name='get-users',help='List all users'"`
	SetCluster     SetClusterCommand     `kong:"cmd,name='set-cluster',help='Create or modify a cluster'"`
	SetContext     SetContextCommand     `kong:"cmd,name='set-context',help='Create or modify a context'"`
	SetCredentials SetCredentialsCommand `kong:"cmd,name='set-credentials',help='Set user credentials'"`
	UseContext     UseContextCommand     `kong:"cmd,name='use-context',help='Set the current context'"`
}

// GetContextsCommand lists all available contexts
type GetContextsCommand struct {
	Output string `flag:"output,o" default:"table" help:"Output format: table, json, yaml, name."`
}

// CurrentContextCommand displays the current context
type CurrentContextCommand struct{}

// CurrentEnvCommand displays the current environment ID
type CurrentEnvCommand struct{}

// UseContextCommand sets the current context
type UseContextCommand struct {
	Context string `arg:"" required:"" help:"Name of the context to use."`
}

// SetContextCommand creates or updates a context
type SetContextCommand struct {
	config.Config
	Context    string `arg:"" required:"" help:"Name of the context."`
	Cluster    string `flag:"cluster" help:"Cluster for this context."`
	User       string `flag:"user" help:"User for this context."`
	ContextEnv string `flag:"env" help:"Environment name, slug, or UUID for this context."`
}

// DeleteContextCommand deletes a context
type DeleteContextCommand struct {
	Context string `arg:"" required:"" help:"Name of the context to delete."`
}

// SetClusterCommand creates or updates a cluster
type SetClusterCommand struct {
	Cluster   string `arg:"" required:"" help:"Name of the cluster."`
	Server    string `flag:"server" help:"API server URL."`
	IssuerURL string `flag:"issuer-url" help:"OIDC issuer URL."`
	ClientID  string `flag:"client-id" help:"OAuth client ID."`
}

// SetCredentialsCommand sets user credentials
type SetCredentialsCommand struct {
	User         string `arg:"" required:"" help:"Name of the user."`
	AccessToken  string `flag:"access-token" help:"Access token."`
	RefreshToken string `flag:"refresh-token" help:"Refresh token."`
	IDToken      string `flag:"id-token" help:"ID token."`
}

func (g *GetContextsCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(userConfig.Contexts) == 0 {
		fmt.Println("No contexts found.")
		return nil
	}

	// Get sorted context names
	names := make([]string, 0, len(userConfig.Contexts))
	for name := range userConfig.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	if g.Output == "name" {
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	}

	// Build context data
	type contextOutput struct {
		Current     bool   `json:"current" yaml:"current"`
		Name        string `json:"name" yaml:"name"`
		Cluster     string `json:"cluster" yaml:"cluster"`
		User        string `json:"user" yaml:"user"`
		Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
	}

	contexts := make([]contextOutput, 0, len(userConfig.Contexts))
	for _, name := range names {
		ctx := userConfig.Contexts[name]

		// Try to get email from user's claims
		userDisplay := ctx.User
		if user, ok := userConfig.Users[ctx.User]; ok && user.Claims != nil {
			if email, ok := (*user.Claims)["email"].(string); ok && email != "" {
				userDisplay = email
			}
		}

		contexts = append(contexts, contextOutput{
			Current:     name == userConfig.CurrentContext,
			Name:        name,
			Cluster:     ctx.Cluster,
			User:        userDisplay,
			Environment: ctx.Environment,
		})
	}

	switch g.Output {
	case "json":
		output, err := json.MarshalIndent(contexts, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
	case "yaml":
		output, err := yaml.Marshal(contexts)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Print(string(output))
	default:
		// Table output
		headers := []string{"Current", "Name", "Cluster", "User", "Environment"}
		data := make([]map[string]interface{}, 0, len(contexts))
		for _, ctx := range contexts {
			current := ""
			if ctx.Current {
				current = "*"
			}
			data = append(data, map[string]interface{}{
				"Current":     current,
				"Name":        ctx.Name,
				"Cluster":     ctx.Cluster,
				"User":        ctx.User,
				"Environment": ctx.Environment,
			})
		}
		displayEntityTable(data, headers)
	}

	return nil
}

func (c *CurrentContextCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if userConfig.CurrentContext == "" {
		return fmt.Errorf("no current context set")
	}

	fmt.Println(userConfig.CurrentContext)
	return nil
}

func (c *CurrentEnvCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if userConfig.CurrentContext == "" {
		return fmt.Errorf("no current context set")
	}

	context, ok := userConfig.Contexts[userConfig.CurrentContext]
	if !ok {
		return fmt.Errorf("context '%s' not found", userConfig.CurrentContext)
	}

	if context.Environment == "" {
		return fmt.Errorf("no environment set for context '%s'", userConfig.CurrentContext)
	}

	fmt.Println(context.Environment)
	return nil
}

func (u *UseContextCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := userConfig.UseContext(u.Context); err != nil {
		return err
	}

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	green := color.New(color.FgGreen)
	green.Printf("Switched to context \"%s\".\n", u.Context)
	return nil
}

func (s *SetContextCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get existing context or create new one
	existingCtx, exists := userConfig.Contexts[s.Context]

	// Determine values to use
	cluster := s.Cluster
	user := s.User
	environment := s.ContextEnv

	if exists {
		// If modifying existing context, preserve values that weren't specified
		if cluster == "" {
			cluster = existingCtx.Cluster
		}
		if user == "" {
			user = existingCtx.User
		}
		if environment == "" {
			environment = existingCtx.Environment
		}
	} else {
		// For new contexts, all values must be provided
		if cluster == "" || user == "" {
			return fmt.Errorf("must specify --cluster and --user when creating a new context")
		}
	}

	// Resolve environment name/slug to UUID if provided
	if environment != "" && s.ContextEnv != "" {
		resolvedEnv, err := util.ResolveEnvironmentUUID(s.Config, s.ContextEnv)
		if err != nil {
			return fmt.Errorf("failed to resolve environment '%s': %w", s.ContextEnv, err)
		}
		environment = resolvedEnv
	}

	userConfig.SetContext(s.Context, cluster, user, environment)

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if exists {
		fmt.Printf("✅ Context '%s' modified.\n", s.Context)
	} else {
		fmt.Printf("✅ Context '%s' created.\n", s.Context)
	}
	return nil
}

func (d *DeleteContextCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := userConfig.DeleteContext(d.Context); err != nil {
		return err
	}

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Deleted context '%s'.\n", d.Context)
	return nil
}

func (s *SetClusterCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get existing cluster or create new one
	existingCluster, exists := userConfig.Clusters[s.Cluster]

	// Determine values to use
	server := s.Server
	issuerURL := s.IssuerURL
	clientID := s.ClientID

	if exists {
		// If modifying existing cluster, preserve values that weren't specified
		if server == "" {
			server = existingCluster.Server
		}
		if issuerURL == "" {
			issuerURL = existingCluster.IssuerURL
		}
		if clientID == "" {
			clientID = existingCluster.ClientID
		}
	} else {
		// For new clusters, server is required
		if server == "" {
			return fmt.Errorf("must specify --server when creating a new cluster")
		}
		// Set defaults for optional fields
		if issuerURL == "" {
			issuerURL = "https://primary-ghoul-65.clerk.accounts.dev"
		}
		if clientID == "" {
			clientID = "I97zD0IQmSFr5pql"
		}
	}

	userConfig.SetCluster(s.Cluster, server, issuerURL, clientID)

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if exists {
		fmt.Printf("✅ Cluster '%s' modified.\n", s.Cluster)
	} else {
		fmt.Printf("✅ Cluster '%s' created.\n", s.Cluster)
	}
	return nil
}

func (s *SetCredentialsCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get existing user or create new one
	existingUser, exists := userConfig.Users[s.User]

	// Determine values to use
	accessToken := s.AccessToken
	refreshToken := s.RefreshToken
	idToken := s.IDToken

	if exists {
		// If modifying existing user, preserve values that weren't specified
		if accessToken == "" {
			accessToken = existingUser.AccessToken
		}
		if refreshToken == "" {
			refreshToken = existingUser.RefreshToken
		}
		if idToken == "" {
			idToken = existingUser.IDToken
		}
	} else {
		// For new users, at least one token should be provided
		if accessToken == "" && refreshToken == "" && idToken == "" {
			return fmt.Errorf("must specify at least one token (--access-token, --refresh-token, or --id-token)")
		}
	}

	var claims *jwt.MapClaims
	if exists && existingUser.Claims != nil {
		claims = existingUser.Claims
	}

	userConfig.SetUser(s.User, accessToken, refreshToken, idToken, claims)

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if exists {
		fmt.Printf("✅ User '%s' modified.\n", s.User)
	} else {
		fmt.Printf("✅ User '%s' created.\n", s.User)
	}
	return nil
}

// GetClustersCommand lists all clusters
type GetClustersCommand struct {
	Output string `flag:"output,o" default:"table" help:"Output format: table, name."`
}

func (g *GetClustersCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(userConfig.Clusters) == 0 {
		fmt.Println("No clusters found.")
		return nil
	}

	if g.Output == "name" {
		for name := range userConfig.Clusters {
			fmt.Println(name)
		}
		return nil
	}

	// Table output
	headers := []string{"Name", "Server", "Issuer URL", "Client ID"}
	data := make([]map[string]interface{}, 0, len(userConfig.Clusters))

	for name, cluster := range userConfig.Clusters {
		data = append(data, map[string]interface{}{
			"Name":       name,
			"Server":     cluster.Server,
			"Issuer URL": cluster.IssuerURL,
			"Client ID":  cluster.ClientID,
		})
	}

	displayEntityTable(data, headers)
	return nil
}

// DeleteClusterCommand deletes a cluster
type DeleteClusterCommand struct {
	Cluster string `arg:"" required:"" help:"Name of the cluster to delete."`
}

func (d *DeleteClusterCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if cluster exists
	if _, ok := userConfig.Clusters[d.Cluster]; !ok {
		return fmt.Errorf("cluster '%s' not found", d.Cluster)
	}

	// Check if any contexts use this cluster
	for contextName, ctx := range userConfig.Contexts {
		if ctx.Cluster == d.Cluster {
			return fmt.Errorf("cannot delete cluster '%s': used by context '%s'", d.Cluster, contextName)
		}
	}

	delete(userConfig.Clusters, d.Cluster)

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Deleted cluster '%s'.\n", d.Cluster)
	return nil
}

// GetUsersCommand lists all users
type GetUsersCommand struct {
	Output string `flag:"output,o" default:"table" help:"Output format: table, name."`
}

func (g *GetUsersCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(userConfig.Users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	if g.Output == "name" {
		for name := range userConfig.Users {
			fmt.Println(name)
		}
		return nil
	}

	// Table output
	headers := []string{"Name", "Has Tokens"}
	data := make([]map[string]interface{}, 0, len(userConfig.Users))

	for name, user := range userConfig.Users {
		hasTokens := "No"
		if user.AccessToken != "" || user.IDToken != "" {
			hasTokens = "Yes"
		}

		data = append(data, map[string]interface{}{
			"Name":       name,
			"Has Tokens": hasTokens,
		})
	}

	displayEntityTable(data, headers)
	return nil
}

// DeleteUserCommand deletes a user
type DeleteUserCommand struct {
	User string `arg:"" required:"" help:"Name of the user to delete."`
}

func (d *DeleteUserCommand) Run() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if user exists
	if _, ok := userConfig.Users[d.User]; !ok {
		return fmt.Errorf("user '%s' not found", d.User)
	}

	// Check if any contexts use this user
	for contextName, ctx := range userConfig.Contexts {
		if ctx.User == d.User {
			return fmt.Errorf("cannot delete user '%s': used by context '%s'", d.User, contextName)
		}
	}

	delete(userConfig.Users, d.User)

	if err := config.SaveUserConfig(userConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Deleted user '%s'.\n", d.User)
	return nil
}
