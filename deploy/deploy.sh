#!/usr/bin/env bash
# Deploy a Traditional Builders release tarball to a systemd host.
#
# Prereqs (one-time on the target host):
#   - user `traditionalbuilders` and dir /opt/traditionalbuilders
#   - deploy/traditionalbuilders.service installed at
#     /etc/systemd/system/traditionalbuilders.service
#   - database migrated: goose -dir migrations sqlite3 traditionbuilders.db up
#   - database seeded:   ./server -seed-only   (applies the embedded zip dataset)
#
# The zip code dataset is embedded in the server binary, so `deploy.sh` runs
# `server -seed-only` after migrations automatically. No network or manual
# data step is required on the host.
#
# Usage:
#   DEPLOY_HOST=deploy@example.com deploy/deploy.sh
#
# Env:
#   DEPLOY_HOST  (required)  ssh target, e.g. user@host
#   DEPLOY_ARCH  (optional)  amd64 | arm64            (default: amd64)
#   DEPLOY_DIR   (optional)  install dir on host       (default: /opt/traditionalbuilders)
#   DEPLOY_SERVICE (optional) systemd unit name        (default: traditionalbuilders)
set -euo pipefail

: "${DEPLOY_HOST:?set DEPLOY_HOST, e.g. DEPLOY_HOST=user@host}"
DEPLOY_ARCH="${DEPLOY_ARCH:-amd64}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/traditionalbuilders}"
DEPLOY_SERVICE="${DEPLOY_SERVICE:-traditionalbuilders}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "==> Building release (linux/$DEPLOY_ARCH)"
make release

TARBALL="dist/traditionalbuilders-linux-${DEPLOY_ARCH}.tar.gz"
[ -f "$TARBALL" ] || { echo "missing $TARBALL"; exit 1; }

echo "==> Shipping $TARBALL to $DEPLOY_HOST"
scp "$TARBALL" "$DEPLOY_HOST:/tmp/traditionalbuilders.tar.gz"

echo "==> Installing on $DEPLOY_HOST:$DEPLOY_DIR"
ssh "$DEPLOY_HOST" bash -s -- "$DEPLOY_DIR" "$DEPLOY_SERVICE" <<'REMOTE'
set -euo pipefail
dir="$1"
service="$2"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
tar -xzf /tmp/traditionalbuilders.tar.gz -C "$tmp"

# Preserve the database: it lives in $dir and is not part of the tarball.
sudo install -d -o traditionalbuilders -g traditionalbuilders "$dir"
sudo install -o traditionalbuilders -g traditionalbuilders "$tmp"/traditionalbuilders-linux-*/server "$dir/server.new"
sudo rm -rf "$dir/static" "$dir/migrations"
sudo cp -r "$tmp"/traditionalbuilders-linux-*/static "$dir/static"
sudo cp -r "$tmp"/traditionalbuilders-linux-*/migrations "$dir/migrations"
sudo chown -R traditionalbuilders:traditionalbuilders "$dir"
sudo mv "$dir/server.new" "$dir/server"

# Apply the embedded zip dataset after shipping the new binary (idempotent;
# a checksum match is a no-op). Runs before restart so the server starts with
# the current data already in place.
echo "==> Seeding data (server -seed-only)"
sudo -u traditionalbuilders bash -c "cd '$dir' && ./server -seed-only"

sudo systemctl restart "$service"
sleep 1
sudo systemctl --no-pager --lines=10 status "$service" || true
REMOTE

echo "==> Deployed. Verify: curl http://<host>:8080/healthz"
