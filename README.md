# DSL

Небольшой экспериментальный DSL на Go. Внутри есть lexer, Pratt parser и evaluator.
Язык сейчас похож на смесь выражений, map literal и простого `select` по Go/JSON данным.

Основная идея: из Go создается parser, в context кладутся значения и функции, потом DSL
выражения парсятся и выполняются.

```go
p := parser.NewParser(`
    users = select * from json(get("https://jsonplaceholder.typicode.com/users/"))
    out = {id: 1} any users
    stringify(out)
`)
```

## Запуск

```bash
go run .
go test ./...
```

## Значения

Поддерживаются:

- числа: `1`, `1.5`
- строки: `"hello"`
- bool: `true`, `false`
- `nil` и `null`
- `nan`
- `undefined` и alias `undefind`
- map: `{id: 1, name: "user"}`
- значения из Go context: `users`, `somevar`, `r`
- вызовы Go-функций из context: `json(...)`, `get(...)`, `stringify(...)`

`null` и `nil` оба дают nil-значение:

```dsl
r = null == nil
```

`nan == nan` в DSL считается true.

Если member-ключа реально нет, возвращается `undefined`:

```dsl
somevar = {a: {id: 1}}
r = somevar.a.ida
```

Если member применяется не к map-объекту, возвращается `nil`.

## Assignment

```dsl
answer = 1 + 2
user = {id: 1, name: "Leanne Graham"}
```

Левая часть assignment должна быть identifier.

## Map

Map literal:

```dsl
somevar = {
    id: 1,
    name: "",
    age: r.age,
    nested: {
        ok: true
    }
}
```

Ключ может быть identifier или string literal. Значения внутри map являются обычными
выражениями.

## Member

Member chain работает через точку:

```dsl
street = user.address.street
lng = user.address.geo.lng
```

Цепочка может быть любой длины. Если на каком-то уровне ключ не найден, результат будет
`undefined`, а не panic.

## Unary

```dsl
a = -5
b = +5
c = !true
d = !!!!!true
```

Для `!` можно писать цепочку. Четное количество `!` возвращает исходный bool, нечетное
инвертирует.

Truthy/falsy cast для logical:

- false: `0`, `-0`, `nil`, `null`, `undefined`
- true: почти все остальное, включая map и slice

## Binary Math

`syntax/binary` сейчас отвечает только за math:

```dsl
1 + 2
1 - 2
2 * 3
6 / 3
7 % 4
```

`+` со string делает concat:

```dsl
1 + ""
```

Результат: `"1"`.

Если math-оператор получает неподходящий тип, результат `nan`, а не ошибка:

```dsl
1 / ""
1 + {a: 1}
1 * {a: 1}
```

## Comparison

Сравнения лежат отдельно в `syntax/comparison`:

```dsl
1 < 2
1 <= 2
2 > 1
2 >= 1
1 == 1.0
name != "admin"
```

Числа разных Go numeric types сравниваются по значению, например `int64(1)` и
`float64(1)` считаются равными.

## Logical

Logical лежит отдельно в `syntax/logical`:

```dsl
r = value || true
ok = active && enabled
a = left or right
b = left and right
```

Перед `&&`, `||`, `and`, `or` значения кастятся в bool.

## Call

DSL может вызывать функции, которые зарегистрированы в Go context:

```dsl
sum(1, 2 + 3)
json(get("https://example.com/data.json"))
stringify(user)
```

Если функции нет в context, evaluation вернет ошибку.

### Регистрация функций

Функции регистрируются из Go через `SetFunc`. Аргументы приходят как
`...value.Type`, вернуть нужно тоже `value.Type`.

```go
p.SetFunc("sum", func(vals ...value.Type) (value.Type, error) {
    var total float64

    for _, val := range vals {
        total += val.UnsafeCastFloat64()
    }

    return value.NewType(total), nil
})
```

После этого функцию можно вызвать из DSL:

```dsl
r = sum(1, 2 + 3)
```

