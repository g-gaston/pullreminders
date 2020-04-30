package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"os/user"

	"github.com/cli/cli/auth"
	"github.com/fatih/color"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const (
	oauthHost        = "github.com"
	configPath       = "~/.config/pullreminders"
	configFileName   = "config.yml"
	oauthSuccessPage = `
<!doctype html>
<meta charset="utf-8">
<title>Success: GitHub CLI</title>
<style type="text/css">
body {
  color: #1B1F23;
  background: #F6F8FA;
  font-size: 14px;
  font-family: -apple-system, "Segoe UI", Helvetica, Arial, sans-serif;
  line-height: 1.5;
  max-width: 620px;
  margin: 28px auto;
  text-align: center;
}
h1 {
  font-size: 24px;
  margin-bottom: 0;
}
p {
  margin-top: 0;
}
.box {
  border: 1px solid #E1E4E8;
  background: white;
  padding: 24px;
  margin: 28px;
}
</style>
<body>
  <svg height="52" class="octicon octicon-mark-github" viewBox="0 0 16 16" version="1.1" width="52" aria-hidden="true"><path fill-rule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path></svg>
  <div class="box">
    <h1>Successfully authenticated GitHub pullreminders CLI</h1>
    <p>You may now close this tab and return to the terminal.</p>
  </div>
</body>
`
)

var (
	// The "pullreminders" OAuth app
	oauthClientID = "798d08bc0f3e41ea9625"
	// This value is safe to be embedded in version control
	oauthClientSecret = "b14ca9798078b3af4c84c886529211ea0f3ba241"
)

type Config struct {
	Login string
	Token string
}

func main() {
	data, err := readFile(configFile(configPath, configFileName))
	var config *Config
	if err != nil && errors.Is(err, os.ErrNotExist) {
		config, err = setupConfigFile(configPath, configFileName, oauthSuccessPage)
		if err != nil {
			fmt.Printf("Error setting up config: %s\n", err)
			os.Exit(1)
		}
	} else if err == nil {
		config = &Config{}
		err = yaml.Unmarshal(data, config)
		if err != nil {
			fmt.Printf("Error unmarshaling config: %s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Error readin config file: %s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	searchQuery := fmt.Sprintf("type:pr state:open review-requested:%s", config.Login)
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
		} `graphql:"search(query: $query, type: ISSUE, first: 100)"`
	}

	err = client.Query(ctx, &pullsQuery, variables)

	if err != nil {
		fmt.Printf("Error running query: %s\n", err)
		return
	}

	whiteBold := color.New(color.FgHiWhite, color.Bold).PrintfFunc()
	white := color.New(color.FgHiWhite).PrintfFunc()
	whiteUnderline := color.New(color.FgHiWhite, color.Underline).PrintfFunc()
	green := color.New(color.FgGreen).PrintfFunc()
	blue := color.New(color.FgBlue).PrintfFunc()

	whiteBold("Pulls requesting a code review from you\n\n")
	for _, pull := range pullsQuery.Search.PullRequests {
		p := pull.PullRequest
		green("\t#%d ", p.Number)
		white(p.Title)
		blue(" [%s]\n", p.HeadRefName)
		fmt.Printf("\t\t")
		whiteUnderline("%s\n", p.Url)
		white("\t\t@%s\n\n", p.Author.Login)
	}

}

func waitForEnter(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	return scanner.Err()
}

func readFile(filePath string) (data []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err = ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func setupConfigFile(configPath, configFileName, oauthSuccessPage string) (config *Config, err error) {
	authFlow := &auth.OAuthFlow{
		Hostname:     oauthHost,
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		Scopes:       []string{"repo"},
		WriteSuccessHTML: func(w io.Writer) {
			fmt.Fprintln(w, oauthSuccessPage)
		},
		VerboseStream: nil,
	}

	fmt.Printf("Press Enter to open %s in your browser... ", authFlow.Hostname)
	_ = waitForEnter(os.Stdin)
	githubToken, err := authFlow.ObtainAccessToken()
	if err != nil {
		return nil, err
	}

	if githubToken == "" {
		return nil, errors.New("Can't find auth token")
	}

	ctx := context.Background()

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	var userQuery struct {
		Viewer struct {
			Login string
		}
	}

	err = client.Query(ctx, &userQuery, nil)

	if err != nil {
		return nil, err
	}

	config = &Config{
		Login: userQuery.Viewer.Login,
		Token: githubToken,
	}

	// write to file
	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	configPath = expandPath(configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(configPath, os.ModePerm)
	}

	err = ioutil.WriteFile(configFile(configPath, configFileName), data, 0644)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func configFile(folderPath, name string) string {
	abs, _ := filepath.Abs(path.Join(expandPath(folderPath), name))
	return abs
}

func expandPath(folderPath string) string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	if folderPath == "~" {
		// In case of "~", which won't be caught by the "else if"
		folderPath = homeDir
	} else if strings.HasPrefix(folderPath, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		folderPath = filepath.Join(homeDir, folderPath[2:])
	}
	return folderPath
}
