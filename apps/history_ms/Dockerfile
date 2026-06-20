FROM golang:tip-alpine3.24 AS build

WORKDIR /

COPY gen/ gen/
COPY apps/history_ms/go.mod apps/history_ms/go.sum apps/history_ms/

WORKDIR /apps/history_ms
RUN go mod download

COPY apps/history_ms/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /

COPY --from=build /apps/history_ms/main .

EXPOSE 50051

USER nonroot:nonroot

ENTRYPOINT [ "./main" ]
