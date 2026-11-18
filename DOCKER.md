# Docker

Build the local control-plane image from the repository root:

```sh
docker compose build naeos-core
docker compose up -d naeos-core
```

The service listens on `http://localhost:8081`. To build only the CLI image,
run `docker build -t naeos:local .`. The multi-platform release build is
available through `make docker` and pushes the image, so it is not a local-only
build command.
