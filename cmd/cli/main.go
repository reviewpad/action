// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/reviewpad/action/v3/agent"
)

var (
	token            = flag.String("github-token", "", "GitHub token")
	semanticEndpoint = flag.String("semantic-endpoint", "0.0.0.0:3008", "Semantic client endpoint")
	eventFilePath    = flag.String("event-payload", "", "File path to github action event")
)

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "help" {
		usage()
	}

	if *token == "" {
		log.Printf("missing argument token")
		usage()
	}

	if *eventFilePath == "" {
		log.Printf("missing argument event")
		usage()
	}

	content, err := ioutil.ReadFile(*eventFilePath)
	if err != nil {
		log.Fatal(err)
	}

	rawEvent := string(content)

	agent.RunAction(*semanticEndpoint, rawEvent, *token)
}
