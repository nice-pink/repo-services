FROM cgr.dev/chainguard/go:latest-dev AS builder

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/"

WORKDIR /app

# get go module ready
COPY go.mod ./
RUN go mod tidy
RUN go mod download

# copy module code
COPY build .
COPY cmd/ .
COPY pkg/ .
COPY test/ .

# build all
RUN ./build

FROM cgr.dev/chainguard/glibc-dynamic:latest AS runner

# Info
LABEL org.opencontainers.image.authors="r@nice.pink"
LABEL org.opencontainers.image.source="https://github.com/nice-pink/"

WORKDIR /app

# copy executable
COPY --from=builder /app/bin/* /app/
