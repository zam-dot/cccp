#include <stdio.h>
#include <string.h>
#include <stdlib.h>

char* concat_strings(const char* a, const char* b) {
    char* result = malloc(strlen(a) + strlen(b) + 1);
    strcpy(result, a);
    strcat(result, b);
    return result;
}

int add(int a, int b);
int multiply(int a, int b);

int add(int a, int b) {
    return a + b;
}

int multiply(int a, int b) {
    return a * b;
}

int main() {
    int sum = add(5, 3);
    int product = multiply(4, 6);
    printf("%d\n", sum);
    printf("%d\n", product);
    return 0;
}
