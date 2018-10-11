/*
 * =========================================
 * Example 2: FIFO Barbershop Problem (Exercise 5.3)
 * Implementation B
 * =========================================
 * Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
 * Spencer Rose (ID V00124060)
 * =========================================
 * SUMMARY: A barbershop consists of a waiting room with n chairs, and the barber room
 * containing the barber chair. If there are no customers to be served, the barber goes
 * to sleep. If a customer enters the barbershop and all chairs are occupied, then the
 * customer leaves the shop. If the barber is busy, but chairs are available, then the
 * customer sits in one of the free chairs. If the barber is asleep, the customer wakes
 * up the barber. Write a program to coordinate the barber and the customers.
 *
 * References
 * Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 101-111.
 * Implementation is an adaption of Downey's 5.2.2 Barbershop solution
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

typedef struct customer customer;
struct customer {
    int id;
    sem_t * sem;
    customer *next;
};

typedef struct queue queue;
struct queue {
    int n_customers;
    int tally;
    int n_seats;
    int customers;
    customer *head;
    customer *tail;
    pthread_mutex_t mtx;
    sem_t * customer;
    sem_t * customerDone;
    sem_t * barberDone;
    tracker * trk;
};

/* Interface */
void * run_barber(void *);
void * run_customer(void *);
customer *init_customer(int);
queue *init_queue(int, int);
int append(queue **, customer *);
customer *pop(queue **);
void display(queue **);
void error(int, char *);

/*
 * =========================================
 * Main Routine
 * =========================================
*/
int main(int argc, char *argv[]) {

    int n_customers = atoi(argv[1]); /* Number of customers to be generated (input) */
    printf("Creating %d customers...\n\n", n_customers);

    int n_seats = 4;
    int i, e;

    pthread_t tid[n_customers + 1]; /* thread-ID array */

    /* initialize thread attributes */
    pthread_attr_t attr;
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    queue *q = init_queue(n_seats, n_customers);

    /* Launch barber thread (as detached) */
    if ((e = pthread_create(&tid[0], &attr, run_barber, &q))) {
        printf("ERROR: pthread_create() has error code %d\n", e);
    }

    /* Launch customer threads */
    for (i = 1; i <= n_customers; i++) {
        if ((e = pthread_create(&tid[i], &attr, run_customer, &q))) {
            printf("ERROR: pthread_create() has error code %d\n", e);
        }
    }

    /* Join customer threads */
    for (i = 0; i < n_customers + 1; i++) {
        if(pthread_join(tid[i], NULL)) {
            fprintf(stderr, "Error joining thread %d\n", i);
            return 2;
        }
    }
    print_tracker(q->trk);

    /* exit the main thread */
    free(q);
    exit(0);
}


/*
  -------------------------------------------
   error message
*/
void error(int e, char *msg) {
    printf("ERROR: %s has error code %d\n", msg, e);
}

void get_cut_hair(int id) {
    /* printf("Get haircut: customer #%d.\n", id); */
}

void cut_hair(int id) {
    /* printf("Cutting hair of customer #%d.\n", id); */
}

void balk(int id) {
    /* printf("Waiting room is full! Customer #%d has left.\n", id); */
}

/*
===========================================
 * Customer thread
===========================================
*/
void* run_customer(void *arg) {

    queue **q = (queue **) arg;
    struct timespec start;

    pthread_mutex_lock(&(*q)->mtx);
    int i = ++(*q)->tally; /* get customer number */
    customer *c = init_customer(i);

    /* Try to enqueue customer; balk if seats are full */
    if (append(q, c)) {
        pthread_mutex_unlock(&(*q)->mtx);
        balk(c->id);
        free(c);
        pthread_exit(NULL);
    }
    pthread_mutex_unlock(&(*q)->mtx);

    /* signal barber that customer is ready */
    sem_post((*q)->customer);

    /* Wait for barber's signal on self semaphore */
    timer_start(&start);
    sem_wait(c->sem);

    /* get_cut_hair(c->id); */

    /* Wait for barber to finish and signal customer exits */
    sem_wait((*q)->barberDone);
    timer_stop(i, start, &(*q)->trk);
    sem_post((*q)->customerDone);

    free(c);
    pthread_exit(NULL);
}

