# Technologies

This document is supposed to give an overview of the technologies that will be used in development of the Karman software.

There is a strict separation between the frontend and the backend of the software. The frontend communicates with the backend via a REST api (see the `api` folder in this repository). In the future there may be additional frontends (e.g. mobile apps) but currently only a web frontend is planned.

## Backend

The backend will be developed using **Python 3** and the [**FastAPI**](https://fastapi.tiangolo.com) framework.

## Frontend

The frontend will be developed using the [**React**](https://reactjs.org) framework. As webserver a simple static fileserver might be used (e.g. nginx) or we might use a [Node.js](https://nodejs.org/en/) server in order to take advantage of server-side-rendering in the future.

## Deployment

There is currently no fixed deployment strategy. However Karman will be developed in a way that makes it easy to run it in containers. For example configuring the application should be possible via the environment and there should be support for standard methods of injecting secrets into the application.