FROM node:18 AS frontend-builder
WORKDIR /app
COPY frontend/ .
RUN npm install && npm run build

FROM golang:1.24 AS backend-builder
WORKDIR /app
COPY backend/ .
COPY --from=frontend-builder /app/dist ./frontend/dist
RUN go build -o fsm .

FROM debian:bookworm-slim

ARG PUID=1000
ARG PGID=1000

RUN apt-get update && apt-get install -y ca-certificates xz-utils && rm -rf /var/lib/apt/lists/*

RUN addgroup --gid $PGID fsm && \
    adduser --disabled-password --gecos "" --uid $PUID --ingroup fsm fsm

WORKDIR /app

RUN chown -R fsm:fsm /app
USER fsm


COPY --from=backend-builder /app/fsm .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist
# Unraid Optimization
EXPOSE 8888 27015 34197/udp

CMD ["./fsm"]