# How to solve palindromes and fixed stuff

```go

func palin(line string) bool {
    left, right := 0, len(line)-1
    for left < right {
        if line[left] != line[right] {
            return false
        }
        left++
        right--
    }
    return true
}

/*
Временная сложность: O(n)
Память: O(1)
*/
```
