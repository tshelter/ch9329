package main

import (
	"log"
	"time"

	"github.com/tarm/serial"

	ch "github.com/tshelter/goch9329/internal"
)

func main() {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600, ReadTimeout: time.Millisecond * 500}
	port, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer func(port *serial.Port) {
		err := port.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(port)

	keyboard := ch.NewKeyboardSender(port)
	mouse := ch.NewMouseSender(port)
	baseCfg := ch.NewBaseCfg(port)

	_ = keyboard
	_ = mouse
	_ = baseCfg
}
