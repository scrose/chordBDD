/*
 * ===========================================
 * Search-Insert-Delete (data structure)
 * ===========================================
 * Provides searcher-inserter-deleter implementation.
 */

#include "list.h"

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

/*
  -------------------------------------------
   error message
*/
void error(int e, char *msg) {
    printf("ERROR: %s has error code %d\n", msg, e);
}


linkedlist *new_linkedlist() {

    linkedlist *l = (linkedlist *) emalloc(sizeof(linkedlist));

    /* get input arguments */
    l->head = NULL; /* head of list */
    l->tail = NULL; /* end of list */

    int e; /* error code */

    /* initialize mutex and conditional variables */

    l->d = 0; /*  deleters */
    l->s = 0; /*  searchers */
    l->i = 0; /*  inserters */

    /* initialize mutex and conditional variables */
    if ((e = pthread_mutex_init(&l->ts_mtx, NULL))) {
        error(e, "mtx_turnstile");
    }
    if ((e = pthread_condattr_init(&l->ts_cond_attr))) {
        error(e, "cond_attr_turnstile");
    }
    if ((e = pthread_cond_init(&l->ts_cond, &l->ts_cond_attr))) {
        error(e, "cond_turnstile");
    }

    if ((e = pthread_mutex_init(&l->inserter_mtx, NULL))) {
        error(e, "mtx_list");
    }
    if ((e = pthread_condattr_init(&l->inserter_cond_attr))) {
        error(e, "cond_attr_list");
    }
    if ((e = pthread_cond_init(&l->inserter_cond, &l->inserter_cond_attr))) {
        error(e, "cond_list");
    }

    if ((e = pthread_mutex_init(&l->searcher_mtx, NULL))) {
        error(e, "mtx_list");
    }
    if ((e = pthread_condattr_init(&l->searcher_cond_attr))) {
        error(e, "cond_attr_list");
    }
    if ((e = pthread_cond_init(&l->searcher_cond, &l->searcher_cond_attr))) {
        error(e, "cond_list");
    }

    if ((e = pthread_mutex_init(&l->deleter_mtx, NULL))) {
        error(e, "mtx_list");
    }
    if ((e = pthread_condattr_init(&l->deleter_cond_attr))) {
        error(e, "cond_attr_list");
    }
    if ((e = pthread_cond_init(&l->deleter_cond, &l->deleter_cond_attr))) {
        error(e, "cond_list");
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
  add element to linkedlist
*/
void add(linkedlist **l, int val) {

    if ((*l)->head == NULL) { /* empty list */
        (*l)->head = new_element(val);
    }
    else { /* add element to end of list */
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
    else { /* search for element */
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

    element *curr = (*l)->head; /* element pointer to head of list */

    if (curr == NULL) { /* empty list */
        return;
    } else if (curr->value == val) { /* element at head of list */
        (*l)->head = curr->next;
        return;
    } else { /* element is in middle of list */
        element *prev = (*l)->head; /* previous pointer */
        for (; curr->next != NULL; curr = curr->next, prev = curr) {
            if (curr->value == val) { /* found the element */
                prev->next = curr->next;
                if (curr->next == NULL) { /* element is at the end of list */
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
