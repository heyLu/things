FROM docker.io/golang:1.22-alpine3.20 as builder

# gcc and libc-dev for sqlite, git for vcs listing in /stats page
RUN apk add --no-cache gcc libc-dev git make

WORKDIR /build

COPY . .
RUN make statics
RUN go build .

FROM alpine:3.20

RUN apk add --no-cache qalc && qalc --exrates '1+1'

RUN apk add --no-cache shadow && useradd --home-dir /dev/null --shell /bin/false things && apk del shadow
USER things

VOLUME /app/data

CMD /app/things -addr 0.0.0.0:5000

COPY --from=builder /build/things /app/
