FROM golang:1.20.1-alpine

RUN apk add gcc musl-dev

WORKDIR /code

COPY ./go.mod ./
COPY ./go.sum ./
COPY ./main.go ./

RUN go get
RUN CGO_ENABLED=1 go build -o main

ENTRYPOINT [ "/code/main" ]
