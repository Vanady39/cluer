# syntax=docker/dockerfile:1

# ---------------------------------------------------------------- deps
# Dependencies get their own stage so the install layer is reused as long as the
# lockfile is untouched, no matter how much of src/ changed.
FROM node:22-alpine AS deps

WORKDIR /app

COPY package.json package-lock.json ./
# npm ci installs exactly what the lockfile pins and fails if the two disagree,
# which is what makes the build reproducible.
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# ---------------------------------------------------------------- build
FROM node:22-alpine AS build

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Vite inlines env values at build time, which is exactly why the default is a
# relative path: it keeps the image environment-independent, and the runtime
# API address stays in nginx's hands (API_UPSTREAM below). Override this only to
# build a bundle that bypasses the proxy and calls an API elsewhere.
ARG VITE_API_URL=/v1
ARG VITE_OIDC_ISSUER=http://auth.localhost:9000/application/o/cluer/
ARG VITE_OIDC_CLIENT_ID=cluer-admin
ARG VITE_CLUER_APP_KEY

ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_OIDC_ISSUER=${VITE_OIDC_ISSUER}
ENV VITE_OIDC_CLIENT_ID=${VITE_OIDC_CLIENT_ID}
ENV VITE_CLUER_APP_KEY=${VITE_CLUER_APP_KEY}

RUN npm run build

# ---------------------------------------------------------------- runtime
FROM nginx:1.29-alpine

COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html

# Where the API lives on the container network. A runtime value, not a build
# one: unlike VITE_API_URL it is read by nginx, so it can be repointed without
# rebuilding the image.
ENV API_UPSTREAM=http://onboarding:8080
# Demo service, reached under the /demo/v1 prefix. Named API_* so the envsubst
# filter below picks it up.
ENV API_DEMO_UPSTREAM=http://demo:8080
# Restrict envsubst to API_* so nginx's own $variables in the template are left
# alone.
ENV NGINX_ENVSUBST_FILTER=API_

# The stock image expects to run as root. Handing the paths nginx writes to over
# to the unprivileged user it already ships lets the container drop root: a
# static file server has no reason to hold it. conf.d is included because the
# entrypoint renders the template into it before nginx starts.
RUN chown -R nginx:nginx /var/cache/nginx /usr/share/nginx/html /etc/nginx/conf.d \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid

USER nginx

EXPOSE 8080

# Foreground mode: the process has to be PID 1 for Docker to see it stop.
CMD ["nginx", "-g", "daemon off;"]
