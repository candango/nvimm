# Candango NvimM

A lightweight command-line utility written in **Go** to manage **neovim**
versions directly from official releases. `nvimm` simplifies the process of
listing, installing and switching between different neovim builds.

---

## Features

* **List:** Fetch and display available versions from official neovim releases.
* **Install:** Download and extract specific versions (e.g., `v0.11.3`, `nightly`).
* **Switch:** Update symlinks to set the active `nvim` version on your machine.
* **Fast:** Compiled Go binary with no external runtime dependencies.

---

## Installation

### From Source
Ensure you have **Go** installed on your system:

```bash
go install github.com/candango/nvimm@latest

```

---

## Usage

```bash
# Usage: nvimm
# Please specify one command of: current, install or list
# Usage:
#   nvimm [Options] command <current | install | list>
#
# Application Options:
#   -v, --verbose           Enable verbose mode
#   -C, --cache-path=       Cache directory [$NVIMM_CACHE_PATH]
#   -c, --config=           Configuration file path [$NVIMM_CONFIG_PATH]
#   -d, --config-dir=       Configuration file directory [$NVIMM_CONFIG_DIR]
#   -n, --config-file-name= Configuration file name (default: nvimm.yml) [$NVIMM_CONFIG_FILE_NAME]
#   -p, --path=             Path where Neovim releases are installed [$NVIMM_PATH]
#   -r, --min-release=      Neovim minimal release (default: 0.7.0) [$NVIMM_MIN_RELEASE]
#
# Help Options:
#   -h, --help              Show this help message
#
# Available commands:
#   current  Display the active or installed Neovim version
#   install  Install the latest or a specific Neovim version
#   list     List Neovim installed versions
```

### List installed and available versions

Show installed and available Neovim versions:

```bash
nvimm list

Installed versions
  nightly
* 0.11.5 (stable)
  0.11.4
  0.11.3
  0.11.2
  0.11.1
  0.11.0
  0.10.4
  0.10.3
  0.10.1
  0.10.0

Available versions
  0.10.2
  0.9.5
  0.9.4
  0.9.2
  0.9.1
  0.9.0
```

### Show current version

Display the active Neovim version:

```bash
nvimm current

* 0.11.5
```

### Install a specific version

Download and install a specific tag or build:

```bash
nvimm install 0.11.5

Download completed. [OK]
Downloaded file: /home/fpiraz/.cache/nvimm/nvim-linux-x86_64.tar.gz

Checksum calculated. [OK]
Calculated checksum: sha256:_a_really_trust_me_bro_hash_
Expected checksum:   sha256:_a_really_trust_me_bro_hash_

Extraction completed. [OK]

Installation completed. [OK]
Installed at: /opt/nvim/0.11.5
```

### Set the current version

Switch the active `nvim` binary to a previously installed version:

```bash
nvimm use v0.11.3
```

---

## Development

To contribute or build the project locally:

1. Clone the repository:
```bash
git clone https://github.com/candango/nvimm.git

```

2. Navigate to the directory:
```bash
cd nvimm

```


3. Build the binary:

```bash
go build -o nvimm .

```
---

## License

This project is licensed under the MIT License. See the
[LICENSE](https://mit-license.org/) file for details.
