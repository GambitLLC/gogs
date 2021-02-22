FROM golang:1.16

WORKDIR /app

ENV GO111MODULE=on

COPY . .

RUN go build -o gogs .

CMD ["/app/gogs"]