typedef struct
{
    const int i;
    const int a;
} User;

void *foo(const User aa) {}

void asd()
{
    User q = {
        .a = 1,
        .i = 2,
    };

    foo(q);
}
