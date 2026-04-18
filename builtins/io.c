#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>

void print_int(int64_t val) {
    if (val < 0) {
        putchar('-');
        val = -val;
    }
    if (val == 0) {
        putchar('0');
        return;
    }

    // 10 max digits i suppsoe but better safe than sorry!
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

/*
 * Prints a formatted string to stdout.
 * Format specifiers: %b (bool), %d (int32), %y (int8), %w (int16), %l(int64)
 * Prefix with: %f to include format specifiers at the end (%fd = 7i32, %d = 7)
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

        switch (*fmt++) {
            case 'b': {
                int val = va_arg(args, int);
                if (val) {
                    fputs("true", stdout);
                } else {
                    fputs("false", stdout);
                }
                break;
            }
            case 'y':
            case 'w':
            case 'd': {
                print_int(va_arg(args, int));
                break;
            }
            case 'l': {
                print_int(va_arg(args, int64_t));
                break;
            }
            case 'f': {
                switch (*fmt++) {
                    case 'd': {
                        print_int(va_arg(args, int));
                        fputs("i32", stdout);
                        break;
                    }
                    case 'y': {
                        print_int(va_arg(args, int));
                        fputs("i8", stdout);
                        break;
                    }
                    case 'w': {
                        print_int(va_arg(args, int));
                        fputs("i16", stdout);
                        break;
                    }
                    case 'l': {
                        print_int(va_arg(args, uint64_t));
                        fputs("i64", stdout);
                        break;
                    }
                }
                break;
            }
            case '%': {
                putchar('%');
                break;
            }
        }
    }

    va_end(args);
}
