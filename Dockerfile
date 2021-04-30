# docker build -t paulfdunn/rest-app:v0.0.0 .
# docker login --username paulfdunn
# docker push paulfdunn/rest-app:v0.0.0
# docker run --name prod1 --network host -it prod1 /bin/ash
# docker container stop prod1
# docker container rm prod1

###########################################################################
# 2 stage with Alpine - image size: 56MB
###########################################################################
FROM golang:alpine as builder
COPY ./ /go/src/github.com/paulfdunn/rest-app-prod1
WORKDIR /go/src/github.com/paulfdunn/rest-app-prod1
RUN go mod download
#RUN go test -v ./... >test.log 2>&1 
# https://github.com/mattn/go-sqlite3#alpine
RUN apk add --update gcc musl-dev
RUN GOOS=linux GOARCH=amd64 go build

FROM alpine:latest AS prod1
RUN apk --no-cache add sqlite
RUN apk --no-cache add curl
# add perl-utils for json_pp
RUN apk --no-cache add perl-utils
EXPOSE 8000
WORKDIR /app
COPY --from=builder /go/src/github.com/paulfdunn/rest-app-prod1/prod1 /app/prod1
COPY --from=builder /go/src/github.com/paulfdunn/rest-app-prod1/key /app/key
# DO NOT use "-log-filepath=" with Docker, just let it go to STDOUT.
CMD ["./prod1",  "-https-port=8000", "-log-level=0", "persistent-directory=/opt/rest-app/data"]

###########################################################################
# 2 stage with Ubuntu - image size: 178 MB
# First stage only results in 609MB container.
###########################################################################
# FROM golang:1.16.3-buster AS builder
# RUN apt-get update
# COPY ./ /go/src/github.com/paulfdunn/rest-app-prod1
# WORKDIR /go/src/github.com/paulfdunn/rest-app-prod1
# RUN go mod download
# #RUN go test -v ./... >test.log 2>&1 
# RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build

# # New stage to create application container
# FROM ubuntu:latest AS prod1
# RUN apt-get update
# RUN apt-get install -y sqlite 
# RUN apt-get install -y curl
# # add perl for json_pp
# RUN apt-get install -y perl
# EXPOSE 8000
# WORKDIR /app
# COPY --from=builder /go/src/github.com/paulfdunn/rest-app-prod1/prod1 /app/prod1
# COPY --from=builder /go/src/github.com/paulfdunn/rest-app-prod1/key /app/key
# DO NOT use "-log-filepath=" with Docker, just let it go to STDOUT.
# CMD ["./prod1",  "-https-port=8000", "-log-level=0"]

###########################################################################
# PROBLEMS - Ubuntu build and run on Alpine
###########################################################################

# FROM golang:1.16.3-buster AS builder
# RUN apt-get update
# COPY ./ /go/src/github.com/paulfdunn/rest-app-prod1
# WORKDIR /go/src/github.com/paulfdunn/rest-app-prod1
# RUN go mod download
#RUN go test -v ./... >test.log 2>&1 
# Was trying static linking for use with Alpine, but that has a linker error
# RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -tags sqlite_omit_load_extension
# RUN GOOS=linux GOARCH=amd64 go build

# New stage to create application container

# Alpine has issues - it doesn't use glibc so dynamic linking doesn't work, and static
# linking the go app has linker errors.
# FROM alpine:latest AS prod1
# RUN apk --no-cache add sqlite
# RUN apk --no-cache add curl
# # add perl-utils for json_pp
# RUN apk --no-cache add perl-utils
# EXPOSE 8000
# WORKDIR /app
# COPY --from=builder /go/src/github.com/paulfdunn/rest-app-prod1/prod1 /app/prod1
# COPY --from=builder /go/src/github.com/paulfdunn/rest-app-prod1/key /app/key
# CMD ["./prod1",  "-https-port=8000", "-log-level=0", "-log-filepath=./prod1.log"]


