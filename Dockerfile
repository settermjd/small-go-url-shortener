ARG GO_VERSION=1
ARG ALPINE_VERSION=3.17

# Stage one - build the binary
FROM golang:${GO_VERSION}-alpine AS builder

#RUN apk --no-cache add gcc g++ make git sqlite

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod tidy && go mod verify

COPY . .
RUN GOOS=linux go build -v -ldflags="-s -w" -o /gourlshortener .

# Stage two - deploy the binary
FROM alpine:${ALPINE_VERSION}

WORKDIR /opt

# Copy over the static assets, database migrations, templates, and scripts
COPY ./bin bin
COPY ./db db
COPY ./static static
COPY ./templates templates

# Ensure that the migrations script is executable
RUN chmod ug+x ./bin/run-migrations.sh

# Install dbmate (the database migration tool) in a way that works on Alpine Linux.
RUN apk --no-cache add npm sqlite \
    && npm install --save-dev dbmate

# Copy over the Go binary and set it as the command to run on boot
COPY --from=builder /gourlshortener /usr/local/bin/
CMD ["gourlshortener"]