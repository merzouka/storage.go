FROM golang:1.22 as base
WORKDIR /app
VOLUME [ "/files" ]

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -o /server .

FROM gcr.io/distroless/base-debian11 as build-release
WORKDIR /

COPY --from=base /server /server

EXPOSE 8081
# USER nonroot:nonroot

CMD [ "/server" ]
