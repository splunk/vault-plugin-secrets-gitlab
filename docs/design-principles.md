# Gitlab Secrets Engine

The Gitlab Vault secrets plugin dynamically generates gitlab project access token based on passed parameters. This enables users to gain access to Gitlab projects without needing to create or manage a dedicated services account.

## Design Principles

This plugin supports two ways to generate a token in `/token` path

1. At root of `/token` path, a user requests a token by passing parameters.
2. (WIP): A user predefines roles with parameters. Then, a user can request a role's token at `/token/:<role-name>`

Parameters are same from Gitlab's [Project Access Token API]

[Project Access Token API]: https://docs.gitlab.com/ee/api/resource_access_tokens.html

## Things to Note
