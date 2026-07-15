# NvimIM rename follow-up

This document tracks the manual and external follow-up required after the in-repo rename from `nvimm` to `nvimim`.

## Naming rules

- Human-facing brand casing: `NvimIM`
- Technical identifiers stay lowercase: `nvimim`
- CLI command: `nvimim`
- Go module path: `github.com/candango/nvimim`
- Environment variable namespace: `NVIMIM_*`

## Manual follow-up checklist

### GitHub repository and project metadata

- Verify the GitHub repository name is `candango/nvimim`
- Update the repository description to use the new name `NvimIM`
- Review pinned repositories, profile links, and organization references
- Update issue templates, PR templates, and labels if they mention `nvimm`
- [x] Fix issue `#18` body formatting after shell escaping corrupted Markdown

### Releases and distribution

- Review release titles and release notes for `NvimIM` branding
- Ensure release automation publishes binaries under the `nvimim` command name
- Confirm install instructions use:
  - `go install github.com/candango/nvimim/cmd/nvimim@latest`
- Update any package manager formulas, manifests, or third-party install docs if they exist

### Module and downstream consumers

- Update import paths in downstream repos from `github.com/candango/nvimm` to `github.com/candango/nvimim`
- Check README badges, blog posts, gists, dotfiles, and snippets that still reference the old module path
- Review CI or automation in external repos that invoke `nvimm`

### Local user migration

This rename is a clean break. No compatibility layer is provided.

Users should update:

- command references from `nvimm` to `nvimim`
- managed install path references from `~/.nvimm` to `~/.nvimim`
- cache path references from `~/.cache/nvimm` to `~/.cache/nvimim`
- config directory references from `~/.config/nvimm` to `~/.config/nvimim`
- environment variables from `NVIMM_*` to `NVIMIM_*`

Custom user-managed runtime locations remain valid, such as `/opt/nvim/current`, as long as the surrounding configuration points to the correct renamed `NVIMIM_*` settings or equivalent config.

### Local clone maintenance

For existing clones, verify the remote:

```bash
git remote -v
```

If needed:

```bash
git remote set-url origin git@github.com:candango/nvimim.git
```

## Validation checklist

- `go test ./...` passes
- no tracked files outside migration documentation still reference `nvimm`
- release notes and GitHub metadata use `NvimIM` where human-facing branding appears
- technical paths and commands use `nvimim`
