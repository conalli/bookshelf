FROM golang:1.17-alpine as build

WORKDIR /go/src/github.com/bookshelf-backend

COPY go.* .

RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:3.15

WORKDIR /app

COPY --from=build /go/src/github.com/bookshelf-backend/main .
COPY --from=build /go/src/github.com/bookshelf-backend/.env .

EXPOSE 8080

CMD ["./main"]