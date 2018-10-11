
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

/* clock parameters */
enum { NS_PER_SECOND = 1000000000 };
struct timespec start;

typedef struct tracker tracker;
struct tracker{
    int n;
    pthread_mutex_t mtx;
    double* tid; /* elapsed wait times by thread */
};

/* interface */
struct timespec timer_elapsed(struct timespec, struct timespec);
double get_elapsed(struct timespec, struct timespec);

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

/* initialize tracker */
tracker *init_tracker(int n) {
    tracker *t = (tracker *) emalloc(sizeof(tracker));
    t->n = n;
    t->tid = (double *) emalloc(sizeof(double[n]));

    /* initialize mutex */
    if (pthread_mutex_init(&t->mtx, NULL)) {
        fprintf(stderr, "Tracker mutex initialization failed.");
    }

    /* exit the thread */
    return t;
}

/* record wait time */
void print_tracker(tracker * t) {
    int i;
    double total_wt = 0;
    for (i = 0; i < t->n; i++) {
        /* printf("Thread %d: %04.10lf\n", i, t->tid[i]); */
        total_wt += t->tid[i];
    }
    printf("Total Wait Time: %04.10lf\n", total_wt);

}

 /* ===========================================
  function: clock timer (start/stop/elapsed)
  - Parameters: <train *> train queue
  - Return: timer_start ~ start time mark in ns
            timer_stop ~ elapsed time in ns
=========================================== */
void timer_start(struct timespec *start){
    if (clock_gettime(CLOCK_MONOTONIC, start) == -1) {
        perror( "clock gettime" );
        exit(1);
    }
}

/*
  subfunction timer_stop()
  -------------------------------------------
  stop the event timer
*/
void timer_stop(int pid, struct timespec start, tracker ** t) {
    struct timespec stop;
    if (clock_gettime(CLOCK_MONOTONIC, &stop) == -1) {
        perror("clock gettime");
        exit(1);
    }
    double elapsed = get_elapsed(start, stop);

    /* add wait time to tracker */
    pthread_mutex_lock(&(*t)->mtx);
    (*t)->tid[pid] += elapsed;
    pthread_mutex_unlock(&(*t)->mtx);
}

/*
  subfunction timer_elapsed()
  -------------------------------------------
  gets the elapsed time of start -> stop
  http://www.devcoons.com/determine-elapsed-time-using-monotonic-clocks-linux/
*/
struct timespec timer_elapsed(struct timespec start, struct timespec stop) {
    struct timespec elapsed_time;
    if ((stop.tv_nsec - start.tv_nsec) < 0)
    {
        elapsed_time.tv_sec = stop.tv_sec - start.tv_sec - 1;
        elapsed_time.tv_nsec = stop.tv_nsec - start.tv_nsec + NS_PER_SECOND;
    }
    else
    {
        elapsed_time.tv_sec = stop.tv_sec - start.tv_sec;
        elapsed_time.tv_nsec = stop.tv_nsec - start.tv_nsec;
    }
    return elapsed_time;
}

/*
  subfunction timer_mark()
  stores elapsed time in timespec pointer
  -------------------------------------------
  formats elapsed time as hh:mm:ss.ds
*/
double get_elapsed(struct timespec start, struct timespec stop) {

    struct timespec td = timer_elapsed(start, stop);

    double ns = 0;
    double ss = 0;

    /* format marked time */
    ns = ((double)td.tv_nsec) / (double) NS_PER_SECOND;
    ss = (double)(((int)td.tv_sec) % 60);
    return ns + ss;
}

/*
  subfunction timer_compare()
  compares two timespec values
  -------------------------------------------
  gets the difference in times
  return values  -1 t1 < t2
                  0 t1 = t2
                  1 t1 > t1
*/
int timer_compare(struct timespec t1, struct timespec t2){

    /* convert times to double */
    double t1_dsec = ((double)(t1.tv_nsec)) / (double) NS_PER_SECOND;;
    double t1_sec = (double)(t1.tv_sec);
    double t2_dsec = ((double)(t2.tv_nsec)) / (double) NS_PER_SECOND;;
    double t2_sec = (double)(t2.tv_sec);

    /* printf("\nT1:%012.9f THEAD:%012.9f\n\n", (t1_sec + t1_dsec), (t2_sec + t2_dsec)); */

    /* correct negative differences */
    if ((t1_sec + t1_dsec) < (t2_sec + t2_dsec)) {
        return -1;
    }
    else if ((t1_sec + t1_dsec) > (t2_sec + t2_dsec)) {
        return 1;
    }
    else {
        return 0;
    }

}