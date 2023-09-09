FROM gcr.io/distroless/static-debian12
COPY aichat /
EXPOSE 8080
ENTRYPOINT ["/aichat"]