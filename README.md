# grianlang3

GL3 is an LLVM-based compiled systems programming language. The compiler is written in Go and uses the llir/llvm library for code generation.

## Philosophy

GL3 makes deliberate design choices that differ from many modern languages:

**No Type Inference** - All types must be explicitly specified. Variables require type annotations, and literals carry type suffixes. This eliminates ambiguity and makes the programmer's intentions explicit in the code.

**No Static Arrays** - Static arrays complicate the type system and offer little practical benefit when heap memory is readily available. Use dynamic arrays from the `arrays` standard library module instead. The `__asm__salloc` function in the `asm` module exists for extreme low-level scenarios (such as OS development before an allocator exists), but its use is discouraged outside of those specific cases. The reasoning behind this is that in practice, while

## Building

The compiler produces LLVM IR which is then compiled to a native executable using `clang`. Clang is required for a full compilation pipeline.

### Command Line Flags

| Flag            | Description                                                                                  |
| --------------- | -------------------------------------------------------------------------------------------- |
| `--keepll`      | Preserve temporary `.ll` files for IR inspection                                             |
| `--dbg`         | Print AST output for each parsed file, along with the executed clang command for compilation |
| `--noexecbuild` | Generate IR only, do not invoke clang                                                        |

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
        cur = *cur.next
    }

    str_free(*a.name)
    str_free(*b.name)
    free(a)
    free(b)
    return 0i32
}
```

## Documentation

- [SYNTAX.md](SYNTAX.md) - Complete language syntax reference
- [STDLIB.md](STDLIB.md) - Standard library documentation
