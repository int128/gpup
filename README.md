# gpup

A command line tool to upload medium to your Google Photos Library or album.

This depends on the official [Google Photos Library API](https://developers.google.com/photos/library/guides/get-started).


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


## Known Issues

There are some issues due to Google Photos Library API for now.

If you upload an image without timestamp in the EXIF header, timestamp of the image will be now.
Also timestamp of a movie will be now.
Google Photos Library API does not provide setting timestamp for now.

You cannot control order of uploading items.
Google Photos Library API does not provide ordering items for now.


## Contribution

Feel free to open issues or pull requests.
