# GiraFiles
[![Docker Pulls](https://img.shields.io/docker/pulls/mattfly/girafiles.svg)](https://hub.docker.com/repository/docker/mattfly/girafiles/general)
[![Live at](https://img.shields.io/badge/Demo%20at-filebin.cloud.mattf.one-007ACC)](https://filebin.cloud.mattf.one/)

Darn minimal filebin with preview and API.

Everything is stored in local disk. No database is used. No fancy things, just files in the filesystem.

## Features
- Upload files
- Simple API
- Avoids duplication
- Preview images in browser
- Automatic deletion of files after a certain time or when storage limit is reached
- Optional Basic Auth
- Limited Customization

## API
- `POST /api/` - Upload a file. Response looks like:
    ```json
    {
        "status": "success",
        "url": "http://localhost:8000/ufa.png"
    }
    ```
    Or
    ```json
    {
        "status": "error",
        "message": "File too large"
    }
    ```
    Or
    ```json
    {
        "status": "error",
        "message": "Hourly rate limit exceeded"
    }
    ```
- `GET /ufa.png` - Download or preview a file

## Usage
You can clone this repository and run it with:
```bash
go run main.go
```

Alternatively with docker:
```bash
docker run -d -p 8000:8000 -it mattfly/girafiles
```

## Configuration
You may want to set `STORE_PATH` to have permanency between restarts.

See [.env.example](.env.example) for configuration options. You can use the same ones listed there
as environment variables.


## Disclaimer
This project is meant for quickly allowing files to be shared and previewed with them only lasting for
a very short period of time. It is not meant to be a permanent file storage solution. So keep in mind:

1. Toy project warning. Very little testing has been done.
2. This is 100% md5 hash collision vulnerable, meaning someone can replace your file.
3. No encryption is used for the files.
4. There is no privacy for the files. Anyone could easily guess valid url's. I wanted them to be short, not secure.
5. Running multiple instances in the same `STORE_PATH` might work, but it's not tested.
6. Code sucks because I'm not a Go developer.
7. I do not have the need to fix any of the above myself but if you do PR's are welcome.
