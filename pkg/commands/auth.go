package commands

import (
	"fmt"

	"github.com/arctir/devgraph-cli/pkg/auth"
	"github.com/arctir/devgraph-cli/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

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

	return nil
}
