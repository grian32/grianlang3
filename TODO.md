# TODO

Future wants:
- Switch Statement
- Global Variable Definitions - .data for non const, .rodata for const
- LSP & Tree-Sitter - error reporting is there, honestly.
- "io" std module
    - "print", reimplementation with specifiers for my base data types, structs not allowed, add a `l` prefix to specifier to print along with literal prefix like `%lu` on `4u64` == '4u64' printed, while `%u` on `4u64` == '4', only accepts a constant .rodata string as the first argument, all datatypes checked at compiletime before emitting, use fputs on stdout for actual output to be platform agnostic since libc dependency already exists for other factors
    - "println", "print" + "\n"
    - does not require var args to be implemented as a language feature
- "arenas" std module
- "calloc" std module - links malloc, free, from c, read as "c's allocation"
- Extend checker to handle more cases for better error msgs, basic type checking, etc.