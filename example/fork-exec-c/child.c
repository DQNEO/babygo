//  child.c: the child program
#include <stdio.h>
#include <stdlib.h>

int main (int argc, char **argv)
{
    int i, j;
    long x;

    // Some arbitrary work done by the child

    printf ("child: Hello, World!\n");

    for (j = 0; j < 1000; j++ ) {
        for (i =0; i < 900000; i++) {
            x = i * j;
        }
    }

    printf("child: Work completed!\n");
    printf("child: Bye.\n");

    exit(0);
}
