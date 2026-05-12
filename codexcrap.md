# grianlang3 Security Scan Report

Scan target: repository-wide scan of `/Users/grian/Projects/grianlang3` at commit `897dffb`.

## Finding: Build input path escapes `lltemp` and writes LLVM IR outside the temp directory

- Priority: P2
- Severity: medium
- Confidence: high
- CWE: CWE-22 Path Traversal, CWE-73 External Control of File Name or Path
- Affected lines: `/Users/grian/Projects/grianlang3/cli/build.go:34`, `/Users/grian/Projects/grianlang3/cli/build.go:81`

### Summary

`RunBuildCmd` reads raw user-supplied build input paths and then uses the same string to build an intermediate output path with `fmt.Sprintf("./lltemp/%s.ll", file)`. A path containing `../` escapes `./lltemp`, so `os.Create` writes generated LLVM IR outside the temporary directory.

### Validation

Validated through the real CLI: `go run . build ../../../../../private/tmp/gl3scan/input.gl3 --noexecbuild` created `/private/tmp/gl3scan/input.gl3.ll` containing generated LLVM IR.

### Reachability Analysis

This crosses the local compiler CLI filesystem boundary. It is not remotely reachable by default, but it matters for CI, scripted builds, or workflows that compile attacker-influenced project paths.

### Attack Path

1. Attacker influences the path passed to `gl3 build`.
2. `cli/build.go:35` reads that path.
3. `cli/build.go:81` interpolates the raw path under `./lltemp`.
4. `cli/build.go:82` creates the normalized escaped path and writes LLVM IR.

### Severity Analysis

Medium: local arbitrary file write with generated content and a `.ll` suffix, constrained by attacker influence over build arguments.

### Remediation

Create intermediates in an OS temp directory, derive safe filenames with `filepath.Base` or a hash, and verify the cleaned output path remains under the temp directory before creating it.

## Finding: Source ending with line comment and no newline hangs the lexer

- Priority: P2
- Severity: medium
- Confidence: high
- CWE: CWE-835 Loop with Unreachable Exit Condition
- Affected lines: `/Users/grian/Projects/grianlang3/lexer/lexer.go:72`

### Summary

The line-comment skipper loops until it sees `\n`, but EOF is represented as `0`. A source file ending in `// comment` without a trailing newline causes the lexer to spin forever before parser error handling can run.

### Validation

A targeted harness using `lexer.New("// no newline")` and `parser.New(l).ParseProgram()` printed `hung` after 500ms. Artifact: `/tmp/codex-security-scans/grianlang3/897dffb_20260512T194526Z/artifacts/validation_artifacts_hang.go`.

### Reachability Analysis

Attacker-controlled `.gl3` source is the compiler's primary input. Impact is availability-only, but it can stall CI or any service that compiles untrusted GL3 code.

### Attack Path

1. Attacker supplies GL3 source ending in a line comment without newline.
2. `cli/build.go` reads the file and constructs the lexer/parser.
3. `lexer/lexer.go:72-75` enters the comment loop.
4. EOF never satisfies the loop condition, so compilation hangs.

### Severity Analysis

Medium for automated or hosted compilation contexts; low for trusted local-only use.

### Remediation

Change the loop to stop on either newline or EOF, and add lexer/parser regression tests for EOF after a line comment.

## Finding: Cyclic `#define` aliases crash `exdef` with stack exhaustion

- Priority: P3
- Severity: low
- Confidence: high
- CWE: CWE-674 Uncontrolled Recursion
- Affected lines: `/Users/grian/Projects/grianlang3/cli/exdef.go:93`, `/Users/grian/Projects/grianlang3/cli/exdef.go:111`

### Summary

`resolveDefine` recursively retries any macro whose inferred type remains `None`. Cyclic aliases such as `#define A B` and `#define B A` never converge and recurse until the Go runtime aborts with stack overflow.

### Validation

Validated through the real CLI using a two-line header. `go run . exdef -i /private/tmp/gl3scan/cycle.h -o /private/tmp/gl3scan/cycle.gl3` crashed with `runtime: goroutine stack exceeds ...`, with repeated frames at `cli/exdef.go:111`.

### Reachability Analysis

This affects the shipped `exdef` subcommand when processing attacker-controlled or third-party headers. It is availability-only and does not affect normal `gl3 build`.

### Attack Path

1. Attacker supplies a C header with cyclic macro aliases.
2. `exdef` asks `clang` to dump macros.
3. `resolveDefine` repeatedly substitutes aliases.
4. Type inference never succeeds, so recursion continues until stack exhaustion.

### Severity Analysis

Low: deterministic CLI crash with no demonstrated confidentiality or integrity impact.

### Remediation

Use an iterative resolver with a visited set and recursion depth limit; emit diagnostics and skip unresolved or cyclic macros.

## Finding: Empty `print` or `println` call crashes the checker

- Priority: P3
- Severity: low
- Confidence: high
- CWE: CWE-129 Improper Validation of Array Index
- Affected lines: `/Users/grian/Projects/grianlang3/checker/checker.go:193`, `/Users/grian/Projects/grianlang3/checker/checker.go:195`

### Summary

When the `io` module is imported, calls to `print` or `println` are routed into `checkPrintArgs`. That function immediately reads `node.Params[0]` without checking that at least one argument exists. A valid parsed call expression like `print()` therefore panics the compiler.

### Validation

Validated through the real CLI with a source file containing `import "io"` and `print()`. `go run . build /private/tmp/gl3scan/print_empty.gl3 --noexecbuild` failed with `panic: runtime error: index out of range [0] with length 0` at `checker.go:195`.

### Reachability Analysis

This is attacker-controlled through GL3 source, but availability-only and limited to compiler execution.

### Attack Path

1. Attacker supplies GL3 source importing `io`.
2. The source calls `print()` or `println()` with no arguments.
3. `checker.go:101` calls `checkPrintArgs`.
4. `checker.go:195` indexes `node.Params[0]` and panics.

### Severity Analysis

Low: deterministic compiler crash from a tiny source file, with no demonstrated confidentiality or integrity impact.

### Remediation

Check `len(node.Params) > 0` before reading the format argument and report a normal checker error when the format string is missing.

## Coverage Closure

- Suppressed command injection for `gl3 build` and `exdef`: subprocess calls use `exec.Command` without a shell.
- Suppressed GL3 import arbitrary-file-read as a top finding: `import "path.gl3"` is a documented file import feature and no direct disclosure sink was found.
- Deferred deeper exploitability review for `builtins/arrays.c`: memory-safety-sensitive code exists, but practical security impact depends on compiled-program ABI/use rather than the compiler CLI boundary.
- Noted but not promoted: `checker.go` appears to validate `%l` / `%ul` print specifiers against 32-bit integer types while `builtins/io.c` reads 64-bit varargs. A quick local run printed correctly on this platform, so I am leaving it as suspicious ABI/correctness risk rather than a confirmed security finding.
