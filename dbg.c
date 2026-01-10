//go:build ignore
#include <stdio.h>
#include <stdbool.h>

void dbg_i64(long long val) {
    printf("dbgi64: %lld\n", val);
}

void dbg_i32(int val) {
    printf("dbgi32: %d\n", val);
}

void dbg_bool(bool val) {
    printf("dbgbool: %s\n", val == 1 ? "true" : "false");
}

void dbg_float(float val) {
    printf("dbgfloat: %f\n", val);
}
