package ch

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/tarm/serial"
)

var (
	Modif = []string{
		"ctrl",
		"ctrl_left",
		"ctrl_right",
		"shift",
		"shift_left",
		"shift_right",
		"alt",
		"alt_left",
		"alt_right",
		"win",
		"win_left",
		"win_right",
	}
	ModifMap = map[string]byte{
		"ctrl":        0b00000001,
		"ctrl_left":   0b00000001,
		"shift":       0b00000010,
		"shift_left":  0b00000010,
		"alt":         0b00000100,
		"alt_left":    0b00000100,
		"win":         0b00001000,
		"win_left":    0b00001000,
		"ctrl_right":  0b00010000,
		"shift_right": 0b00100000,
		"alt_right":   0b01000000,
		"win_right":   0b10000000,
	}
	Head = []byte{0x57, 0xAB}
	Addr = []byte{0x00}
	Cmd  = []byte{0x02}
	Len  = []byte{0x08}
)

type keyboardSender struct {
	ser *serial.Port
}

func NewKeyboardSender(ser *serial.Port) *keyboardSender {
	return &keyboardSender{ser: ser}
}

func sumBytes(data []byte) int {
	sum := 0
	for _, b := range data {
		sum += int(b)
	}
	return sum
}

func getPacket(head, addr, cmd, length, data []byte) []byte {
	headSum := sumBytes(head)
	dataSum := sumBytes(data)

	addrInt := int(addr[0])
	cmdInt := int(cmd[0])
	lengthInt := int(length[0])

	checksum := (headSum + addrInt + cmdInt + lengthInt + dataSum) % 256

	packet := append(
		append(append(append(append([]byte{}, head...), addr...), cmd...), length...),
		data...)
	packet = append(packet, byte(checksum))

	return packet
}

func (k *keyboardSender) Send(keys [6]string, modifs []string) error {
	var dat []byte
	var modifB byte = 0x00
	for _, m := range modifs {
		val, ok := ModifMap[m]
		if !ok {
			return fmt.Errorf("%w, modifier: %s", ErrInvalidModifer, m)
		}
		modifB |= val
	}
	dat = append(dat, modifB, 0x00)
	for _, key := range keys {
		mapping, ok := HIDKeyMap[key]
		if !ok {
			return fmt.Errorf("%w, key: %s", ErrInvalidKey, key)
		}
		dat = append(dat, mapping.Code)
	}
	packet := getPacket(Head, Addr, Cmd, Len, dat)
	_, err := k.ser.Write(packet)
	return err
}

func (k *keyboardSender) Press(key string, modifs []string) error {
	mapping, ok := HIDKeyMap[key]
	if !ok {
		return fmt.Errorf("%w, key: %s", ErrInvalidKey, key)
	}
	if mapping.Shift {
		newModifiers := make([]string, len(modifs))
		copy(newModifiers, modifs)
		newModifiers = append(newModifiers, "shift")
		modifs = newModifiers
	}
	var keys [6]string
	keys[0] = key
	return k.Send(keys, modifs)
}

func (k *keyboardSender) Release() error {
	var keys [6]string
	return k.Send(keys, nil)
}

func (k *keyboardSender) PressAndRelease(
	key string,
	modifs []string,
	minInterval, maxInterval float64,
) error {
	if err := k.Press(key, modifs); err != nil {
		return err
	}

	sleepDuration := minInterval + rand.Float64()*(maxInterval-minInterval)
	time.Sleep(time.Duration(sleepDuration * float64(time.Second)))

	return k.Release()
}

func (k *keyboardSender) TriggerKeys(keys []string, modifiers []string) error {
	keySet := make(map[string]struct{})
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	modSet := make(map[string]struct{})
	for _, m := range modifiers {
		modSet[m] = struct{}{}
	}

	uniqueKeys := make([]string, 0, len(keySet))
	for k := range keySet {
		if k != "" {
			uniqueKeys = append(uniqueKeys, k)
		}
	}
	uniqueModifiers := make([]string, 0, len(modSet))
	for m := range modSet {
		uniqueModifiers = append(uniqueModifiers, m)
	}

	if len(uniqueKeys) > 6 {
		return fmt.Errorf(
			"%w, error: %s",
			ErrTooManyKeys,
			"CH9329 supports maximum of 6 keys to be pressed at once.",
		)
	}
	if len(uniqueModifiers) > 8 {
		return fmt.Errorf(
			"%w, error: %s",
			ErrTooManyKeys,
			"CH9329 supports maximum of 8 control keys to be pressed at once.",
		)
	}

	for len(uniqueKeys) < 6 {
		uniqueKeys = append(uniqueKeys, "")
	}

	var keysArray [6]string
	for i := 0; i < 6; i++ {
		keysArray[i] = uniqueKeys[i]
	}

	return k.Send(keysArray, uniqueModifiers)
}

func (k *keyboardSender) Write(text string, minInterval, maxInterval float64) error {
	for _, ch := range text {
		if err := k.PressAndRelease(string(ch), nil, minInterval, maxInterval); err != nil {
			return err
		}
		sleepDuration := minInterval + rand.Float64()*(maxInterval-minInterval)
		time.Sleep(time.Duration(sleepDuration * float64(time.Second)))
	}
	return nil
}
