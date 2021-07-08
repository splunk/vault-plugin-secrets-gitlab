# Contributing Guidelines

Thank you for your interest in contributing to our project! Whether it's a bug report, new feature, question, or additional documentation, we greatly value feedback and contributions from our community. Read through this document before submitting any issues or pull requests to ensure we have all the necessary information to effectively respond to your bug report or contribution.

In addition to this document, please review our [Code of Conduct](CODE_OF_CONDUCT.md). For any code of conduct questions or comments please leave an issue.

## Reporting Bugs/Feature Requests

We welcome you to use the [Gitlab issue tracker] to report bugs or suggest features. When filing an issue, please check existing open, or recently closed, issues to make sure somebody else hasn't already reported the issue. Please try to include as much information as you can. Details like these are incredibly useful:

- A reproducible test case or series of steps
- The version of this plugin
- The version of Vault Server
- The version of Gitlab Server
- Any modifications you've made relevant to the bug
- Anything unusual about your environment or deployment
- Any known workarounds

When filing an issue, please do *NOT* include:

- Internal identifiers such as JIRA tickets
- Any sensitive information related to your environment, users, etc.

## Contributing via Merge Requests

Contributions via Merge Requests (MRs) are much appreciated. Before sending us a merge request, please ensure that:

1. You are working against the latest source on the `main` branch.
2. You check existing open, and recently merged, merge requests to make sure
   someone else hasn't addressed the problem already.
3. You open an issue to discuss any significant work - we would hate for your
   time to be wasted.
4. You submit MRs that are easy to review and ideally less 500 lines of code.
   Multiple MRs can be submitted for larger contributions.

To send us a merge request, please:

1. Fork the project.
2. Modify the source; please ensure a single change per MR. If you also
   reformat all the code, it will be hard for us to focus on your change.
3. Ensure local tests pass and add new tests related to the contribution.
4. Commit to your fork using clear commit messages.
5. Send us a merge request, answering any default questions in the merge request
   interface.
6. Pay attention to any automated CI failures reported in the merge request, and
   stay involved in the conversation.

GitLab provides additional documentation on [forking a
project](https://docs.gitlab.com/ee/user/project/repository/forking_workflow.html) and [creating a merge
request](https://docs.gitlab.com/ee/user/project/merge_requests/creating_merge_requests.html).

## Finding contributions to work on

Looking at the existing issues is a great way to find something to contribute
on. As our projects, by default, use the default GitLab issue labels
(enhancement/bug/duplicate/help wanted/invalid/question/wontfix), looking at
any 'help wanted' issues is a great place to start.

## Licensing

See the [LICENSE](LICENSE) file for our project's licensing. We will ask you to
confirm the licensing of your contribution.

[Gitlab issue tracker]: https://gitlab.com/m0rosan/vault-plugin-secrets-gitlab/-/issues
