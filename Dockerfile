# Stdio MCP server for Glama / local Docker use.
# Build: docker build -t behave .
# Run:   docker run --rm -i behave
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/behave ./cmd/behave

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/behave /usr/local/bin/behave
ENTRYPOINT ["behave", "mcp"]
