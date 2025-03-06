package main

import (
	"fmt"
	"log"

	"github.com/tarm/serial"
	"github.com/tshelter/ch9329/internal"
)

func main() {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	keyboard := &ch.Keyboard{Ser: *s}

	text := "echo 'Hello world'\n"
	minInterval := 0.01
	maxInterval := 0.03

	err = keyboard.Write(text, minInterval, maxInterval)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Typed 'HelloWorld!' successfully")
}
