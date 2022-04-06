# Backlogs

## Replace Gitlab Library

For speed of development, I reused <https://github.com/xanzy/go-gitlab> library. This library is actively maintained by community. However, with the far more active development on Gitlab side, library sometimes sits behind. Also since this plugin only uses limited API endpoints, we can possibly maintain our gitlab clinet.

## Feasibility with Gitlab's Native Vault Support

This plugin is not initially created with a mindset of compatibility with Gitlab's native vault support for static credentials.

## Granular Control on Token Expiry

Gitlab currently doesn't have granular control on token expiry. A token is expired at midnight UTC of a chosen day. We should have a shorter-lived token like 15 mins in case a user failed to revoke a token after use.

Gitlab issue to have [granular control on token expiry]

## Pipeline

Setup CICD pipeline in gitlab and do the following at least.

- lint
- unit test
- OSS scan
- SAST scan
- binary build and publish

For comprehensive CI,

- DAST scan
- acceptance testing

## Acceptance Testing

Running test against real servers doesn't seem good idea. Create an isolated environment by spinning up vault and gitlab in docker in CI. Then, run full suite of testing there. *Self-hosted GitLab has project/group access token available from free version*

[granular control on token expiry]: https://gitlab.com/gitlab-org/gitlab/-/issues/335535
