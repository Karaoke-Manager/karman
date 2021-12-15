# The Karman API

At its core the Karman software communicates via a JSON REST API with its backend. This document outlines the design of the API as well as underlying priciples that will define further API endpoints and their behavior. You can find a more in-depth technical documentation on some endpoints in the other files in this directory.

## API Root

The Karman API can be located at any domain (e.g. `karman.example.com`) or path (e.g. `example.com/karman`). Any endpoints referenced in this documentation reside under the API root which can be specified when deploying the application. For example the endpoint `/health` endpoint would be accessible via `http://karman.example.com/api/health` if you have deployed the application under `http://karman.example.com/api`.

## API Versioning

The Karman API is explicitly versioned, meaning that every request must contain the API version that should be used. The API version is included in the path:

```http
GET <api-root>/v0.1/...
```

For simplicity the API version is omitted in the conceptual documentation.

## URL Schema

The typical way to interact with the Karman API is via the standard REST endpoints in the following formats:

- **Entity List**: `/<entities>/` returns a paginated list of `entity` (possibly filtered)
- **Entity Details**: `/<entities>/<id>` returns details for a single entity with the specified id

## Pagination

The *entity list* endpoints `/<entities>` return a (possibly paginated) list of all entities of the respective type. The returned JSON has the following format:
```json
{
    "total": 123,
    "offset": 50,
    "items": [

    ]
}
```
| Field    | Meaning                                                      |
| -------- | ------------------------------------------------------------ |
| `total`  | The total number of entities in the list.                    |
| `offset` | The index of the first result. An offset of `0` means that the first element is actually the first element. In the above example the items would be item `50`, `51`, … |
| `items`  | The array of entities on this page.                          |

Every list of entities can be paginated using query parameters:
- `/...?size=<count>` sets the maximum number of returned entities per page. The default value is `20`. Note that the last page may disrespect this parameter and return slightly more entities than specified.
- `/api/...?offset=<page>` specifies the index of the first result. Negative indices are invalid. If you specify a page index greater than the number of available items an empty list is returned. `0` is always a valid page index (although the resulting list might still be empty).

## Searching

Searching is implemented on *entity lists* via the query parameter `search`. Any searchable list of entities can be searched using this query parameter. The implementation of the search however depends on the entity. For example to search all songs one would use the URL `/api/songs?search=<term>` where `<term>` is the search term. The search query is included in the returned response as a `query` field. An empty search term generally returns an empty list of entities.

==RFC==: It this kind of search implementation too simple for our purposes? E.g. do we need to implement search that is restricted to searching only song’s titles or only artists?

## Filtering

*Entity lists* can filtered in order to scope lists to specific criteria. E.g. one might want to find all songs by a specific artist. This can be done using the `/songs/` endpoint filtered using that artist.

Currently there is no plan on how a filter should be implemented.

==RFC==: How could such a filter be implemented?

## Limiting Returned Fields

Sometimes you are only interested in very specific fields of an entity or of multiple entities and do not care about details. In those cases the response can efficiently be truncated to only include explicitly specified fields. Those fields are specified using the `?include=...` query parameter. The value for the parameter is a comma separated list of fields that you want to be included in the response. This parameter can be applied to *entity detail* endpoints as well as to *entity list* endpoints. For list endpoints the parameter is applied to the respective entities (and not to the list wrapper).

The `include` query parameter implicitly excludes fields not specified by the parameter. Note the API may still decide to include fields not specified in the `include` parameter.

## Authentication

The API supports HTTP Basic Auth for authentication. Not all requests need to be authenticated.

==RFC==: Should we support session based authentication for browsers?

==RFC==: Would OAuth2 be a sensible choice for authenticating Mafiasi users?