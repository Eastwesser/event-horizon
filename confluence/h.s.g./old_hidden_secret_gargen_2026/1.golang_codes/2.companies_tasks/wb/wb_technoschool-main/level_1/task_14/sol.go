package main

import (
	"fmt"
	"reflect"
)

func detectType(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		if reflect.TypeOf(v).Kind() == reflect.Chan {
			return "chan"
		}
		return "unknown"
	}
}

func main() {
	fmt.Println(detectType(10))
	fmt.Println(detectType("go"))
	fmt.Println(detectType(true))
	ch := make(chan int)
	fmt.Println(detectType(ch))
}
