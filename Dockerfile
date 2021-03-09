FROM golang:1.13 as build
RUN mkdir -p /build/demo_api
WORKDIR /build/demo_api/

# Force the go compiler to use modules
ENV GO111MODULE=on

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
# COPY go.sum .

# This is the ‘magic’ step that will download all the dependencies that are specified in 
# the go.mod and go.sum file.
# Because of how the layer caching system works in Docker, the  go mod download 
# command will _ only_ be re-run when the go.mod or go.sum file change 
# (or when we add another docker instruction below this line)
RUN go mod download

COPY . .
# RUN go get -v -d ./...

RUN CGO_ENABLED=0 go build -a -installsuffix cgo --ldflags "-s -w" -o /usr/bin/demo_api

FROM alpine:3.9

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build /usr/bin/demo_api /root/
COPY --from=build /build/demo_api/.config.yml /root/

ENV LOG_LEVEL="debug"
ENV LOG_FORMAT="json"

EXPOSE 6060
EXPOSE 80
ENV PORT 80
WORKDIR /root/

CMD ["./demo_api", "start"]