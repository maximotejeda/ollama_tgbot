FROM golang:1.21 as build

WORKDIR /telegram_bot

COPY . .

RUN go mod download && go mod tidy
RUN go build -o /bin/tgbot ./cmd/telegram_bot/main.go

FROM debian:unstable-slim
COPY --from=build /bin/tgbot /usr/bin/tgbot

ENTRYPOINT ["tgbot"]
