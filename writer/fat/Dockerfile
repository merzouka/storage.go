FROM golang:1.22 as base
WORKDIR /app
VOLUME [ "/files" ]

COPY go.mod go.sum ./
RUN go mod download

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
CMD [ "/server" ]
