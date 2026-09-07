# GL3

GL3 is an LLVM-based compiler for a systems programming language. The compiler is written in Go and uses the llir/llvm library for code generation.

## Philosophy

GL3 makes deliberate design choices that differ from many modern languages:

**No Type Inference** - All types must be explicitly specified. Variables require type annotations, and literals carry type suffixes. This eliminates ambiguity and makes the programmer's intentions explicit in the code.

**No Static Arrays** - Static arrays complicate the type system and offer little practical benefit when heap memory is readily available. Use dynamic arrays from the `arrays` standard library module instead. The `__asm__salloc` function in the `asm` module exists for extreme low-level scenarios (such as OS development before an allocator exists), but its use is discouraged outside of those specific cases.

## Building

The compiler produces LLVM IR which is then compiled to a native executable using `clang`. Clang is required for a full compilation pipeline.

### Requirements

- Go >=1.25.5
- Clang (tested with version 22.1.1)

### Building the compiler

```bash
make
```

This compiles the stdlib builtin modules to LLVM IR and builds the `gl3` compiler binary. You can also cleanup any build files using `make clean`

## Testing

Run the full suite with `go test ./...`, or only the compiler end-to-end tests
with `go test ./e2e -count=1`. The e2e suite requires Go, Make, and Clang;
setup regenerates stale builtin LLVM files and builds a temporary compiler.
Use `go test -short ./...` to skip compiler builds and e2e execution.

Fixtures live in `e2e/testdata` and are registered in `e2e/e2e_test.go`.
Each fixture builds `main.gl3` (or its registered source list) in an isolated
temporary directory. Missing `stdout.txt` and `stderr.txt` mean empty output;
the default expected program exit is zero. `compile_error.txt` requires a
failed build, its diagnostic substring, and no output executable.
`checker_error.txt` requires a fatal `checker error:` diagnostic containing
the expected substring, a failed build, and no output executable or emitter
error. These fixtures assert the intended checker behavior ahead of its rewrite.
Selected fixtures also run with `--O3`, using the same expected results as their default build.
These variants avoid builtin imports because builtin processing currently
mutates the build optimization options.

Regression fixtures assert intended results even when the compiler currently
fails them; they are not skipped or changed to accept incorrect behavior.
Boolean short-circuit behavior is not tested until its language contract is defined.

## Example Program

```gl3
import "io"
import "ralloc"
import "strings"

struct Node {
    char* name
    Node* next
}

fnc main() -> int32 {
    def Node* a = malloc(sizeof Node) as Node*
    def Node* b = malloc(sizeof Node) as Node*

    a.name = dynstr("alice")
    a.next = b
    b.name = dynstr("bob")
    b.next = (0 as Node*)

    def Node* cur = a
    while (cur as int) != 0 {
        println("hello, %s", *cur.name)
        cur = cur.next
    }

    str_free(a.name)
    str_free(b.name)
    free(a)
    free(b)
    return 0i32
}
```

### Compile & run the example

```bash
./gl3 build example.gl3 -o output
./output
```

## Documentation

- [SYNTAX.md](SYNTAX.md) - Complete language syntax reference
- [STDLIB.md](STDLIB.md) - Standard library documentation
- [CLI.md](CLI.md) - Command line interface reference
