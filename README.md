# gpup [![CircleCI](https://circleci.com/gh/int128/gpup.svg?style=shield)](https://circleci.com/gh/int128/gpup)

A command to upload photos and movies to your Google Photos Library.


## Getting Started

### Setup

You can install this from brew tap or [releases](https://github.com/int128/gpup/releases).

```sh
brew tap int128/gpup
brew install gpup
```

Setup your API access by the following steps:

1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
1. Enable Photos Library API.
1. Open https://console.cloud.google.com/apis/credentials
1. Create an OAuth client ID where the application type is other.
1. Run `gpup` and follow the instruction as follows.

```
% gpup
2018/09/13 15:38:13 Skip reading ~/.gpupconfig: Could not open ~/.gpupconfig: open /user/.gpupconfig: no such file or directory
2018/09/13 15:38:13 Setup your API access by the following steps:

1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
1. Enable Photos Library API.
1. Open https://console.cloud.google.com/apis/credentials
1. Create an OAuth client ID where the application type is other.

Enter your OAuth client ID (e.g. xxx.apps.googleusercontent.com): YOUR_CLIENT_ID.apps.googleusercontent.com
Enter your OAuth client secret: YOUR_CLIENT_SECRET
2018/09/13 15:38:22 Saved credentials to ~/.gpupconfig
2018/09/13 15:38:22 Error: Nothing to upload
```

### Upload files to the library

To upload files in a folder to your Google Photos library:

```
$ gpup my-photos/
2018/06/14 10:28:40 The following 2 files will be uploaded:
  1: my-photos/travel.jpg
  2: my-photos/lunch.jpg
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
And then it uploads files concurrently.

You can specify URLs as well.

```sh
gpup https://www.example.com/image.jpg
```

### Upload files to an album

You can upload files to the album by `-a` option.
If the album does not exist, it will be created.

```sh
gpup -a "My Album" my-photos/
```

You can upload files to a new album by `-n` option.

```sh
gpup -n "My Album" my-photos/
```


## Usage

```
Usage:
  gpup [OPTIONS] <FILE | DIRECTORY | URL>...

Application Options:
  -a, --album=TITLE                 Add files to the album or a new album if it does not exist
  -n, --new-album=TITLE             Add files to a new album
      --request-header=KEY:VALUE    Add the header on fetching URLs
      --request-auth=USER:PASS      Add the basic auth header on fetching URLs
      --gpupconfig=                 Path to the config file (default: ~/.gpupconfig) [$GPUPCONFIG]
      --debug                       Enable request and response logging [$DEBUG]

Options read from gpupconfig:
      --google-client-id=           Google API client ID [$GOOGLE_CLIENT_ID]
      --google-client-secret=       Google API client secret [$GOOGLE_CLIENT_SECRET]
      --google-token=               Google API token [$GOOGLE_TOKEN]

Help Options:
  -h, --help                        Show this help message
```


## Known issues

See [the Google Issue Tracker](https://issuetracker.google.com/issues?q=componentid:385336%20status:open) for the known issues.

### Timestamp

If you upload an photo or movie without timestamp in the header, timestamp of the image will be current time.
Google Photos Library API does not provide setting timestamp for now.

### Photo storage and quality

By using gpup, files are uploaded in original quality, which consumes user's storage. It's the limitation of Photos Library API, as stated in [the offical document](https://developers.google.com/photos/library/guides/api-limits-quotas).

```
All media items uploaded to Google Photos using the API are stored in full resolution at original quality. They count toward the user’s storage.
```

A workaround for this is to use [recover storage feature](https://support.google.com/photos/answer/6220791) of Google Photos, which lets you free up the storage space by converting all the files uploaded in original quality to high quality.

## Contribution

Feel free to open issues or pull requests.
