/*
 * =========================================
 * Example 4: Baboon Crossing Problem (Exercise 6.3)
 * Implementation B
 * =========================================
 * Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
 * Author: Spencer Rose
 *
 * SUMMARY: There is a deep canyon and a single rope that spans the canyon. Baboons
 * can cross the canyon by swinging on the rope, but if two baboons going in
 * opposite directions meet in the middle, they will fight and drop to their deaths.
 * Once a baboon has begun to cross, it is guaranteed to get to the other side
 * without running into a baboon going the other way. There are never more than 5
 * baboons on the rope.
 *
 * USAGE: ./main [Number of Threads]
 *
 * References
 * Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 101-111.
*/

#define _POSIX_C_SOURCE 199309L /* for clock_gettime */
#include <assert.h>
#include <unistd.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <semaphore.h>
#include "tracker.h"

typedef enum {EAST, WEST, NONE} dir;
static const char *arrows[] = {"<---", "--->", "None"};

/* Lighthouse Pattern: adapted from Downey, p. 70 */
typedef struct Lightswitch {
    int count;
    pthread_mutex_t mtx;
} lightswitch;

typedef struct Rope {
    int n;
    int dir;
    int west_waiting;
    int east_waiting;
    int crossing;
    int crossed;
    tracker * trk;
    pthread_mutex_t mtx;
    pthread_condattr_t cond_attr;
    pthread_cond_t update;
    lightswitch *ls_east;
    lightswitch *ls_west;
    sem_t * multiplex;
    sem_t * turnstile;
    sem_t * rope;
} rope;


/* Baboon Crossing interface */
rope *new_rope(int n);
void* run_baboon_east(void *arg);
void* run_baboon_west(void *arg);
void* run_rope(void *arg);
void clean_up(rope **);
sem_t* new_semaphore(char *, int);
lightswitch *init_lightswitch();
void sem_lock(lightswitch **, sem_t **);
void sem_unlock(lightswitch **, sem_t **);

/*
  -------------------------------------------
   error message
*/
void error(int e, char *msg) {
    printf("ERROR: %s has error code %d\n", msg, e);
}


/* Semaphore constructor */
sem_t *new_semaphore(char *name, int n) {
    sem_unlink(name);
    sem_t *s = sem_open(name, O_CREAT, 0644, n);
    if (s == SEM_FAILED)
        perror("%s sem initialization failed.");
    return s;
}

/*
  -------------------------------------------
  new rope (constructor)
*/
rope *new_rope(int n) {

    rope *r = (rope *) emalloc(sizeof(rope));

    int e;

    r->n = n;
    r->dir = NONE;
    r->crossing = 0;
    r->west_waiting = 0;
    r->east_waiting = 0;
    r->crossed = 0;
    r->trk = init_tracker(2);

    r->ls_east = init_lightswitch();
    r->ls_west = init_lightswitch();

    r->multiplex = new_semaphore("multiplex", 5);
    r->turnstile = new_semaphore("turnstile", 1);
    r->rope = new_semaphore("rope", 1);

    /* initialize mutex and conditional variables */
    if ((e = pthread_mutex_init(&r->mtx, NULL))) {
        error(e, "rope_mtx");
    }
    if ((e = pthread_condattr_init(&r->cond_attr))) {
        error(e, "cond_attr_rope");
    }
    if ((e = pthread_cond_init(&r->update, &r->cond_attr))) {
        error(e, "cond_update");
    }
    return r;
}

/*
  -------------------------------------------
   close semaphores
*/
void clean_up(rope **r) {
    sem_close((*r)->turnstile);
    sem_close((*r)->rope);
    sem_close((*r)->multiplex);
    free(*r);
}

