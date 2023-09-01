#  Welcome to the Karman contributing guide

Thank you for your interest in contributing to the Karman software. This guide will give you an overview of the contribution workflow.

## Issues and Pull Requests

Issues and pull requests are the main way of communicating changes to the codebase. Issues are usually more related to features and ideas whereas pull requests are related to an actual set of changes to the code. Feel free to open issues and start discussions.

If you want to propose your own changes to the code please do open a pull request. Note that we follow [GitHub Flow](https://docs.github.com/en/get-started/quickstart/github-flow) so please read that document to get an overview of the pull request process. We recommend that you name your branches `feature/<name>` where `<name>` is an indication of the feature or change you are working on.

If you are working on a new feature that others might be working on as well, we recommend opening an issue to let others know what you’re up to. This avoids situations where multiple people are working on the same feature without knowing about each other. Likewise before you start working on a feature, search for open issues of people that might have already started.

The issues in this repositories can have multiple flavors indicated by their labels:

- **Proposals** (`proposal`): A proposed new feature for Karman. See below for how to work with proposals.
- **Documentation Issues** (`documentation`): Issues related to the documentation that is being held in this repository (the actual code).
- **Questions** (`question`): Any questions about the Karman software or about the development of the software.
- **Bugs** (`bug`): A bug report.
- **Other** (`other`): Other issues that do not belong in any of the above categories.

### Working with Feature Proposals

The Karman roadmap is managed through proposals. A proposal is in essence just a GitHub issue with the `proposal` label. A proposal can propose any kinds of changes to the Karman API, frontend or both (which usually is the case). The typical lifecycle of a proposal is as follows:

1. The proposal is created and the `proposal` label is assigned. At first the proposal gets assigned the `status/draft` label as well to indicate that there may still be some details to be discussed.
2. After the proposal has been reviewed (which may involve multiple iterations of changes to the proposal) it may either get accepted or rejected which is reflected by the `status/accepted` or `status/rejected` label. If the proposal gets rejected, the issue will be closed.
3. Either immediately after being accepted or some time later the proposal is put on the roadmap. This involves being assigned to a corresponding milestone and the `status/planned` label. At this point there should be corresponding issues in the frontend and backend repositories.
4. The issue may be closed before the implementation is complete, depending on the nature of the proposal. This is usually the case if the proposal is either being tracked somewhere else or if the remaining challenges are purely technical and are represented through other issues.

A proposal with `status/draft` or `status/accepted`  may also receive the `far future` label. This label indicates that although the proposal fits the design, it will probably not be implemented for quite some time. When the proposal reaches the `status/planned` the `far future` label should be removed.

Right now there is no preferred format for a proposal. However, you should make an attempt at describing the proposed change as clearly as you can. Some proposals make take a lot of time to get implemented. The clearer you can describe the feature the lower are the chances that the understanding of the proposal changes during development.

There is an issue template for feature proposals that we recommend you use. The finalized proposal should always be kept in the original issue. If the proposal has to be amended the issue is edited correspondingly. Maintainers may edit proposals to make notes about implications or implementation details. Ideally the issue itself contains only the proposal and all discussions are held in the comments (including any initial request for comments).

### API Changes

When contributing code be aware of the fact, that the public API of the Karman backend cannot change arbitrarily. The API is the interface between the backend and the frontend so any changes need to be properly coordinated between the two. Because of this pull requests that involve API changes usually take longer and need to be reviewed by someone working on the frontend as well.

We recommend that you split your pull requests whenever possible to separate API changes from implementation changes.

## Milestones and Projects

We use milestones and GitHub projects to track the Karman roadmap. For each version there will be a milestone. Proposals and issues get assigned to these milestones to put them on the roadmap for that version.

For some versions there may also be a GitHub project keeping track of the progress until that version is released. The project usually aggregates the issues for a specific milestone from multiple repositories.

## Working with the code

Editing the code should be straight forward. Use your favorite editor or IDE and just start typing.

Before submitting a pull request please make sure that your code passes all tests and adheres to the coding style enforced by `golangci-lint`. These things are checked in CI and your pull request will only be merged if the CI passes.

### Tests

Please ensure that you add appropriate tests alongside your features. We do not aim for a specific coverage but all features should be accompanied by useful tests. When writing your tests, please follow these guidelines:

- Tests should be written using the Go Standard Library. We want to keep test-only dependencies to a minimum.
- When testing database queries or when running integration tests use the `test.NewDB(*testing.T)` function to prepare a clean database environment for the test.
- Tests that rely on external dependencies (such as database) should be annotated with an appropriate build tag. Currently only the `database` tag is used for tests that require a database connection.
- Do not hesitate to include long-running tests but skip those if `testing.Short()` is `true`.
