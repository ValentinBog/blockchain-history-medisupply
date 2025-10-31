# Dockerfile pour HistorialBlockchain
FROM golang:1.21-alpine AS builder

# Installer les dépendances du système
RUN apk add --no-cache git ca-certificates

# Définir le répertoire de travail
WORKDIR /app

# Copier go mod et go sum
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier le code source
COPY . .

# Construire l'application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Stage de production
FROM alpine:latest

# Installer ca-certificates pour les connexions HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copier l'exécutable depuis le builder
COPY --from=builder /app/main .

# Exposer le port
EXPOSE 8081

# Commande par défaut
CMD ["./main"]
