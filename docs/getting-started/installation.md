# Installation

This guide walks you through installing Aether on your system.

## Prerequisites

Before installing Aether, ensure you have:

- **Go 1.21 or later** (required)
- **make** (for building from source)
- **git** (for cloning the repository)

Optional tools:
- TORCH server access (for FHIR data extraction)
- DIMP service (for pseudonymization features)

## Installation from Source

### 1. Clone the Repository

```bash
git clone https://github.com/trobanga/aether.git
cd aether
```

### 2. Build Aether

Using the Makefile:

```bash
make build
```

This compiles the Aether binary for your system.

### 3. Install to System PATH

**Option A: System-wide installation (requires sudo)**

```bash
sudo make install
# Installs to /usr/local/bin
```

**Option B: User-local installation (no sudo required)**

```bash
make install-local
# Installs to ~/.local/bin
# Ensure ~/.local/bin is in your PATH
```

To verify ~/.local/bin is in your PATH:

```bash
echo $PATH | grep '.local/bin'
```

If not present, add to your shell configuration file (~/.bashrc, ~/.zshrc, etc.):

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Verification

Verify the installation by checking the version:

```bash
aether --help
```

You should see the help output with available commands and options.

## Shell Completions (Optional)

For improved command-line experience, install shell completions.

### For Zsh (oh-my-zsh)

Automatic installation:

```bash
./scripts/install-completions.sh
```

Or manual installation:

```bash
mkdir -p ~/.oh-my-zsh/custom/plugins/aether
aether completion zsh > ~/.oh-my-zsh/custom/plugins/aether/_aether
```

Then add 'aether' to the plugins array in ~/.zshrc:

```bash
plugins=(... aether)
```

Reload your shell:

```bash
exec zsh
```

### For Bash

```bash
aether completion bash | sudo tee /etc/bash_completion.d/aether
```

Then reload your shell:

```bash
exec bash
```

### For Fish or Other Shells

For comprehensive shell completion setup for fish, bash, zsh, and other shells:

```bash
./scripts/install-completions.sh
```

Or see manual instructions:

```bash
aether completion --help
```

## Next Steps

- [Quick Start Guide](./quick-start.md) - Get started with your first pipeline
- [Configuration Guide](./configuration.md) - Learn how to configure Aether for your environment
