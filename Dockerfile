FROM ubuntu:22.04

ARG RELEASE=prod
ENV COMMENTO_BIND_ADDRESS="0.0.0.0"

# Install CA certificates (for sending mail via SMTP TLS), tzdata (to allow timezone conversions)
RUN \
    export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true && \
    apt-get update && \
    apt-get install -y ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

# Copy the previously built artifacts
COPY \
    build/${RELEASE}/commento \
    build/${RELEASE}/js \
    build/${RELEASE}/css \
    build/${RELEASE}/images \
    build/${RELEASE}/fonts \
    build/${RELEASE}/templates \
    build/${RELEASE}/db \
    build/${RELEASE}/*.html \
    /commento/

EXPOSE 8080
WORKDIR /commento/
ENTRYPOINT ["/commento/commento"]
