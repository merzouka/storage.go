FROM golang:1.22 as base
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /check .

FROM alpine:3.20 as release
WORKDIR /

RUN apk update
RUN apk upgrade
RUN apk --no-cache add curl

COPY --from=base /check /check

EXPOSE 8080

ENV GIN_MODE=release
ENV LOGS_PATH=/logs
CMD [ "sleep", "10m" ]
