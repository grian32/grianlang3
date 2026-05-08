#!/usr/bin/env bash

cp ../../gl3 .

./gl3 build --shared -o libadd.so add.gl3

javac Main.java
jar cfe app.jar Main Main.class
java -Djava.library.path=. -jar app.jar

rm -f Main.class app.jar gl3 libadd.so
