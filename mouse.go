package ch9329

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/tarm/serial"
)

type MouseCtrl string

const (
	CtrlNull   MouseCtrl = "null"
	CtrlLeft   MouseCtrl = "left"
	CtrlRight  MouseCtrl = "right"
	CtrlCenter MouseCtrl = "center"
)

var ctrlToHexMapping = map[MouseCtrl]byte{
	CtrlNull:   0x00,
	CtrlLeft:   0x01,
	CtrlRight:  0x02,
	CtrlCenter: 0x04,
}

var (
	MouseHead = []byte{0x57, 0xAB}
	MouseAddr = byte(0x00)
	CmdAbs    = byte(0x04)
	CmdRel    = byte(0x05)
	LenAbs    = byte(0x07)
	LenRel    = byte(0x05)
)

type MouseSender struct {
	ser *serial.Port
}

func NewMouseSender(ser *serial.Port) *MouseSender {
	return &MouseSender{ser: ser}
}

func wheelIntToBytes(wheelDelta int) (byte, error) {
	if math.Abs(float64(wheelDelta)) > 127 {
		return 0, fmt.Errorf("error maximum wheel delta allowed is 127")
	}
	if wheelDelta >= 0 {
		return byte(wheelDelta), nil
	}
	return byte(256 + wheelDelta), nil
}

func (ms *MouseSender) sendDataAbsolute(
	x, y int,
	ctrl MouseCtrl,
	xMax, yMax, wheelDelta int,
) error {
	data := []byte{0x02, ctrlToHexMapping[ctrl]}

	xCur := (4096 * x) / xMax
	yCur := (4096 * y) / yMax

	xBytes := make([]byte, 2)
	yBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(xBytes, uint16(xCur))
	binary.LittleEndian.PutUint16(yBytes, uint16(yCur))

	b, err := wheelIntToBytes(wheelDelta)
	if err != nil {
		return err
	}

	data = append(data, xBytes...)
	data = append(data, yBytes...)
	data = append(data, b)

	packet := append(MouseHead, MouseAddr, CmdAbs, LenAbs)
	packet = append(packet, data...)
	if _, err := ms.ser.Write(packet); err != nil {
		return err
	}
	return nil
}

func (ms *MouseSender) sendDataRelative(x, y int, ctrl MouseCtrl, wheelDelta int) error {
	data := []byte{0x01, ctrlToHexMapping[ctrl]}

	b, err := wheelIntToBytes(wheelDelta)
	if err != nil {
		return err
	}

	data = append(data, byte(x), byte(y), b)

	packet := append(MouseHead, MouseAddr, CmdRel, LenRel)
	packet = append(packet, data...)
	if _, err := ms.ser.Write(packet); err != nil {
		return err
	}
	return nil
}

func (ms *MouseSender) Move(x, y int, relative bool, monitorWidth, monitorHeight int) error {
	if relative {
		if err := ms.sendDataRelative(x, y, CtrlNull, 0); err != nil {
			return err
		}
	} else {
		if err := ms.sendDataAbsolute(x, y, CtrlNull, monitorWidth, monitorHeight, 0); err != nil {
			return err
		}
	}
	return nil
}

func (ms *MouseSender) Press(button MouseCtrl) error {
	return ms.sendDataRelative(0, 0, button, 0)
}

func (ms *MouseSender) Release() error {
	return ms.sendDataRelative(0, 0, CtrlNull, 0)
}

func (ms *MouseSender) Click(button MouseCtrl) error {
	if err := ms.Press(button); err != nil {
		return err
	}
	delay := time.Duration(rand.Intn(350)+100) * time.Millisecond
	time.Sleep(delay)
	return ms.Release()
}

func (ms *MouseSender) Wheel(wheelDelta int) error {
	return ms.sendDataRelative(0, 0, CtrlNull, wheelDelta)
}
