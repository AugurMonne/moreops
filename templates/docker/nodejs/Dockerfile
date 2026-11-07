# Multi-stage secure build for Node.js applications
FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build --if-present

# Minimal production runtime
FROM node:20-alpine

WORKDIR /app
ENV NODE_ENV=production

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder --chown=appuser:appgroup /app ./

USER appuser
EXPOSE 3000

CMD ["npm", "start"]