/*
===========================================
 * Barber thread
===========================================
*/
void* run_barber(void *arg) {

    queue **q = (queue **) arg;
    struct timespec start;

    while((*q)->tally < (*q)->n_customers || (*q)->customers > 0) {

        /* wait for customers to enqueue */
        while((*q)->customers == 0)
            sem_wait((*q)->customer);

        pthread_mutex_lock(&(*q)->mtx);
        /* display(q); */
        customer *c = pop(q);
        /* printf("Tally: %d| Last customer: %d|waiting: %d\n", (*q)->tally, c->id, (*q)->customers); */

        pthread_mutex_unlock(&(*q)->mtx);
        sem_post(c->sem);

        /* cut_hair(c->id); */

        sem_post((*q)->barberDone);

        timer_start(&start);
        sem_wait((*q)->customerDone);
        timer_stop(0, start, &(*q)->trk);

    }
    pthread_exit(NULL);
}


/*
 * ===========================================
 * Queue (data structure)
 * ===========================================
 */

queue *init_queue(int n_seats, int n_customers) {

    queue *q = (queue *) emalloc(sizeof(queue));

    /* set default values */
    q->n_customers = n_customers; /* absolute number of customers */
    q->tally = 0; /* tally of number of customers */
    q->customers = 0; /* number of customers in waiting room */
    q->n_seats = n_seats; /* number of seats in waiting room */
    q->head = NULL; /* head of list */
    q->tail = NULL; /* end of list */
    q->trk = init_tracker(n_customers); /* semaphore tracker */

    int e; /* error code */

    /* initialize mutex */
    if ((e = pthread_mutex_init(&q->mtx, NULL))) {
        error(e, "mtx");
    }
    sem_unlink("customers");
    q->customer = sem_open("customers", O_CREAT, 0644, 0);
    if (q->customer == SEM_FAILED) {
        perror("customers sem initialization failed.");
    }
    sem_unlink("customerDone");
    q->customerDone = sem_open("customerDone", O_CREAT, 0644, 0);
    if (q->customerDone == SEM_FAILED) {
        perror("customerDone sem initialization failed.");
    }
    sem_unlink("barberDone");
    q->barberDone = sem_open("barberDone", O_CREAT, 0644, 0);
    if (q->barberDone == SEM_FAILED) {
        perror("barberDone sem initialization failed.");
    }

    return q;
}

/*
  -------------------------------------------
  new customer (constructor)
*/
customer *init_customer(int id) {

    customer *c = (customer *) emalloc(sizeof(customer));
    c->id = id;
    c->next = NULL;
    char *name = (char *) emalloc(sizeof(char)*10);
    sprintf(name, "c_%d", id);

    c->sem = sem_open(name, O_CREAT, 0644, 0);
    if (c->sem == SEM_FAILED) {
        perror("self sem initialization failed.");
    }
    return c;
}

/*
  -------------------------------------------
  add customer to end of queue; returns 0 on success
*/
int append(queue **q, customer *c) {

    if ((*q)->head == NULL) { /* empty list */
        (*q)->head = c;
        (*q)->tail = c;
    } else if ((*q)->customers < (*q)->n_seats) {
        (*q)->tail->next = c;
        (*q)->tail = c;
    } else { /* seats are full! */
        return(1);
    }
    /* increment wait counter */
    (*q)->customers++;
    return(0);
}

/*
  -------------------------------------------
  pop customer from FIFO queue; returns top customer
*/
customer *pop(queue **q) {

    customer *top = (*q)->head;

    if (top == NULL) { /* empty queue */
        perror("Trying to pop an empty queue.");
        return NULL;
    } else if (top->next == NULL) { /* one customer in queue */
        (*q)->head = NULL;
        (*q)->tail = NULL;
    } else {
        (*q)->head = top->next;
    }
    /* Decrement wait counter */
    (*q)->customers--;
    return top;
}

/*
  -------------------------------------------
  display queue
*/
void display(queue **l) {

    customer *curr = (*l)->head;
    if (curr == NULL) { /* empty list */
        return;
    } else {
        printf("\nQUEUE\n");
        int i;
        for (i = 1; curr != NULL; curr = curr->next, i++){
            printf("%d. Customer %d|%d\n", i, curr->id, (*l)->tally);
        }
    }
}