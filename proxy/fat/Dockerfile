FROM golang:1.22 as base
WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

ENV GOPATH=/go
COPY . ./
RUN CGO_ENABLED=0 go build -o /server .

FROM alpine:3.20 as release
WORKDIR /

RUN apk update
RUN apk upgrade
RUN apk --no-cache add curl

COPY --from=base /server /server

EXPOSE 8080

ENV LOGS_PATH=/logs
ENV GIN_MODE=release
CMD [ "/server" ]
