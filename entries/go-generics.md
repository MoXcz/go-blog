---
date: 30-Nov-2025
title: Go generics in action
---

Generics in Go arrived really late on version 1.18, and with reason, as its *usage should be reserved strictly when necessary*. There's very few cases when making use of a generic makes sense, for example an `assert`-like type of function:
```go
func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
```

> `t.Helper()` *returns* the error back up to the caller to avoid showing the `file:line` error message on this function

A useless example, but that helps explain type restriction:
```go
func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
    var s V
    for _, v := range m {
        s += v
    }
    return s
}
```

Note that `comparable` is just an interface similar to this one:
```go
type Number interface {
    int64 | float64
}
```

Which just helps *restrict* the kind of values this generic can be.

