# TRANSLIT

```go
package main

import (
	"fmt"
	"strings"
)

// translit converts Russian text to English transliteration.
func translit(s string) string {
	replacer := strings.NewReplacer(
		"а", "a", "б", "b", "в", "v", "г", "g", "д", "d",
		"е", "e", "ё", "e", "ж", "zh", "з", "z", "и", "i",
		"й", "y", "к", "k", "л", "l", "м", "m", "н", "n",
		"о", "o", "п", "p", "р", "r", "с", "s", "т", "t",
		"у", "u", "ф", "f", "х", "kh", "ц", "ts", "ч", "ch",
		"ш", "sh", "щ", "shch", "ъ", "", "ы", "y", "ь", "",
		"э", "e", "ю", "yu", "я", "ya",
	)
	return replacer.Replace(strings.ToLower(s))
}

func main() {
	input := "вкусняшка для кли"
	output := translit(input)
	fmt.Println(output) // vkusnyashka dlya kli
}


```
--- с заглавными ---
```go
package main

import (
    "fmt"
    "strings"
)

func translit(rus string) string {
	rules := map[string]string{
		"а": "a", "б": "b", "в": "v", "г": "g", "д": "d",
		"е": "e", "ё": "e", "ж": "zh", "з": "z", "и": "i",
		"й": "y", "к": "k", "л": "l", "м": "m", "н": "n",
		"о": "o", "п": "p", "р": "r", "с": "s", "т": "t",
		"у": "u", "ф": "f", "х": "kh", "ц": "ts", "ч": "ch",
		"ш": "sh", "щ": "shch", "ъ": "", "ы": "y", "ь": "",
		"э": "e", "ю": "yu", "я": "ya",
	}

	result := strings.Builder{}
	for _, ch := range rus {
		lower := strings.ToLower(string(ch))
		if eng, ok := rules[lower]; ok {
			if ch >= 'А' && ch <= 'Я' || ch == 'Ё' {
				result.WriteString(strings.Title(eng))
			} else {
				result.WriteString(eng)
			}
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func main() {
	fmt.Println(translit("Вкусняшка для Кли")) // Vkusnyashka dlya Kli
}
```
