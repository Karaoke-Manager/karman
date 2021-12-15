# Workflow

This document details how the Karman project manages its projects and what code workflows will be used. Note: This is unrelated to any continuous integration that projects may have set up.

## Project Management

The Karman project is managed mainly through GitHub issues and pull requests. There are two main rules to follow:

- Every idea should be accompanied by a GitHub issue
- Every code change should be done through a reviewed pull request

The rationale behind this approach is that every idea can be discussed conceptually in its issue and every code change can be discussed and reviewed in its pull request. This avoids unneccessary work on ideas that will not be implemented and allows us to use the tools GitHub offers to help with this process.

There are three main repositories on which issues can be opened:

- The **[concept](https://github.com/Karaoke-Manager/concept)** repository contains conceptual documentation. Here we mainly discuss the API design and the feature set of Karman. Pull requests usually create or modify markdown files that describe different aspects of the project or the Karman API.
- The **[frontend](https://github.com/Karaoke-Manager/frontend)** repository contains the web frontend of Karman. Issues in this repository should address features or bugs in the frontend whose solution do not require the Karman API to be modified.
- The **[backend](https://github.com/Karaoke-Manager/backend)** repository contains the python backend of Karman. Issues in this repository should address features or bugs in the backend whose solution do not require the Karman API to be modified.

For new features the typical workflow is as follows:

1. The API design is discussed in the **concept** repository. This may involve multiple issues and pull requests.
2. After the API design has been decided on issues should be created in the **frontend** and **backend** related to the implementation of the respective change.
3. Only after all issues related to the original concept/idea have been closed the feature can be considered implemented.

In addition to GitHub issues we may use GitHub projects to keep track of the progress across multiple repositories.

## Development Workflow

### Background

As a starting point I recommend reading the blog post [A successful Git branching model](https://nvie.com/posts/a-successful-git-branching-model/) as it explains many of the challenges that a good source control workflow can solve. The workflow described in that post became to be known as **Git Flow**.

### Our Workflow

We actually follow a simplified version of Git Flow called [**GitHub Flow**](https://docs.github.com/en/get-started/quickstart/github-flow). For a full description of the workflow I recommend reading GitHub’s documentation. The workflow can be summarized as follows:

1. The `main` branch contains the most current development version of the software. Everyone should make an effort that the `HEAD` of the `main` branch can be built and run without errors.
2. Any changes are made by creating a so called **feature branch**. Typically these branch off from the `main` branch (but you can branch off of another feature branch as well). A feature branch should be scoped to a single issue (or sometimes only part of an issue) and thus be relatively small. Despite the name feature branch these kinds of branches should be used not only for new features but for enhancements, bug fixes and really any kind of change. A feature branch for a feature `<name>` is usually called `feature/<name>`.
3. When a feature is *complete* (as defined by its author), a **pull request** is made to merge the changes into the `main` branch. Now one of the maintainers will review the changes and give feedback. At this point several things can happen:
   1. If your pull request does not work (the CI fails or someone discovers an issue) you should fix that issue by pushing respective fixes to the feature branch.
   2. You might receive comments about functionality, code style, … Usually these must be incorporated before the pull request can be merged.
   3. Your pull request might be **closed** (read: denied). If for some reason the maintainers decide that the pull request should not be merged into the `main` branch the request might be closed. Usually you will receive a comment about why the request was denied.
   4. Your pull request might be **merged** (read: accepted). In that case congratulations.
4. New versions of a software are defined using **git tags**. Each version is identified by a tag on the `main` branch and may be accompanied by a GitHub release.
5. In contrast to Git Flow, releases and hotfixes are treated as feature branches.