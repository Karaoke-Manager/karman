#  Welcome to the Karman contributing guide

Thank you for your interest in contributing to the Karman software. This guide will give you an overview of the contribution workflow.

This document describes the contribution workflow for this repository. If you want to contribute to the actual implementation of Karman, see the [Frontend](https://github.com/Karaoke-Manager/frontend) or [Backend](https://github.com/Karaoke-Manager/backend) repositories instead.

## Code and Pull Requests

Although this is a git repository it currently contains very litte code. This is intentional. Instead we use this repository as an issue tracker for the Karman software. Because of the design of the software the more technical issues and PRs are managed in their own repositories.

The little code that does exist in this repository serves as a high-level overview of the Karman software, mainly directed at developers that want to contribute.

## Working with Issues

The issues in this repositories can have multiple flavors indicated by their labels:

- **Proposals** (`proposal`): A proposed new feature for Karman. See below for how to work with proposals.
- **Documentation Issues** (`documentation`): Issues related to the documentation that is being held in this repository (the actual code).
- **Questions** (`question`): Any questions about the Karman software or about the development of the software.
- **Other** (`other`): Other issues that do not belong in any of the above categories.

### Working with Feature Proposals

The Karman roadmap is managed through proposals. A proposal is in essence just a GitHub issue with the `proposal` label. A proposal can propose any kinds of changes to the Karman API, frontend or both (which usually is the case). The typical lifecycle of a proposal is as follows:

1. The proposal is created and the `proposal` label is assigned. At first the proposal gets assigned the `status/draft` label as well to indicate that there may still be some details to be discussed.
2. After the proposal has been reviewed (which may involve multiple iterations of changes to the proposal) it may either get accepted or rejected which is reflected by the `status/accepted` or `status/rejected` label. If the proposal gets rejected, the issue will be closed.
3. Either immediately after being accepted or some time later the proposal is put on the roadmap. This involves being assigned to a corresponding milestone and the `status/planned` label. At this point there should be corresponding issues in the frontend and backend repositories.
4. The issue may be closed before the implementation is complete, depending on the nature of the proposal. This is usually the case if the proposal is either being tracked somewhere else or if the remaining challenges are purely technical and are represented through other issues.

A proposal with `status/draft` or `status/accepted`  may also receive the `far future` label. This label indicates that although the proposal fits the design, it will probably not be implemented for quite some time. When the proposal reaches the `status/planned` the `far future` label should be removed.

### Anatomy of a proposal

Currently there is no preferred format for a proposal. However you should make an attempt at describing the proposed change as clearly as you can. Some proposals make take a lot of time to get implemented. The clearer you can describe the feature the lower are the chances that the understanding of the proposal changes during development.

There is an issue template for feature proposals that we recommend you use. The finalized proposal should always be kept in the original issue. If the proposal has to be amended the issue is edited correspondingly. Maintainers may edit proposals to make notes about implications or implementation details. Ideally the issue itself contains only the proposal and all discussions are held in the comments (including any initial request for comments).

## Milestones and Projects

We use milestones and GitHub projects to track the Karman roadmap. For each version there will be a milestone. Proposals and issues get assigned to these milestones to put them on the roadmap for that version.

For some versions there may also be a GitHub project keeping track of the progress until that version is released. The project usually aggregates the issues for a specific milestone from multiple repositories.

## Building the documentation

The documentation in this repository is written in ReStructuredText and can be compiled into various formats using Sphinx. The documentation is automatically generated and published on [ReadTheDocs](https://karman.readthedocs.io/). During development the documentation can be built locally as follows:

```shell
# Create a virtual environment and activate it. If you are using an IDE such as PyCharm this can also be done by configuring a project interpreter and using the integrated shell.
python3 -m venv .venv
source .venv/bin/activate

# Install the required dependencies
pip3 install -r docs/requirements.txt
# Install dev requirements if needed
pip3 install -r docs/requirements-dev.txt

# Build the documentation
make html

# Start a development server with autoreloading (dev requirements needed)
make serve
```

