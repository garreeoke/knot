package main

import (

	"log"
	"os"
	"github.com/garreeoke/knot/work"
)

func main() {

	knot := work.Knot{
		Action: "create",
		Auth: os.Getenv("KNOT_AUTH"),
		//KubeConfigPath: os.Getenv("KNOT_K8_CFG_PATH"),
	}

	if knot.Auth == "" {
		knot.Auth = work.OnCluster
	} else if knot.Auth == work.Local {
		/*
		if knot.KubeConfigPath == "" {
			log.Fatalln("KNOT_K8_CFG env variable not set")
		}
		*/
	}

	// Get source at path for yamls
	knotType := os.Getenv("KNOT_TYPE")
	knotURI := os.Getenv("KNOT_URI")

	if knotType == "" {
		log.Fatalln("No source type given")
	}
	if knotURI == "" {
		log.Fatalln("No URI given")
	}

	var err error

	log.Println("Knot type: ", knotType)
	switch knotType {
	case work.TypeGitHub:
		g := work.GitHub{
			User: os.Getenv("GITHUB_USER"),
			Token: os.Getenv("GITHUB_TOKEN"),
			Path: knotURI,
		}
		if g.User == "" || g.Token == "" || g.Path == "" {
			log.Println("Not all environment variables for github provided")
			log.Println("GITHUB_USER: ", g.User)
			log.Println("GITHUB_TOKEN: ", g.Token)
			log.Println("GITHUB_PATH: ", g.Path)
			log.Fatalln(" Please set variables and try again")
		}
		knot.WorkDir, err = g.GetFiles()
		if err != nil {
			log.Fatalln("Error getting files: ", err)
		}
	}

	log.Println("Workdir: ", knot.WorkDir)

	err = knot.GetK8Client()
	if err != nil {
		log.Fatalln("No client avaialble: ", err)
	}

	err = knot.AddOns()
	if err != nil {
		log.Fatalln("Error with add ons: ")
	}

}
