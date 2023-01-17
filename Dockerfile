FROM alpine:3

ARG RELEASE=prod
ENV COMMENTO_BIND_ADDRESS="0.0.0.0"

# Install CA certificates (for sending mail via SMTP TLS), tzdata (to allow timezone conversions)
RUN apk add --no-cache --update ca-certificates

# Copy the previously built artifacts
COPY build/${RELEASE} /commento/

EXPOSE 8080
WORKDIR /commento/
ENTRYPOINT ["/commento/commento"]
