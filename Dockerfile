# Stage one - build the binary
FROM golang:alpine AS build

ARG DATABASE_DIR
ARG DATABASE_FILE
ENV DATABASE_DIR=$DATABASE_DIR
ENV DATABASE_FILE=$DATABASE_FILE

RUN apk --no-cache add gcc g++ make git sqlite

WORKDIR /go/src/app

COPY . .

RUN go mod tidy
RUN GOOS=linux go build -ldflags="-s -w" -o ./bin/gourlshortener ./main.go
# RUN chmod +x ./bin/build-db.sh && ./bin/build-db.sh

# Stage two - deploy the binary
FROM alpine:3.17

# ARG DATABASE_DIR
# ARG DATABASE_FILE

RUN apk --no-cache add ca-certificates

WORKDIR /usr/bin

COPY --from=build /go/src/app/bin /go/bin
COPY --from=build /go/src/app/db /go/bin/db
# COPY --from=build /go/src/app/${DATABASE_FILE} ${DATABASE_DIR}${DATABASE_FILE}
COPY --from=build /go/src/app/static /go/bin/static
COPY --from=build /go/src/app/templates /go/bin/templates

EXPOSE 80

ENTRYPOINT /go/bin/gourlshortener --port 80
