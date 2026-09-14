package task14

func Or(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		c := make(chan interface{})
		close(c)
		return c
	case 1:
		return channels[0]
	case 2:
		out := make(chan interface{})
		go func(a, b <-chan interface{}) {
			defer close(out)
			select {
			case <-a:
			case <-b:
			}
		}(channels[0], channels[1])
		return out
	default:
		mid := len(channels) / 2
		return Or(Or(channels[:mid]...), Or(channels[mid:]...))
	}
}

var or = Or
