# Reviewpad GitHub Action ![x-ray-badge](https://img.shields.io/badge/Time%20to%20Merge-Strong%20team-ee9b00?link=https://xray.reviewpad.com/analysis?repository=https%3A%2F%2Fgithub.com%2Freviewpad%2Faction&style=plastic.svg) [![CIDeploy](https://github.com/reviewpad/action/actions/workflows/cideploy.yml/badge.svg)](https://github.com/reviewpad/action/actions/workflows/cideploy.yml)

**Latest Stable Version**: v2.x (Lisbon Edition)

For **questions**, check out the [discussions](https://github.com/reviewpad/reviewpad/discussions).

For **documentation**, check out this document and the [official documentation](https://docs.reviewpad.com).

**If you think Reviewpad is or could be useful for you, join our community on [Discord](https://reviewpad.com/discord).**

____

This action runs the docker image [reviewpad/action](https://hub.docker.com/repository/docker/reviewpad/action).

The docker image is automatically pushed to Docker Hub on every commit to the main branch.

It reads and automates the pull request workflows specified in the `reviewpad.yml` file at the root of your GitHub repository.

These workflows can be used to automatically label, assign reviewers, comment, merge and close pull requests.

For example, the following `reviewpad.yml` file:

```yaml
api-version: reviewpad.com/v2.x

rules:
  - name: is-small
    kind: patch
    description: small pull request
    spec: '$size() <= 50'

  - name: is-medium
    kind: patch
    description: medium-sized pull request
    spec: '$size() > 50 && $size() <= 150'

  - rule: is-large
    kind: patch
    description: large-sized pull request
    spec: '$size() > 150'

workflows:
  - name: label-pull-request-with-size
    description: Label pull request with size
    if:
      - rule: is-small
        extra-actions:
          - $addLabel("small")
      - rule: is-medium
        extra-actions:
          - $addLabel("medium")
      - rule: is-large
        extra-actions:
          - $addLabel("large")
```

Specifies a workflow to automatically add a label based on the size of the pull request.

For more information on the release procedure, check the [RELEASE.md](./RELEASE.md) document.

## Inputs

- **repository**: Uses default `${{ github.repository }}`
- **prnumber**: Uses default `${{ github.event.pull_request.number }}`
- **token**: Uses default `${{ github.token }}`

## Outputs

None.

## Usage examples

**This action should only be used on `pull_request` related events.**

Add the following step to a GitHub Action job:

```yaml
- name: Run reviewpad action
  uses: reviewpad/action@v2.x
```


By default this action uses the `github-actions[bot]` PAT.

As described in the [official GitHub documentation](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#using-the-github_token-in-a-workflow):

> When you use the repository's GITHUB_TOKEN to perform tasks, events triggered by the GITHUB_TOKEN will not create a new workflow run.

If you want to use more advanced features such as the auto-merge feature, we recommend that you explicitly pass a PAT to run this action:

```yaml
- name: Run reviewpad action
  uses: reviewpad/action@v2.x
  with:
    token: ${{ secrets.GH_TOKEN }}
```

[Please follow this link to know more](https://docs.reviewpad.com/docs/install-github-action-with-github-token).
