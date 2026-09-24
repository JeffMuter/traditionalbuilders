# Build stage: Nix is not used inside the image; toolchain is installed
# explicitly and pinned to the same versions as shell.nix and CI.
FROM golang:1.25-bookworm AS build

WORKDIR /src

RUN go install github.com/a-h/templ/cmd/templ@v0.3.977

RUN apt-get update \
    && apt-get install -y --no-install-recommends nodejs npm \
    && rm -rf /var/lib/apt/lists/*
RUN npm install -g tailwindcss@3.4.17

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Regenerate artifacts (they are gitignored) before building.
RUN templ generate \
    && tailwindcss -i static/css/input.css -o static/css/output.css \
    && CGO_ENABLED=1 go build -o /out/server ./cmd/server

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# SQLite is a local file opened relative to the working directory, so keep the
# database next to the app and persist it with a bind mount:
#   docker run -v "$PWD/data/traditionbuilders.db:/app/traditionbuilders.db" ...
#
# The full zip code dataset is embedded in the binary. On startup the server
# applies it automatically when the database is already migrated; to apply it
# after migrations without starting the web server, run:
#   docker run ... /app/server -seed-only
WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /src/static /app/static
COPY --from=build /src/db/migrations /app/db/migrations

ENV ADDR=:8080 \
    LOG_LEVEL=info

EXPOSE 8080

ENTRYPOINT ["/app/server"]
