package main

import (
	"fmt"

	"github.com/taigrr/temper"
)

func main() {
	descriptor := "/dev/hidraw18"
	temperDev, _ := temper.New(descriptor)
	for {
		f, _ := temperDev.ReadF()
		c, _ := temperDev.ReadC()
		fmt.Printf("F: %f C: %f\n", f, c)
	}
}
