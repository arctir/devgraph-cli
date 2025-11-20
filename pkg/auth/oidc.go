package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/arctir/devgraph-cli/pkg/config"
	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/int128/oauth2cli"
	"github.com/int128/oauth2cli/oauth2params"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

// WellKnownConfig represents the standard OpenID Connect discovery document
type WellKnownConfig struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	// Add more fields as needed (e.g., "end_session_endpoint")
}

func getWellKnownEndpoints(issuerURL string) (*WellKnownConfig, error) {
	// Parse the issuer URL
	u, err := url.Parse(issuerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer URL: %w", err)
	}

	// Append the well-known path
	u.Path += "/.well-known/openid-configuration"

	// Make the HTTP request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch well-known config: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the JSON response
	var config WellKnownConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &config, nil
}

// UpdateClerkSessionForEnvironment updates the Clerk session for a specific environment
// This is a placeholder for Clerk-specific organization switching logic
func UpdateClerkSessionForEnvironment(config config.Config, environmentID string) error {
	// For now, this is a no-op since the environment switching is handled
	// by your API layer through the Devgraph-Environment header
	// In the future, if you need to update the Clerk session's active organization,
	// you would implement the Clerk API calls here

	fmt.Printf("Debug: Would update Clerk session for environment %s\n", environmentID)
	return nil
}

// Logout performs OIDC logout by calling the end session endpoint and clearing local credentials
func Logout(config config.Config) error {
	// Load current credentials
	creds, err := LoadCredentials()
	if err != nil || creds.IDToken == "" {
		// If no credentials, just clear any local state
		fmt.Println("No active session found.")
		return ClearCredentials()
	}

	// Get OIDC well-known configuration
	wellKnown, err := getWellKnownEndpoints(config.IssuerURL)
	if err != nil {
		fmt.Printf("Warning: Could not retrieve logout endpoint: %v\n", err)
		fmt.Println("Clearing local credentials...")
		return ClearCredentials()
	}

	// Try to call the OIDC end session endpoint if available
	if endSessionURL := getEndSessionEndpoint(wellKnown); endSessionURL != "" {
		err := callEndSessionEndpoint(endSessionURL, creds.IDToken, DefaultRedirectURL)
		if err != nil {
			fmt.Printf("Warning: OIDC logout failed: %v\n", err)
			fmt.Println("Clearing local credentials anyway...")
		} else {
			fmt.Println("Successfully logged out of OIDC provider.")
		}
	} else {
		fmt.Println("No OIDC logout endpoint found. Clearing local credentials...")
	}

	// Clear local credentials
	return ClearCredentials()
}

// ClearCredentials removes all stored authentication credentials
func ClearCredentials() error {
	userConfig, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	// Clear credentials
	userConfig.Credentials = config.Credentials{}

	err = config.SaveUserConfig(userConfig)
	if err != nil {
		return fmt.Errorf("failed to save cleared credentials: %w", err)
	}

	fmt.Println("Local credentials cleared.")
	return nil
}

// getEndSessionEndpoint extracts the end session endpoint from well-known config
func getEndSessionEndpoint(wellKnown *WellKnownConfig) string {
	// This would need to be added to WellKnownConfig struct
	// For now, construct the standard Clerk logout URL
	if wellKnown.Issuer != "" {
		return wellKnown.Issuer + "/v1/client/sign_outs"
	}
	return ""
}

// callEndSessionEndpoint calls the OIDC end session endpoint
func callEndSessionEndpoint(endSessionURL, idToken, redirectURL string) error {
	client := &http.Client{}

	// Build logout URL with parameters
	u, err := url.Parse(endSessionURL)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("id_token_hint", idToken)
	if redirectURL != "" {
		q.Set("post_logout_redirect_uri", redirectURL)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("logout request failed with status: %d", resp.StatusCode)
	}

	return nil
}

// GetCurrentUser returns information about the currently authenticated user
func GetCurrentUser() (*UserInfo, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds.IDToken == "" {
		return nil, fmt.Errorf("no active session found")
	}

	if creds.Claims == nil {
		return nil, fmt.Errorf("no user claims available")
	}

	claims := *creds.Claims
	userInfo := &UserInfo{}

	// Extract standard OIDC claims
	if sub, ok := claims["sub"].(string); ok {
		userInfo.Subject = sub
	}
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	}
	if givenName, ok := claims["given_name"].(string); ok {
		userInfo.GivenName = givenName
	}
	if familyName, ok := claims["family_name"].(string); ok {
		userInfo.FamilyName = familyName
	}
	if picture, ok := claims["picture"].(string); ok {
		userInfo.Picture = picture
	}

	// Extract Clerk-specific organization information
	if orgData, ok := claims["org_metadata"]; ok {
		userInfo.OrganizationMetadata = orgData
	}
	if orgId, ok := claims["org_id"].(string); ok {
		userInfo.OrganizationID = orgId
	}
	if orgSlug, ok := claims["org_slug"].(string); ok {
		userInfo.OrganizationSlug = orgSlug
	}

	return userInfo, nil
}

