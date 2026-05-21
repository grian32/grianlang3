all:
	go build -o gl3 .

clean:
	rm -f builtins/*.ll
	rm -f gl3