Функции обычно проверяют количество и тип аргументов сами:

```go
p.SetFunc("json", func(vals ...value.Type) (value.Type, error) {
    if len(vals) != 1 {
        return value.NewTypeNil(), fmt.Errorf("json() expects exactly 1 argument")
    }

    str, ok := vals[0].CastString()
    if !ok {
        return value.NewTypeNil(), fmt.Errorf("json() expects string")
    }

    var data any
    if err := json.Unmarshal([]byte(str), &data); err != nil {
        return value.NewTypeNil(), err
    }

    return value.NewType(data), nil
})
```

## Select

`select` берет map или slice maps и возвращает только нужные поля.

```dsl
users = select
    id,
    name,
    username
from json(get("https://jsonplaceholder.typicode.com/users/"))
```

### select *

```dsl
users = select * from users
```

`*` возвращает всю строку целиком.

### Member fields

```dsl
users = select
    id,
    address.street,
    address.geo.lng
from users
```

В output ключом становится последний segment:

```dsl
select address.geo.lng from users
```

даст поле `lng`.

### Alias

```dsl
user1 = select
    id as pin,
    1 + 2 as sum,
    address.geo
from users
```

Для expression field лучше указывать alias, иначе имя будет выведено из типа выражения
вроде `binary`.

### where и limit

```dsl
activeUsers = select * from users where active == true limit 1
```

В `where` доступны поля текущей строки как identifiers:

```dsl
select name as username from users where address.city == "Gwenborough"
```

## any

`any` ищет любое совпадение. Для map это значит: достаточно одного общего ключа с
одинаковым значением.

```dsl
sample = {id: 1, a: 2, b: 3}
user = {id: 1, name: "Leanne Graham"}
r = sample any user
```

Результат: right map `user`, потому что `id` совпал.

Если right side array, результат тоже array с найденными map:

```dsl
sample = {id: 1}
users = select * from json(get("https://jsonplaceholder.typicode.com/users/"))
r = sample any users
```

Результат: array всех rows, где найдено хотя бы одно совпадение по ключу и значению.

Если right side map, результат map:

```dsl
r = {id: 1} any {id: 1, name: "user", age: 20}
```

Результат:

```json
{"id": 1, "name": "user", "age": 20}
```

Для primitives `any` ищет совпавшее значение:

```dsl
r = 2 any nums
```

Если `nums` содержит `2`, результат будет `2`. Если совпадений нет, результат `nil`.

## all

`all` строже чем `any`: left должен полностью содержаться в right.

Для map это значит: все ключи и значения из left должны быть в right.

```dsl
sample = {id: 1, name: "user"}
target = {id: 1, name: "user", age: 20}
r = sample all target
```

Результат: right map `target`.

Если left содержит лишние ключи, которых нет в right, результат `nil`:

```dsl
sample = {id: 1, a: 2, b: 3, c: 4}
users = select * from json(get("https://jsonplaceholder.typicode.com/users/"))
out = sample all users
```

Тут `out` будет `nil`, если в users нет row, где есть все ключи `id`, `a`, `b`, `c`
с такими же значениями.

Для primitives:

```dsl
r = 2 all nums
```

Если `nums` содержит `2`, результат будет `2`.

Для slices `all` проверяет, что все значения из left есть в right.

## Текущее разделение syntax-пакетов

- `syntax/literal`: literals, identifiers, format strings
- `syntax/assignment`: `=`
- `syntax/map`: `{key: value}`
- `syntax/member`: `a.b.c`
- `syntax/unary`: `!`, unary `-`, unary `+`
- `syntax/binary`: math only
- `syntax/comparison`: `>`, `<`, `>=`, `<=`, `==`, `!=`
- `syntax/logical`: `&&`, `||`, `and`, `or`
- `syntax/any`: `any`
- `syntax/all`: `all`
- `syntax/call`: `fn(...)`
- `syntax/select`: `select ... from ... where ... limit ...`
