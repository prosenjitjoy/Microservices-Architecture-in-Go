FROM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o metadata metadata/cmd/main.go
RUN go build -o rating rating/cmd/main.go
RUN go build -o movie movie/cmd/main.go

FROM alpine:latest AS metadata
WORKDIR /app
COPY --from=builder /src/metadata/main metadata
COPY .env .
EXPOSE 8081
CMD [ "/app/metadata" ]

FROM alpine:latest AS rating
WORKDIR /app
COPY --from=builder /src/rating/main rating
COPY .env .
EXPOSE 8082
CMD [ "/app/rating" ]

FROM alpine:latest AS movie
WORKDIR /app
COPY --from=builder /src/movie/main movie
COPY .env .
EXPOSE 8083
CMD [ "/app/movie" ]