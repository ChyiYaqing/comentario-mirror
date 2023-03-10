FROM alpine:3

# Install CA certificates (for sending mail via SMTP TLS)
RUN apk add --no-cache --update ca-certificates

# Copy the previously built artifacts
COPY build /comentario/

# Make sure files were built and are available
RUN test -x /comentario/comentario && \
    test -d /comentario/db && \
    test -s /comentario/frontend/comentario.css && \
    test -s /comentario/frontend/comentario.js && \
    test -d /comentario/frontend/en/fonts && \
    test -d /comentario/frontend/en/images && \
    test -s /comentario/frontend/en/index.html && \
    test -d /comentario/templates

WORKDIR /comentario/
ENTRYPOINT ["/comentario/comentario"]
CMD ["--host=0.0.0.0", "--port=80", "-v"]
