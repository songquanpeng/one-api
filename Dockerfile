FROM --platform=$BUILDPLATFORM node:16 AS builder

WORKDIR /web
COPY ./VERSION .
COPY ./web .

RUN npm install --prefix /web/default & \
    npm install --prefix /web/berry & \
    npm install --prefix /web/air & \
    wait

RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat /web/default/VERSION) npm run build --prefix /web/default & \
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat /web/berry/VERSION) npm run build --prefix /web/berry & \
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat /web/air/VERSION) npm run build --prefix /web/air & \
    wait

FROM golang:alpine AS builder2

RUN apk add --no-cache g++

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder /web/build ./web/build
RUN go build -trimpath -ldflags "-s -w -X 'github.com/songquanpeng/one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o one-api

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates 2>/dev/null || true

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]