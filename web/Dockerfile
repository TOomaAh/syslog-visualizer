# Multi-stage Dockerfile for Syslog Visualizer Frontend (Next.js + nginx)

# Stage 1: Dependencies
FROM node:20-alpine AS deps

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./
RUN npm ci

# Stage 2: Builder
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source code
COPY . .

# Build Next.js application (standalone mode)
RUN npm run build

# Stage 3: Runner with nginx
FROM node:20-alpine AS runner

WORKDIR /app

# Install nginx and gettext (for envsubst)
RUN apk add --no-cache nginx gettext

ENV NODE_ENV=production
# Default backend URL (can be overridden via environment variable)
ENV BACKEND_URL=http://backend:8080

# Copy necessary files from builder
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
# Note: public folder is not copied as it doesn't exist yet
# Add "COPY --from=builder /app/public ./public" if you add static assets

# Copy nginx configuration template
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Create nginx run directory
RUN mkdir -p /run/nginx

# Expose port 80 (nginx entry point)
EXPOSE 80

# Start script to run both nginx and Next.js
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'set -e' >> /start.sh && \
    echo 'echo "Configuring nginx with BACKEND_URL=${BACKEND_URL}..."' >> /start.sh && \
    echo 'envsubst '"'"'${BACKEND_URL}'"'"' < /etc/nginx/templates/default.conf.template > /etc/nginx/http.d/default.conf' >> /start.sh && \
    echo 'echo "Starting Next.js on port 3000..."' >> /start.sh && \
    echo 'node server.js &' >> /start.sh && \
    echo 'echo "Starting nginx on port 80..."' >> /start.sh && \
    echo 'nginx -g "daemon off;"' >> /start.sh && \
    chmod +x /start.sh

CMD ["/start.sh"]
