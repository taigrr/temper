package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/taigrr/temper"
)

func findTempers() ([]*temper.Temper, error) {
	tempers := []*temper.Temper{}
	// list over dev folder for temperXX devices
	dirEnts, err := os.ReadDir("/dev")
	if err != nil {
		return tempers, err
	}
	for _, d := range dirEnts {
		if name := d.Name(); strings.HasPrefix(name, "temper") {
			temper, err := temper.New(filepath.Join("/dev", name))
			if err != nil {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*250)
			_, err = temper.ReadCWithContext(ctx)
			if err == nil {
				tempers = append(tempers, temper)
			} else {
				temper.Close()
			}
			cancel()

		}
	}

	return tempers, nil
}

func main() {
	tempers, err := findTempers()
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
		time.Sleep(time.Second)
	}
}
