Напишите функцию, которая получает на вход map вида map[string]int 
и возвращает новую map инвертированного, где ключи и значения исходной map меняются местами
// Пример
input := map[string]int{
    "a": 1,
    "b": 2,
    "c": 1,
}
output := InvertMap(input)
// output: map[int][]string{1: {"a","c"}, 2: {"b"}}


func InvertMap(input map[string]int) map[int][]string {
invert := make(map[int][]string, len(input)) //
for k, v :=range input{
    invert[v] = append(invert[v], k) //
}
    
return invert    
}

------------------------------------
//Хорошая задача...
Напишите функцию UnionAllInts, которая объединяет произвольное число срезов []int, убирая дубликаты и 
сохраняя порядок первого появления.
// UnionAllInts(nil, []int{1,1,2})                    // -> []int{1,2}
func UnionAllInts(slices ...[]int) []int {
    // ваш код
    for _, s ;= range slices {
        
    }
    
    
    res := make([]int, 0) // можно оптимизировать
    
    uniqMap := make(map[int]struct{})
    for _, slice := range slices {
        if slice == nil { // 
            continue
        }
        for _, v := range slice {
          if  _, ok := uniqMap[v]; ok {
              continue
          } else {
              res := append(res, v)
          }
          
        }
    }
    return res
}

/*
res := []int
for s range slices{
    totalLen += 
}


*/
 
// Пример
UnionAllInts([]int{1,2,3}, []int{3,4}, []int{2,5}) // -> []int{1,2,3,4,5}
UnionAllInts()                                     // -> []int{}
UnionAllInts(nil, []int{1,1,2})                    // -> []int{1,2}







--------
//var res strings.Builder
//res.WriteString(fmt.Sprintf(),...)
Написать реализацию функции по сжатию строки
func compactString(s sting) string {
    slice:=[]byte(s)
    res:=make([]byte)
    
    i:=0
    var vtemp byte // А чему оно равно перед входом в цикл?
    for _,v:=range slice{
       if vtemp!=v{
           if i!=0{
               st:=srtconv.String(i)
               res=append(res,st)
           }
           res=append(res,v)
           vtemp=v
                      
       }else{
           i++
       }
    }
    
    return string(res)
    // Реализация
}  

///
//var res strings.Builder
//res.WriteString(fmt.Sprintf(%c,%d),...)
func compactString(s sting) string {
    var builder strings.Builder
    
    //for i := 0; i < len(s); i++ { if i < len(i) - 1 && s[i] == s[i+1] //выскочим за пределы слайса, может лучше итерироваться от i=1, а нулевой i-1. Нужно делать такб чтобы код был читаемым. 0 потом -1, сложно быстро понять как это работает? а как я нулевой индекс строки возьму тогда?  {cnt++} else {builder..WriteString(fmt.Sprintf(%c,%d), s[i], cnt) cnt=1 }  } // Это байты
    var prev byte 
    var currIdx int // 4
    
    for _, currB := range []byte(s) {
        currIdx++
        
        if currIdx == 1 {
           prev = currB
        }
    
        if perv != currB {
            builder.WriteString(fmt.Sprintf("%s%d", prev, currIdx))
            currIdx = 0
        }
    }
    
    // нужно не потерять последнюю последовательность
    
    
    return builder.String()
}
var cnt int
for i := 0; i < len(s); i++{ //Может len(s-1)?
    var builder strings.Builder
    
    if i < len(i) - 1 && s[i] == s[i+1] {
        cnt++
    } else {
        builder..WriteString(fmt.Sprintf(%c%d), s[i], cnt)
        cnt=1 
    }
    
    return builder.String()
}

 
func main() {
    fmt.Println(compactString("hhhhgggrreeee")) // h4g3r2e4
    fmt.Println(compactString("zzzzz"))         // z5
    //fmt.Println(compactString("hhhhgggrreeeehhhh")) //h4g3r2e4h4  
 
}



---------

// ЗАДАЧА 17: Group (аналог sync.WaitGroup)
// Реализуйте методы у структуры Group (аналог sync.WaitGroup), чтобы код не приводил к панике

package main

import (
    "reflect"
    "sort"
    "sync"
)

type Group struct {
    c    chan struct{}
    size int
}

func New(size int) *Group {    
    return &Group{
        c: make(chan struct{}, size),
        size: size,
    }
}

func (s *Group) Done() {
    s.c <- struct{}{} 
}
 
func (s *Group) Wait() {
   for i := s.size; i > 0 ; i-- {
       <- s.c
   }
}

func main() {
    numbers := []int{1, 2, 3, 4, 5}
    n := len(numbers)

    var res []int
    var mu sync.Mutex

    group := New(n)

    for _, num := range numbers {
        // А здесь не нужна функция Add? Если Add нет, то канал нужно предзаполнять
        go func(num int) {
            defer group.Done()

            mu.Lock()
            res = append(res, num)
            mu.Unlock()
        }(num)
    }

    group.Wait()

    sort.IntSlice(res).Sort()

    if !reflect.DeepEqual(res, numbers) {
        panic("wrong code")
    }
}













