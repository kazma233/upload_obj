package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type (
	PathType string
)

var (
	DAY    PathType = "DAY"
	HOUR   PathType = "HOUR"
	SECOND PathType = "SECOND"
	NONE   PathType = "NONE"
)

type (
	GithubConfig struct {
		Token string `json:"token"` // github write token

		Repo           string   `json:"repo"`
		Owner          string   `json:"owner"`
		Branch         string   `json:"branch"`
		PrefixPathType PathType `json:"prefix"`
		Commit         string   `json:"commit"`
		Author         Author   `json:"author"`
	}

	Author struct {
		Name string `json:"name"`
		Mail string `json:"mail"`
	}
)

type (
	GithubBed struct {
		config GithubConfig
		client *github.Client
	}
)

func NewGithubBed(githubConfig GithubConfig) (g GithubBed, err error) {
	ctx, df := context.WithTimeout(context.Background(), time.Second*30)
	defer df()

	err = g.Vailed(githubConfig)
	if err != nil {
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubConfig.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	g.config = githubConfig
	g.client = client

	return
}

func (g *GithubBed) Type() BedType {
	return GITHUB
}

func (g *GithubBed) Close() {

}

// UploadByPath impl bed.Bed
func (g *GithubBed) UploadByPath(filePath string) (string, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fb, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return g.UploadByBytes(fb, fi.Name())
}

func (g *GithubBed) UploadByBytes(bs []byte, fileName string) (string, error) {
	conf := g.config

	message := fmt.Sprintf("%v(%v)", conf.Commit, time.Now().Format("2006-01-02 15:04:05"))
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: bs,
		Branch:  github.String(conf.Branch),
		Committer: &github.CommitAuthor{
			Name:  github.String(conf.Author.Name),
			Email: github.String(conf.Author.Mail),
		},
	}

	resp, _, err := g.client.Repositories.CreateFile(
		context.Background(),
		conf.Owner,
		conf.Repo,
		path.Join(conf.PrefixPathType.Path(), fileName),
		opts,
	)

	if err != nil {
		return "", err
	}

	return resp.Content.GetDownloadURL(), nil
}

func (g *GithubBed) Vailed(config GithubConfig) error {
	if config.Token == "" {
		return errors.New("token must not be empty")
	}

	return nil
}

func (pt PathType) Path() string {
	switch pt {
	case DAY:
		return time.Now().Format("2006/01/02")
	case HOUR:
		return time.Now().Format("2006/01/02/15/04")
	case SECOND:
		return time.Now().Format("2006/01/02/15/04/05")
	case NONE:
		return ""
	default:
		return ""
	}
}
