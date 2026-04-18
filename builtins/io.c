#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdbool.h>


typedef struct {
    bool ok;
    bool show_suffix;
    bool is_unsigned;
    char width; // y, w, d, l; 8, 16, 32, 64
} IntegerSpecifier;

IntegerSpecifier parse_int_specifier(const char** fmt) {
    IntegerSpecifier spec = { .ok = true, .show_suffix = false, .is_unsigned = false };
    if (**fmt == 'f') {
        spec.show_suffix = true;
        (*fmt)++;
    }

    if (**fmt == 'u') {
        spec.is_unsigned = true;
        (*fmt)++;
    }

    char c = **fmt;
    if (c == 'y' || c == 'w' || c == 'd' || c == 'l') {
        spec.width = c;
        (*fmt)++;
    } else {
        spec.ok = false;
    }

    return spec;
}

void print_uint(uint64_t val) {
    if (val == 0) {
        putchar('0');
        return;
    }

    // 19 max digits i suppsoe but better safe than sorry!
    char buf[40];
    int i = 0;
    while (val) {
        buf[i++] = '0' + (val % 10);
        val /= 10;
    }
    while (i--) {
        putchar(buf[i]);
    }
}

void print_int(int64_t val) {
    if (val < 0) {
        putchar('-');
        val = -val;
    }
    print_uint((uint64_t)val);
}

const char* format_suffix(char width, bool is_unsigned) {
    if (is_unsigned) {
        switch (width) {
            case 'y': return "u8";
            case 'w': return "u16";
            case 'd': return "u32";
            case 'l': return "u64";
            default: return "";
        }
    }
    switch (width) {
        case 'y': return "i8";
        case 'w': return "i16";
        case 'd': return "i32";
        case 'l': return "i64";
        default: return "";
    }

    return "";
}

/*
 * Prints a formatted string to stdout.
 * Format specifiers: %b (bool), %d (int32), %y (int8), %w (int16), %l(int64), %s (string), %c (char)
 * Prefix with %f to include format specifiers at the end (%fd = 7i32, %d = 7)
 * Prefix with %u for unsigned integers
 *
 * Examples: %fud = format specifier at the end, unsigned, 32-bit integer
 * Note: The order of the specifiers must be 'f', and then 'u', and then the width specifier
 */
void print(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);

    while (*fmt) {
        if (*fmt != '%') {
            putchar(*fmt++);
            continue;
        }
        fmt++;

        if (*fmt == '%') {
            putchar('%');
            fmt++;
            continue;
        }

        if (*fmt == 'b') {
            int val = va_arg(args, int);
            fputs(val ? "true" : "false", stdout);
            fmt++;
            continue;
        }

        if (*fmt == 'c') {
            int val = va_arg(args, int);
            putchar(val);
            fmt++;
            continue;
        }

        if (*fmt == 's') {
            char* val = va_arg(args, char*);
            fputs(val, stdout);
            fmt++;
            continue;
        }

        IntegerSpecifier spec = parse_int_specifier(&fmt);
        if (!spec.ok) {
            putchar('%');
            fmt++;
            continue;
        }

        if (spec.is_unsigned) {
            switch (spec.width) {
                case 'y': {
                    print_uint((uint8_t)va_arg(args, unsigned int));
                    break;
                }
                case 'w': {
                    print_uint((uint16_t)va_arg(args, unsigned int));
                    break;
                }
                case 'd': {
                    print_uint(va_arg(args, unsigned int));
                    break;
                }
                case 'l': {
                    print_uint(va_arg(args, uint64_t));
                    break;
                }
            }
        } else {
            switch (spec.width) {
                case 'y': {
                    print_int((int8_t)va_arg(args, int));
                    break;
                }
                case 'w': {
                    print_int((int16_t)va_arg(args, int));
                    break;
                }
                case 'd': {
                    print_int(va_arg(args, int));
                    break;
                }
                case 'l': {
                    print_int(va_arg(args, int64_t));
                    break;
                }
            }
        }
        if (spec.show_suffix) {
            fputs(format_suffix(spec.width, spec.is_unsigned), stdout);
        }
    }

    putchar('\n');

    va_end(args);
}
