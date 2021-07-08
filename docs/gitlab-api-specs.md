# Gitlab API Specs

Gitlab project access token API docs can be found [here][pat doc].

## API behavior

### Creating a Token

#### Create a token with same name that exists in a project

when you create a token with same name that already exists in a project, it creates another token. Name field is actually obsolete in terms of uniqueness. Every call is POST and it creates a new token regardless of name provided.  There's currently no update(PUT) API for existing project access tokens.

```bash
➜ curl --header "PRIVATE-TOKEN: MYTOKEN" https://my.gitlab.com/api/v4/projects/1/access_tokens/ -XPOST --header "Content-Type:application/json" \
--data '{ "name":"test_token", "scopes":["api"] }'
{"id":10,"name":"test_token","revoked":false,"created_at":"2021-01-01T00:00:0.000Z","scopes":["api"],"user_id":10,"active":true,"expires_at":null,"token":"XXX"}
➜ curl --header "PRIVATE-TOKEN: MYTOKEN" https://my.gitlab.com/api/v4/projects/1/access_tokens/ -XPOST --header "Content-Type:application/json" \
--data '{ "name":"test_token", "scopes":["api"] }'
{"id":11,"name":"test_token","revoked":false,"created_at":"2021-0101T00:00:1.000Z","scopes":["api"],"user_id":11,"active":true,"expires_at":null,"token":"YYY"} 
```

### Create a token in a project where its parent group disables creation of project access token

when a parent group disables `Allow project access token creation` like [this image](./disabled-pat-setting-in-group.md). (You can visit thsi in groups settings > genera > Permissions)

```bash
➜ curl --header "PRIVATE-TOKEN: MYTOKEN" https://my.gitlab.com/api/v4/projects/1/access_tokens/ -XPOST --header "Content-Type:application/json" \
--data '{ "name":"test_token", "scopes":["api"] }'
{"message":"400 Bad request - User does not have permission to create project access token"}
```

### Create a token with scope that's not available for the project

If a project/group/instance doesn't enable certain scopes such as container registry, it gets 400

```bash
➜ curl --header "PRIVATE-TOKEN: MYTOKEN" https://my.gitlab.com/api/v4/projects/1/access_tokens/ -XPOST --header "Content-Type:application/json" --data '{ "name":"test_token_developer", "scopes":["api", "read_repository","read_registry"], "access_level": 40 }'
{"message":"400 Bad request - Scopes can only contain available scopes"}%   
```

### Revoking a Token

#### Revoking a token that has been revoked

```bash
➜ curl --header "PRIVATE-TOKEN: MYTOKEN" https://my.gitlab.com/api/v4/projects/1/access_tokens/1 -XDELETE
{"message":"404 Could not find project access token with token_id: 1 Not Found"}
```

#### Revoking a token in another project

Say, a token is created in project 1. What happens if we try to delete the generated token in another project, say project 2

[pat doc]: https://docs.gitlab.com/ee/api/resource_access_tokens.html
