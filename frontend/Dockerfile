# Build stage
FROM node:20-alpine AS builder
WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# VITE_* vars must be present at build time — they get baked into the static bundle
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

COPY . .
RUN pnpm run build

# Runtime stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# SPA routing: unknown paths fall back to index.html instead of 404
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80