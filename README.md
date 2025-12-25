# Candango nviman

A lightweight command-line utility written in **Go** to manage **neovim**
versions directly from official releases. `nviman` simplifies the process of
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
go install [github.com/candango/nviman@latest](https://github.com/candango/nviman@latest)

```

---

## Usage

### List available versions

Retrieve a list of versions available from the official neovim repository:

```bash
nviman list

```

### Install a specific version

Download and install a specific tag or build:

```bash
nviman install v0.11.3

```

### Set the current version

Switch the active `nvim` binary to a previously installed version:

```bash
nviman use v0.11.3

```

---

## Development

To contribute or build the project locally:

1. Clone the repository:
```bash
git clone [https://github.com/candango/nviman.git](https://github.com/candango/nviman.git)

```

2. Navigate to the directory:
```bash
cd nviman

```


3. Build the binary:

```bash
go build -o nviman .

```



---

## License

This project is licensed under the MIT License. See the
[LICENSE](https://mit-license.org/) file for details.
