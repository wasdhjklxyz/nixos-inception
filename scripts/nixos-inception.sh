#!/usr/bin/env bash

set -euo pipefail

FLAKE_PATH=""
CONFIG_NAME=""
AGE_KEY=""
PORT=""
BOOT_MODE=""
CLEANUP_PIPE=""
CLEANUP_DIR=""
SYSTEM_TOPLEVEL=""
CLOSURE_FILE=""
CERT_DURATION="10m" # FIXME: See architect/flags.go
CERT_SKEW="5m" # FIXME: See architect/flags.go
DISKO_SCRIPT=""
DISKO_DEVICE=""
DISK_SELECTION=""

print_error() {
  echo -e "\033[1;31merror:\033[0m $1" >&2
}

print_warning() {
  echo -e "\033[1;35mwarning:\033[0m $1" >&2
}

print_info() {
  echo $1 >&2
}

show_help() {
  cat << EOF
Usage: nix run github:wasdhjklxyz/nixos-inception -- [OPTIONS]

OPTIONS:
    --help                    Show this help message
    --flake PATH#CONFIG       Path to flake configuration
    --age-key PATH            Age identity file
    --port NUM                Listen port
    --netboot                 Use net boot
    --cert-duration DURATION  Certificate validity duration
    --cert-skew DURATION      Certificate start time offset

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
      --netboot)
        BOOT_MODE="netboot"
        shift
        ;;
      --cert-duration)
        if [[ -z "${2:-}" ]]; then
          print_error "--cert-duration requires an argument"
          exit 1
        fi
        CERT_DURATION="$2"
        shift 2
        ;;
      --cert-skew)
        if [[ -z "${2:-}" ]]; then
          print_error "--cert-skew requires an argument"
          exit 1
        fi
        CERT_SKEW="$2"
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
  if [[ -z "$FLAKE_PATH" ]]; then
    FLAKE_PATH="."
  fi

  if [[ "$FLAKE_PATH" == *"#"* ]]; then
    CONFIG_NAME="${FLAKE_PATH##*#}"
    FLAKE_PATH="${FLAKE_PATH%#*}"
  fi

  if [[ -z "$CONFIG_NAME" ]]; then
    local configs
    configs=$(nix eval --json "$FLAKE_PATH#nixosConfigurations" \
      --apply "builtins.attrNames") || exit 1

    local count=$(echo "$configs" | jq "length")
    if [[ "$count" -eq 1 ]]; then
      CONFIG_NAME=$(echo "$configs" | jq -r '.[0]')
      print_warning "using only available configuration '$CONFIG_NAME'"
    else
      # NOTE: Trusting nix to fail if no configs found - this runs if it doesnt
      print_error "multiple configurations found:"
      echo "$configs" | jq -r ".[]" | sed "s/^/  /"
      exit 1
    fi
  fi
}

validate_config() {
  if ! nix eval "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME" \
    --apply 'x: true' >/dev/null; then
    exit 1
  fi

  if ! nix eval "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception" \
    --apply 'x: true' >/dev/null; then
    exit 1
  fi

  local deployment_config=$(nix eval \
    --json "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.deploymentConfig" \
    2>/dev/null || echo "{}")

  if [[ -z "$AGE_KEY" ]]; then
    AGE_KEY=$(echo "$deployment_config" | jq -r '.ageKeyFile // empty' \
      2>/dev/null || true)
    if [[ -n "$AGE_KEY" && "$AGE_KEY" != /* ]]; then
      AGE_KEY="${FLAKE_PATH}/${AGE_KEY}"
    fi
  fi

  if [[ -z "$PORT" ]]; then
    PORT=$(echo "$deployment_config" | jq -r '.serverPort // empty' \
      2>/dev/null || true)
  fi

  if [[ -z "$BOOT_MODE" ]]; then
    BOOT_MODE=$(echo "$deployment_config" | jq -r '.bootMode // "iso"')
  fi
}

cleanup() {
  [[ -n "$CLEANUP_PIPE" ]] && rm -f "$CLEANUP_PIPE"
  [[ -n "$CLEANUP_DIR" ]] && rm -rf "$CLEANUP_DIR"
  [[ -n "$CLOSURE_FILE" ]] && rm -f "$CLOSURE_FILE"
}

start_architect() {
  trap cleanup EXIT

  print_info "preparing control pipe..."
  CLEANUP_PIPE=$(mktemp -u --suffix=".nixos-inception-ctl")
  mkfifo "$CLEANUP_PIPE"

  print_info "building system toplevel..."
  SYSTEM_TOPLEVEL=$(nix build --print-out-paths \
    "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME.config.system.build.toplevel")

  print_info "querying disk info..."
  DISKO_SCRIPT=$(nix build --print-out-paths \
    "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME.config.system.build.diskoScript")
  rm result # FIXME: above makes symlink. dogshit fix. for system build later
  DISKO_DEVICE=$(nix eval --raw \
    "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.diskoDevice")
  DISK_SELECTION=$(nix eval --raw \
    "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.deploymentConfig.diskSelection")

  print_info "querying requisites..."
  CLOSURE_FILE=$(mktemp)
  nix-store -qR "$SYSTEM_TOPLEVEL" > "$CLOSURE_FILE"
  nix-store -qR "$DISKO_SCRIPT" >> "$CLOSURE_FILE"

  coproc ARCHITECT { \
    architect --age-key "$AGE_KEY" --ctl-pipe "$CLEANUP_PIPE" --lport "$PORT" \
    --toplevel "$SYSTEM_TOPLEVEL" --closure "$CLOSURE_FILE" \
    --cert-duration "$CERT_DURATION" --cert-skew "$CERT_SKEW" \
    --disko-script "$DISKO_SCRIPT" --disko-device "$DISKO_DEVICE" \
    --disk-selection "$DISK_SELECTION"; \
  }

  read -r CLEANUP_DIR <&${ARCHITECT[0]}

  print_info "building bootable image..."
  if [[ "$BOOT_MODE" == "netboot" ]]; then
    NIXOS_INCEPTION_CERT_DIR="$CLEANUP_DIR" \
      nix build --impure \
      "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.netboot.config.system.build.kexecTree" \
    || { kill "$ARCHITECT_PID" 2>/dev/null; exit 1; }
    print_info "Done. kexec tree at ./result"
  else
    NIXOS_INCEPTION_CERT_DIR="$CLEANUP_DIR" \
      nix build --impure \
      "$FLAKE_PATH#nixosConfigurations.$CONFIG_NAME._inception.iso.config.system.build.isoImage" \
    || { kill "$ARCHITECT_PID" 2>/dev/null; exit 1; }
    print_info "Done. iso at ./result/iso/"
  fi

  echo "START" >&${ARCHITECT[1]}

  trap 'echo "stopping server..."; kill "$ARCHITECT_PID" 2>/dev/null; exit 0' INT TERM
  wait "$ARCHITECT_PID"
}

parse_args "$@"
resolve_flake
validate_config
start_architect
