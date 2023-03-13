package main

import (
	"fmt"
	"log"
	"time"

	"github.com/taigrr/temper"
)

func main() {
	// this is an example hidraw file descriptor, but if the udev rules
	// work on your system, you can use /dev/temper
	descriptor := "/dev/hidraw18"
	temperDev, err := temper.New(descriptor)
	if err != nil {
		panic(err)
	}
	for {
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
		fmt.Printf("F: %f C: %f\n", f, c)
		time.Sleep(time.Second)
	}
}
