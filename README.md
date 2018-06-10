# gpup

gpup uploads medium to your Google Photos Library.


## Getting Started

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
