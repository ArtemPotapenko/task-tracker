FROM golang:1.25-alpine AS build

ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=local

WORKDIR /src

COPY go.work ./
COPY account-service account-service
COPY task-service task-service
COPY email-service email-service
COPY gateway gateway
COPY scheduler-service scheduler-service
COPY shared-libs shared-libs
COPY proto-lib proto-lib

RUN go build -o /out/service ./email-service/cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/service /app/service
ENTRYPOINT ["/app/service"]
