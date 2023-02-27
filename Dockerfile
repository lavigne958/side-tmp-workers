FROM golang:1.20.1-alpine

RUN apk add gcc musl-dev

WORKDIR /code

COPY ./go.mod ./
COPY ./go.sum ./
# pre-install go-sqlite3 to save build time
RUN go install github.com/mattn/go-sqlite3
COPY ./main.go ./

RUN go get
RUN CGO_ENABLED=1 go build -o main

ENTRYPOINT [ "/code/main" ]
