#include <stdio.h>

int main(int argc, char *argv[]) {
  printf("Hello world!\n");
  fflush(stdout);

  int a, b;
  scanf("%d %d", &a, &b);
  printf("%d\n", a + b);
  return 0;
}
