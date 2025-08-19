package commands

import (
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/arctir/devgraph-cli/pkg/util"
	"github.com/golang-jwt/jwt/v5"
)

type AuthLoginCommand struct {
	config.Config
}

type AuthLogoutCommand struct {
	config.Config
}

type AuthWhoamiCommand struct {
	config.Config
}

type AuthCommand struct {
	Login  *AuthLoginCommand  `cmd:"login" help:"Authenticate with your Devgraph account"`
	Logout *AuthLogoutCommand `cmd:"logout" help:"Log out and clear authentication credentials"`  
	Whoami *AuthWhoamiCommand `cmd:"whoami" help:"Show information about the currently authenticated user"`
}

// Keep the old Auth struct for backward compatibility
type Auth struct {
	config.Config
}

func parseJWT(tokenString string) (*jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Println("Error parsing token:", err)
		return nil, err
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from token")
	}
	return &claims, nil
}

func (a *Auth) Run() error {
	// Step 1: Authenticate with OIDC
	token, err := auth.Authenticate(a.Config)
	if err != nil {
		return err
	}

	claims, err := parseJWT(token.Extra("id_token").(string))
	if err != nil {
		fmt.Println("Error parsing JWT:", err)
	}

	creds := auth.Credentials{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      token.Extra("id_token").(string),
		Claims:       claims,
	}

	err = auth.SaveCredentials(creds)
	if err != nil {
		return err
	}

	// Step 2: After successful authentication, handle environment selection
	tempConfig := a.Config
	envSelected, err := util.CheckEnvironment(&tempConfig)
	if err != nil {
		// Don't fail the auth process if environment selection fails
		fmt.Printf("Warning: Environment selection failed: %v\n", err)
		fmt.Println("You can set your environment later using 'devgraph env switch'")
		return nil
	}

	if envSelected && tempConfig.Environment != "" {
		// Save the selected environment to user config
		userConfig, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Warning: Failed to save environment selection: %v\n", err)
			return nil
		}
		
		userConfig.Settings.DefaultEnvironment = tempConfig.Environment
		err = config.SaveUserConfig(userConfig)
		if err != nil {
			fmt.Printf("Warning: Failed to save environment selection: %v\n", err)
		} else {
			fmt.Printf("Environment saved as default: %s\n", tempConfig.Environment)
		}
	}

	fmt.Println("Authentication successful!")
	return nil
}

func (a *AuthLoginCommand) Run() error {
	// Just use the same logic as the original Auth command
	oldAuth := &Auth{Config: a.Config}
	return oldAuth.Run()
}

func (a *AuthLogoutCommand) Run() error {
	return auth.Logout(a.Config)
}

func (a *AuthWhoamiCommand) Run() error {
	userInfo, err := auth.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to get user information: %v", err)
	}

	// Display user information
	fmt.Println("Current User Information:")
	fmt.Println("========================")
	
	if userInfo.Name != "" {
		fmt.Printf("Name: %s\n", userInfo.Name)
	} else if userInfo.GivenName != "" || userInfo.FamilyName != "" {
		fmt.Printf("Name: %s %s\n", userInfo.GivenName, userInfo.FamilyName)
	}
	
	if userInfo.Email != "" {
		fmt.Printf("Email: %s\n", userInfo.Email)
	}
	
	if userInfo.Subject != "" {
		fmt.Printf("User ID: %s\n", userInfo.Subject)
	}
	
	if userInfo.OrganizationID != "" {
		fmt.Printf("Organization ID: %s\n", userInfo.OrganizationID)
	}
	
	if userInfo.OrganizationSlug != "" {
		fmt.Printf("Organization Slug: %s\n", userInfo.OrganizationSlug)
	}

	// Show current environment
	userConfig, err := config.LoadUserConfig()
	if err == nil && userConfig.Settings.DefaultEnvironment != "" {
		fmt.Printf("Current Environment: %s\n", userConfig.Settings.DefaultEnvironment)
	}

	return nil
}
