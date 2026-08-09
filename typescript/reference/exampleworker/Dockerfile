# syntax=docker/dockerfile:1
FROM node:24.18.0-bookworm-slim AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run verify

FROM node:24.18.0-bookworm-slim AS production-dependencies
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev && npm cache clean --force

FROM node:24.18.0-bookworm-slim AS runtime
ARG VERSION=dev
ARG COMMIT=unknown
ARG CREATED
LABEL org.opencontainers.image.source="https://github.com/BonzTM/coding-handbook" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${CREATED}"
ENV NODE_ENV=production
WORKDIR /app
COPY --from=production-dependencies --chown=10001:10001 /app/node_modules ./node_modules
COPY --from=build --chown=10001:10001 /app/dist ./dist
COPY --from=build --chown=10001:10001 /app/package.json ./package.json
USER 10001:10001
EXPOSE 3001
ENTRYPOINT ["node", "dist/main.js"]
