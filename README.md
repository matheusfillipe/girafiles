# GiraFiles
[![Docker Pulls](https://img.shields.io/docker/pulls/mattfly/girafiles.svg)](https://hub.docker.com/repository/docker/mattfly/girafiles/general)

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
- `POST /api/files/` - Upload a file. Response looks like:
    ```json
    {
        "status": "success",
        "url": "http://localhost:8000/files/ufa.png"
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
- `GET /files/ufa.png` - Download a file

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
