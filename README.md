# nixos-inception

**_From nothing to NixOS. One command. Zero access._**

<img src="https://raw.githubusercontent.com/wasdhjklxyz/nixos-inception/main/docs/logo.png" width="150" height="150">

Zero-touch NixOS deployment with secrets management. Boot an ISO or netboot
image, walk away, come back to a fully configured system.

**Work in progress!!** The core flow works but rough edges remain. Expect breaking
changes.

## What is this?

nixos-inception is a deployment framework that turns NixOS installation into a
single command. You define your system configuration as a normal NixOS flake,
add a `deployment` block, and run `nix run github:wasdhjklxyz/nixos-inception --
--flake .#myhost`. The framework:
1. Generates ephemeral mTLS certificates (CA never touches disk)
2. Builds a bootable installer image with the certs baked in
3. Starts a server waiting for the installer to phone home
4. The installer authenticates, receives the system over mTLS
5. Partitions disks (via disko), installs NixOS, and reboots

No USB drives to prepare. No interactive prompts. No manually copying SSH keys
around. One command, one reboot, done.

## Why?

Existing NixOS deployment tools either require SSH access to an already-running
system (nixos-rebuild, deploy-rs, colmena) or need manual intervention during
install (nixos-anywhere needs you to provide SSH keys and run commands).
nixos-inception handles the entire lifecycle from bare metal to running system,
including secrets provisioning, without any pre-existing access to the target
machine.

It also integrates with sops-nix so your secrets are encrypted at rest in your
repo and only decrypted on the target machine. The age key is generated during
deployment.

## How it works

The project is themed after the movie *Inception* since the deployment flow is
analogous to the film's structure.

### Components

| Component | Inception Role | What it does |
| --------- | -------------- | ------------ |
| Architect | Designs the dream | Go binary that generates certs, builds closures, serves configs over mTLS |
| Dreamer | Experiences the dream | Go binary that runs on the target, phones home, receives config, installs |
| Totem | Validates reality | mTLS certificates that prove both sides are who they claim to be |
| The Kick | Wakes you up | The reboot into the fully installed system |
| Limbo | The deepest level | The mTLS server layer inside architect, waiting for dreamers |

### Flow

```txt
                  YOU                          TARGET MACHINE
                   │                                 │
       nix run . -- --flake .#host                   │
                   │                                 │
          ┌────────▼────────┐                        │
          │ Generate certs  │  CA stays in memory    │
          │ (architect)     │  client certs to disk  │
          └────────┬────────┘                        │
          ┌────────▼────────┐                        │
          │ Build installer │  certs baked into      │
          │ image           │  ISO/netboot image     │
          └────────┬────────┘                        │
          ┌────────▼────────┐              ┌─────────▼──────────┐
          │ Start mTLS      │◄────mTLS────►│ Boot installer     │
          │ server (limbo)  │              │ (dreamer wakes up) │
          └────────┬────────┘              └─────────┬──────────┘
                   │                                 │
          ┌────────▼────────┐              ┌─────────▼──────────┐
          │ Send flake +    │─────mTLS────►│ Receive flake,     │
          │ secrets         │              │ build & install    │
          └────────┬────────┘              └─────────┬──────────┘
                   │                       ┌─────────▼──────────┐
                   │                       │ disko + install    │
                   │                       └─────────┬──────────┘
                   │                       ┌─────────▼──────────┐
                   │                       │ THE KICK (reboot)  │
                   │                       └────────────────────┘
                   │
                 done
```

### Security model

The mTLS handshake ensures mutual authentication. The dreamer proves it booted
from a legitimate installer image, and the architect proves it's the real
deployment server. The CA private key exists only in the architect process's
memory and is never written to disk.

Secrets flow:
- Your age key is passed to architect at invocation
- Architect generates a host-specific age key for the target
- Secrets are re-encrypted for the target's key and sent over mTLS
- sops-nix on the installed system decrypts secrets at activation time using the
  host key

## Quick start

### Prerequisites

- NixOS (or Nix with flakes enabled)
- A target machine you can boot from USB/network
- An age key for sops-nix secrets (optional)

### 1. Setup your flake

Your flake looks almost identical to a normal NixOS config — just swap
`nixpkgs.lib.nixosSystem` for `nixos-inception.lib.nixosSystem` and add a
`deployment` block:

```nix
# flake.nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    disko = {
      url = "github:nix-community/disko";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    sops-nix = {
      url = "github:Mic92/sops-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixos-inception = {
      url = "github:wasdhjklxyz/nixos-inception";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, disko, sops-nix, nixos-inception }: {
    nixosConfigurations.myhost = nixos-inception.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        disko.nixosModules.disko
        sops-nix.nixosModules.sops
        ./config.nix
      ];
      deployment = {  # see lib/deployment.nix for all options
        serverAddr = "192.168.0.123";  # your machine's IP
        serverPort = 12345;
        diskSelection = "auto";  # or "prompt" to confirm installation disk
      };
    };
  };
}
```

