/*
   =========================================
   Example 2: The Search-Insert-Delete Problem (Exercise 6.1)
   =========================================
   Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
   Spencer Rose (ID V00124060)
*/
#define _POSIX_C_SOURCE 199309L /* for clock_gettime */
#include <assert.h>
#include <unistd.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <semaphore.h>
#include "tracker.h"

/* Lighthouse Pattern: adapted from Downey, p. 70 */
typedef struct Lightswitch {
    int count;
    pthread_mutex_t mtx;
} lightswitch;

typedef struct element element;
struct element {
    int value; /* integer value of element */
    element *next;
};

typedef struct linkedlist linkedlist;
struct linkedlist {
    element *head; /* number of deleters */
    element *tail; /* number of searchers */

    int s, i, d;

    lightswitch * ls_searchers;
    sem_t * empty;
    tracker * trk;
    pthread_mutex_t inserter_mtx;
};

/* search-insert-delete interface */
void* searcher(void *arg);
void* inserter(void *arg);
void* deleter(void *arg);
lightswitch *init_lightswitch();
element *new_element(int);
linkedlist *init_linkedlist();
void sem_lock(lightswitch **, sem_t **);
void sem_unlock(lightswitch **, sem_t **);
void search(linkedlist **, int);
void add(linkedlist **, int);
void delete(linkedlist **, int);
void display(linkedlist **);
void *emalloc(size_t);
void error(int, char *);
void clean_up(linkedlist **);

/* linkelist constructor */
linkedlist *init_linkedlist() {
    int e;
    linkedlist *l = (linkedlist *) emalloc(sizeof(linkedlist));
    l->head = NULL;
    l->tail = NULL;

    l->s = l->i = l->d = 0;
    l->ls_searchers = init_lightswitch();

    l->trk = init_tracker(3);

    sem_unlink("empty");
    l->empty = sem_open("empty", O_CREAT, 0644, 1);
    if (l->empty == SEM_FAILED) {
        perror("empty sem initialization failed.");
    }
    if ((e = pthread_mutex_init(&l->inserter_mtx, NULL))) {
        error(e, "mtx_list");
    }
    return l;
}

/*
  -------------------------------------------
  new element (constructor)
*/
element *new_element(int val) {

    element *e = (element *) emalloc(sizeof(element));
    e->value = val;
    e->next = NULL; /* head of list */
    return e;
}

/*
  -------------------------------------------
   error message
*/
void error(int e, char *msg) {
    printf("ERROR: %s has error code %d\n", msg, e);
}

/*
  -------------------------------------------
   close down
*/
void clean_up(linkedlist **l) {
    sem_close((*l)->empty);
    free(*l);
}

/*
===========================================
Main Routine
===========================================
*/
int main(int argc, char *argv[]) {

    int n = atoi(argv[1]);
    printf("\n\nCreating %d searchers, %d inserters, %d deleters...\n\n", n, n, n);

    srand(time(0));

    int i, j, e;
    pthread_t tid[n*3];
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    linkedlist *l = init_linkedlist();

    /* launch searcher, deleter, inserter threads */
    for (i = 0, j = 0; i < n; i++) {

        if ((e = pthread_create(&tid[j++], &attr, searcher, (void *) &l))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
        if ((e = pthread_create(&tid[j++], &attr, inserter, (void *) &l))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
        if ((e = pthread_create(&tid[j++], &attr, deleter, (void *) &l))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
    }

    /* Join exited threads */
    for (i = 0; i < 3*n; i++) {
        if(pthread_join(tid[i], NULL)) {
            fprintf(stderr, "Error joining thread\n");
            return(2);
        }
    }
    print_tracker(l->trk);
    clean_up(&l);
    exit(0);
}

/*
===========================================
// Searchers merely examine the list; hence can execute concurrently
===========================================
*/
void* searcher(void *arg) {

    rand();
    int r = rand() % 10;
    struct timespec start;

    linkedlist **l = (linkedlist **)arg;

    timer_start(&start);
    sem_lock(&(*l)->ls_searchers, &(*l)->empty);
    timer_stop(0, start, &(*l)->trk);

    (*l)->s++;
    search(l, r);
    (*l)->s--;
    sem_unlock(&(*l)->ls_searchers, &(*l)->empty);

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

    rand();
    int r = rand() % 10;
    /* struct timespec start; */

    linkedlist **l = (linkedlist **)arg;

    pthread_mutex_lock(&(*l)->inserter_mtx);
    (*l)->i++;
    add(l, r);
    (*l)->i--;
    pthread_mutex_unlock(&(*l)->inserter_mtx);

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

    rand();
    int r = rand() % 10;
    struct timespec start;

    linkedlist **l = (linkedlist **)arg;

    timer_start(&start);
    sem_wait((*l)->empty);
    pthread_mutex_lock(&(*l)->inserter_mtx);
    timer_stop(2, start, &(*l)->trk);

    (*l)->d++;
    delete(l, r);
    (*l)->d--;
    pthread_mutex_unlock(&(*l)->inserter_mtx);
    sem_post((*l)->empty);

    pthread_exit(NULL);
}



/*
  -------------------------------------------
  add element to linkedlist
*/
void add(linkedlist **l, int val) {

    if ((*l)->head == NULL) {
        (*l)->head = new_element(val);
    }
    else {
        element *curr = (*l)->head;
        for (; curr->next != NULL; curr = curr->next);
        curr->next = new_element(val);
    }
}

/*
  -------------------------------------------
  search for value in linkedlist
*/
void search(linkedlist **l, int val) {

    if ((*l)->head == NULL) { /* empty list */
        return;
    }
    else {
        element *curr = (*l)->head;
        for (; curr->next != NULL; curr = curr->next){
            if (curr->value == val) {
                return;
            }
        }
    }
}


/*
  -------------------------------------------
  remove element from linkedlist
*/
void delete(linkedlist **l, int val) {

element *curr = (*l)->head;

    if (curr == NULL) {
        return;
    } else if (curr->value == val) {
        (*l)->head = curr->next;
        return;
    } else {
        element *prev = (*l)->head;
        for (; curr->next != NULL; curr = curr->next, prev = curr) {
            if (curr->value == val) {
                prev->next = curr->next;
                if (curr->next == NULL) {
                    (*l)->tail = prev;
                }
                return;
            }
        }
    }
}

/*
  -------------------------------------------
  display linkedlist
*/
void display(linkedlist **l) {

    element *curr = (*l)->head;
    if (curr == NULL) { /* empty list */
        return;
    } else {
        int i = 0;
        printf("\nLINKEDLIST\n");
        for (; curr->next != NULL; curr = curr->next){
            printf("-> Element %d: Value: %d\n", i++, curr->value);
        }
    }
}

/*
 * =============================================
 * Lighthouse Pattern: adapted from Downey p.70)
 * =============================================
 * */
lightswitch *init_lightswitch() {
    lightswitch *l = (lightswitch *) emalloc(sizeof(lightswitch));
    l->count = 0;
    if (pthread_mutex_init(&l->mtx, NULL))
        perror("empty mtx");
    return l;
}

void sem_lock(lightswitch **l, sem_t **control) {
    pthread_mutex_lock(&(*l)->mtx);
    (*l)->count++;
    if ((*l)->count == 1) {
        sem_wait(*control);
    }

    pthread_mutex_unlock(&(*l)->mtx);
}

void sem_unlock(lightswitch **l, sem_t **control) {
    pthread_mutex_lock(&(*l)->mtx);
    (*l)->count--;
    if ((*l)->count == 0)
        sem_post(*control);
    pthread_mutex_unlock(&(*l)->mtx);
}