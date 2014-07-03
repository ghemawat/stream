package pipe

import (
	"sync"
)

type parItem struct {
	index int
	value string
}

// ParallelMap calls fn(x) for every item x in a pool of n
// goroutines and yields the outputs of the fn calls. The output order
// matches the input order.
func ParallelMap(n int, fn func(string) string) Filter {
	return func(arg Arg) {
		// Attach a sequence number to each item.
		source := make(chan parItem, 10000)
		go func() {
			i := 0
			for s := range arg.In {
				source <- parItem{i, s}
				i++
			}
			close(source)
		}()

		// We keep track of outputs in a map indexed by the
		// sequence number of the item.  These items are
		// yielded in order.
		var mu sync.Mutex
		buffered := make(map[int]string)
		next := 0
		nextItemReady := func() bool {
			_, ok := buffered[next]
			return ok
		}

		// Process the items in n go routines.
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				for item := range source {
					s := fn(item.value)
					mu.Lock()
					buffered[item.index] = s
					for nextItemReady() {
						arg.Out <- buffered[next]
						delete(buffered, next)
						next++
					}
					mu.Unlock()
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
