FROM golang:1.22-alpine AS build

ARG SERVICE
ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/service ./${SERVICE}

FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/service /app/service
EXPOSE 8080 50051 50052
ENTRYPOINT ["/app/service"]
