FROM golang:1.21

ARG ENVIRONMENT
ENV ENVIRONMENT=$ENVIRONMENT

ARG DISCORD_AUTHORIZATION_BOT_TOKEN
ENV DISCORD_AUTHORIZATION_BOT_TOKEN=$DISCORD_AUTHORIZATION_BOT_TOKEN

WORKDIR /opt/src

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

ADD . .

RUN go build -o /go/bin/app ./cmd/authorization_bot/discord

CMD ["/go/bin/app"]