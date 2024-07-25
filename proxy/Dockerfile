FROM golang:1.22 as base
WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY *.go models ./
RUN CGO_ENABLED=0 go build -o /server .

FROM gcr.io/distroless/base-debian11 as release
WORKDIR /
COPY --from=base /server /server

USER nonroot

EXPOSE 8080

ENV GIN_MODE=release
CMD [ "/server" ]
