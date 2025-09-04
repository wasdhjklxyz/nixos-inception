#!/usr/bin/env bash

set -euo pipefail

FLAKE_PATH=""
CONFIG_NAME=""
AGE_KEY=""
PORT=""

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
          echo "Error: --flake requires a path argument" >&2
          exit 1
        fi
        FLAKE_PATH="$2"
        shift 2
        ;;
      --age-key)
        if [[ -z "${2:-}" ]]; then
          echo "Error: --age-key requires a path argument" >&2
          exit 1
        fi
        AGE_KEY="$2"
        shift 2
        ;;
      --port)
        if [[ -z "${2:-}" ]]; then
          echo "Error: --port requires a number argument" >&2
          exit 1
        fi
        if ! [[ "$2" =~ ^[0-9]+$ ]]; then
          echo "Error: --port must be a number" >&2
          exit 1
        fi
        PORT="$2"
        shift 2
        ;;
      *)
        echo "Error: Unkown option $1" >&2
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
      echo "Error: No flake.nix found in current directory and no --flake specified" >&2
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
    echo "Error: Flake not found at $flake" >&2
    exit 1
  fi

  if [[ -z "$config" ]]; then
    echo "No configuration specified, detecting available nixosConfigurations..." >&2

    local configs
    configs=$(nix eval \
      --json "$flake#nixosConfigurations" \
      --apply 'builtins.attrNames' \
      2>/dev/null || echo "[]")

    if [[ "$configs" == "[]" ]]; then
      echo "Error: No nixosConfigurations found in flake" >&2
      exit 1
    fi

    config=$(echo "$configs" | jq -r '.[0]')
    echo "Using configuration: $config" >&2
  fi

  FLAKE_PATH="$flake"
  CONFIG_NAME="$config"
}

validate_config() {
  local has_inception
  has_inception=$(nix eval \
    --json "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception" \
    >/dev/null 2>&1 && echo "true" || echo "false")

  if "$has_inception" != "true" ]]; then
    echo "Error: Configuration '$CONFIG_NAME' was not created with nixos-inception.lib.nixosSystem" >&2
    echo "Make sure your flake uses nixos-inception.lib.nixosSystem instead of nixpkgs.lib.nixosSystem" >&2
    exit 1
  fi

  local deployment_config
  deployment_config=$(nix eval \
    --json "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.deploymentConfig" \
    2>/dev/null || echo "{}")

  if [[ -n "$AGE_KEY" ]]; then
    echo "Using CLI-provided age key: $AGE_KEY" >&2
  else
    local flake_age_key
    flake_age_key=$(echo "$deployment_config" | jq -r '.ageKeyFile // empty' \
      2>/dev/null || true)

    if [[ -n "$flake_age_key" ]]; then
      echo "Found age key in flake: $flake_age_key" >&2
    fi
  fi

  if [[ -n "$PORT" ]]; then
    echo "Using CLI-provided port: $PORT" >&2
  else
    local flake_port
    flake_port=$(echo "$deployment_config" | jq -r '.serverPort // empty' \
      2>/dev/null || true)
    if [[ -n "$flake_port" ]]; then
      echo "Found server port in flake: $flake_port" >&2
    fi
  fi
}

parse_args "$@"
resolve_flake
validate_config
