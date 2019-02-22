FROM golang AS builder
WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download
ADD ./ ./
RUN ./script/build.sh

FROM alpine:latest
RUN apk -U add ca-certificates \
 && mkdir -p /app
COPY --from=builder /app/output/ /app/
RUN adduser -D app
ADD ./docker-entrypoint.sh /
USER app
ENTRYPOINT [ "/docker-entrypoint.sh" ]
