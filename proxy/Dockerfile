FROM golang:1.22 as base
WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

ENV GOPATH=/go
COPY . ./
RUN CGO_ENABLED=0 go build -o /server .

FROM gcr.io/distroless/base-debian11 as release
WORKDIR /
COPY --from=base /server /server

EXPOSE 8080

ENV LOGS_PATH=/logs
ENV GIN_MODE=release
CMD [ "/server" ]
