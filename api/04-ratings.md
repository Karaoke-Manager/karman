# Song Ratings Entities

Users can rate songs on a scala from 1 to 10. Song ratings do not inherently have a specific meaning.

Songs reside at the API endpoint

```http
GET <api-root>/<api-version>/songs/<id>/ratings
```

Where `<id>` is the ID of a song. Ratings are always attached to a specific song.

## Fields

The following fields are returned by the API when querying the *entity details* endpoint for a song.

| Field    | Meaning                                                  |
| -------- | -------------------------------------------------------- |
| `userID` | The ID of the user that made the rating.                 |
| `songID` | The ID of the song that is being rated.                  |
| `value`  | The rating of the song (an integer between `1` and `10`) |

## Example Response

```json
{
  "userID": 5,
  "songID": 123,
  "value": 7.5
}
```

## Creating and Updating Ratings

Ratings are created by `POST`ing to the above API endpoint. A rating can be updated via a `PATCH` request or deleted via a `DELETE` request:

````http
PATCH <api-root>/<api-version>/songs/<id>/ratings/<userID>
````

Every user can have at most one rating for a given song. Depending on the permissions of a user it can be possible to change or remove another user’s rating.

