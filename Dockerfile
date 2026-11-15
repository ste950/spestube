# syntax=docker/dockerfile:1

# Il frontend e il binario si compilano sulla macchina che costruisce l'immagine
# ($BUILDPLATFORM), non in emulazione: Go compila da sé per l'architettura di destinazione,
# e i file del frontend sono gli stessi ovunque. Così l'immagine per il Raspberry Pi
# (linux/arm64) si costruisce in fretta anche da un PC amd64.

FROM --platform=$BUILDPLATFORM node:24-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS backend
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Il frontend compilato entra nel binario con go:embed.
COPY --from=web /src/web/dist/ ./internal/webui/dist/
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/spestube ./cmd/spestube
# La cartella dei dati nasce qui con il proprietario giusto: un volume nuovo ne eredita i
# permessi, e il processo, che non è root, può scriverci.
RUN mkdir -p /out/data

# Nessuna shell e nessun gestore di pacchetti: solo i certificati per parlare con YouTube
# in HTTPS, e un utente senza privilegi.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend /out/spestube /spestube
COPY --from=backend --chown=nonroot:nonroot /out/data /data
ENV SPESTUBE_PORT=8080 \
    SPESTUBE_DB_PATH=/data/spestube.db
EXPOSE 8080
VOLUME /data
USER nonroot
ENTRYPOINT ["/spestube"]
