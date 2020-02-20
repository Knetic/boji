# boji

boji is a **self-cloud webdav server** which aims to be as straightforward and effective as possible.

* Leaves everything as readable and writeable as it found it
* Doesn't use databases
* Doesn't litter your system with lockfiles or temp directories
* TLS support
* No hassle with users or config files
* Optionally transparently compresses directories
* Optionally transparently encrypts your files

WebDAV is supported by all major OS's, and there are a ton of client apps/libraries that interface with it. By hosting your files with a webdav server, you're giving yourself access to mount them as a regular filesystem from any device.

## Why not owncloud, or nextcloud?

Other self-hosting solutions aim to replicate multi-tenant features common to dropbox or google drive. However, most of the time you never actually use those features; file comments, folder permissions, plugin systems, ACL's, and other features aren't necessary to solve the immediate problem of "centralized self-cloud storage". And, in the author's experience, the expansion of features leads to a rapid increase in system complexity (requiring multiple databases), and a severe degradation in performance and user satisfaction - while removing the ability for the user to just _access their files normally_.

`boji` is meant to be as practical as possible - serving to accomplish the aims of having centralized personalized cloud storage, while not compromising the fundamental experience in favor of feature creep.

## This seems really basic, what other applications should I use with this?

* The author backs up his pictures/screenshots/videos/contacts/etc from his Android phone with [FolderSync](https://play.google.com/store/apps/details?id=dk.tacit.android.foldersync.lite&hl=en_US).
* The author uses the [Enpass](https://www.enpass.io/) password manager, syncing password vault file through boji.
* The author browses his cloud storage on Android with [Cx](https://play.google.com/store/apps/details?id=com.cxinventor.file.explorer&hl=en_US)
* For day-to-day one-off uploads or downloads, the author uses Linux Mint, which supports the `dav://` protocol in its file browser.
* On Windows, the author maps the `B:` drive of his Windows machine following the [University of Leicester tutorial](https://www2.le.ac.uk/offices/itservices/ithelp/my-computer/files-and-security/work-off-campus/webdav/webdav-on-windows-10)

## How do I use it?

Just run the executable. 

The root path to use must be specified with the `-r` flag.
By default the BASIC auth is "boji:boji", but this can be changed by setting `BOJI_USER` and `BOJI_PASS` environment variables. 

Example;

```
BOJI_USER=boji2 \
BOJI_PASS=boji3 \
boji -r /tmp/boji
``` 

By default it runs on port `5170`, but this can be configured with the `-p` flag. Further options are available by just running `boji` with no arguments, or `boji -h`.

The author recommends using a docker-compose file that looks like this;

```
version: '2'

services:
  boji:
    image: knetic/boji:2.0
    container_name: boji
    command: boji -r /mnt/boji
    restart: always
    ports:
    - 5170:5170
    environment:
    - BOJI_USER=boji
    - BOJI_PASS=boji
    volumes:
    - /var/lib/boji/data:/mnt/boji:z
```

## Transparent compression

`boji` can read an `archive.zip` from any directory, and serve them as if they weren't zipped. This allows large directories of uncompressed files to be compressed at rest, but still accessed normally. Reads, writes, renames, copies, deletes, and all other calls are handled normally in archived and unarchived directories.

archive zips must only contain one level of files. They do not need to be written by this system, but there's not a lot of reason not to do so.

`POST`ing to a valid path, with the querystring `compress=true`, will cause the server to compress all files in that directory into a single `archive.zip`.
`POST`ing to any archived path with the querystring `compress=false` will unzip all files, and remove the archive.

It's recommended to only compress directories that are written infrequently.

## TLS

If given a path to appropriate key/cert files, `boji` can run over TLS ("davs" protocol). Specify the `-c` and `-k` flags, and the system will run on TLS. If not specified, the system will work over plain HTTP ("dav" protocol). Authentication is unchanged, but TLS is recommended because it encrypts all communications - especially usernames and passwords.

## Transparent encryption

When a client connects to `boji` and provides a valid admin password, a symmetric encryption key can also be provided. If provided, boji can use it to transparently read and write encrypted files in any directory it serves. Reads, writes, renames, copies, deletes, and all other calls are handled normally in encrypted and unencrypted directories, but on disk, the contents will be encrypted. The key given by the user is _not stored_, so the user can specify different keys for different files, and must remember the key themselves.

From a user point of view, just add a colon and the key after your password when logging in to provide the key.

This should result basic auth which is formed in this fashion; `Authorization: Basic <username>:<password>:<key>` wherein the username/password/key are base64-encoded as usual.

`POST`ing to a valid path, with the querystring `encrypt=true`, will cause the server to encrypt all files in that directory (although their names and sizes are not obfuscated).
`POST`ing to any encrypted path with the querystring `encrypt=false` will unencrypt all the files, and leave them that way.

The on-disk effect of this is that all files that are to be encrypted will _not_ be written to the name specified, and instead to that name postfixed with `.pgp-boji`. Any reads to any file in boji will check to see if such a file exists, and try to decrypt it, before checking for a file of the given name.

_Unlike compression_, encryption is not a "setting" applied to a directory. It is dependent on the user specifying a key with their credentials. If no key is provided, no decryption or encryption can occur, and everything will be read and written plaintext. Any file written by a request that includes a key _will be encrypted_.

If encryption and compression are both specified for a directory, the archive will be made _and then encrypted_, rather than each file being encrypted before being compressed.

If the user does not provide a key when requesting reads or lists, boji will not serve or list any encrypted files whatsoever. They will not exist, as far as the user is concerned.

If the user provides an incorrect key when requesting reads or lists, boji will return an error to the user, not serve garbage data.

The encryption is PGP, with AES-256 as the cipher. The user does not need to use boji to decrypt the files - files are written to disk such that `gpg` or other pgp tools are able to manipulate them normally. Encrypted files are stored with a suffix of `.pgp-boji`, which indicates that it's a pgp archive.

## What does "boji" mean?

 It's a loose transliteration of the word for "duplicate" in Korean (한극: 복제). Korean speakers will probably be horrified at this butchered pronounciation, but it's unique and easy to say.

 ## Errata

 This project is technically a fork from [hacdias webdav](https://github.com/hacdias/webdav). The author of boji has no affiliation with hacdias, and the resultant code for `boji` doesn't actually use anything from that project. It's a fork purely as an anomaly of history.