# Candango NvimIM

Pronounced like: **Nvim I am!**

A lightweight command-line installer written in **Go** for downloading
**Neovim** versions directly from official GitHub releases. `nvimim` lets you
install, switch, and upgrade Neovim builds without touching your system package
manager.

> **Rename notice:** This project was renamed from `nvimm` to `nvimim` as a clean break. There is no compatibility layer for the old CLI command, module path, environment variables, or default directories. Update `nvimm` references to `nvimim`, `NVIMM_*` to `NVIMIM_*`, and `~/.nvimm`-style paths to their `~/.nvimim` equivalents before upgrading.

---

## Quick Start

**1. Install nvimim:**

```bash
go install github.com/candango/nvimim/cmd/nvimim@latest
```

**2. Install the latest stable Neovim:**

```bash
nvimim install stable -y
```

**3. Add the installed binary to your PATH** (add to your shell profile):

```bash
export PATH="$HOME/.nvimim/current/bin:$PATH"
```

That's it. Run `nvim --version` to confirm.

---

## Commands

### `nvimim list` — Browse available versions

Shows installed versions and what is available to install:

```
nvimim list

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

### `nvimim install <version>` — Install a version

Installs a specific version, the current stable, or the nightly build:

```bash
nvimim install 0.11.5     # specific version
nvimim install stable     # current stable release
nvimim install nightly    # latest nightly build
```

`nvimim` will prompt before downloading and before setting as current:

```
nvimim install 0.11.5

download 0.11.5? (y/n) y
Download completed. [OK]
...
Installation completed. [OK]
Installed at: /home/user/.nvimim/0.11.5
set 0.11.5 as current? (y/n) y
Version 0.11.5 set as current.
```

Skip all prompts with `-y`:

```bash
nvimim install -y stable
```

---

### `nvimim upgrade` — Upgrade to the latest stable

Checks if a newer stable release is available and installs it:

```
nvimim upgrade

Checking for latest stable...
0.11.6 not installed
download 0.11.6? (y/n) y
...
Installation completed. [OK]
Installed at: /home/user/.nvimim/0.11.6
set 0.11.6 as current? (y/n) y
Version 0.11.6 set as current.
```

If you are already on the latest stable:

```
nvimim upgrade

Checking for latest stable...
Already up to date: 0.11.5
```

Skip all prompts with `-y`:

```bash
nvimim upgrade -y
```

---

### `nvimim current` — Show or switch the active version

Show which version is currently active:

```bash
nvimim current

* 0.11.5
```

Switch to a different installed version:

```bash
nvimim current 0.11.3
```

---

## Configuration

`nvimim` can be configured via environment variables or flags:

| Variable | Flag | Description | Default |
|---|---|---|---|
| `NVIMIM_PATH` | `-p` | Where releases are installed | `~/.nvimim` |
| `NVIMIM_CACHE_PATH` | `-C` | Cache directory | `~/.cache/nvimim` |
| `NVIMIM_MIN_RELEASE` | `-r` | Oldest release to show | `0.7.0` |
| `NVIMIM_CONFIG_DIR` | `-d` | Config file directory | `~/.config/nvimim` |

---

## Development

Clone and build locally:

```bash
git clone https://github.com/candango/nvimim.git
cd nvimim
go build -o nvimim ./cmd/nvimim
```

Run the tests:

```bash
go test ./...
```

---

## Support

NvimIM is one of the [Candango Open Source Group
](https://www.candango.org/projects/) initiatives.

For bug reports, feature requests, and questions, please open an
[issue](https://github.com/candango/nvimim/issues) on GitHub.

## License

NvimIM is distributed under the MIT License. See the
[LICENSE](LICENSE) file for details.
