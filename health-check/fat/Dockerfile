FROM golang:1.22 as base
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /check .

FROM alpine:3.20 as release
WORKDIR /

COPY --from=base /check /check

EXPOSE 8080

ENV INSTANCES=writer-service1,writer-service2,writer-service3
ENV NAMESPACE=default
CMD [ "/check" ]
