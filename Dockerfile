# ====== 构建后端 ======
FROM golang:1.25-alpine AS backend

ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o forks .

# ====== 运行镜像 ======
FROM alpine:3.21

RUN apk add --no-cache git ca-certificates && \
    git config --global --add safe.directory '*'
WORKDIR /app
COPY --from=backend /build/forks /app/forks

ENV FORKS_HOME=/data
ENV FORKS_REPO_PATH=/data/repos
ENV FORKS_PORT=8080
EXPOSE ${FORKS_PORT}
VOLUME /data

ENTRYPOINT ["/app/forks"]
CMD ["serve"]
