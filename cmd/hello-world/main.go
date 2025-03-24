package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
	ch "github.com/tshelter/ch9329"
)

func main() {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600, ReadTimeout: time.Millisecond * 500}
	port, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer func(port *serial.Port) {
		if err := port.Close(); err != nil {
			log.Fatal(err)
		}
	}(port)

	keyboard := ch.NewKeyboardSender(port)
	text := "echo 'Hello world'\n"
	minInterval := 0.01
	maxInterval := 0.03

	if err := keyboard.Write(text, minInterval, maxInterval); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Typed 'Hello World!' successfully")
}
