FROM golang:1.22.2 AS chat-api-builder

RUN go version

COPY . /github.com/Woodfyn/chat-api-backend-go/
WORKDIR /github.com/Woodfyn/chat-api-backend-go/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/app ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=chat-api-builder /github.com/Woodfyn/chat-api-backend-go/.bin/app .
COPY main.env .
COPY configs ./configs

CMD ["./app"]