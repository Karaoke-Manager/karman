# User Entities

This document describes the `users` API entity located at

```http
GET <api-root>/<api-version>/users
```

## Accessing Users

In addition the the standard list and details endpoints a special endpoint `.../users/me` is supported. This special endpoint returns data on the user that is associated with the corresponding request.

# Fields

The following fields are returned in the *entity details*:

| Field         | Meaning                                                      |
| ------------- | ------------------------------------------------------------ |
| `id`          | The ID of the user.                                          |
| `username`    | The unique username of the user.                             |
| `firstName`   | The first name of the user. May be empty.                    |
| `lastName`    | The last name of the user. May be empty.                     |
| `email`       | The E-Mail adress of the user. May be empty.                 |
| `admin`       | A boolean value indicating whether the user has administrator (aka superuser) status. |
| `permissions` | An array of permissions in the form of strings the user has. Only included if the request is authenticated by the respective user. |

The `permissions` array consists of strings identifying the permissions of a user. Permissions that the user does not currently have are absent from the array.

==RFC==: Does this permissions design make sense?

## Example Response

```json
{
  "id": 123,
  "username": "johndoe",
  "firstName": "John",
  "lastName": "Doe",
  "email": "john_doe@example.com",
  "admin": false,
  "permissions": [
      "add_song", "change_song", "delete_song"
  ]
}
```