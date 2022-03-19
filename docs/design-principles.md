# Gitlab Secrets Engine

The Gitlab Vault secrets plugin dynamically generates gitlab project access token based on passed parameters. This enables users to gain access to Gitlab projects without needing to create or manage project access tokens manually.

## Design Principles

This plugin supports two ways to generate a token in `/token` path

1. At root of `/token` path, a user requests a token by passing parameters.
2. (WIP): A user predefines roles with parameters. Then, a user can request a role's token at `/token/:<role-name>`

Parameters are same from Gitlab's [Project Access Token API] and [Group Access Token API], make sure to pass `type` field with project/group

path `/token`

- Create/Update: generate a project/group access token with given parameters

path `/roles/:<role_name>`

- Create/Update: create/update vault resource with given parameters. This won't do anything against Gitlab API
- Delete: delete vault resource
- Get: return stored parameters for the role
- List: list all roles

path `/token/:<role_name>`

- Create/Update: generate a project/group access token with stored parameters for the role

## Things to Note

### Access Control

There are 2 kinds of access control in this plugins.

1. permissions attaches to the configured token
1. Vault resource access control - path access and capabilities

Root `/token` path can be used to request a project/group access token for any projects/groups and any scopes as long as the configured token to generate access tokens have necessary permissions in these projects/groups. 2nd kind of access token can't limit parameters to pass.

With that being said, it's better to use **roles**, which predefines a project/groups and scopes; then, requesting a project/group access token for a role. You can further limit access to path via 2nd kind of access control imposed by Vault

[Project Access Token API]: https://docs.gitlab.com/ee/api/resource_access_tokens.html
[Group Access Token API]: https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html
