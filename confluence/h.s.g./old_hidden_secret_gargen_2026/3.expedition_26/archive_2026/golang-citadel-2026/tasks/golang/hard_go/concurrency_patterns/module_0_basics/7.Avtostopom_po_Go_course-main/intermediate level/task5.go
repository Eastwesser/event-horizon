// time.Sleep использовать нельзя. это будет не валидным ответом на собеседовании
// 1. Иногда приходят нули. В чем проблема? Исправь ее
// 2. Если функция bank_network_call выполняется 5 секунд, то за сколько выполнится balance()?

func balance() int {
	x := make(map[int]int, 1)
	var m sync.Mutex

	// call bank
	for i := 0; i < 5; i++ {
		i := i
		go func () {
			m.Lock()
			b := bank_network_call(i)

			x[i] = b
			m.Unlock()
		}()
	}

// Как-то считается сумма значений в мапе и возвращается
return sumOfMap
}
