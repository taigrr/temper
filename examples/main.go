package main

import (
	"fmt"
	"log"
	"time"

	"github.com/taigrr/temper"
)

func main() {
	tempers, err := temper.FindTempers()
	if err != nil {
		panic(err)
	}
	for {
		for _, temperDev := range tempers {
			f, fErr := temperDev.ReadF()
			if fErr != nil {
				log.Println(fErr)
				time.Sleep(time.Second)
				continue
			}
			c, cErr := temperDev.ReadC()
			if cErr != nil {
				log.Println(cErr)
				time.Sleep(time.Second)
				continue
			}
			fmt.Printf("Read from %s: F: %f C: %f\n", temperDev.Descriptor(), f, c)
		}
	}
}
