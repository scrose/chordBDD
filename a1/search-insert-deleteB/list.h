/*
===========================================
Trains (data structure) ~ header
===========================================
Provides queue of trains
*/

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pthread.h>
#include <string.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <signal.h>
#include <time.h>

#ifndef _LIST_H_
#define _LIST_H_


typedef struct element element;
struct element {
    int value; /* integer value of element */
    element *next;
};

typedef struct linkedlist linkedlist;
struct linkedlist {
    element *head; /* number of deleters */
    element *tail; /* number of searchers */

    /* linkedlist mutex & condition variables */

    int d; /* number of deleters */
    int s; /* number of searchers */
    int i; /* number of inserters */

    /* turnstile mutex & condition variables */
    pthread_mutex_t ts_mtx;
    pthread_condattr_t ts_cond_attr;
    pthread_cond_t ts_cond;

    pthread_mutex_t searcher_mtx;
    pthread_condattr_t searcher_cond_attr;
    pthread_cond_t searcher_cond;

    pthread_mutex_t inserter_mtx;
    pthread_condattr_t inserter_cond_attr;
    pthread_cond_t inserter_cond;

    pthread_mutex_t deleter_mtx;
    pthread_condattr_t deleter_cond_attr;
    pthread_cond_t deleter_cond;
};

/* Searcher-Inserter-Deleter interface */
element *new_element(int);
linkedlist *new_linkedlist();
void search(linkedlist **, int);
void add(linkedlist **, int);
void delete(linkedlist **, int);
void display(linkedlist **);
void *emalloc(size_t);
void error(int, char *);

#endif
