# Vault Plugin for Gitlab Project Access Token

[![build-status-badge]][actions-page]
[![go-report-card-badge]][go-report-card]
[![codecov-badge]][codecov]
![go-version-badge]

This is a backend plugin to be used with Vault. This plugin generates [Project Access Tokens][pat] and [Group Access Tokens][gat]

- [Requirements](#requirements)
- [Getting Started](#getting-started)
  - [Usage](#usage)
- [Design Principles](#design-principles)
- [Development](#development)
- [Contribution](#contribution)
- [License](#license)

## Requirements

- Gitlab instance with **13.10** or later for API compatibility for [Project Access Tokens][pat] and **14.7** or later for [Group Access Tokens][gat]
- You need **14.1** or later to have access level
- Self-managed instances on Free and above. Or, GitLab SaaS Premium and above
- a token of a user with maintainer or higher permission in a project or group

- Lifting API rate limit for the user whose token will be used in this plugin to generate/revoke project/group access tokens. Admin of self-hosted can check [this doc][lift rate limit] to allow specific users to bypass authenticated request rate limiting. For SaaS Gitlab, I have not confirmed how to lift API limit yet.

## Getting Started

This is a [Vault plugin] meant to work with Vault. This guide assumes you have already installed
Vault and have a basic understanding of how Vault works.

Otherwise, first read [how to get started with Vault][vault-getting-started].

To learn specifically about how plugins work, see documentation on [Vault
plugins][vault plugin].

### Usage

```sh
# Please mount a plugin, then you can enable a secret
$ vault secrets enable -path=gitlab vault-plugin-secrets-gitlab
Success! Enabled the vault-plugin-secrets-gitlab secrets engine at: gitlab/

# configure the /config backend. You must supply a token which can generate project access tokens
$ vault write gitlab/config base_url="https://gitlab.example.com" token=$GITLAB_TOKEN 

# see supported paths
$ vault path-help gitlab/
$ vault path-help gitlab/config

# generate an ephemeral gitlab token
$ vault write gitlab/token id=1 name=ci-token scopes=api,write_repository
Key           Value
---           -----
id            12345
name          ci-token
scopes        [api write_repository]
token         REDACTED_TOKEN

# create a role
$ vault write gitlab/roles/ci-role id=1 name=project1-role scopes=read_api,read_repository
Key           Value
---           -----
role_name     ci-role
id            1
name          project1-role
scopes        [read_api read_repository]
token_ttl     86400s

# generate an ephemeral gitlab token for ci-role
$ vault write gitlab/token/ci-role
Key           Value
---           -----
id            12346
name          project1-role
scopes        [read_api read_repository]
token         REDACTED_TOKEN
expires_at    2021-09-13
```

## Design Principles

The Gitlab Vault secrets plugin dynamically generates gitlab project access token based on passed parameters. This enables users to gain access to Gitlab projects without needing to create or manage project access tokens manually.

You can find [detail design principles](docs/design-principles.md)

## Development

## Full dev environment

To be coming...

TODO: spin up a gitlab instance in docker

## Developing with an existing Gitlab instance

Requirements:

- vault

```sh
# Build binary in plugins directory, and spin up dev vault
make vault-only

# In New Terminal
export VAULT_ADDR=http://localhost:8200
export GITLAB_URL="https://gitlab.example.com"
export GITLAB_TOKEN=TOKEN


# enable secrets backend and configuration
./scripts/setup_dev_vault.sh
```

You can then issue a project access following above usage.

### Tests

```sh
# run unit tests
make test

# run subset of tests
make test TESTARGS='-run=TestConfig'

# run acceptance tests (uses Vault and Gitlab Docker containers against the compiled plugin)
make acc-test

# generate a code coverage report
make report
open coverage.html

```

## Contribution

This plugin was initially created as Hackathon project to enahance ephemeral credential suite. Another example is [vault-plugin-secrets-artifactory]. Contribution in a form of `issue`, `merge request` and donation will always be welcome.

Please refer [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## License

[Apache Software License version 2.0](LICENSE)

[pat]: https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html
[gat]: https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html
[lift rate limit]: https://docs.gitlab.com/ee/user/admin_area/settings/user_and_ip_rate_limits.html#allow-specific-users-to-bypass-authenticated-request-rate-limiting
[vault-plugin-secrets-artifactory]: https://github.com/splunk/vault-plugin-secrets-artifactory
[vault plugin]:https://www.vaultproject.io/docs/internals/plugins.html
[vault-getting-started]:https://www.vaultproject.io/intro/getting-started/install.html
[actions-page]:https://github.com/splunk/vault-plugin-secrets-gitlab/actions
[build-status-badge]:https://github.com/splunk/vault-plugin-secrets-gitlab/workflows/test.yml/badge.svg
[codecov]:https://codecov.io/gh/splunk/vault-plugin-secrets-gitlab
[codecov-badge]:https://codecov.io/gh/splunk/vault-plugin-secrets-gitlab/branch/main/graph/badge.svg
[go-report-card]:https://goreportcard.com/report/github.com/splunk/vault-plugin-secrets-gitlab
[go-report-card-badge]:https://goreportcard.com/badge/github.com/splunk/vault-plugin-secrets-gitlab
[go-version-badge]:https://img.shields.io/github/go-mod/go-version/splunk/vault-plugin-secrets-gitlab
