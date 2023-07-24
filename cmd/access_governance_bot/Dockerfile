FROM golang:1.20

ENV ENVIRONMENT ${ENVIRONMENT}
ENV TELEGRAM_BOT_TOKEN ${TELEGRAM_BOT_TOKEN}
ENV TELEGRAM_BOT_UPDATE_TIMEOUT ${TELEGRAM_BOT_UPDATE_TIMEOUT}

WORKDIR /opt/src

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

ADD . .

RUN go build -o /go/bin/app ./cmd/access_governance_bot/

CMD ["/go/bin/app"]