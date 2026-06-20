FROM golang:tip-alpine3.24 AS build

WORKDIR /

COPY gen/ gen/
COPY apps/content_processing_ms/go.mod apps/content_processing_ms/go.sum apps/content_processing_ms/

WORKDIR /apps/content_processing_ms
RUN go mod download

COPY apps/content_processing_ms/ ./

RUN CGO_ENABLED=0 go build -o main ./cmd/worker/main.go

FROM alpine:3.24.1

WORKDIR /tmp

RUN wget https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz && \
  tar -xJf ffmpeg-git-amd64-static.tar.xz && \
  mv ffmpeg-*-amd64-static/ffmpeg ffmpeg-*-amd64-static/ffprobe /usr/local/bin/ && \
  rm -rf /tmp/*

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

WORKDIR /

COPY --from=build /apps/content_processing_ms/main ./

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "./main" ]
