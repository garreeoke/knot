package main

import (

	"log"
	"os"
	//"github.com/garreeoke/kates"
)

func main() {

	// Get source at path for yamls
	sourceType := os.Getenv("KNOT_SOURCE_TYPE")
	sourcePath := os.Getenv("KNOT_SOURCE_PATH")

	if sourceType == "" {
		log.Fatalln("No source type given")
	}



}
