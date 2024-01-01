FROM node:16 as builder

WORKDIR /build
COPY ./web .
COPY ./VERSION .
RUN themes=$(cat THEMES) \
    && IFS=$'\n' \
        && for theme in $themes; do \
            theme_path="web/$theme" \
            && echo "Building theme: $theme" \
            && cd $theme_path \
            && npm install \
            && DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat VERSION) npm run build \
            && cd /app \
        done

FROM golang AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=builder /build/build ./web/build
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o one-api

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates 2>/dev/null || true

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
