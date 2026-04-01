# Candango NvimM

A lightweight command-line utility written in **Go** to manage **Neovim**
versions directly from official GitHub releases. `nvimm` lets you install,
switch, and upgrade Neovim builds without touching your system package manager.

---

## Quick Start

**1. Install nvimm:**

```bash
go install github.com/candango/nvimm/cmd/nvimm@latest
```

**2. Install the latest stable Neovim:**

```bash
nvimm install stable -y
```

**3. Add the managed binary to your PATH** (add to your shell profile):

```bash
export PATH="$HOME/.nvimm/current/bin:$PATH"
```

That's it. Run `nvim --version` to confirm.

---

## Commands

### `nvimm list` — Browse available versions

Shows installed versions and what is available to install:

```
nvimm list

Installed versions
  nightly
* 0.11.5 (stable)
  0.11.4
  0.11.3

Available versions
  0.10.2
  0.9.5
  ...
```

The `*` marks the version currently set as active.

---

### `nvimm install <version>` — Install a version

Installs a specific version, the current stable, or the nightly build:

```bash
nvimm install 0.11.5     # specific version
nvimm install stable     # current stable release
nvimm install nightly    # latest nightly build
```

`nvimm` will prompt before downloading and before setting as current:

```
nvimm install 0.11.5

download 0.11.5? (y/n) y
Download completed. [OK]
...
Installation completed. [OK]
Installed at: /home/user/.nvimm/0.11.5
set 0.11.5 as current? (y/n) y
Version 0.11.5 set as current.
```

Skip all prompts with `-y`:

```bash
nvimm install -y stable
```

---

### `nvimm upgrade` — Upgrade to the latest stable

Checks if a newer stable release is available and installs it:

```
nvimm upgrade

Checking for latest stable...
0.11.6 not installed
download 0.11.6? (y/n) y
...
Installation completed. [OK]
Installed at: /home/user/.nvimm/0.11.6
set 0.11.6 as current? (y/n) y
Version 0.11.6 set as current.
```

If you are already on the latest stable:

```
nvimm upgrade

Checking for latest stable...
Already up to date: 0.11.5
```

Skip all prompts with `-y`:

```bash
nvimm upgrade -y
```

---

### `nvimm current` — Show or switch the active version

Show which version is currently active:

```bash
nvimm current

* 0.11.5
```

Switch to a different installed version:

```bash
nvimm current 0.11.3
```

---

## Configuration

`nvimm` can be configured via environment variables or flags:

| Variable | Flag | Description | Default |
|---|---|---|---|
| `NVIMM_PATH` | `-p` | Where releases are installed | `~/.nvimm` |
| `NVIMM_CACHE_PATH` | `-C` | Cache directory | `~/.cache/nvimm` |
| `NVIMM_MIN_RELEASE` | `-r` | Oldest release to show | `0.7.0` |
| `NVIMM_CONFIG_DIR` | `-d` | Config file directory | `~/.config/nvimm` |

---

## Development

Clone and build locally:

```bash
git clone https://github.com/candango/nvimm.git
cd nvimm
go build -o nvimm ./cmd/nvimm
```

Run the tests:

```bash
go test ./...
```

---

## License

This project is licensed under the MIT License. See the
[LICENSE](https://mit-license.org/) file for details.