/*
===========================================
Main Routine
===========================================
*/
int main(int argc, char *argv[]) {

    int n = atoi(argv[1]);
    printf("Creating %d baboon threads...\n\n", n);

    int i = 0;
    int e;
    pthread_t tid[n + 1]; /* thread-ID array */

    /* initialize thread attribute object */
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    rope *r = new_rope(n);

    /* display scoreboard */
    /*
    if ((e = pthread_create(&tid[0], &attr, run_rope, (void *) &r))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    } */

    for (i = 1; i <= n; i++) {

        /* create eastward baboon thread */
        if ((e = pthread_create(&tid[i], &attr, run_baboon_east, &r))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }

        /* create westward baboon thread */
        if ((e = pthread_create(&tid[i], &attr, run_baboon_west, &r))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
    }

    /* Join exited threads */
    for (i = 1; i <= n; i++) {
        if(pthread_join(tid[i], NULL)) {
            fprintf(stderr, "Error joining thread %d\n", i);
        }
    }
    print_tracker(r->trk);
    clean_up(&r);
    pthread_exit(NULL);

}


/*
===========================================
// Rope monitors requests to cross
===========================================
*/
void* run_rope(void *arg) {

    rope **r = (rope **) arg;

    pthread_mutex_lock(&(*r)->mtx);

    while ((*r)->crossed < (*r)->n) {
        printf("Crossed: %-5d | Waiting [E]: %-5d | %-5s %-5d | Waiting [W]: %-5d\n",
               (*r)->crossed, (*r)->east_waiting, arrows[(*r)->dir], (*r)->crossing, (*r)->west_waiting);
        pthread_cond_wait(&(*r)->update, &(*r)->mtx);

    }
    pthread_mutex_unlock(&(*r)->mtx);

    pthread_exit(NULL);

}

/*
===========================================
// Baboons crossing rope eastward
===========================================
*/
void* run_baboon_east(void *arg) {

    rope **r = (rope **)arg;
    struct timespec start;

    pthread_mutex_lock(&(*r)->mtx);
    (*r)->east_waiting++;
    pthread_mutex_unlock(&(*r)->mtx);

    timer_start(&start);
    sem_wait((*r)->turnstile);
    sem_lock(&(*r)->ls_east, &(*r)->rope);
    sem_post((*r)->turnstile);
    sem_wait((*r)->multiplex);
    timer_stop(0, start, &(*r)->trk);

    pthread_mutex_lock(&(*r)->mtx);
    (*r)->east_waiting--;
    (*r)->crossing++;
    (*r)->dir = EAST;
    /* pthread_cond_signal(&(*r)->update); */
    pthread_mutex_unlock(&(*r)->mtx);

    sem_post((*r)->multiplex);

    sem_unlock(&(*r)->ls_east, &(*r)->rope);

    pthread_mutex_lock(&(*r)->mtx);
    (*r)->crossing--;
    (*r)->crossed++;
    pthread_mutex_unlock(&(*r)->mtx);

    pthread_exit(NULL);
}

/*
===========================================
// Baboons crossing rope westward
===========================================
*/
void* run_baboon_west(void *arg) {

    rope **r = (rope **)arg;
    struct timespec start;

    pthread_mutex_lock(&(*r)->mtx);
    (*r)->west_waiting++;
    pthread_mutex_unlock(&(*r)->mtx);

    timer_start(&start);
    sem_wait((*r)->turnstile);
    sem_lock(&(*r)->ls_west, &(*r)->rope);
    sem_post((*r)->turnstile);
    sem_wait((*r)->multiplex);
    timer_stop(1, start, &(*r)->trk);

    pthread_mutex_lock(&(*r)->mtx);
    (*r)->west_waiting--;
    (*r)->crossing++;
    (*r)->dir = WEST;
    /* pthread_cond_signal(&(*r)->update); */
    pthread_mutex_unlock(&(*r)->mtx);

    sem_post((*r)->multiplex);

    pthread_mutex_lock(&(*r)->mtx);
    (*r)->crossing--;
    (*r)->crossed++;
    pthread_mutex_unlock(&(*r)->mtx);

    sem_unlock(&(*r)->ls_west, &(*r)->rope);

    pthread_exit(NULL);
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
        perror("rope_mtx");
    return l;
}

void sem_lock(lightswitch **l, sem_t **control) {
    pthread_mutex_lock(&(*l)->mtx);
    (*l)->count++;
    if ((*l)->count == 1)
        sem_wait(*control);
    pthread_mutex_unlock(&(*l)->mtx);
}

void sem_unlock(lightswitch **l, sem_t **control) {
    pthread_mutex_lock(&(*l)->mtx);
    (*l)->count--;
    if ((*l)->count == 0)
        sem_post(*control);
    pthread_mutex_unlock(&(*l)->mtx);
}