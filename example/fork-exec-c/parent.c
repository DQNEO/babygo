// parent.c: the parent program

/* Usage:
# compile parent
$ gcc parent.c -o parent
# compile child
$ gcc child.c -o child
# run parent (and child)
$ ./parent
Parent: Hello, World!
Parent: Waiting for Child to complete.
Child: Hello, World!
Child: Work completed!
Child: Bye now.
Parent: Child process waited for.
*/
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <unistd.h>
#include <string.h>
#include <sys/wait.h>

int main (int argc, char **argv)
{
    int i = 0;
    int pid, status, ret;
    char *myargs [] = { NULL };
    char *myenv [] = { NULL };

    printf("parent: Hello\n");
    pid = fork();
    if (pid == 0) {
        // I am the child
        execve ("child", myargs, myenv);
    }

    // I am the parent
    printf ("parent: Waiting for Child to complete.\n");
    if ((ret = waitpid(pid, &status, 0)) == -1)
         printf ("parent:error\n");

    if (ret == pid)
        printf ("parent: Bye.\n");
}