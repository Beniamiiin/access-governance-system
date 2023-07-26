FROM golang:1.20

WORKDIR /opt/src

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

ADD . .

RUN go build -o /go/bin/app ./cmd/access_governance_bot/

CMD ["/go/bin/app"]