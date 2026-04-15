FROM golang:1.25.3-alpine3.21 AS build

WORKDIR /app

COPY apps/recommendation_ms/go.mod apps/recommendation_ms/go.sum ./
COPY proto/ /proto/
RUN go mod download

COPY apps/recommendation_ms/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /

COPY --from=build /app/main .

EXPOSE 50052

USER nonroot:nonroot

ENTRYPOINT [ "/main" ]
