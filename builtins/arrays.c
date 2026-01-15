//go:build ignore
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void* arr_new(size_t elem_size) {
    size_t cap = 16;
    void* meta = malloc((sizeof(size_t) * 3) + (elem_size * cap));
    ((size_t*)meta)[0] = cap;
    ((size_t*)meta)[1] = elem_size;
    ((size_t*)meta)[2] = 0;

    return ((uint8_t*)meta) + (sizeof(size_t) * 3);
}

void arr_push(void** arr, void* elem) {
    size_t* meta = (size_t*)(((uint8_t*)(*arr)) - (sizeof(size_t) * 3));
    size_t cap = meta[0];
    size_t elem_size = meta[1];
    size_t len = meta[2];

    if (len >= cap) {
        cap *= 2;
        void* new_meta =  realloc(meta, (sizeof(size_t) * 3) + (elem_size * cap));
        if (!new_meta) {
            printf("arr_push: failed to realloc\n");
        }
        ((size_t*)new_meta)[0] = cap;
        meta = new_meta;
        *arr = ((uint8_t*)meta) + (sizeof(size_t) * 3);
    }

    uint8_t* dest = (((uint8_t*)(*arr)) + (len++ * elem_size));
    memcpy((void*)dest, &elem, elem_size);
    meta[2] = len;
}

void arr_free(void** arr) {
    uint8_t* meta = (((uint8_t*)(*arr)) - (sizeof(size_t) * 3));
    free(meta);
}
