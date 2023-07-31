FROM alpine:3.18 AS build-env
RUN apk --no-cache add ca-certificates tzdata
COPY aichat /usr/local/bin
RUN set -ex; \
	chmod +x /usr/local/bin/aichat


FROM scratch
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /usr/share/zoneinfo /usr/share/
COPY --from=build-env /usr/local/bin/aichat /

EXPOSE 8080

VOLUME ["/tmp"]
ENTRYPOINT ["/aichat"]