// UserInfo represents user information extracted from OIDC claims
type UserInfo struct {
	Subject              string      `json:"subject"`
	Email                string      `json:"email"`
	Name                 string      `json:"name"`
	GivenName            string      `json:"given_name"`
	FamilyName           string      `json:"family_name"`
	Picture              string      `json:"picture"`
	OrganizationID       string      `json:"org_id"`
	OrganizationSlug     string      `json:"org_slug"`
	OrganizationMetadata interface{} `json:"org_metadata"`
}

func Authenticate(a config.Config) (*oauth2.Token, error) {
	ctx := context.Background()
	ready := make(chan string, 1)
	tokenChan := make(chan *oauth2.Token, 1)
	defer close(ready)
	defer close(tokenChan)

	providerConfig, err := getWellKnownEndpoints(a.IssuerURL)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// Configure OAuth2 with PKCE
	oauth2Config := oauth2.Config{
		ClientID:    a.ClientID,
		RedirectURL: DefaultRedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  providerConfig.AuthorizationEndpoint,
			TokenURL: providerConfig.TokenEndpoint,
		},
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	pkce, err := oauth2params.NewPKCE()
	if err != nil {
		return nil, fmt.Errorf("error: %s", err)
	}

	// Set up oauth2cli configuration
	cliConfig := oauth2cli.Config{
		OAuth2Config:           oauth2Config,
		LocalServerBindAddress: []string{"0.0.0.0:40000", "0.0.0.0:40001", "0.0.0.0:40002", "0.0.0.0:40003", "0.0.0.0:40004", "0.0.0.0:40005"},

		AuthCodeOptions:      pkce.AuthCodeOptions(),
		TokenRequestOptions:  pkce.TokenRequestOptions(),
		LocalServerReadyChan: ready,
		LocalServerSuccessHTML: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Devgraph - Authentication Successful</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
            background: #070d17;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 48px;
            text-align: center;
            max-width: 420px;
            width: 100%;
        }
        .logo {
            font-size: 24px;
            font-weight: 700;
            color: #ffffff;
            margin-bottom: 32px;
            letter-spacing: -0.5px;
        }
        .success-icon {
            width: 64px;
            height: 64px;
            background: #10b981;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
        }
        .success-icon svg {
            width: 32px;
            height: 32px;
            fill: white;
        }
        h1 {
            color: #ffffff;
            font-size: 24px;
            margin-bottom: 12px;
            font-weight: 600;
        }
        p {
            color: rgba(255, 255, 255, 0.7);
            font-size: 15px;
            line-height: 1.6;
            margin-bottom: 16px;
        }
        .close-hint {
            font-size: 13px;
            color: rgba(255, 255, 255, 0.4);
            margin-bottom: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">Devgraph</div>
        <div class="success-icon">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
            </svg>
        </div>
        <h1>Authentication Successful</h1>
        <p>You have successfully signed in to Devgraph.</p>
        <p class="close-hint">You can close this window and return to your terminal.</p>
    </div>
</body>
</html>`,
		//Logf:                 log.Printf,
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		select {
		case url := <-ready:
			fmt.Println("\n" + "============================================================")
			fmt.Println("🔐 Devgraph Authentication")
			fmt.Println("============================================================")
			fmt.Println("Opening browser for authentication...")
			fmt.Printf("URL: %s\n", url)
			if err := browser.OpenURL(url); err != nil {
				fmt.Printf("⚠️  Could not open browser automatically: %s\n", err)
				fmt.Println("Please open the URL above manually in your browser.")
			}
			fmt.Println("⏳ Waiting for authentication to complete...")
			fmt.Println()
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context done while waiting for authorization: %w", ctx.Err())
		}
	})
	eg.Go(func() error {
		token, err := oauth2cli.GetToken(ctx, cliConfig)
		if err != nil {
			return fmt.Errorf("could not get a token: %w", err)
		}
		tokenChan <- token
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("authorization error: %s", err)
	}

	token := <-tokenChan
	return token, nil
}
