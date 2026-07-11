#!/usr/bin/env sh
set -eu

MERMAID_ASCII_VERSION="${MERMAID_ASCII_VERSION:-v1.4.0}"
MERMAID_ASCII_MODULE="github.com/AlexanderGrooff/mermaid-ascii"

log() {
	printf '%s\n' "$*"
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

repo_root() {
	script_dir=$(CDPATH= cd "$(dirname "$0")" && pwd)
	CDPATH= cd "$script_dir/.." && pwd
}

copy_binary() {
	src=$1
	dst=$2

	if command -v install >/dev/null 2>&1; then
		install -m 0755 "$src" "$dst"
	else
		cp "$src" "$dst"
		chmod 0755 "$dst"
	fi
}

copy_binary_with_privilege() {
	src=$1
	dst=$2

	if copy_binary "$src" "$dst" 2>/dev/null; then
		return 0
	fi

	if command -v sudo >/dev/null 2>&1; then
		log "Need elevated permissions to write $dst"
		if command -v install >/dev/null 2>&1; then
			sudo install -m 0755 "$src" "$dst"
		else
			sudo cp "$src" "$dst"
			sudo chmod 0755 "$dst"
		fi
		return 0
	fi

	die "cannot write $dst and sudo is unavailable"
}

ensure_dir() {
	dir=$1

	if [ -d "$dir" ]; then
		return 0
	fi

	if mkdir -p "$dir" 2>/dev/null; then
		return 0
	fi

	if command -v sudo >/dev/null 2>&1; then
		sudo mkdir -p "$dir"
		return 0
	fi

	die "cannot create $dir and sudo is unavailable"
}

backup_existing() {
	path=$1
	stamp=$(date +%Y%m%d%H%M%S)
	backup="${path}.backup-${stamp}"

	if [ -e "$path" ] || [ -L "$path" ]; then
		log "Backing up existing glow to $backup"
		if mv "$path" "$backup" 2>/dev/null; then
			return 0
		fi

		if command -v sudo >/dev/null 2>&1; then
			sudo mv "$path" "$backup"
			return 0
		fi

		die "cannot back up $path and sudo is unavailable"
	fi
}

command_path() {
	name=$1
	command -v "$name" 2>/dev/null || true
}

install_mermaid_ascii() {
	if command -v mermaid-ascii >/dev/null 2>&1; then
		log "mermaid-ascii already available"
		return 0
	fi

	if [ "${GLOW_UPDATE_SKIP_MERMAID_INSTALL:-0}" = "1" ]; then
		log "Skipping mermaid-ascii install"
		return 0
	fi

	log "Installing mermaid-ascii $MERMAID_ASCII_VERSION"
	if go install "$MERMAID_ASCII_MODULE@$MERMAID_ASCII_VERSION"; then
		return 0
	fi

	log "Could not install mermaid-ascii; continuing without Mermaid rendering on PATH."
	log "Install later with: go install $MERMAID_ASCII_MODULE@$MERMAID_ASCII_VERSION"
}

ROOT=$(repo_root)
cd "$ROOT"

command -v go >/dev/null 2>&1 || die "Go is required"

GOOS=$(go env GOOS)
case "$GOOS" in
	darwin|linux|freebsd|openbsd|windows|android)
		;;
	*)
		die "unsupported GOOS=$GOOS; this script targets the platforms listed in README.md"
		;;
esac

EXE=
case "$GOOS" in
	windows)
		EXE=.exe
		;;
esac

TMPDIR=${TMPDIR:-/tmp}
BUILD_DIR=$(mktemp -d "$TMPDIR/glow-update.XXXXXX")
trap 'rm -rf "$BUILD_DIR"' EXIT INT TERM

install_mermaid_ascii

if [ "${GLOW_UPDATE_SKIP_TESTS:-0}" != "1" ]; then
	log "Running tests"
	go test ./...
fi

BIN="$BUILD_DIR/glow$EXE"
log "Building $BIN"
go build -o "$BIN" .

TARGET=${GLOW_UPDATE_TARGET:-}
if [ -z "$TARGET" ]; then
	TARGET=$(command_path "glow$EXE")
fi
if [ -z "$TARGET" ] && [ "$EXE" = ".exe" ]; then
	TARGET=$(command_path glow)
fi
if [ -z "$TARGET" ]; then
	GOBIN=$(go env GOBIN)
	if [ -z "$GOBIN" ]; then
		GOPATH=$(go env GOPATH)
		TARGET="$GOPATH/bin/glow$EXE"
	else
		TARGET="$GOBIN/glow$EXE"
	fi
	log "No existing glow found; installing to $TARGET"
fi

ensure_dir "$(dirname "$TARGET")"
backup_existing "$TARGET"
copy_binary_with_privilege "$BIN" "$TARGET"

log "Installed $TARGET"
"$TARGET" --version || true

if ! command -v mermaid-ascii >/dev/null 2>&1; then
	log "mermaid-ascii was installed by Go, but it is not on PATH."
	log "Add \$(go env GOPATH)/bin or \$(go env GOBIN) to PATH before running glow."
fi
