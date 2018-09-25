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
#include "queue.h"

// Pthread routines
void searcher(list **, int);
void inserter(list **, int);
void deleter(list **, int);

// Singly-linked integer list

typedef struct list list;
struct list {
   int value; /* item value */
   list *next; /* next item in list */
   list *head; /* first item in list */
   list *last; /* last item in list */
 };

list *new_list();
void search(list **, int);
void insert(list **, int);
void delete(list **, int);
void display(list **);

// Utility functions
void *emalloc(size_t);


/*
===========================================
Main Routine
===========================================
*/
int main(int argc, char *argv[]) {

  // Build singly-linked list
  // Generate random threads
  exit(0);
}


/*
  new singly-linked list (constructor)
*/
list *new_list() {

  list *l = (list *) emalloc(sizeof(list));

  l->value = 0; /*  default value of head */
  l->next = NULL; /* next item pointer */
  l->head = NULL; /* first item pointer */
  l->last = NULL; /* last item pointer */

  return l;
}

/*
===========================================
// Searchers merely examine the list; hence they can execute concurrently
// with each other.
===========================================
*/
void searcher(list **l) {

}

/*
===========================================
// Inserters add new items to the end of the list; insertions must be mutually
// exclusive to preclude two inserters from inserting new items at about the
// same time. However, one insert can proceed in parallel with any number of
// searches.
===========================================
*/
void inserter(list **l) {

}

/*
===========================================
deleters remove items from anywhere in the list. At most one deleter process
can access the list at a time, and deletion must also be mutually exclusive
with searches and insertions.
===========================================
*/
void deleter(list **l) {

}


/*
===========================================
inserts an item into end of singly-linked list
parameters: singly-linked list, integer value
return: none
===========================================
*/
void insert(list **l, int x) {

  /* insert into empty list */
  if ((*l)->head == NULL) {
    (*l)->head = l;
    (*l)->last = l;
    return;
  }
  /* otherwise create new item and append to list */
  else {
    list *n = (list *) emalloc(sizeof(list));
    (*l)->last->next = n;
    (*l)->last = n;
  }

  (*l)->last->value = x;
  return;
}

/*
===========================================
deletes a random item from singly-linked list
parameters: singly-linked list, integer value
return: none
===========================================
*/
void delete(list **l) {

  /* delete from empty list */
  if ((*l)->head == NULL) {
    fprintf(stderr, "Cannot delete from an empty list.");
    _exit(1);
    return;
  }
  /* otherwise delete randomly selected item from list */
  else {
    list *item = *l;
    list *prev = item;
    int i, n;
    for (i = 0; curr->next != NULL; i++, item = item->next, prev=item);
    prev->next = item->next;
    if (item->last == item) {
      (*l)
    }
    item->next = NULL;
    item->head = NULL;
    item->last = NULL;

  item->last->value = x;
  return;
}


/* ===========================================
displays the list
=========================================== */
void show_queue(list *l) {

  /* write list to stdout */
  for (; l != NULL; l = l->next) {
    printf(l);
  }
  printf("\n");
}

/* show application usage */
void usage() {
    fprintf(stderr, "usage: mts <inputfile> \n"
    );
}
