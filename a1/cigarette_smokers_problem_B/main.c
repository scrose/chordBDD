/*
 * =========================================
 * Example 1: Cigarette Smokers Problem (Exercise 4.5)
 * Implementation B
 * =========================================
 * Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
 * Spencer Rose (ID V00124060)
 * =========================================
 * SUMMARY: Four threads are involved: an agent and three smokers. The smokers loop
 * forever, first waiting for ingredients, then making and smoking cigarettes. The
 * ingredients are tobacco, paper, and matches. We assume that the agent has an infinite
 * supply of all three ingredients, and each smoker has an infinite supply of one of the
 * ingredients; that is, one smoker has matches, another has paper, and the third has
 * tobacco. The agent repeatedly chooses two different ingredients at random and makes
 * them available to the smokers. Depending on which ingredients are chosen, the smoker
 * with the complementary ingredient should pick up both resources and proceed.
 *
 * Reference: Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 101-111.
 */
#define _POSIX_C_SOURCE 199309L
#define _BSD_SOURCE
#include <assert.h>
#include <unistd.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <semaphore.h>

typedef struct {
    int n;
    pthread_mutex_t mtx;
    pthread_mutex_t smokes_mtx;
    int tobacco;
    int papers;
    int matches;
    int smokes;
    sem_t * tobacco_sem;
    sem_t * papers_sem;
    sem_t * matches_sem;
    sem_t * tobacco_helper;
    sem_t * papers_helper;
    sem_t * matches_helper;
    sem_t * ready;
} ingredients;

void* agent(void *arg);
void* smokerA(void *arg);
void* smokerB(void *arg);
void* smokerC(void *arg);
void* helperA(void *arg);
void* helperB(void *arg);
void* helperC(void *arg);
ingredients* init_ingredients(int);


