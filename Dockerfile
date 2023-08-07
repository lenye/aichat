FROM gcr.io/distroless/static-debian11
COPY aichat /
EXPOSE 8080
ENTRYPOINT ["/aichat"]