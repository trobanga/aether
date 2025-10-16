#!/usr/bin/env bash

# Install shell completions for aether CLI
# Supports bash, zsh, and oh-my-zsh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Check if aether binary exists
if ! command -v aether &> /dev/null; then
    # Try local binary
    if [ -f "$PROJECT_ROOT/bin/aether" ]; then
        AETHER_BIN="$PROJECT_ROOT/bin/aether"
        print_info "Using local aether binary: $AETHER_BIN"
    else
        print_error "aether command not found. Please build or install aether first."
        echo "  Run: make build && make install"
        exit 1
    fi
else
    AETHER_BIN="aether"
    print_success "Found aether in PATH"
fi

# Detect shell
SHELL_NAME=$(basename "$SHELL")
print_info "Detected shell: $SHELL_NAME"

install_bash() {
    print_info "Installing bash completion..."

    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        COMPLETION_DIR="/etc/bash_completion.d"
        if [ -d "$COMPLETION_DIR" ]; then
            sudo $AETHER_BIN completion bash > "$COMPLETION_DIR/aether"
            print_success "Installed to $COMPLETION_DIR/aether"
        else
            # Fallback to user directory
            mkdir -p ~/.bash_completion.d
            $AETHER_BIN completion bash > ~/.bash_completion.d/aether
            print_success "Installed to ~/.bash_completion.d/aether"

            # Add to bashrc if not already present
            if ! grep -q "bash_completion.d/aether" ~/.bashrc; then
                echo "source ~/.bash_completion.d/aether" >> ~/.bashrc
                print_info "Added source command to ~/.bashrc"
            fi
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            COMPLETION_DIR="$(brew --prefix)/etc/bash_completion.d"
            mkdir -p "$COMPLETION_DIR"
            $AETHER_BIN completion bash > "$COMPLETION_DIR/aether"
            print_success "Installed to $COMPLETION_DIR/aether"
        else
            print_error "Homebrew not found. Cannot determine completion directory."
            exit 1
        fi
    fi

    print_info "Restart your shell or run: source ~/.bashrc"
}

install_zsh() {
    print_info "Installing zsh completion..."

    # Standard zsh completion directory
    mkdir -p "${HOME}/.zsh/completions"
    $AETHER_BIN completion zsh > "${HOME}/.zsh/completions/_aether"
    print_success "Installed to ~/.zsh/completions/_aether"

    # Add to fpath in .zshrc if not present
    if ! grep -q ".zsh/completions" ~/.zshrc 2>/dev/null; then
        echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
        echo 'autoload -U compinit && compinit' >> ~/.zshrc
        print_info "Added completion setup to ~/.zshrc"
    fi

    print_info "Restart your shell or run: exec zsh"
}

install_ohmyzsh() {
    print_info "Installing oh-my-zsh completion..."

    if [ ! -d "$HOME/.oh-my-zsh" ]; then
        print_error "oh-my-zsh not found at ~/.oh-my-zsh"
        print_info "Falling back to standard zsh installation..."
        install_zsh
        return
    fi

    # Create custom plugin directory
    PLUGIN_DIR="$HOME/.oh-my-zsh/custom/plugins/aether"
    mkdir -p "$PLUGIN_DIR"

    # Generate completion file
    $AETHER_BIN completion zsh > "$PLUGIN_DIR/_aether"
    print_success "Installed to $PLUGIN_DIR/_aether"

    # Check if plugin is already in .zshrc
    if grep -q "plugins=(" ~/.zshrc; then
        if grep -q "plugins=.*aether" ~/.zshrc; then
            print_success "Plugin 'aether' already in ~/.zshrc plugins array"
        else
            print_info "Add 'aether' to your plugins array in ~/.zshrc:"
            echo ""
            echo "  plugins=(...existing plugins... aether)"
            echo ""
            print_info "Then restart your shell: exec zsh"
        fi
    else
        print_info "Add the following to your ~/.zshrc:"
        echo ""
        echo "  plugins=(aether)"
        echo ""
        print_info "Then restart your shell: exec zsh"
    fi
}

install_fish() {
    print_info "Installing fish completion..."

    mkdir -p ~/.config/fish/completions
    $AETHER_BIN completion fish > ~/.config/fish/completions/aether.fish
    print_success "Installed to ~/.config/fish/completions/aether.fish"
    print_info "Restart your fish shell"
}

# Main installation logic
case "$SHELL_NAME" in
    bash)
        install_bash
        ;;
    zsh)
        # Check if oh-my-zsh is installed
        if [ -d "$HOME/.oh-my-zsh" ]; then
            echo ""
            echo "Detected oh-my-zsh installation."
            read -p "Install as oh-my-zsh plugin? [Y/n] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
                install_ohmyzsh
            else
                install_zsh
            fi
        else
            install_zsh
        fi
        ;;
    fish)
        install_fish
        ;;
    *)
        print_error "Unsupported shell: $SHELL_NAME"
        echo ""
        echo "Supported shells: bash, zsh, fish"
        echo ""
        echo "To manually generate completions:"
        echo "  aether completion [bash|zsh|fish] > /path/to/completion/file"
        exit 1
        ;;
esac

echo ""
print_success "Installation complete!"
