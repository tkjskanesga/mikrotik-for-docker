FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
ARG TARGETARCH
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=linux go build -o entrypoint main.go

FROM alpine:3.19
ARG TARGETARCH
RUN apk add --no-cache qemu-system-x86_64 qemu-img iproute2 iptables bash wget

ARG CHR_VERSION=7.14.3

WORKDIR /images
RUN CHR_VERSION_SUFFIX="" && if [ "$TARGETARCH" = "arm64" ]; then CHR_VERSION_SUFFIX="-arm64"; fi && \
    CHR_FILE_VERSION="${CHR_VERSION}${CHR_VERSION_SUFFIX}" && \
    wget "https://download.mikrotik.com/routeros/${CHR_FILE_VERSION}/chr-${CHR_FILE_VERSION}.img.zip" && \
    unzip "chr-${CHR_FILE_VERSION}.img.zip" && \
    mv "chr-${CHR_FILE_VERSION}.img" /images/chr.img && \
    rm "chr-${CHR_FILE_VERSION}.img.zip"

RUN mkdir -p /storage

WORKDIR /app
COPY --from=builder /app/entrypoint .

ENTRYPOINT ["./entrypoint"]