package main

import (
	"fmt"
	"github.com/koyeo/nest/test/second"
	"github.com/koyeo/nest/test/second/third"
	"time"
)

func main() {
	fmt.Println("Hello world3")
	second.Second()
	third.Third()
	test := make(chan string)

	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(i)
			test <- fmt.Sprintf("test-%d", i)
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			select {
			case m := <-test:
				fmt.Println(m)
			}
		}
	}()




	time.Sleep(10 * time.Second)

}
