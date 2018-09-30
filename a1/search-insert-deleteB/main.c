/*
   =========================================
   Example 2: The Search-Insert-Delete Problem (Exercise 6.1)
   =========================================
   Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
   Spencer Rose (ID V00124060)
*/

#include <assert.h>
#include <unistd.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "list.h"


/* Pthread routines */
void* searcher(void *arg);
void* inserter(void *arg);
void* deleter(void *arg);


/*
===========================================
Main Routine
===========================================
*/
int main(int argc, char *argv[]) {

    /* number of thread triplets */
    int n = 31;
    int i, j, e;
    pthread_t tid[n*3]; /* thread-ID array */

    /* initialize thread attribute object */
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    /* create new linkedlist */
    linkedlist *l = new_linkedlist();

    /* launch searcher, deleter, inserter threads */
    for (i = 0, j = 0; i < n; i++) {


        /* create searcher thread */
        if ((e = pthread_create(&tid[j++], &attr, searcher, (void *) &l))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }

        /* create inserter thread */
        if ((e = pthread_create(&tid[j++], &attr, inserter, (void *) &l))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }

        /* create deleter thread */
        if ((e = pthread_create(&tid[j++], &attr, deleter, (void *) &l))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }

    }

    /* Join exited threads */
    for (i = 0; i < j-1; i++) {
        if(pthread_join(tid[i], NULL)) {
            fprintf(stderr, "Error joining thread\n");
            return 2;
        }
    }

    /* exit the thread */
    free(l);
    pthread_exit(NULL);

}


/*
===========================================
// Searchers merely examine the list; hence can execute concurrently
===========================================
*/
void* searcher(void *arg) {
    srand(time(0));
    rand();
    int r = rand() % 10;

    linkedlist **l = (linkedlist **)arg;

    pthread_mutex_lock(&(*l)->ts_mtx);
    while ((**l).d > 0) {
        pthread_cond_wait(&(*l)->ts_cond, &(*l)->ts_mtx);
    }
    pthread_cond_signal(&(*l)->ts_cond);
    pthread_mutex_unlock(&(*l)->ts_mtx);

    /* increment searcher count */
    pthread_mutex_lock(&(*l)->searcher_mtx);
    (**l).s++;

    printf("# Searchers: %d\n", (**l).s);
    pthread_mutex_unlock(&(*l)->searcher_mtx);


    printf("Search started.\n");
    search(l, r);
    printf("Search finished.\n");

    /* decrement searcher count */
    pthread_mutex_lock(&(*l)->searcher_mtx);
    (**l).s--;
    printf("# Searchers: %d\n", (**l).s);

    pthread_mutex_unlock(&(*l)->searcher_mtx);

    /* exit the thread */
    pthread_exit(NULL);
}

/*
===========================================
// Inserters adds new item to the end of the list; insertions must be mutually
// exclusive to preclude two inserters from inserting new items at about the
// same time. However, one insert can proceed in parallel with any number of
// searches.
===========================================
*/
void* inserter(void *arg) {
    srand(time(0));
    rand();
    int r = rand() % 10;

    linkedlist **l = (linkedlist **)arg;

    pthread_mutex_lock(&(*l)->inserter_mtx);
    printf("Insert element of value %d.\n", r);
    add(l, r);
    printf("Inserted element of value %d.\n", r);
    pthread_mutex_unlock(&(*l)->inserter_mtx);

    /* exit the thread */
    pthread_exit(NULL);

}

/*
===========================================
deleters remove items from anywhere in the list. At most one deleter process
can access the list at a time, and deletion must also be mutually exclusive
with searches and insertions.
===========================================
*/
void* deleter(void *arg) {
    srand(time(0));
    rand();
    int r = rand() % 10;

    linkedlist **l = (linkedlist **)arg;

    /* Lock out searchers */
    pthread_mutex_lock(&(*l)->ts_mtx);

    /* increment deleter count */
    (**l).d++;
    printf("# Deleters: %d\n", (**l).d);

    /* Wait for no searchers */
    pthread_mutex_lock(&(*l)->searcher_mtx);
    while((**l).s > 0) {
        pthread_cond_wait(&(*l)->searcher_cond, &(*l)->searcher_mtx);
    }

    /* Lock out inserters */
    pthread_mutex_lock(&(*l)->inserter_mtx);
    printf("Delete element of value: %d.\n", r);
    delete(l, r);
    printf("Deleted element of value: %d.\n", r);
    pthread_mutex_unlock(&(*l)->inserter_mtx);

    /* decrement deleter count */
    (**l).d--;
    printf("# Deleters: %d\n", (**l).d);
    pthread_mutex_unlock(&(*l)->ts_mtx);

    /* signal no writers */
    pthread_cond_signal(&(*l)->searcher_cond);

    /* exit the thread */
    pthread_exit(NULL);
}
