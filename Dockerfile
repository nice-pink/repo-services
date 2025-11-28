FROM golang:1.25.4-trixie AS builder

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/"

RUN mkdir -p /app

# get go module ready
COPY go.mod go.sum /app/
RUN go mod download

# copy module code
COPY build /app/
COPY cmd/ /app/cmd/
COPY pkg/ /app/pkg/
# COPY test/ /app/test/

# build all
RUN /app/build

ENTRYPOINT ["/app/deploy"]