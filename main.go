package main

import (
	"context"
	"fmt"
	"github.com/g-gaston/pullreminders/app"
	"os"
)

const (
	configPath     = "~/.config/pullreminders"
	configFileName = "config.yml"
	// The "pullreminders" OAuth app
	oauthClientID = "798d08bc0f3e41ea9625"
	// This value is safe to be embedded in version control
	oauthClientSecret = "b14ca9798078b3af4c84c886529211ea0f3ba241"
)

func main() {
	a, err := app.Init(configFileName, configPath, oauthClientID, oauthClientSecret, context.Background())
	if err != nil {
		fmt.Printf("Error initializing app: %s\n", err)
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	a.ShowPullRequestsPendingForReview()
}
