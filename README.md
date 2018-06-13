# gpup

A command line tool to upload medium to your Google Photos Library or album.


## Features

- Depending on the official [Google Photos Library API](https://developers.google.com/photos/library/guides/get-started).
- Faster concurrent upload.
- Exponential backoff retry.


## Getting Started

Prepare your environment:

1. Open https://console.cloud.google.com/apis/library/photoslibrary.googleapis.com/
1. Enable Photos Library API.
1. Open https://console.cloud.google.com/apis/credentials
1. Create an OAuth client ID where the application type is other.
1. Set the following environment variables:

```
export GOOGLE_CLIENT_ID=
export GOOGLE_CLIENT_SECRET=
```

To create an album with files:

```sh
gpup --album-name "My Album" *.jpg
```

To add files to your library (not create an album):

```sh
gpup *.jpg
```
