package github

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
)

type Client struct {
	user string
	*SimpleClient
}

func New(user, accessToken string, ctx context.Context) *Client {
	return &Client{
		user:         user,
		SimpleClient: NewSimpleClient(accessToken, ctx),
	}
}

type PullRequest struct {
	Id          string
	Title       string
	Url         string
	HeadRefName string
	Number      int
	Author      string
}

func (c *Client) GetPullsPendingForReview() (pulls []PullRequest, err error) {
	searchQuery := fmt.Sprintf("type:pr state:open review-requested:%s", c.user)
	variables := map[string]interface{}{
		"query": githubv4.String(searchQuery),
	}

	var pullsQuery struct {
		Search struct {
			IssueCount int
			PageInfo   struct {
				EndCursor   string
				HasNextPage bool
			}
			PullRequests []struct {
				PullRequest struct {
					Id          string
					Title       string
					Url         string
					HeadRefName string
					Number      int
					Author      struct {
						Login string
					}
				} `graphql:"... on PullRequest"`
			} `graphql:"nodes"`
		} `graphql:"search(query: $query, type: ISSUE, first: 100)"` //TODO: handle pagination
	}

	err = c.githubApiClient.Query(c.ctx, &pullsQuery, variables)

	if err != nil {
		return nil, errors.Wrap(err, "errors retrieving pull requests pending for review")
	}

	pulls = make([]PullRequest, len(pullsQuery.Search.PullRequests), 0)

	for _, pull := range pullsQuery.Search.PullRequests {
		p := pull.PullRequest
		pulls = append(
			pulls,
			PullRequest{
				Id:          p.Id,
				Title:       p.Title,
				Url:         p.Url,
				HeadRefName: p.HeadRefName,
				Number:      p.Number,
				Author:      p.Author.Login,
			},
		)
	}
	return pulls, nil
}
