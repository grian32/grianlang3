all: stdlib
	go build -o gl3 .

clean:
	rm -f builtins/*.ll
	rm -f grianlang3

builtins/%.ll: builtins/%.c
	clang -S -O3 -emit-llvm $< -o $@

stdlib: $(patsubst builtins/%.c, builtins/%.ll, $(wildcard builtins/*.c))
