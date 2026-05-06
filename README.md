# Immich SmartAlbum

Immich SmartAlbum is a standalone implementation of smart albums for Immich.
Its job is to automatically add named people to an album, because Immich lacks this ability natively.
It's written in Go and runs as a standalone binary or as a Docker container.

## Prerequisites
- Go 1.26
- An available Immich server
- Immich API keys for the user's albums you want to manage
- Docker (optional)

## API Keys

The API keys need the following permissions:
* asset.read
* album.read
* albumAsset.create
* person.read

## Installation

### Docker

```
docker run --rm -v $(pwd)/config.yaml:/config.yaml ghcr.io/nsheridan/immich-smartalbum:latest
```

Or with Docker Compose, copy [`compose.yaml`](compose.yaml) from this repo and run:

```
docker compose up -d
```

### Installation using `go` tools

```
go install nsheridan.dev/immich-smartalbum@latest
```

```
immich-smartalbum -h
Usage of immich-smartalbum:
  -config string
        path to config file (default "config.yaml")

immich-smartalbum -config /etc/immich-smartalbum/config.yaml
```

### Configuration

The program is configured with a yaml file with the following structure:

```yaml
# server: Your Immich server hostname
server: https://your-immich-server.example.com
# interval: How often to search for new photos, must be a valid Go duration.
interval: 1h

# Optional: log_level controls verbosity. Valid values: info (default), debug.
log_level: info

users:
  - name: alice
    api_key: your-api-key-here
    albums:
      - name: "Family"
        # If specified, album_id finds the album by ID instead of looking it up by name.
        # Required when the album is shared and owned by another user.
        # Find the ID in the album's URL in the Immich web UI.
        album_id: "abcdef-abcdef-0000"
        people:  # If any of the named people are present in the photo, the photo will be added to the album
          - Alice
          - Bob
          - Carol
      - name: "Friends"  # Without album_id the album is found by name.
        people:
          - Dan
          - Frank
  - name: bob
    api_key: bobs-api-key
    albums:
      - name: "Family"
        album_id: "abcdef-abcdef-0000"
        people:
          - Alice
          - Bob
          - Carol
```

Note that you can specify multiple users, each with their own API key. 

## Development

### Running tests

```
go test ./...
```

### Running locally

```
go build .
./immich-smartalbum [-config /path/to/config.yaml]
```

```
docker build -t immich-smartalbum .
docker run --rm -v $(pwd)/config.yaml:/config.yaml immich-smartalbum
```

Or with Docker Compose:

```
docker compose up
```

The `compose.override.yaml` will build the image locally instead of pulling.
