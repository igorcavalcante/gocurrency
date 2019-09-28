package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	workers := 2

	done := make(chan struct{})
	generatedData := gen(done)

	var funOutData []chan int

	for i := 0; i < workers; i++ {
		funOutData = append(funOutData, worker(done, i, generatedData))
	}

	funInData := merge(done, funOutData)

	timer := time.NewTimer(5 * time.Second).C

	func() {
		for {
			select {
			case out, more := <-funInData:
				if more {
					log.Println("final output", out)
				} else {
					log.Println("Hasta la vista baby!")
					return
				}
			case <-timer:
				close(done)
			}
		}
	}()
}

func gen(done <-chan struct{}) chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		i := 0
		for {
			select {
			case <-done:
				log.Println("generator leaving")
				return
			case out <- i:
				i++
			}
		}
	}()

	return out
}

func worker(done <-chan struct{}, number int, c <-chan int) chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for i := range c {
			log.Println("ZZzzzzZZ worker:", number)
			time.Sleep(1 * time.Second)
			log.Println("WoW worker:", number)
			select {
			case <-done:
				break
			case out <- i * i:
				log.Printf("Msg sent worker %d, value %d", number, i*i)
			}
		}
		log.Println("leaving worker", number)
	}()

	return out

}

func merge(done <-chan struct{}, cs []chan int) chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	wg.Add(len(cs))

	funIn := func(c chan int) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case out <- i:
			}
		}
	}

	for _, c := range cs {
		go funIn(c)
	}

	go func() {
		wg.Wait()
		log.Println("Leaving merge")
		close(out)
	}()

	return out
}
