#
# "Makefile" for output
# Executable file: main
#

CC=gcc
CFLAGS1=-Wall -ansi -g -rdynamic

.PHONY all: main

main: main.c
	$(CC) $(CFLAGS1) -pthread main.c -o main -lrt

.PHONY clean:
clean:
	rm -rf *.o *.exe *.out