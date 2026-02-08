//go:build ignore
#include <stdio.h>
#include <stdbool.h>
#include <stdint.h>
#include <inttypes.h>

void dbg_i64(int64_t val) {
    printf("dbgi64: %"PRId64"\n", val);
}

void dbg_i32(int32_t val) {
    printf("dbgi32: %"PRId32"\n", val);
}

void dbg_i16(int16_t val) {
    printf("dbgi16: %"PRId16"\n", val);
}

void dbg_i8(int8_t val) {
    printf("dbgi8: %"PRId8"\n", val);
}

void dbg_u64(uint64_t val) {
    printf("dbgu64: %"PRIu64"\n", val);
}

void dbg_u32(uint32_t val) {
    printf("dbgu32: %"PRIu32"\n", val);
}

void dbg_u16(uint16_t val) {
    printf("dbgu16: %"PRIu16"\n", val);
}

void dbg_u8(uint8_t val) {
    printf("dbgu8: %"PRIu8"\n", val);
}

void dbg_bool(bool val) {
    printf("dbgbool: %s\n", val == 1 ? "true" : "false");
}

void dbg_float(float val) {
    printf("dbgfloat: %.50f\n", val);
}

void dbg_str(const char* val) {
    printf("dbgstr: %s\n", val);
}

void dbg_char(char val) {
    printf("dbgchar: %c(%d)\n", val, val);
}
