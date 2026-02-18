# grianlang3

GL3 is an LLVM-based compiled systems programming language. The compiler is written in Go and uses the llir/llvm library for code generation.

> **Note**: This compiler is still in development. In particular, users may encounter issues due to a fragile parser at time of writing.

## Philosophy

GL3 makes deliberate design choices that differ from many modern languages:

**No Type Inference** - All types must be explicitly specified. Variables require type annotations, and literals carry type suffixes. This eliminates ambiguity and makes the programmer's intentions explicit in the code.

**No Static Arrays** - Static arrays are not a language feature. Use dynamic arrays from the `arrays` standard library module instead. The `__asm__salloc` function in the `asm` module exists for extreme low-level scenarios (such as OS development before an allocator exists), but its use is discouraged outside of those specific cases.

## Building

The compiler produces LLVM IR which is then compiled to a native executable using `clang`. Clang is required for a full compilation pipeline.

### Command Line Flags

| Flag         | Description                                          |
|--------------|------------------------------------------------------|
| `--keepll`   | Preserve temporary `.ll` files for IR inspection     |
| `--dbg`      | Print AST output for each parsed file                |
| `--noexecbuild` | Generate IR only, do not invoke clang              |

## Example Program

```gl3
import "arrays"
import "strings"

struct Point {
    int32 x
    int32 y
}

fnc distance(Point* p) -> int32 {
    // approximate manhattan distance from origin
    def int32 dx = p.x
    def int32 dy = p.y
    if dx < 0i32 {
        dx = -dx
    }
    if dy < 0i32 {
        dy = -dy
    }
    return dx + dy
}

fnc main() -> int32 {
    def Point origin = Point{ 0i32, 0i32 }
    def Point p = Point{ 3i32, 4i32 }
    
    def int32 dist = distance(&p)
    
    // dynamic string example
    def char* greeting = dynstr("distance: ")
    def char* result = str_append(greeting, "7 units")
    
    return 0i32
}
```

## Documentation

- [SYNTAX.md](SYNTAX.md) - Complete language syntax reference
- [STDLIB.md](STDLIB.md) - Standard library documentation
