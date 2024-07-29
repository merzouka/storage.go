FROM golang:1.22 as base
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /check .

FROM gcr.io/distroless/base-debian11 as release
WORKDIR /

COPY --from=base /check /check

EXPOSE 8080
ENV GIN_MODE=release
ENV LOGS_PATH=/logs
CMD [ "/check" ]
