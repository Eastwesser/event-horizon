# SLICE_QUESTION_1

Типичный screening про **header слайса**: len/cap, общий underlying array.

Вопросы в духе:
- что напечатает `a := []int{1,2,3}; b := a[:2]; b[0]=9; fmt.Println(a)`?
- чем `append` в переполненный cap отличается от append с местом?

Готовь устно: «слайс — ptr/len/cap; reslice шарит массив, пока не сработает grow».

Теория: [`../go/SLICE.md`](../go/SLICE.md). Отдельного runnable pack нет — смотри PRINTABLE / dojo.
