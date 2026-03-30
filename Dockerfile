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

ENTRYPOINT ["/app/bin/deploy"]