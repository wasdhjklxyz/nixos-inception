#!/usr/bin/env bash

set -euo pipefail

FLAKE_PATH=""
CONFIG_NAME=""
AGE_KEY=""
PORT=""

print_error() {
  echo -e "\033[1;31merror:\033[0m $1" >&2
}

print_warning() {
  echo -e "\033[1;35mwarning:\033[0m $1" >&2
}

show_help() {
  cat << EOF
Usage: nix run github:wasdhjklxyz/nixos-inception -- [OPTIONS]

OPTIONS:
    --help              Show this help message
    --flake PATH#CONFIG Path to flake configuration (e.g., ./flake/path#config)
    --age-key PATH      Age identity file
    --port NUM          Listen port

EXAMPLES:
    nix run github:wasdhjklxyz/nixos-inception -- --flake ./path/to/flake#config
    nix run github:wasdhjklxyz/nixos-inception -- --age-key key.txt --port 8080

If no --flake is provided, searches for flake.nix in current directory.
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --help|-h)
        show_help
        exit 0
        ;;
      --flake)
        if [[ -z "${2:-}" ]]; then
          print_error "--flake requires a path argument"
          exit 1
        fi
        FLAKE_PATH="$2"
        shift 2
        ;;
      --age-key)
        if [[ -z "${2:-}" ]]; then
          print_error "--age-key requires a path argument"
          exit 1
        fi
        AGE_KEY="$2"
        shift 2
        ;;
      --port)
        if [[ -z "${2:-}" ]]; then
          print_error "--port requires a number argument"
          exit 1
        fi
        if ! [[ "$2" =~ ^[0-9]+$ ]]; then
          print_error "--port must be a number"
          exit 1
        fi
        PORT="$2"
        shift 2
        ;;
      *)
        print_error "unkown option $1"
        exit 1
        ;;
    esac
  done
}

resolve_flake() {
  local flake config
  if [[ -z "$FLAKE_PATH" ]]; then
    if [[ -f "./flake.nix" ]]; then
      FLAKE_PATH="."
    else
      print_error "no flake.nix found in current directory and no --flake specified"
      exit 1
    fi
  fi

  if [[ "$FLAKE_PATH" == *"#"* ]]; then
    flake="${FLAKE_PATH%#*}"
    config="${FLAKE_PATH%*#}"
  else
    flake="$FLAKE_PATH"
    config=""
  fi

  if [[ ! -f "$flake/flake.nix" && ! -f "$flake" ]]; then
    print_error "flake not found at $flake"
    exit 1
  fi

  if [[ -z "$config" ]]; then
    print_warning "no configuration specified, detecting available nixosConfigurations..."

    local configs=$(nix eval \
      --json "$flake#nixosConfigurations" \
      --apply 'builtins.attrNames' \
      2>/dev/null || echo "[]")

    if [[ "$configs" == "[]" ]]; then
      print_error "no nixosConfigurations found in flake $flake"
      exit 1
    fi

    config=$(echo "$configs" | jq -r '.[0]')
    print_warning "using first detected configuration '$config'"
  fi

  FLAKE_PATH="$flake"
  CONFIG_NAME="$config"
}

validate_config() {
  local has_inception=$(nix eval \
    --json "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception" \
    >/dev/null 2>&1 && echo "true" || echo "false")

  if "$has_inception" != "true" ]]; then
    print_error "configuration '$CONFIG_NAME' was not created with nixos-inception.lib.nixosSystem"
    exit 1
  fi

  local deployment_config=$(nix eval \
    --json "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.deploymentConfig" \
    2>/dev/null || echo "{}")

  if [[ -n "$AGE_KEY" ]]; then
    echo "Using CLI-provided age key: $AGE_KEY" >&2
  else
    AGE_KEY=$(echo "$deployment_config" | jq -r '.ageKeyFile // empty' \
      2>/dev/null || true)

    if [[ -n "$AGE_KEY" ]]; then
      if [[ ! "$AGE_KEY" == /* ]]; then
        AGE_KEY="${FLAKE_PATH}/${AGE_KEY}"
      fi
      echo "Found age key in flake: $AGE_KEY" >&2
    fi
  fi

  if [[ -n "$PORT" ]]; then
    echo "Using CLI-provided port: $PORT" >&2
  else
    PORT=$(echo "$deployment_config" | jq -r '.serverPort // empty' \
      2>/dev/null || true)
    if [[ -n "$PORT" ]]; then
      echo "Found server port in flake: $PORT" >&2
    fi
  fi
}

start_architect() {
  architect plant --age-key "$AGE_KEY"
  exit 0
  #local architect_pid=$!

  sleep 2

  if ! kill -0 "$architect_pid" 2>/dev/null; then
    print_error "server failed to start"
    return 1
  fi

  cat << EOF

Server listening on: http://localhost:$PORT
PID: $architect_pid

Boot your ISO and it will connect automatically.
Use Ctrl+C to stop the server when deployment is complete.

EOF

  trap 'echo "Stopping server..."; kill $architect_pid 2>/dev/null; exit 0' \
    INT TERM
  wait "$architect_pid"
}

parse_args "$@"
resolve_flake
validate_config

if nix build "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.iso.config.system.build.isoImage"; \
then
  echo "ISO available at ./result/iso/*.iso" >&2
  start_architect
else
  exit 1
fi
