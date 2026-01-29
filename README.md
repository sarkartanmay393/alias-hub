# Alias Hub (ah)

**The Ultimate Shell Alias Manager.**  
Manage, share, and sync shell aliases across your machines with conflict detection, live updates, and zero bloat.

## Why `ah`?

Most alias tools are just "dotfile managers" or heavy frameworks like Oh-My-Zsh. `ah` is different:

*   **Live Updates**: Install an alias in one terminal tab, and it works in all other open tabs *instantly*. No `source ~/.zshrc` needed.
*   **Daemon-less**: We don't run a background process. We use a lightweight "Stat-Check Hook" that costs microseconds.
### ⚡ Features
- **Central Registry**: Secure packages from a curated repository.
- **Conflict Resolution UI**: A beautiful web interface (`ah resolve`) to handle alias collisions properly.
- **Strict Security**: Aliases are parsed and sanitized; no arbitrary code execution allowed.
- **Atomic Operations**: Zero race conditions during concurrent installs.
- **Zero Dependency**: Single binary, works on macOS and Linux.
*   **Conflict Safety**: Used `g` for `git`? `ah` won't let you install a package that sets `g` to `grep` without a warning.
*   **Performance**: Compiles all active aliases into a single O(1) source file.

## 🚀 Installation

### 🍺 Homebrew (macOS/Linux)
```bash
brew tap sarkartanmay393/ah https://github.com/sarkartanmay393/ah
brew install ah
```

### 🚀 Automatic Install (Linux/Mac)
```bash
curl -sL https://raw.githubusercontent.com/sarkartanmay393/ah/main/install.sh | bash
```

### 📦 Go Install (Developers)
```bash
go install github.com/sarkartanmay393/ah@latest
```

### 💻 Manual Build
```bash
# Clone and Build
git clone https://github.com/sarkartanmay393/ah
cd ah
go build -o ah main.go
```

### ✨ Initialize
Initialize `ah` (Automatically updates .zshrc/.bashrc)
```bash
ah init
```

### 📦 Install a Package
Install any package from the registry.
```bash
ah install clawdbot
```
*Prompts you to review aliases before enabling.*

```bash
ah disable my-package
# Aliases gone instantly in all tabs

ah enable my-package
# Aliases back instantly
```

### 🔍 Search
Find packages in the registry.
```bash
ah search git
```

### 🛠 Management
```bash
ah list                 # List installed packages
ah remove my-package       # Delete package & symlinks
ah doctor --fix         # Fix broken paths/permissions
```

## How it Works

1.  **Storage**: Packages are cloned to `~/.ah/packages`.
2.  **Activation**: Enabled packages are symlinked to `~/.ah/active`.
3.  **Compilation**: `ah` compiles all active scripts into `~/.ah/aliases.compiled.sh`.
4.  **Live Sync**: Your shell prompt checks the timestamp of `~/.ah/state`. If it changed, it re-sources the compiled file.

## Directory Structure
```
~/.ah/
├── active/              # Symlinks to enabled packages
├── packages/            # Git clones of installed repos
├── aliases.compiled.sh  # The single file your shell sources
└── state                # 0-byte timestamp file for sync
```
