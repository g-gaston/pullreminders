package github

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type SimpleClient struct {
	githubApiClient *githubv4.Client
	ctx context.Context
}

type User struct {
	Username string
}

func NewSimpleClient(accessToken string, ctx context.Context) *SimpleClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	return &SimpleClient{
		githubApiClient: client,
		ctx: ctx, // I'm not sure about this, maybe I should allow to pass a different context on every api call
	}
}

func (c *SimpleClient) GetUser() (user *User, err error) {
	var userQuery struct {
		Viewer struct {
			Login string
		}
	}

	err = c.githubApiClient.Query(c.ctx, &userQuery, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error when running github api query to get current user")
	}

	return &User {
		Username: userQuery.Viewer.Login,
	}, nil
}
