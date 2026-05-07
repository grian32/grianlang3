# CLI

`gl3` provides subcommands for building programs and generating `.gl3` definitions from C headers.

## Overview

```text
gl3 [command] [--flags]
```

### Top-level commands

| Command | Description |
| ------- | ----------- |
| `build` | Compile GL3 files to an executable |
| `exdef` | Extract `#define` values from a C header into a `.gl3` file |
| `help` | Show help for a command |

### Top-level flags

| Flag | Description |
| ---- | ----------- |
| `-h`, `--help` | Show help for `gl3` |
| `-v`, `--version` | Show the compiler version |

## `build`

Compile GL3 files to an executable.

### Usage

```text
gl3 build [--flags]
```

### Flags

| Flag | Description |
| ---- | ----------- |
| `--dbg` | Print the AST for all compiled files, along with the `clang` command used for compilation |
| `--keepll` | Save the `.ll` files produced by compilation |
| `--noexecbuild` | Generate LLVM IR without running the `clang` build step |
| `-h`, `--help` | Show help for `build` |

### Example

```bash
./gl3 build example.gl3 -o output
./output
```

## `exdef`

Extract `#define` statements from a C header file and emit them as constants in a `.gl3` file.

### Usage

```text
gl3 exdef [--flags]
```

### Flags

| Flag | Description |
| ---- | ----------- |
| `-i`, `--input` | Path to the input header file |
| `-o`, `--output` | Path to the output `.gl3` file |
| `-h`, `--help` | Show help for `exdef` |

### Example

```bash
./gl3 exdef -i example.h -o example.gl3
```
