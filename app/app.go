package app

import (
	"context"
	nativeErrors "errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/g-gaston/pullreminders/app/config"
	"github.com/g-gaston/pullreminders/github"
	"github.com/g-gaston/pullreminders/oauth"
	"github.com/pkg/errors"
)

var (
	whiteBold = color.New(color.FgHiWhite, color.Bold).PrintfFunc()
	white = color.New(color.FgHiWhite).PrintfFunc()
	whiteUnderline = color.New(color.FgHiWhite, color.Underline).PrintfFunc()
	green = color.New(color.FgGreen).PrintfFunc()
	blue = color.New(color.FgBlue).PrintfFunc()
)

type App struct {
	client *github.Client
}

func Init(configFileName, configFolderPath, oauthClientID, oauthClientSecret string, ctx context.Context) (app *App, err error) {
	storedConfig, err := config.NewFromFile(configFileName, configFolderPath)

	if err != nil && nativeErrors.Is(err, config.ErrNotExist) {
		storedConfig, err = buildAndStoreConfig(configFileName, configFolderPath, oauthClientID, oauthClientSecret, ctx)
	}

	if err != nil {
		return nil, errors.WithMessagef(err, "error when building app config")
	}

	return &App{
		client: github.New(storedConfig.Username, storedConfig.AccessToken, ctx),
	}, nil
}

func (app *App) ShowPullRequestsPendingForReview() {
	pulls, err := app.client.GetPullsPendingForReview()
	if err != nil {
		//TODO: what's the cleanest why of handling this?
		// Since App is responsible for the command output,
		// it makes sense to output errors from here as well
		// But should we notify the error to the caller as well?
		fmt.Printf("Error getting pending pull requests: %v\n", err)
		fmt.Printf("%+v\n", err)
		return
	}

	whiteBold("Pulls requesting a code review from you\n\n")
	for _, pull := range pulls {
		printPull(&pull)
	}
}

func buildAndStoreConfig(configFileName, configFolderPath, oauthClientID, oauthClientSecret string, ctx context.Context) (*config.Config, error) {
	accessToken, err := oauth.ExecuteOAuthFlow(oauthClientID, oauthClientSecret)
	if err != nil {
		return nil, errors.WithMessage(err, "error building new config")
	}

	apiClient := github.NewSimpleClient(accessToken, ctx)

	user, err := apiClient.GetUser()
	if err != nil {
		return nil, errors.WithMessage(err, "error getting user to build new config")
	}

	storedConfig, err := config.StoreInFile(user.Username, accessToken, configFileName, configFolderPath)
	if err != nil {
		return nil, errors.WithMessage(err, "error storing new config")
	}

	return storedConfig, nil
}

//TODO: move this to a printer service
func printPull(p *github.PullRequest) {
	green("\t#%d ", p.Number)
	white(p.Title)
	blue(" [%s]\n", p.HeadRefName)
	fmt.Printf("\t\t")
	whiteUnderline("%s\n", p.Url)
	white("\t\t@%s\n\n", p.Author)
}
