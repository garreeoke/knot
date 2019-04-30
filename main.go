package main

import (
	"github.com/garreeoke/knot/work"
	"log"
	"os"
	"strings"
)

func main() {

	knot := work.Knot{
		Operation: os.Getenv("KNOT_ACTION"),
		Auth:      os.Getenv("KNOT_AUTH"),
	}

	if knot.Auth == "" {
		knot.Auth = work.OnCluster
	}

	// Get source at path for yamls
	knotType := os.Getenv("KNOT_TYPE")
	knotURI := os.Getenv("KNOT_URI")

	if knotType == "" {
		log.Fatalln("No source type given")
	}
	if knot.Operation == "" {
		log.Fatalln("No operation given: create, update, dynamic")
	}
	if knotURI == "" && knotType != work.TypeLocal {
		log.Fatalln("No URI given")
	}

	if os.Getenv("KNOT_WHITELIST") != "" {
		wl := os.Getenv("KNOT_WHITELIST")
		knot.WhiteList = strings.Split(wl,",")
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
		if g.Path == "" {
			log.Println("Not all environment variables for github provided")
			log.Println("GITHUB_PATH: ", g.Path)
			log.Fatalln(" Please set variables and try again")
		}
		knot.WorkDir, err = g.GetFiles()
		if err != nil {
			log.Fatalln("Error getting files: ", err)
		}
	case work.TypeLocal:
		log.Println("Using directory mounted to: ", )
		knot.WorkDir = work.FileDir
	}

	log.Println("Workdir: ", knot.WorkDir)

	err = knot.GetK8Client()
	if err != nil {
		log.Fatalln("No client avaialble: ", err)
	}

	err = knot.Tie()
	if err != nil {
		log.Fatalln("Error with add ons: ")
	}

	// Figure out what to do with the output later

}
