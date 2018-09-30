#
# "Makefile" for the SSI: Simple Shell Interpreter
# (UVic CSC 360, Spring 2018, assignment #1)
# Executable file: mts
#

CC=gcc
CFLAGS1=-c -Wall -ansi -g

.PHONY all: main

main: main.o list.o
	$(CC) -pthread main.o list.o -o main

main.o: main.c list.h
	$(CC) $(CFLAGS1) main.c

list.o: list.c list.h
	$(CC) $(CFLAGS1) list.c

.PHONY clean:

clean:
	rm -rf *.o *.exe