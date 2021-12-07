# Karman Concept Documents
Conceptual documents on the development and the roadmap of Karman.

## Where to begin?

In order to get an overview of the Karman Software you should continue reading this readme. Below we will explain the idea behind the software and what problems we hope it will solve. If you want to dive deeper into the technical details see the other files and folders in this repository:
- **API Design** (```api``` folder): This folder is used to develop an API specification for the Karman backend so that any clients can rely on a consistent API.
- **Technologies** (`technologies.md` file): This document summarizes the technologies that will be used for the Karman software.

## Motivation

Everyone who has tried to manage an UltraStar karaoke library has probably experienced how difficult it can be to ensure that all songs in the library adhere to a certain quality standard. Most libraries contain low-bitrate audio, difficult file encodings, low-res or even missing artwork, incorrectly synchronized songs and a multitude of other problems. The **Karman** [ˈkaɾmɛ̃n] project aims to make it easier to manage large karaoke libraries.

## Karman Overview

The most important aspect of the architecture of Karman is that it takes full control over your entire karaoke library. When using Karman you will not be moving songs into specific folders but instead you will use the import feature of the software. When importing songs, Karman does a lot of things for you:

- Validate that your files actually contain an UltraStar song
- Scan the songs for obvious problems (low bitrate audio, missing files, …)
- Rename files to a standard naming scheme
- Discard unused files
- …

From this point on Karman manages the karaoke files.

Users interact with their libraries through the Karman web interface which allows them to edit and delete their songs. Karman ensures that the files on the filesystem stay in sync with its internal state.

When singing karaoke all you need to do is point UltraStar to the location of your Karman library.

## Additional Features

In the beginning Karman will only support importing songs. However there are many features that can be implemented to make Karman more effective at managing libraries. Some of these features include:

- Allow for automatically fixing common issues (e.g. encoding issues, overlapping syllables, …)
- Include a live editor for songs that allows for pitch correction, synchronization, …
- Include an audio/video editor for media files.
- A possibility to suggest new songs
- Per-User workspaces where users can develop their own songs and then submit them for review for inclusion in the main library. Possibly a feature where multiple users can coordinate who works on which song suggestion.
- An option to automatically search for song artwork
