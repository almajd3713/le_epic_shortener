# Build stage
FROM node:20-alpine AS builder
WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm run build

# Runtime stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# SPA routing: unknown paths fall back to index.html instead of 404
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80