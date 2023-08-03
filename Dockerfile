FROM alpine:3.18 AS build-env
RUN apk --no-cache add ca-certificates tzdata
RUN set -ex; \
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
		armhf) arch='armv6' ;; \
		armv7) arch='armv7' ;; \
		aarch64) arch='arm64' ;; \
		x86_64) arch='amd64' ;; \
		s390x) arch='s390x' ;; \
		*) echo >&2 "error: unsupported architecture: $apkArch"; exit 1 ;; \
	esac; \
    wget --quiet -O /tmp/app.tar.gz "https://github.com/lenye/aichat/releases/download/v0.3.0/aichat_v0.3.0_linux_$arch.tar.gz"; \
    tar xzvf /tmp/app.tar.gz -C /usr/local/bin aichat; \
    rm -f /tmp/app.tar.gz; \
	chmod +x /usr/local/bin/aichat


FROM scratch
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /usr/share/zoneinfo /usr/share/
COPY --from=build-env /usr/local/bin/aichat /

EXPOSE 8080

VOLUME ["/tmp"]
ENTRYPOINT ["/aichat"]