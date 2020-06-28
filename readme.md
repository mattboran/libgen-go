# Libgen

Libgen is a Cobra-based CLI application for querying and downloading books from Library Genesis.

## Building and Running 

To build this application, clone this repository and run `make build`. To run this application, run `make run` or execute the resulting executable `./bin/libegen`. You must have `golang` installed.

## Available Commands

### Article

Search for a scientific article on Library Genesis.

```
  libgen article [string to search for] [flags]
```

#### Flags

- `page` - Page number to query for. Default 1.

---

### Fiction

Search for a fiction book on Library Genesis.

```
libgen fiction [string to search for] [flags]
```
#### Flags
- `criteria` - Search criteria. Can be `author`, `title`, `series`. Default any.
- `format` - Ebook format. Can be `epub`, `mobi`, `azw`, `azw3`, `fb2`, `pdf`, `rtf`, `txt`. Default any.
- `page` - Page number to query for. Default 1.

---

### Textbook

Search for a textbook on Library Genesis.

```
libgen textbook [string to search for] [flags]
```
#### Flags
- `criteria` - Search criteria. Can be `author`, `title`. Default any.
- `page` - Page number to query for. Default 1.
- `sort` - Sort results by this field. Can be `author`, `title`, `publisher`, `year`, `pages`, `language`, `id`, `extension`, `size`. Default `title`. 
- `reverse` - Sort in descending order instead.

---

### Dl

Set default download path.

```
libgen dl
```

You will be prompted for a default download location. The result is stored in `$HOME/.libgen.yaml`. This config location can be modified by the following command:

```
libgen --config <path/file>
```

---

#### Disclaimer

All information provided on this website is produced strictly for educational purposes. We do not condone piracy and are not responsible for how you decide to use the information provided. This application is intended only to search for and download content that is in the public domain.

We do not have any control over the links on Library Genesis. If you see any form of infringement, please contact appropriate media file owners or host sites immediately. [DMCA Legislation](https://www.copyright.gov/dmca-directory/)