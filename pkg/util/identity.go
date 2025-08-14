package util

import "github.com/arctir/devgraph-cli/pkg/auth"

func GetUsername() (string, error) {
	var username string
	creds, err := auth.LoadCredentials()
	if err != nil {
		return "", err
	}
	if creds.Claims == nil {
		username = "localuser"
	}

	var ok bool
	username, ok = (*creds.Claims)["preferred_username"].(string)
	if !ok {
		username, ok = (*creds.Claims)["email"].(string)
		if !ok {
			username = "localuser"
		}
	}
	return username, nil
}
