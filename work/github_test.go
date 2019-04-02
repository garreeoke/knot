package work

import (
	"log"
	"os"
	"testing"
)

func TestGetGithub(t *testing.T) {

	//NEW way: owner/applariat/repository/acme-air/branch/master
	g := GitHub{
		Path: os.Getenv("GITHUB_PATH"),
		User: os.Getenv("GITHUB_USER"),
		Token: os.Getenv("GITHUB_TOKEN"),
	}
	artifactRoot := "/tmp/knot"
	err := os.MkdirAll(artifactRoot, 0700)
	if err != nil {
		log.Println(err)
	}

	wd, err := g.GetFiles()
	if err != nil {
		log.Println(err)
	}
	log.Println("Workdir: ", wd)

}
