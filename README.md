# Vault Plugin for Gitlab Project Access Token

This is a backend pluing to be used with Vault. This plugin generates [Gitlab Project Access Tokens][pat]

- [Requirements](#requirements)
- [Design Principles](#design-principles)
- [Contribution](#contribution)
- [License](#license)

## Requirements

- Gitlab instance wiht **13.10** or later for API compatibility
- Self-managed instances on Free and above. Or, GitLab SaaS Premium and above
- a token of a user with maintainer or higher permission in a project

- Lifting API rate limit for the user whose token will be used in this plugin to generate/revoke project access tokens. Admin of self-hosted can check [this doc][lift rate limit] to allow specific users to bypass authenticated request rate limiting. For SaaS Gitlab, I have not confirmed how to lift API limit yet.

## Design Principles

The Gitlab Vault secrets plugin dynamically generates gitlab project access token based on passed parameters. This enables users to gain access to Gitlab projects without needing to create or manage project access tokens manually.

You can find [detail design principles](docs/design-principles.md)

## Contribution

This plugin was initially created as Hackathon project to enahance ephemeral credential suite. Another example is [vault-plugin-secrets-artifactory]. Contribution in a form of `issue`, `merge request` and donation will always be welcome.

Please refer [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## License

[Apache Software License version 2.0](LICENSE)

[pat]: https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html
[lift rate limit]: https://docs.gitlab.com/ee/user/admin_area/settings/user_and_ip_rate_limits.html#allow-specific-users-to-bypass-authenticated-request-rate-limiting
[vault-plugin-secrets-artifactory]: https://github.com/splunk/vault-plugin-secrets-artifactory
