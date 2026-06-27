#FROM --platform=$BUILDPLATFORM node:18-alpine AS builder
FROM node:18-alpine AS builder

WORKDIR /web
COPY ./VERSION .
COPY ./web .

# 配置 npm 国内镜像加速
RUN npm config set registry https://registry.npmmirror.com \
    && npm install --prefix /web/default \
    && npm install --prefix /web/berry \
    && npm install --prefix /web/air

RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/default \
    && DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/berry \
    && DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/air

FROM golang:1.22-alpine AS builder2

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    build-base

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /web/build ./web/build

RUN go build -trimpath -ldflags "-s -w -X 'github.com/songquanpeng/one-api/common.Version=$(cat VERSION)' -linkmode external -extldflags '-static'" -o one-api

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder2 /build/one-api /

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
