FROM golang:1.21.5 AS build-stage

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /generate

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM scratch

WORKDIR  /

COPY --from=build-stage /generate /generate

# Run
ENTRYPOINT [ "/generate" ]