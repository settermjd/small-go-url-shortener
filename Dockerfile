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

COPY --from=builder /gourlshortener /usr/local/bin/
CMD ["gourlshortener"]