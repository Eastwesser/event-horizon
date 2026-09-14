package main

import "fmt"

// setBit устанавливает i-й бит числа в указанное значение (0 или 1)
func setBit(num int64, position int, value int) int64 {
	if value == 1 {
		// Установка бита в 1 через операцию ИЛИ
		return num | (1 << position)
	} else {
		// Установка бита в 0 через операцию И НЕ
		return num &^ (1 << position)
	}
}

// getBit возвращает значение i-го бита
func getBit(num int64, position int) int {
	return int((num >> position) & 1)
}

func main() {
	var num int64 = 11 // 0101₂

	fmt.Printf("Исходное число: %d (двоичное: %b)\n", num, num)

	// Пример из задания: установка 1-го бита в 0
	result := setBit(num, 1, 0)
	fmt.Printf("После установки 1-го бита в 0: %d (двоичное: %b)\n", result, result)

	// Дополнительные примеры
	fmt.Printf("Установка 0-го бита в 0: %d -> %d\n", num, setBit(num, 0, 0))
	fmt.Printf("Установка 2-го бита в 1: %d -> %d\n", num, setBit(num, 2, 1))
	fmt.Printf("Установка 3-го бита в 1: %d -> %d\n", num, setBit(num, 3, 1))
}
