package work

import (
	"log"
	"io/ioutil"
	"fmt"
	"strings"
	"errors"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"
	"context"
)

func (g *GitHub) GetFiles () (string,error) {

	log.Println("Getting artifacts from github: ", g.Path)

	// use vanilla HttpClient
	tc := http.DefaultClient
	if (g.Token != "") {
		// if credential has a token then create oauth HttpClient
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.Token},
		)
		tc = oauth2.NewClient(context.Background(), ts)
	}
	client := github.NewClient(tc)

	newInfo := strings.Split(g.Path, "/")
	repoOptions := github.RepositoryContentGetOptions{}
	if len(newInfo) == 6 {
		//          0     1         2          3        4      5
		// NEW way: owner/applariat/repository/acme-air/branch/master
		// NEW way: owner/applariat/repository/acme-air/commit/1234908sha
		// NEW way: owner/applariat/repository/acme-air/tag/v22.0
		repoOptions.Ref = newInfo[5]
	}

	url, _, err := client.Repositories.GetArchiveLink(context.Background(),newInfo[1], newInfo[3], github.Zipball, &repoOptions)
	if err != nil {
		return "",err
	}
	resp, err := tc.Get(url.String())
	if err != nil {
		return "",err
	}

	if resp.StatusCode == 404 {
		return "",errors.New(fmt.Sprintf("File not found: %v", url))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "",err
	}

	g.Path = g.Path + ".zip"
	err = ProcessFile(body, g)
	if err != nil {
		return "",err
	}

	return g.WorkDir,nil
}

func (g *GitHub) GetPath() string {
	return g.Path
}

func (g *GitHub) SetWorkDir(name string) {
	g.WorkDir = name
}

func (g *GitHub) GetWorkDir() string {
	return g.WorkDir
}