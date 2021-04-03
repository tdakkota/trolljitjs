# Create buidler image.
FROM golang:latest AS builder

ENV GOFLAGS='-trimpath -mod=readonly'
ENV CGO_ENABLED=0
ENV GO_EXTLINK_ENABLED=0

# Set the Current Working Directory inside the container.
WORKDIR /x

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Build all dependencies. Build cache will be cached if the go.mod and go.sum files are not changed.
RUN go build $(go list -f '{{if and (not .Main) (not .Indirect)}}{{.Path}}/...{{end}}' -m all)
# Copy source code and build.
COPY . .
RUN go build -v -o app ./


# Create runtime image.
FROM alpine:latest
# Set the Current Working Directory inside the container.
WORKDIR /x

# Setup certificates.
RUN apk --no-cache add ca-certificates
# Copy the Pre-built binary file from the previous stage.
COPY --from=builder /x/app .
ENTRYPOINT ["./app"]