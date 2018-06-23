# boji

(built on top of a fork from [https://github.com/hacdias/webdav](hacdias webdav))

boji is a self-cloud server which aims to be as straightforward and effective as possible.

## Why not owncloud, or nextcloud?

Other self-hosting solutions aim to replicate multi-tenant features common to dropbox or google drive. However, most of the time you never actually use those features; file comments, folder permissions, plugin systems, and other features aren't necessary to solve the immediate problem of "centralized self-cloud storage". And, in the author's experience, the expansion of features leads to a rapid increase in system complexity (requiring multiple databases), and a severe degradation in performance and user satisfaction - while removing the ability for the user to just _access their files normally_.

This is meant to be as practical as possible - serving to accomplish the aims of 

## How do I use it?

Just run the executable, the first positional argument must be the root path to use. By default the BASIC auth is "boji:boji", but this can be changed by setting `BOJI_USER` and `BOJI_PASS` environment variables.

Example;

```
export BOJI_USER=boji2
export BOJI_PASS=boji3
boji /tmp/boji
``` 

By default it runs on port `5157`, but this can be configured with the `-p` flag. Further options are available by just running `boji` with no arguments, or `boji -h`.

## What does "boji" mean?

 It's a loose transliteration of the word for "duplicate" in Korean (한극: 복제). Korean speakers will probably be horrified at this butchered pronounciation, but it's unique and easy to say.