# boji

boji is a self-cloud webdav server which aims to be as straightforward and effective as possible.

## Why not owncloud, or nextcloud?

Other self-hosting solutions aim to replicate multi-tenant features common to dropbox or google drive. However, most of the time you never actually use those features; file comments, folder permissions, plugin systems, and other features aren't necessary to solve the immediate problem of "centralized self-cloud storage". And, in the author's experience, the expansion of features leads to a rapid increase in system complexity (requiring multiple databases), and a severe degradation in performance and user satisfaction - while removing the ability for the user to just _access their files normally_.

This is meant to be as practical as possible - serving to accomplish the aims of having centralized personalized cloud storage, while not compromising the fundamental experience in favor of feature creep.

## How do I use it?

Just run the executable, the first positional argument must be the root path to use. By default the BASIC auth is "boji:boji", but this can be changed by setting `BOJI_USER` and `BOJI_PASS` environment variables.

Example;

```
export BOJI_USER=boji2
export BOJI_PASS=boji3
boji /tmp/boji
``` 

By default it runs on port `5157`, but this can be configured with the `-p` flag. Further options are available by just running `boji` with no arguments, or `boji -h`.

## This seems really basic, what other applications should I use with this?

* The author backs up his pictures/screenshots/videos/contacts/etc from his Android phone with [FolderSync](https://play.google.com/store/apps/details?id=dk.tacit.android.foldersync.lite&hl=en_US).
* For day-to-day one-off uploads or downloads, the author uses Linux Mint, which supports the `dav://` protocol in its file browser.
* On Windows, the author maps the `B:` drive of his machine following the [University of Leicester tutorial](https://www2.le.ac.uk/offices/itservices/ithelp/my-computer/files-and-security/work-off-campus/webdav/webdav-on-windows-10)

The author intends to add a very simple web frontend to boji in the near future.

## What does "boji" mean?

 It's a loose transliteration of the word for "duplicate" in Korean (한극: 복제). Korean speakers will probably be horrified at this butchered pronounciation, but it's unique and easy to say.

 ## Errata

 This project is built directly on top of a fork from [hacdias webdav](https://github.com/hacdias/webdav). The author of boji has no affiliation with hacdias.