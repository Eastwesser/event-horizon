# SLICE_QUESTION_3

Продолжение slice-квеста: удаление элемента, `copy` vs append-трюк, утечка памяти при `s = append(s[:i], s[i+1:]...)`.

Скажи на собесе: «для больших объектов обнуляй `s[len-1]` перед trunc»; для int — достаточно re-slice.

Практика compact/filter: [`../code/03_remove_zeros`](../code/03_remove_zeros).  
Теория: [`../go/SLICE.md`](../go/SLICE.md).
