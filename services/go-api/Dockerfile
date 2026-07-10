FROM golang:1.22-alpine AS build

WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/go-api ./cmd/server

FROM alpine:3.20

RUN adduser -D -H appuser
WORKDIR /app

COPY --from=build /out/go-api /app/go-api

EXPOSE 8080
USER appuser

CMD ["/app/go-api"]
