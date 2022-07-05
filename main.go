// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	atlas "github.com/explore-dev/atlas-common/go/api/services"
	"github.com/google/go-github/v42/github"
	reviewpad_premium "github.com/reviewpad/reviewpad-premium/v2"
	"github.com/reviewpad/reviewpad/v2"
	"github.com/reviewpad/reviewpad/v2/collector"
	"github.com/reviewpad/reviewpad/v2/engine"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

var MixpanelToken string

const ReviewpadFile string = "reviewpad.yml"

type Env struct {
	RepoOwner        string
	RepoName         string
	Token            string
	PRNumber         int
	SemanticEndpoint string
}

func getEnv() (*Env, error) {
	repo, ok := os.LookupEnv("INPUT_REPOSITORY")
	if !ok {
		return nil, fmt.Errorf("missing input repository")
	}

	repoStrings := strings.Split(repo, "/")
	if len(repoStrings) != 2 {
		return nil, fmt.Errorf("error parsing repository owner and name")
	}
	repoOwner := repoStrings[0]
	repoName := repoStrings[1]

	token, ok := os.LookupEnv("INPUT_TOKEN")
	if !ok {
		return nil, fmt.Errorf("missing token")
	}

	prnumStr, ok := os.LookupEnv("INPUT_PRNUMBER")
	if !ok {
		return nil, fmt.Errorf("missing pull request number")
	}

	prNum, err := strconv.Atoi(prnumStr)
	if err != nil {
		return nil, fmt.Errorf("error retrieving pull request number %v", err)
	}

	semanticEndpoint, ok := os.LookupEnv("SEMANTIC_SERVICE_ENDPOINT")
	if !ok {
		return nil, fmt.Errorf("missing semantic service endpoint")
	}

	return &Env{
		RepoOwner:        repoOwner,
		RepoName:         repoName,
		Token:            token,
		PRNumber:         prNum,
		SemanticEndpoint: semanticEndpoint,
	}, nil
}

// reviewpad-an: critical
func main() {
	env, err := getEnv()
	if err != nil {
		log.Fatalln(err.Error())
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: env.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	clientGQL := githubv4.NewClient(tc)

	ghPullRequest, _, err := client.PullRequests.Get(ctx, env.RepoOwner, env.RepoName, env.PRNumber)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if ghPullRequest.Merged != nil && *ghPullRequest.Merged {
		log.Print("skip execution for merged pull requests")
		return
	}

	baseBranch := ghPullRequest.Base
	if baseBranch == nil {
		log.Fatalln("base branch is nil")

		if baseBranch.Ref == nil {
			log.Fatalln("base branch ref is nil")
		}

		baseRepo := baseBranch.Repo
		if baseRepo == nil {
			log.Fatalln("base branch repository is nil")
		}

		if baseRepo.Name == nil {
			log.Fatalln("base branch repository name is nil")
		}

		if baseRepo.Owner == nil || baseRepo.Owner.Login == nil {
			log.Fatalln("base branch repository owner (or login) is nil")
		}
	}

	baseRepoOwner := *ghPullRequest.Base.Repo.Owner.Login
	baseRepoName := *ghPullRequest.Base.Repo.Name
	baseRef := *ghPullRequest.Base.Ref

	ioReader, _, err := client.Repositories.DownloadContents(ctx, baseRepoOwner, baseRepoName, ReviewpadFile, &github.RepositoryContentGetOptions{
		Ref: baseRef,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(ioReader)
	if err != nil {
		log.Fatalln(err.Error())
	}

	file, err := reviewpad.Load(buf)
	if err != nil {
		log.Fatalln(err.Error())
	}

	collectorClient := collector.NewCollector(MixpanelToken, baseRepoOwner)

	switch file.Edition {
	case engine.PROFESSIONAL_EDITION:
		defaultOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(419430400))
		semanticConnection, semanticConnectionError := grpc.Dial(env.SemanticEndpoint, grpc.WithInsecure(), defaultOptions)
		if semanticConnectionError != nil {
			log.Fatalf("failed to dial semantic service: %v", err)
		}
		defer semanticConnection.Close()
		semanticClient := atlas.NewSemanticClient(semanticConnection)

		err = reviewpad_premium.Run(ctx, client, clientGQL, collectorClient, semanticClient, ghPullRequest, file, false)
	default:
		_, err = reviewpad.Run(ctx, client, clientGQL, collectorClient, ghPullRequest, file, false)
	}

	if err != nil {
		if file.IgnoreErrors {
			log.Print(err.Error())
			return
		}
		log.Fatal(err.Error())
	}
}
