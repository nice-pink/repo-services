FROM golang:1.25.4-trixie AS builder

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/"

RUN mkdir -p /app && chmod u+x /app

# get go module ready
COPY go.mod go.sum /app/
RUN cd /app && go mod download

# copy module code
COPY build /app/
COPY cmd/ /app/cmd/
COPY pkg/ /app/pkg/
# COPY test/ /app/test/

# build all
RUN cd /app && ./build

# Run as non-root (fixed uid/gid for stable volume/Kubernetes ownership)
RUN groupadd --system --gid 65532 nonroot && \
    useradd --system --uid 65532 --gid nonroot --no-create-home --shell /usr/sbin/nologin nonroot && \
    chown -R nonroot:nonroot /app

USER nonroot:nonroot

ENTRYPOINT ["/app/bin/deploy"]