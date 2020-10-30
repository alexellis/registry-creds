# Build the manager binary
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  GO111MODULE=on go build -ldflags="-w -s" -a -o /usr/bin/controller

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /usr/bin/controller .
USER nonroot:nonroot

ENTRYPOINT ["/controller"]
