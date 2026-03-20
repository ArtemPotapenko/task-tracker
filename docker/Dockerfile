FROM golang:1.25-alpine AS build

ARG SERVICE
ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=local

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/service ./${SERVICE}

FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/service /app/service
COPY --from=build /src/migrations /app/migrations
COPY --from=build /src/gen/public/openapi /app/gen/public/openapi
EXPOSE 8080 50051 50052
ENTRYPOINT ["/app/service"]
