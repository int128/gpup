# gpup [![CircleCI](https://circleci.com/gh/int128/gpup.svg?style=shield)](https://circleci.com/gh/int128/gpup)

`gpup` is a command to upload files to your Google Photos library or album.

This depends on the official [Google Photos Library API](https://developers.google.com/photos/library/guides/get-started).


## Getting Started

Setup your API access by the following steps:

1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
1. Enable Photos Library API.
1. Open https://console.cloud.google.com/apis/credentials
1. Create an OAuth client ID where the application type is other.
1. Create `~/.gpupconfig` with the following content:

```yaml
client-id: YOUR_CLIENT_ID
client-secret: YOUR_CLIENT_SECRET
```

You can install this from brew tap or [releases](https://github.com/int128/gpup/releases).

```sh
brew tap int128/gpup
brew install gpup
```

To upload files in a folder to your Google Photos library:

```
$ gpup my-photos/
2018/06/14 10:28:40 The following 2 files will be uploaded:
  1: travel.jpg
  2: lunch.jpg
2018/06/14 10:28:40 Open http://localhost:8000 for authorization
2018/06/14 10:28:43 GET /
2018/06/14 10:28:49 GET /?state=...&code=...
2018/06/14 10:28:49 Saved token to ~/.gpupconfig
2018/06/14 10:28:49 Queued 2 file(s)
2018/06/14 10:28:49 Uploading travel.jpg
2018/06/14 10:28:49 Uploading lunch.jpg
2018/06/14 10:28:52 Adding 2 file(s) to the library
```

It opens the browser and you can log in to the provider.

You can create a new album and upload files into the album by `-n` option.

```sh
gpup -n "My Album" my-photos/
```


## Usage

```
Usage:
  gpup [OPTIONS] FILE or DIRECTORY...

Application Options:
      --gpupconfig=        Path to the config file (default: ~/.gpupconfig) [$GPUPCONFIG]
  -n, --new-album=TITLE    Create an album and add files into it
      --debug              Enable request and response logging [$DEBUG]

Help Options:
  -h, --help               Show this help message
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
