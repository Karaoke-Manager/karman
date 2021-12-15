# Song Entities

Arguably the most important Karman entity is the song. A song represents a single karaoke song (which correlates to a single UltraStar `*.txt` file). Currently multiple *versions* of a song have no special support and are treated as two different songs.

Songs reside at the API endpoint

```http
GET <api-root>/<api-version>/songs
```

## Premissions

Currently all of the `song` API is only available for authenticated users. In the future parts of the API may be available anonymously.

## Fields

The following fields are returned by the API when querying the *entity details* endpoint for a song.

| Field             | Meaning                                                      |
| ----------------- | ------------------------------------------------------------ |
| `id`              | The ID of the song.                                          |
| `title`           | The title of the song.                                       |
| `artist`          | The (main) artist of the song.                               |
| `featuredArtists` | An array of artists that are involved with the song.<br />==RFC==: Does this representation make sense? |
| `year`            | The year the song was released.                              |
| `genre`           | The genre of the song.                                       |
| `duet`            | A boolean value indicating whether the song is a duet.       |
| `lyrics`          | The complete lyrics of the song. For duets this contains the lyrics of the first voice. |
| `lyrics2`         | Only included for duets. Contains the lyrics of the second voice. |
| `duration`        | The duration of the song in seconds. If the song's duration can not be determined for some reason this field is set to `0`. Note that this field is guaranteed to be a number but not guaranteed to be an integer. |
| `artworkUrl`      | A URL pointing to the song’s artwork, or `null` if the song does not have an artwork. |
| `backgroundUrl`   | A URL pointing to the song’s background image, or `null` if the song does not have one. |
| `audioUrl`        | A URL pointing to the song’s audio file, or `null` if the song does not have one. |
| `videoUrl`        | A URL pointing to the song’s background video, or `null` if the song does not have one. |
| `goldenNotes`     | A boolean value indicating whether the song contains golden notes. |
| `verifiedBy`      | The ID of the user that verified the song (or `null` if the song is not verified yet). |

More fields will possibly be added to allow editing song files online such as `gap`, `medleyStart`, etc.


# Attributes

A Song entity has a number of attributes. Attributes are available using a URL of the form

```http
GET /song/<id>/<attribute>
```

| Attribute    | Value                                                        |
| ------------ | ------------------------------------------------------------ |
| `artwork`    | Returns the artwork of the song as an image. The artwork of a song may not exist. You can add the query parameter `?empty=no` in order to substitute missing artworks for a default image. |
| `file`       | Returns the UltraStar compatible `*.txt` file of this song. This resource is only accessible for authenticated users with the appropriate permissions. |
| `audio`      | Returns the audio file of the song (probably a `*.mp3` file). The audio file may not exist. This resource is only accessible for authenticated users with the appropriate permissions. |
| `video`      | Returns the music video file of the song. The video may not exist. This resource is only accessible for authenticated users with the appropriate permissions. |
| `background` | Returns the background image of the song. The background image may not exist. This resource is only accessible for authenticated users with the appropriate permissions. |
| `archive`    | Returns a `*.zip` archive file containing the UltraStar `*.txt` file as well as any associated files (audio, video, artwork, background) of the song. This resource is only accessible for authenticated users with the appropriate permissions. |

## Related Entities

There are some other potential entities that have a strong relationship with Songs:

- **Song Activities**: Song activities are used to keep track of changes made to a song.
- **Song Issues**: Song problems are used for reporting issues with songs (such as synchronization issues, missing artwork, etc.)

## Example Response

```json
{
  "id": 123,
  "title": "Revolt",
  "artist": "Muse",
  "featuredArtists": [],
  "year": 2015,
  "genre": "Rock",
  "duet": false,
  "lyrics": "How did we get in so much trouble?\nGetting out just seems impossible\n...",
  "duration": 246,
  "artwork": true,
  "background": false,
  "audio": true,
  "video": false,
  "goldenNotes": true,
  "verifiedBy": null
}
```