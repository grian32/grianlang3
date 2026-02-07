//go:build ignore
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

char* dynstr(const char* in) {
    size_t len = strlen(in);
    char* new = malloc((len+1) * sizeof(char)); // size of technically not necessary but just doing it for clarity sake
    memcpy(new, in, len+1);
    return new;
}

char* str_append(char* a, char* b) {
    size_t a_len = strlen(a);
    size_t b_len = strlen(b);
    size_t new_len = a_len + b_len + 1;
    char* new = malloc(new_len * sizeof(char));
    memcpy(new, a, a_len);
    memcpy(new + a_len, b, b_len + 1);
    return new;
}

uint64_t str_len(char* a) {
    return (uint64_t)strlen(a);
}