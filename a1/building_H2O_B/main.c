/*
 * =========================================
 * Example 3: Building H2O (Exercise 5.6)
 * Implementation B (Adapted from Downey, Allen B., The Little Book of Semaphores)
 * =========================================
 * Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
 * Spencer Rose (ID V00124060)
 * =========================================
 * SUMMARY: There are two kinds of threads, oxygen and hydrogen. In order to
 * assemble these threads into water molecules, we have to create a barrier
 * that makes each thread wait until a complete molecule is ready to proceed.
 * As each thread passes the barrier, it should invoke bond. You must guarantee
 * that all the threads from one molecule invoke bond before any of the
 * threads from the next molecule do.
 *
 * References:
 * 1. Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 143-148.
 * 2. https://stackoverflow.com/questions/47522174/reusable-barrier-implementation-using-posix-semaphores
 */

#define _GNU_SOURCE
#include <assert.h>
#include <unistd.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <semaphore.h>
#include "tracker.h"

#define THREAD_COUNT 3

typedef struct scoreboard scoreboard;
struct scoreboard {
    int hydrogen;
    int oxygen;
    sem_t * hydro_sem;
    sem_t * oxy_sem;
    pthread_barrier_t barrier;
    pthread_mutex_t mtx;
    tracker * trk;
};

/*
  -------------------------------------------
   error message
*/
void error(int e, char *msg) {
    printf("ERROR: %s has error code %d\n", msg, e);
}

/*
  -------------------------------------------
  Represents an H2O bond action
*/
void bond(char * atom) {
    printf("%s bonded!\n", atom);
}
/*
 -------------------------------------------
  scoreboard tracks the H2O bonds and counts
*/
scoreboard *init_scoreboard(int n_oxygen, int n_hydrogen) {

    int e;

    scoreboard *s = (scoreboard *) emalloc(sizeof(scoreboard));

    s->oxygen = 0;
    s->hydrogen = 0;
    s->trk = init_tracker(1);

    e = pthread_barrier_init(&s->barrier, NULL, THREAD_COUNT);

    s->hydro_sem = sem_open("hydro_sem", O_CREAT, 0600, 1);
    if (s->hydro_sem == SEM_FAILED) {
        perror("hydro_sem initialization failed.");
    }

    s->oxy_sem = sem_open("oxy_sem", O_CREAT, 0600, 1);
    if (s->oxy_sem == SEM_FAILED) {
        perror("oxy_sem initialization failed.");
    }

    if ((e = pthread_mutex_init(&s->mtx, NULL))) {
        error(e, "Mutex initialization failed.");
    }
    return s;
}

/*
===========================================
 * Oxygen thread
===========================================
*/
void* run_oxygen(void *arg) {

    scoreboard **s = (scoreboard **) arg;

    pthread_mutex_lock(&(*s)->mtx);
    (*s)->oxygen++;

    if ((*s)->hydrogen >= 2) {
        sem_post((*s)->hydro_sem);
        sem_post((*s)->hydro_sem);
        (*s)->hydrogen += -2;
        sem_post((*s)->oxy_sem);
        (*s)->oxygen--;
    } else {
        pthread_mutex_unlock(&(*s)->mtx);
    }
    timer_start(&start);
    sem_wait((*s)->oxy_sem);
    timer_stop(0, start, &(*s)->trk);

    /* bond("O"); */

    timer_start(&start);
    pthread_barrier_wait(&(*s)->barrier);
    timer_stop(0, start, &(*s)->trk);

    pthread_mutex_unlock(&(*s)->mtx);

    pthread_exit(NULL);

}

/*
===========================================
 * Hydrogen thread
===========================================
*/
void* run_hydrogen(void *arg) {

    scoreboard **s = (scoreboard **) arg;

    pthread_mutex_lock(&(*s)->mtx);
    (*s)->hydrogen++;

    if ((*s)->hydrogen >= 2 && (*s)->oxygen >= 1) {
        sem_post((*s)->hydro_sem);
        sem_post((*s)->hydro_sem);
        (*s)->hydrogen += -2;
        sem_post((*s)->oxy_sem);
        (*s)->oxygen--;
    } else {
        pthread_mutex_unlock(&(*s)->mtx);
    }

    timer_start(&start);
    sem_wait((*s)->hydro_sem);
    timer_stop(0, start, &(*s)->trk);

    /* bond("H"); */

    timer_start(&start);
    pthread_barrier_wait(&(*s)->barrier);
    timer_stop(0, start, &(*s)->trk);
    pthread_exit(NULL);

}


/*
 * =========================================
 * Main Routine
 * =========================================
*/
int main(int argc, char *argv[]) {

    int n_oxygen = atoi(argv[1]);
    int n_hydrogen = atoi(argv[2]);
    printf("\n\nCreating %d oxygen atoms and %d hydrogen atoms...\n\n", n_oxygen, n_hydrogen);

    int i, e;
    pthread_t tid[n_oxygen + n_hydrogen]; /* thread-ID array */

    scoreboard *s = init_scoreboard(n_oxygen, n_hydrogen);

    /* initialize thread attribute object */
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    /* launch oxygen threads */
    for (i = 0; i < n_oxygen; i++) {

        if ((e = pthread_create(&tid[i], &attr, run_oxygen, &s))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
    }
    /* launch hydrogen threads */
    for (i = n_oxygen; i < n_hydrogen + n_oxygen; i++) {

        if ((e = pthread_create(&tid[i], &attr, run_hydrogen, &s))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
    }
    /* Join exited threads */
    for (i = 0; i < n_hydrogen + n_oxygen; i++) {
        if(pthread_join(tid[i], NULL)) {
            fprintf(stderr, "Error joining thread %d\n", i);
            return(2);
        }
    }
    print_tracker(s->trk);
    pthread_barrier_destroy(&s->barrier);
    free(s);
    exit(0);
}