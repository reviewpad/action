// Copyright (C) 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package agent

import (
	"bytes"
	"context"
	"log"
	"strings"

	atlas "github.com/explore-dev/atlas-common/go/api/services"
	"github.com/google/go-github/v42/github"
	"github.com/reviewpad/host-event-handler/handler"
	reviewpad_premium "github.com/reviewpad/reviewpad-premium/v3"
	"github.com/reviewpad/reviewpad/v3"
	"github.com/reviewpad/reviewpad/v3/collector"
	"github.com/reviewpad/reviewpad/v3/engine"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

var MixpanelToken string

type Env struct {
	RepoOwner        string
	RepoName         string
	Token            string
	PRNumber         int
	SemanticEndpoint string
	EventPayload     interface{}
}

// reviewpad-an: critical
func runReviewpad(prNum int, e *handler.ActionEvent, semanticEndpoint string, filePath string) {
	repo := *e.Repository
	splittedRepo := strings.Split(repo, "/")
	repoOwner := splittedRepo[0]
	repoName := splittedRepo[1]
	eventPayload, err := github.ParseWebHook(*e.EventName, *e.EventPayload)

	if err != nil {
		log.Print(err)
		return
	}

	env := &Env{
		RepoOwner:        repoOwner,
		RepoName:         repoName,
		Token:            *e.Token,
		PRNumber:         prNum,
		SemanticEndpoint: semanticEndpoint,
		EventPayload:     eventPayload,
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: env.Token})
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	clientGQL := githubv4.NewClient(tc)

	pullRequest, _, err := client.PullRequests.Get(ctx, env.RepoOwner, env.RepoName, env.PRNumber)
	if err != nil {
		log.Print(err)
		return
	}

	if pullRequest.Merged != nil && *pullRequest.Merged {
		log.Print("skip execution for merged pull requests")
		return
	}

	baseBranch := pullRequest.Base
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

		if baseRepo.Owner == nil {
			log.Fatalln("base branch repository owner is nil")
		}

		if baseRepo.Owner.Login == nil {
			log.Fatalln("base branch repository owner login is nil")
		}
	}

	baseRepoOwner := *pullRequest.Base.Repo.Owner.Login
	baseRepoName := *pullRequest.Base.Repo.Name
	baseRef := *pullRequest.Base.Ref

	// We fetch the configuration file from the base branch to prevent misuse
	// of the action by hijacking it through a pull request from a fork.
	ioReader, _, err := client.Repositories.DownloadContents(ctx, baseRepoOwner, baseRepoName, filePath, &github.RepositoryContentGetOptions{
		Ref: baseRef,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(ioReader)

	file, err := reviewpad.Load(buf)
	if err != nil {
		log.Print(err.Error())
		return
	}

	collectorClient := collector.NewCollector(MixpanelToken, baseRepoOwner)

	switch file.Edition {
	case engine.PROFESSIONAL_EDITION:
		err = runReviewpadPremium(ctx, env, client, clientGQL, collectorClient, pullRequest, eventPayload, file, false)
	default:
		_, err = reviewpad.Run(ctx, client, clientGQL, collectorClient, pullRequest, eventPayload, file, false)
	}

	if err != nil {
		if file.IgnoreErrors {
			log.Print(err.Error())
			return
		}
		log.Fatal(err.Error())
	}
}

// reviewpad-an: critical
func runReviewpadPremium(
	ctx context.Context,
	env *Env,
	client *github.Client,
	clientGQL *githubv4.Client,
	collector collector.Collector,
	ghPullRequest *github.PullRequest,
	eventPayload interface{},
	reviewpadFile *engine.ReviewpadFile,
	dryRun bool,
) error {
	defaultOptions := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(419430400))
	semanticConnection, err := grpc.Dial(env.SemanticEndpoint, grpc.WithInsecure(), defaultOptions)
	if err != nil {
		log.Fatalf("failed to dial semantic service: %v", err)
	}
	defer semanticConnection.Close()
	semanticClient := atlas.NewSemanticClient(semanticConnection)

	return reviewpad_premium.Run(ctx, client, clientGQL, collector, semanticClient, ghPullRequest, eventPayload, reviewpadFile, dryRun)
}

// reviewpad-an: critical
func RunAction(semanticEndpoint, rawEvent, token, file string) {
	event, err := handler.ParseEvent(rawEvent)
	if err != nil {
		log.Print(err)
		return
	}

	prs := handler.ProcessEvent(event)

	event.Token = &token

	for _, pr := range prs {
		runReviewpad(pr, event, semanticEndpoint, file)
	}
}
