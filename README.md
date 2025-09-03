# pwgen

Generate passphrases with interactive TUI and CLI modes

---

[![GitHub release](https://img.shields.io/github/v/release/idelchi/pwgen)](https://github.com/idelchi/pwgen/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/idelchi/pwgen.svg)](https://pkg.go.dev/github.com/idelchi/pwgen)
[![Go Report Card](https://goreportcard.com/badge/github.com/idelchi/pwgen)](https://goreportcard.com/report/github.com/idelchi/pwgen)
[![Build Status](https://github.com/idelchi/pwgen/actions/workflows/github-actions.yml/badge.svg)](https://github.com/idelchi/pwgen/actions/workflows/github-actions.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`pwgen` generates passphrases using word-based diceware methods.

- Interactive TUI with slot-machine interface and column locking
- CLI mode for scripting with JSON output
- Customizable separators, word count, digits, symbols, and casing
- External diceware wordlist libraries
- Clipboard integration and entropy calculation

## Installation

For a quick installation, you can use the provided installation script:

```sh
curl -sSL https://raw.githubusercontent.com/idelchi/pwgen/refs/heads/main/install.sh | sh -s -- -d ~/.local/bin
```

## Usage

### Interactive TUI

Launch the interactive TUI (default mode):

```sh
# Start interactive TUI
pwgen
```

**TUI Controls:**

- `Space` – Generate new unlocked columns
- `s` – Cycle separators: `-` → `_` → `.` → `space` → `none`
- `w`/`W` – Increase/decrease word count (1-10)
- `d`/`D` – Increase/decrease digit count (0-5)
- `x`/`X` – Increase/decrease symbol count (0-5)
- `+`/`-` – Universal increase/smart decrease
- `ctrl+s` – Cycle casing styles
- `1-9` – Lock/unlock specific columns
- `Enter` – Lock/unlock focused column
- `c` – Copy to clipboard
- `v` – Toggle visibility (mask/unmask)
- `n` – Generate new passphrase (unlock all)
- `?` – Show detailed help

### CLI Mode

Generate passphrases directly from command line:

```sh
# Generate default passphrase (4 words, mixed case, hyphen-separated)
pwgen gen
```

```sh
# Customize generation parameters
pwgen gen --words 5 --sep "." --caps title --digits 2 --symbols 1
```

```sh
# JSON output for scripting
pwgen gen --json --count 10
```

```sh
# Use pattern syntax for complex generation
pwgen gen --pattern "W:title SEP W:lower SEP DD{2} SEP S"
```

## Dictionary Support

List available wordlist dictionaries:

```sh
pwgen dicts
```

Currently supported:

- **eff** – EFF Large Wordlist (7,776 words, 12.9 bits entropy) using external cryptographic libraries

## Commands

<details>
<summary><strong>gen</strong> — Generate passphrases</summary>

- **Usage:** `pwgen gen [flags]`
- **Flags:**
  - `--words, -w <int>` – Number of words (default: 4)
  - `--digits <int>` – Number of digits (default: 0)
  - `--symbols <int>` – Number of symbols (default: 0)
  - `--sep <string>` – Separator between tokens (default: "-")
  - `--caps <string>` – Casing style: lower, upper, title, mixed (default: "mixed")
  - `--dict <string>` – Dictionary to use (default: "eff")
  - `--pattern <string>` – Custom pattern DSL
  - `--count <int>` – Number of passphrases to generate (default: 1)
  - `--json` – Output in JSON format
  - `--copy` – Copy to clipboard
  - `--kebab` – Use kebab-case separators
  - `--snake` – Use snake_case separators
  - `--camel` – Use camelCase (no separators)

</details>

<details>
<summary><strong>dicts</strong> — List available dictionaries</summary>

- **Usage:** `pwgen dicts`

</details>

<details>
<summary><strong>check</strong> — Analyze passphrase strength</summary>

- **Usage:** `pwgen check [passphrase]`
- **Flags:**
  - `--json` – Output analysis in JSON format

</details>

<details>
<summary><strong>version</strong> — Show version information</summary>

- **Usage:** `pwgen version`

</details>

## Pattern DSL

Use the pattern DSL for complex passphrase generation:

```sh
# Word + separator + word + digits + symbols
pwgen gen --pattern "W:title SEP W:lower SEP DD{2} SEP SS{1}"

# Multiple words with mixed casing
pwgen gen --pattern "W:upper W:lower W:title W:mixed"
```

**Pattern Elements:**

- `W[:style]` – Word with optional casing (lower, upper, title, mixed)
- `D{n}` – n digits
- `S{n}` – n symbols
- `SEP` – Separator token

## Security Features

- Uses `crypto/rand` for random generation
- External diceware wordlist libraries
- Entropy calculation using logarithmic math
- Automatic memory wiping for passphrases

## Output Formats

### Text Output

```text
sQuiRrel-aDmit-conTAin-reaDy
```

### JSON Output

```json
{
  "passphrase": "sQuiRrel-aDmit-conTAin-reaDy",
  "entropy_bits": 51.6,
  "strength": "Strong",
  "word_count": 4,
  "character_count": 29,
  "pattern": "W:mixed SEP W:mixed SEP W:mixed SEP W:mixed"
}
```

## Column Locking System

The TUI features a unique column locking system:

1. **Generate** initial passphrase: `word-word-word-word`
2. **Lock** specific columns you like (press number keys or Enter)
3. **Regenerate** unlocked columns while preserving locked ones
4. **Change separators** - locked columns keep original separators
5. **Mix and match** different parts for perfect passphrases

**Example workflow:**

1. Generate: `apple-banana-cherry-date`
2. Lock "banana" (press `2`): `apple-🔒banana-cherry-date`
3. Regenerate: `grape-🔒banana-lemon-mango`
4. Change separator to `_`: `grape_🔒banana_lemon_mango` (locked keeps `-`)

## Clipboard Integration

Supports multiple clipboard providers:

- **OSC52** for terminal/SSH sessions
- **pbcopy** (macOS)
- **xclip** (Linux X11)
- **wl-copy** (Linux Wayland)

## Examples

```sh
# Simple 4-word passphrase
pwgen gen
# → "sQuiRrel-aDmit-conTAin-reaDy"

# Corporate-friendly with mixed separators
pwgen gen --words 3 --digits 2 --sep "_"
# → "Account_System_Security_47"

# High-security with symbols
pwgen gen --words 6 --symbols 2 --caps lower
# → "correct-horse-battery-staple-mountain-ocean-#$"

# Generate 5 passphrases in JSON for scripting
pwgen gen --count 5 --json | jq '.[] | .passphrase'
```
