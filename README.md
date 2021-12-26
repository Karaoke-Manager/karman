# Karman – The Karaoke Manager

[![Documentation Status](https://readthedocs.org/projects/karman/badge/?version=latest)](https://karman.readthedocs.io/en/latest/?badge=latest)

The **Kar**aoke **Man**ager **Karman** [ˈkaɾmɛ̃n] is a software that manages your UltraStar song library. This repository does not contain any code of the Karman software but serves as a starting point for users and developers that want to contribute to the software.

## Motivation

Everyone who has tried to manage an UltraStar karaoke library has probably experienced how difficult it can be to ensure that all songs in the library adhere to a certain quality standard. Most libraries contain low-bitrate audio, difficult file encodings, low-res or even missing artwork, incorrectly synchronized songs and a multitude of other problems. The **Karman** [ˈkaɾmɛ̃n] project aims to make it easier to manage large karaoke libraries.

## Karman Overview

Karman is an application running on a server that takes full control over your entire karaoke library. When using Karman you will not be moving songs into specific folders but instead you will use the import feature of the software to upload songs to the server. The software then manages the library and moves files into the right places. When you want to sing you can either sync the songs to your machine or if your network connection is fast enough you can just mount the respective folder.

In contrast to existing software (most notably the [UltraStar Manager](https://github.com/UltraStar-Deluxe/UltraStar-Manager)) Karman takes a few different approaches:

- Karman runs server-side and expects to be the only application or user with write access to the song library.
- Karman has the concept of a library you import songs into. Cleaning up metadata of songs and correcting common errors usually takes place on import, not at some arbitrary time.
- Karman can be easily extended to implement more complex optimizations for songs
- Karman includes a library explorer that allows (potentially unauthenticated) users to browse the library. This can be very useful when singing karaoke and people want to know which songs are available.

In the future Karman might do even more, eventually integrating a sophisticated song editor and intelligent library management

## Karman Architecture

Karman is designed to be compatible with modern development and deployment strategies. At its core the Karman software consists of the **Karman API Backend**. The backend provides a well-documented REST API through which clients can read data and make changes. The backend implements the song management logic and most of the import logic.

The user interface is implemented via a **Web Frontend**. The frontend provides the user interface of the software. It communicates with the backend exclusively via the documented REST API.

## Where do I go now?

Currently there are three repositories that coud use your contribution. Depending on your background and what you want to do you should go to one of these repositories and start with that project’s README.md or their CONTRIBUTING.md.

- **Karman** (this repo): This repository contains pretty much no code. Instead is is used as a high-level issue tracker for features and the roadmap of Karman. This is the right place to discuss major changes and completely new features for the software. This is also the place to ask questions about the software. Most communication in this repository happens through issues. Read the [CONTRIBUTING](CONTRIBUTING.md) guide for more information.
- [**Frontend**](https://github.com/Karaoke-Manager/frontend): This repository contains the web frontend of Karman. It is developed using React and the [Next.js](https://nextjs.org) framework. This is the right place if you want to improve the end user experience. Pull requests usually contain some changes to the design or UX of the webapp.
- [**Backend**](https://github.com/Karaoke-Manager/backend): This repository contains both the Karman API as well as the code that acutally interfaces with the UltraStar files. The backend is written in Python using the [FastAPI](https://fastapi.tiangolo.com) framework. This is the right place if you want to start introducing new functionality or want to fix bugs. Pull requests either introduce changes to the Karman API or are concerned with the implementation of the API endpoints. Read the [CONTRIBUTING](https://github.com/Karaoke-Manager/backend/blob/main/CONTRIBUTING.md) guide for more information.

## Frequently asked questions

### Why is Karman a server-side application?

Karman is intended for users with huge libraries consisting of hundreds of gigabytes of songs. These kinds of libraries cannot feasibly be managed manually by a single person. There were multiple factors involved in the decision of going server side: 1. Storage. Growing libraries with thousands of songs quickly exceed the free storage on typical PCs. 2. Backups. By storing songs on a server it is easy to include them in periodic backups that protect your valuable songs. 3. Multi user support. Large libraries often come into existence when many users merge their smaller libraries. We want to offer some kind of collaborative feature that makes it possible to work on new songs together. 4. Public index. When singing it is often very useful to be able to directo users to a website where they can search their favorite songs. This would not be possible without a server.



// TODO: More questions, maybe?