### 2. Write your system config

```nix
# config.nix
{ config, ... }:
{
  disko.devices.disk.main = {
    # When deployment.diskSelection is "auto" or "prompt", this MUST be set to
    # exactly "/dev/disk/by-id/nixos-inception-placeholder". The actual target
    # device is selected at install time by the architect.
    #
    # When deployment.diskSelection is "specific", set this to the actual
    # device path (e.g. "/dev/sda" or "/dev/disk/by-id/...").
    #
    # If you think this is cringe, I agree! See the following to discuss:
    #   https://github.com/wasdhjklxyz/nixos-inception/issues/19
    device = "/dev/disk/by-id/nixos-inception-placeholder";
    type = "disk";
    content = {
      type = "gpt";
      partitions = {
        MBR = {
          type = "EF02";
          size = "1M";
          priority = 1;
        };
        ESP = {
          type = "EF00";
          size = "500M";
          content = {
            type = "filesystem";
            format = "vfat";
            mountpoint = "/boot";
            mountOptions = [ "umask=0077" ];
          };
        };
        root = {
          size = "100%";
          content = {
            type = "filesystem";
            format = "ext4";
            mountpoint = "/";
          };
        };
      };
    };
  };

  sops = {
    defaultSopsFile = ./secrets.yaml;
    # Only supports age.keyFile (for now) a key is generated/written here.
    #   https://github.com/wasdhjklxyz/nixos-inception/issues/24
    age.keyFile = "/var/lib/sops-nix/key.txt";
    secrets.password.neededForUsers = true;
  };

  users.users.user = {
    isNormalUser = true;
    hashedPasswordFile = config.sops.secrets.password.path;
    extraGroups = [ "wheel" ];
  };

  system.stateVersion = "25.11";
}
```

### 3. Deploy

```bash
nix run github:wasdhjklxyz/nixos-inception -- --flake .#myhost
```

## Deployment options

Options are set in the `deployment` attribute of your
`nixos-inception.lib.nixosSystem` call. See `lib/deployment.nix` for the full
schema.

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `serverAddr` | string | `"127.0.0.1"` | IP or hostname the dreamer connects back to |
| `serverPort` | port | `8443` | Port for the mTLS server |
| ~~`squashfsCompression`~~ | ~~string~~ | ~~`"zstd -Xcompression-level 6"`~~ | ~~Squashfs compression for the installer image~~ **NOT WORKING RN** |
| `diskSelection` | enum | `"specific"` | How the target disk is selected. `"auto"` picks the largest disk, `"prompt"` asks for confirmation, `"specific"` uses the exact device in your disko config |
| `shipLock` | bool | `true` | Ship `flake.lock` with the deployment to pin input versions on the target |
| `installerModule` | string | `".../installation-cd-minimal.nix"` | NixOS module for the installer image — determines the boot medium type |
| `bootOverrides` | null or raw | `null` | Override boot-related NixOS options in the installer image |

### Boot modes

The boot medium is determined by `deployment.installerModule`. The default builds
an ISO image. For other platforms, override it with a different NixOS installer
module:

```nix
deployment.installerModule =
  "/nixos/modules/installer/cd-dvd/installation-cd-minimal.nix";

deployment.installerModule =
  "/nixos/modules/installer/sd-card/sd-image-aarch64-installer.nix";
```

Use `bootOverrides` to tweak installer-specific options like kernel parameters
or firmware without writing a custom installer module.

Yes I know this way of specifying an installer sucks - planning to fix this.

## Examples

The `examples/` and `test/` directories contain example configurations. Each of
these is a self-contained flake. To try:
```bash
SOPS_AGE_KEY_FILE=./examples/x86_64/key.txt nix run . -- --flake ./test/x86_64
# FYI user's password is "password"
```

## Cross-compilation

nixos-inception supports cross-architecture builds via binfmt/QEMU emulation. If
your machine has binfmt configured for the target architecture, it Just Works™.

For example, on a `x86_64-linux` host, the following enables cross compilation
for `aarch64-linux` systems (builds may be slow af):
```nix
boot.binfmt.emulatedSystems = [ "aarch64-linux" ];
```

## Contributing

This is a personal project and very much a work in progress. Issues and PRs
welcome. If something breaks, which it probably will, file an issue.

## License

[MIT](https://github.com/wasdhjklxyz/nixos-inception/blob/main/LICENSE)
