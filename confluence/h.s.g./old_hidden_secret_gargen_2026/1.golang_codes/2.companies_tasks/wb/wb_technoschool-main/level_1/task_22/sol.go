package main

import (
	"fmt"
	"math/big"
)

func main() {
	a := new(big.Int).Exp(big.NewInt(2), big.NewInt(50), nil)
	b := new(big.Int).Exp(big.NewInt(2), big.NewInt(45), nil)

	sum := new(big.Int).Add(a, b)
	diff := new(big.Int).Sub(a, b)
	prod := new(big.Int).Mul(a, b)
	quot := new(big.Int).Div(a, b)

	fmt.Println(sum)
	fmt.Println(diff)
	fmt.Println(prod)
	fmt.Println(quot)
}
