FROM golang:1.25.3-alpine3.21 AS build

WORKDIR /

COPY gen/ gen/
COPY apps/api/go.mod apps/api/go.sum apps/api/

WORKDIR /apps/api
RUN go mod download

COPY apps/api/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/server.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /

COPY --from=build /apps/api/main .

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "./main" ]
