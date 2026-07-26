package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/taigrr/temper"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	tempers, err := temper.FindTempersWithContext(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		for _, dev := range tempers {
			dev.Close()
		}
	}()

	if len(tempers) == 0 {
		fmt.Println("no TEMPer devices found")
		return
	}

	for {
		for _, dev := range tempers {
			readCtx, cancel := context.WithTimeout(ctx, time.Second)

			celsius, cErr := dev.ReadCWithContext(readCtx)
			cancel()
			if cErr != nil {
				log.Println(cErr)
				continue
			}

			fahrenheit := celsius*9.0/5.0 + 32.0

			fmt.Printf("Read from %s: F: %.2f C: %.2f\n", dev.Descriptor(), fahrenheit, celsius)
		}

		select {
		case <-ctx.Done():
			fmt.Println("shutting down")
			return
		case <-time.After(time.Second):
		}
	}
}