/*
===========================================
 * Main Routine
===========================================
*/
int main(int argc, char *argv[]) {

    int n = atoi(argv[1]); /* Number of cigarettes to be generated (input) */
    printf("Making %d cigarettes...\n\n", n);

    int i, e;
    srand(time(0));

    /* thread-ID array for agent, helper and 3 smokers */
    pthread_t tid[6];

    /* initialize thread attribute object */
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    ingredients *ing = init_ingredients(n);

    /* launch Agent thread */
    if ((e = pthread_create(&tid[0], &attr, agent, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }
    /* launch Helper A thread */
    if ((e = pthread_create(&tid[1], &attr, helperA, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }
    /* launch Helper B thread */
    if ((e = pthread_create(&tid[2], &attr, helperB, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }
    /* launch Helper C thread */
    if ((e = pthread_create(&tid[3], &attr, helperC, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }
    /* launch Smoker A thread */
    if ((e = pthread_create(&tid[4], &attr, smokerA, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }
    /* launch Smoker B thread */
    if ((e = pthread_create(&tid[5], &attr, smokerB, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }
    /* launch Smoker C thread */
    if ((e = pthread_create(&tid[6], &attr, smokerC, &ing))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }

    /* Join helper and customer threads */
    for (i = 0; i < 6; i++) {
        if(pthread_join(tid[i], NULL)) {
            fprintf(stderr, "Error joining thread %d\n", i);
            return 2;
        }
    }

    /* exit the thread */
    free(ing);
    exit(0);
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
   allocate element memory; Returns error on NULL allocation
*/
void *emalloc(size_t n) {
    void *p;
    p = malloc(n);
    if (p == NULL) {
        fprintf(stderr, "ERROR: Malloc of %zu bytes failed.", n);
        _exit(1);
    }
    return p;
}

void smoke() {
   printf("Smoking...\n");
}


/*
 -------------------------------------------
  scoreboard tracks the H2O bonds and counts
*/
ingredients *init_ingredients(int n) {

    /* create event tracker */

    int e; /* error code */

    ingredients *i = (ingredients *) emalloc(sizeof(ingredients));

    i->n = n;
    i->tobacco = 0;
    i->papers = 0;
    i->matches = 0;
    i->smokes = 0;

    /* initialize mutex and conditional variables */
    if ((e = pthread_mutex_init(&i->mtx, NULL))) {
        error(e, "Mutex initialization failed.");
    }
    if ((e = pthread_mutex_init(&i->smokes_mtx, NULL))) {
        error(e, "Smokes Mutex initialization failed.");
    }

    /* initialize semaphores */
    sem_unlink("tobacco_sem");
    i->tobacco_sem = sem_open("tobacco_sem", O_CREAT, 0644, 0);
    if (i->tobacco_sem == SEM_FAILED) {
        perror("tobacco sem initialization failed.");
    }
    sem_unlink("papers_sem");
    i->papers_sem = sem_open("papers_sem", O_CREAT, 0644, 0);
    if (i->papers_sem == SEM_FAILED) {
        perror("papers sem initialization failed.");
    }
    sem_unlink("matches_sem");
    i->matches_sem = sem_open("matches_sem", O_CREAT, 0644, 0);
    if (i->matches_sem == SEM_FAILED) {
        perror("matches_sem sem initialization failed.");
    }
    sem_unlink("tobacco_helper");
    i->tobacco_helper = sem_open("tobacco_helper", O_CREAT, 0644, 0);
    if (i->tobacco_helper == SEM_FAILED) {
        perror("tobacco_helper sem initialization failed.");
    }
    sem_unlink("papers_helper");
    if ((i->papers_helper = sem_open("papers_helper", O_CREAT, 0644, 0)) == SEM_FAILED) {
        perror("papers_helper sem initialization failed.");
    }
    sem_unlink("matches_helper");
    if ((i->matches_helper = sem_open("matches_helper", O_CREAT, 0644, 0)) == SEM_FAILED) {
        perror("matches_helper sem initialization failed.");
    }
    sem_unlink("ready");
    i->ready = sem_open("ready", O_CREAT, 0644, 0);
    if (i->ready == SEM_FAILED) {
        perror("ready sem initialization failed.");
    }

    return i;
}

/*
===========================================
 * Agent: releases ingredients (resources) to smokers.
===========================================
*/
void* agent(void *arg) {

    int r;
    ingredients **ing = (ingredients **) arg;

    while ((*ing)->smokes < (*ing)->n) {
        /* generate pseudo-random int between 1 and 3 */
        rand();
        r = rand() % 3 + 1;
        printf("%d",r);

        switch (r) {
            case 1:
                /* signal tobacco helper */
                sem_post((*ing)->tobacco_helper);
               break;
            case 2:
                /* signal papers helper */
                sem_post((*ing)->papers_helper);
                break;
            case 3:
                /* signal matches helper */
                sem_post((*ing)->matches_helper);
                break;
            default:
                perror("Agent thread failed.");
        }

        pthread_mutex_lock(&(*ing)->mtx);
        pthread_mutex_lock(&(*ing)->smokes_mtx);
        printf("Released: T: %-10d | P: %-5d | M: %-5d | # Smokes: %d | %-5d \n",
                (*ing)->tobacco, (*ing)->papers, (*ing)->matches, (*ing)->smokes, (*ing)->n);
        pthread_mutex_unlock(&(*ing)->smokes_mtx);
        pthread_mutex_unlock(&(*ing)->mtx);
    }
    sem_post((*ing)->tobacco_helper);
    sem_post((*ing)->papers_helper);
    sem_post((*ing)->matches_helper);
    pthread_exit(NULL);
}

/*
===========================================
 * Smoker A: has tobacco; needs papers and matches.
===========================================
*/
void* smokerA(void *arg) {

    ingredients **ing = (ingredients **) arg;

    while ((*ing)->smokes < (*ing)->n) {

        sem_wait((*ing)->tobacco_sem);
        if ((*ing)->smokes >= (*ing)->n) {
            pthread_exit(NULL);
        }
        /* makeCigarette() */
        smoke();
        pthread_mutex_lock(&(*ing)->smokes_mtx);
        (*ing)->smokes++;
        pthread_mutex_unlock(&(*ing)->smokes_mtx);
    }
    pthread_exit(NULL);
}


/*
===========================================
 * Smoker B: has papers; needs tobacco and matches.
===========================================
*/
void* smokerB(void *arg) {

    ingredients **ing = (ingredients **) arg;

    while ((*ing)->smokes < (*ing)->n) {

        sem_wait((*ing)->papers_sem);
        if ((*ing)->smokes >= (*ing)->n) {
            pthread_exit(NULL);
        }
        /* makeCigarette() */
        smoke();
        pthread_mutex_lock(&(*ing)->smokes_mtx);
        (*ing)->smokes++;
        pthread_mutex_unlock(&(*ing)->smokes_mtx);
    }
    pthread_exit(NULL);
}

/*
===========================================
 * Smoker C: has matches; needs tobacco and papers.
===========================================
*/
void* smokerC(void *arg) {

    ingredients **ing = (ingredients **) arg;

    while ((*ing)->smokes < (*ing)->n) {
        sem_wait((*ing)->matches_sem);

        if ((*ing)->smokes >= (*ing)->n) {
            pthread_exit(NULL);
        }
        /* makeCigarette() */
        smoke();
        pthread_mutex_lock(&(*ing)->smokes_mtx);
        (*ing)->smokes++;
        pthread_mutex_unlock(&(*ing)->smokes_mtx);
    }
    pthread_exit(NULL);
}

/*
===========================================
 * Helpers: communicates between agent and smokers.
===========================================
*/

/*
===========================================
 * Helper A: has tobacco; checks for papers and matches.
===========================================
*/
void* helperA(void *arg) {

    ingredients **ing = (ingredients **) arg;

    sem_post((*ing)->ready);
    while ((*ing)->smokes < (*ing)->n) {

        if ((*ing)->smokes >= (*ing)->n) {
            sem_post((*ing)->tobacco_sem);
            sem_post((*ing)->papers_sem);
            sem_post((*ing)->matches_sem);
            pthread_exit(NULL);
        }

        /* wait for agent to release tobacco */
        sem_wait((*ing)->tobacco_helper);

        pthread_mutex_lock(&(*ing)->mtx);
        if ((*ing)->papers) {
            sem_post((*ing)->matches_sem);
            (*ing)->papers--;
        }
        else if ((*ing)->matches) {
            sem_post((*ing)->papers_sem);
            (*ing)->matches--;
        }
        else {
            (*ing)->tobacco++;
        }
        pthread_mutex_unlock(&(*ing)->mtx);
        sem_post((*ing)->ready);
    }
    sem_post((*ing)->tobacco_sem);
    sem_post((*ing)->papers_sem);
    sem_post((*ing)->matches_sem);
    pthread_exit(NULL);
}

/*
===========================================
 * Helper B: has papers; checks for tobacco and matches.
===========================================
*/
void* helperB(void *arg) {

    ingredients **ing = (ingredients **) arg;

    sem_post((*ing)->ready);
    while ((*ing)->smokes < (*ing)->n) {

        if ((*ing)->smokes == (*ing)->n) {
            sem_post((*ing)->tobacco_sem);
            sem_post((*ing)->papers_sem);
            sem_post((*ing)->matches_sem);
            pthread_exit(NULL);
        }

        /* wait for agent to release papers */
        sem_wait((*ing)->papers_helper);

        pthread_mutex_lock(&(*ing)->mtx);
        if ((*ing)->tobacco) {
            (*ing)->tobacco--;
            sem_post((*ing)->matches_sem);
        } else if ((*ing)->matches) {
            (*ing)->matches--;
            sem_post((*ing)->tobacco_sem);
        } else {
            (*ing)->papers++;
        }
        pthread_mutex_unlock(&(*ing)->mtx);
        sem_post((*ing)->ready);
    }
    sem_post((*ing)->tobacco_sem);
    sem_post((*ing)->papers_sem);
    sem_post((*ing)->matches_sem);
    pthread_exit(NULL);
}

/*
===========================================
 * Helper C: has matches; checks for tobacco and papers.
===========================================
*/
void* helperC(void *arg) {

    ingredients **ing = (ingredients **) arg;
    sem_post((*ing)->ready);
    while ((*ing)->smokes < (*ing)->n) {

        if ((*ing)->smokes == (*ing)->n) {
            sem_post((*ing)->tobacco_sem);
            sem_post((*ing)->papers_sem);
            sem_post((*ing)->matches_sem);
            pthread_exit(NULL);
        }

        /* wait for agent to release matches */
        sem_wait((*ing)->matches_helper);

        pthread_mutex_lock(&(*ing)->mtx);
        if ((*ing)->tobacco) {
            sem_post((*ing)->papers_sem);
            (*ing)->tobacco--;
        } else if ((*ing)->papers) {
            sem_post((*ing)->tobacco_sem);
            (*ing)->papers--;
        } else {
            (*ing)->matches++;
        }
        pthread_mutex_unlock(&(*ing)->mtx);
        sem_post((*ing)->ready);
    }
    sem_post((*ing)->tobacco_sem);
    sem_post((*ing)->papers_sem);
    sem_post((*ing)->matches_sem);
    pthread_exit(NULL);
}