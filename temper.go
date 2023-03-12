package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

func main() {
	tempChan := make(chan float32)
	temperW, err := os.OpenFile("/dev/hidraw18",
		os.O_APPEND|os.O_WRONLY, os.ModeDevice)
	if err != nil {
		panic(err)
	}
	defer temperW.Close()

	temper, err := os.Open("/dev/hidraw18")
	if err != nil {
		panic(err)
	}
	defer temper.Close()

	go func() {
		response := make([]byte, 8)
		_, err = temper.Read(response)
		if err != nil {
			panic(err)
		}
		hexStr := hex.EncodeToString(response)
		if err != nil {
			panic(err)
		}
		temp := hexStr[4:8]
		tempInt, err := strconv.ParseInt(temp, 16, 64)
		if err != nil {
			panic(err)
		}
		float := float32(tempInt) / 100
		tempChan <- float
	}()
	//	// magic byte sequence to request a temperature reading
	_, err = temperW.Write([]byte{0, 1, 128, 51, 1, 0, 0, 0, 0})
	if err != nil {
		panic(err)
	}
	temperature := <-tempChan
	fmt.Printf("temp: %f\n", temperature)
}
