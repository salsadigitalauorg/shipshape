# End-to-end tests using venom

See https://github.com/ovh/venom

## Running the tests

```sh
docker compose up --detach --build
docker compose exec drupal venom run
```

## Build shipshape && run tests

Create a `docker-compose.override.yml` with
```yaml
version: "3"

services:
  drupal:
    volumes:
      - /path/to/shipshape:/shipshape
```

Build shipshape
```sh
docker compose exec drupal bash -c \
  'cd /shipshape && CGO_ENABLED=0 go build -o /usr/local/bin/shipshape .'
```

Running the tests again from the previous section will now use the latest code.
