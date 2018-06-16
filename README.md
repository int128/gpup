# gpup [![CircleCI](https://circleci.com/gh/int128/gpup.svg?style=shield)](https://circleci.com/gh/int128/gpup)

`gpup` is a command to upload files to your Google Photos library or album.

This depends on the official [Google Photos Library API](https://developers.google.com/photos/library/guides/get-started).


## Getting Started

Setup your API access by the following steps:

1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
1. Enable Photos Library API.
1. Open https://console.cloud.google.com/apis/credentials
1. Create an OAuth client ID where the application type is other.
1. Set the following environment variables:

```
export GOOGLE_CLIENT_ID=
export GOOGLE_CLIENT_SECRET=
```

Download the latest release from [releases](releases).

To upload files in a folder to your Google Photos library:

```
$ gpup my-photos/
2018/06/14 10:28:40 The following 2 files will be uploaded:
  1: travel.jpg
  2: lunch.jpg
2018/06/14 10:28:40 Open http://localhost:8000 for authorization
2018/06/14 10:28:43 GET /
2018/06/14 10:28:49 GET /?state=...&code=...
2018/06/14 10:28:49 Storing token cache to /home/user/.gpup_token
2018/06/14 10:28:49 Queued 2 file(s)
2018/06/14 10:28:49 Uploading travel.jpg
2018/06/14 10:28:49 Uploading lunch.jpg
2018/06/14 10:28:52 Adding 2 file(s) to the library
```

Only first time you need to open browser for authorize API access.
`gpup` will store the token to `~/.gpup_token` and use in the next time.

You can create a new album and upload files into the album.

```sh
gpup -n "My Album" my-photos/
```


## Usage

```
Usage:
  gpup [OPTIONS] FILE or DIRECTORY...

Setup:
1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
2. Enable Photos Library API.
3. Open https://console.cloud.google.com/apis/credentials
4. Create an OAuth client ID where the application type is other.
5. Export GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET variables or set the options.

Application Options:
  -n, --new-album=TITLE               Create an album and add files into it
      --oauth-method=[browser|cli]    OAuth authorization method (default: browser)
      --google-client-id=             Google API client ID [$GOOGLE_CLIENT_ID]
      --google-client-secret=         Google API client secret [$GOOGLE_CLIENT_SECRET]

Help Options:
  -h, --help                          Show this help message
```


## Caveats

At this time there are some limitations due to Google Photos Library API.

If you upload an image without timestamp in the EXIF header, timestamp of the image will be current time.
Also timestamp of a movie will be current time.
Google Photos Library API does not provide setting timestamp for now.

You cannot control order of uploading items.
Google Photos Library API does not provide ordering media items for now.


## Contribution

Feel free to open issues or pull requests.
