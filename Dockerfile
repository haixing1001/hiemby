# Multi-arch build: amd64 / arm64
# Stage 1: Go backend
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS gobuild
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/emby-server .

# Stage 2: Vue frontend
FROM node:20-alpine AS webuild
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ ./
RUN npm run build

# Stage 3: runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates ffmpeg tzdata
WORKDIR /app
COPY --from=gobuild /out/emby-server /app/emby-server
COPY --from=webuild /web/dist /app/web/dist
VOLUME ["/app/data", "/media"]
EXPOSE 8096
ENV ADMIN_PASSWORD=admin123
ENTRYPOINT ["/app/emby-server"]
CMD ["-addr", ":8096", "-data", "/app/data"]
