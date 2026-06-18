FROM golang:1.25.3-alpine3.21 AS build

WORKDIR /

COPY gen/ gen/
COPY apps/api/go.mod apps/api/go.sum apps/api/

WORKDIR /apps/api
RUN go mod download

COPY apps/api/ ./

RUN CGO_ENABLED=0 go build -o main ./cmd/server/server.go

FROM alpine:latest

WORKDIR /tmp

RUN wget https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz && \
  tar -xJf ffmpeg-git-amd64-static.tar.xz && \
  mv ffmpeg-*-amd64-static/ffmpeg ffmpeg-*-amd64-static/ffprobe /usr/local/bin/ && \
  rm -rf /tmp/*

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

WORKDIR /

COPY --from=build /apps/api/main ./

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "./main" ]
