package main

import "fmt"

const MonetizationAccess uint32 = 1 << 3

func main() {
	var accessLevel uint32 = MonetizationAccess | (1 << 1)
	ban := true
	if ban {
		accessLevel &^= MonetizationAccess // сброс бита
	} else {
		accessLevel |= MonetizationAccess
	}
	fmt.Printf("%b\n", accessLevel)
	accessLevel |= MonetizationAccess
	fmt.Printf("%b\n", accessLevel)
}
