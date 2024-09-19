FROM scratch
ENTRYPOINT ["/caddy-log-exporter"]
COPY caddy-log-exporter /
