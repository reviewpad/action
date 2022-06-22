// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/reviewpad/action/v2/agent"
)

func main() {
	semanticEndpoint, ok := os.LookupEnv("SEMANTIC_SERVICE_ENDPOINT")
	if !ok {
		log.Print("missing semantic service endpoint")
		return
	}

	rawEvent, ok := os.LookupEnv("INPUT_EVENT")
	if !ok {
		log.Print("missing variable INPUT_EVENT")
		return
	}

	agent.RunAction(&semanticEndpoint, &rawEvent)
}
