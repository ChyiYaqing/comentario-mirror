FROM alpine:3

# Install CA certificates (for sending mail via SMTP TLS)
RUN apk add --no-cache --update ca-certificates

# Copy the previously built artifacts
COPY build /comentario/

# Make sure files were built and are available
RUN test -x /comentario/comentario && \
    test -d /comentario/css && \
    test -d /comentario/db && \
    test -d /comentario/fonts && \
    test -d /comentario/html && \
    test -d /comentario/images && \
    test -d /comentario/js && \
    test -s /comentario/js/comentario.js && \
    test -d /comentario/templates

EXPOSE 8080
WORKDIR /comentario/
ENTRYPOINT ["/comentario/comentario"]
CMD ["--static-path=/comentario", "--db-migration-path=/comentario/db", "--template-path=/comentario/templates"]
