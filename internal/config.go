package ch

import (
	"encoding/binary"
	"errors"

	"github.com/tarm/serial"
)

type USBStringDescriptor byte

const (
	Manufacturer USBStringDescriptor = 0x00
	Product      USBStringDescriptor = 0x01
	SerialNumber USBStringDescriptor = 0x02
)

var (
	CfgHead             = []byte{0x57, 0xAB}
	CfgAddr             = byte(0x00)
	CmdGetParaCfg       = byte(0x08)
	LenGetParaCfg       = byte(0x00)
	CmdSetParaCfg       = byte(0x09)
	LenSetParaCfg       = byte(0x32)
	UsbStringEnableFlag = byte(0x87)
	CmdGetUsbString     = byte(0x0A)
	LenGetUsbString     = byte(0x01)
	CmdSetUsbString     = byte(0x0B)
)

type Configuration struct {
	Data []byte
}

type baseCfg struct {
	ser *serial.Port
}

func NewBaseCfg(ser *serial.Port) *baseCfg { return &baseCfg{ser: ser} }

func (b *baseCfg) setDeviceDescriptors(
	descriptorType USBStringDescriptor,
	description string,
) error {
	if len(description) > 23 {
		return errors.New("length of description should not be more than 23")
	}
	descriptionBytes := []byte(description)
	packet := append(CfgHead, CfgAddr, CmdGetUsbString, LenGetUsbString, byte(descriptorType))
	b.ser.Write(packet)
	b.ser.Read(make([]byte, 128))
	descriptorLength := len(descriptionBytes)
	if descriptorLength == 0 {
		descriptorLength = 1
	}

	modifiedData := append(
		[]byte{byte(descriptorType), byte(descriptorLength)},
		descriptionBytes...)
	modifiedPacket := append(CfgHead, CfgAddr, CmdSetUsbString, byte(len(modifiedData)))
	modifiedPacket = append(modifiedPacket, modifiedData...)
	b.ser.Write(modifiedPacket)

	returnPacket := make([]byte, 7)
	b.ser.Read(returnPacket)
	expectedPacket := []byte{0x57, 0xAB, 0x00, 0x8B, 0x01, 0x00, 0x8E}
	if string(returnPacket) != string(expectedPacket) {
		return errors.New("unexpected response from device")
	}
	return nil
}

func (b *baseCfg) getParameters() ([]byte, error) {
	packet := append(MouseHead, MouseAddr, CmdGetParaCfg, LenGetParaCfg)
	if _, err := b.ser.Write(packet); err != nil {
		return nil, err
	}
	buffer := make([]byte, 128)
	n, err := b.ser.Read(buffer)
	if err != nil || n == 0 {
		return nil, errors.New("no response received")
	}
	return buffer[:n], nil
}

func (b *baseCfg) getUSBString(descriptor USBStringDescriptor) (string, error) {
	if _, err := b.ser.Read(make([]byte, 128)); err != nil {
		return "", err
	}
	packet := append(MouseHead, MouseAddr, CmdGetUsbString, LenGetUsbString, byte(descriptor))
	if _, err := b.ser.Write(packet); err != nil {
		return "", err
	}

	buffer := make([]byte, 128)
	n, err := b.ser.Read(buffer)
	if err != nil || n < 7 {
		return "", errors.New("invalid response length")
	}
	length := buffer[6]
	return string(buffer[7 : 7+length]), nil
}

func (b *baseCfg) getSerialNumber() (string, error) {
	return b.getUSBString(SerialNumber)
}

func (b *baseCfg) getManufacturer() (string, error) {
	return b.getUSBString(Manufacturer)
}

func (b *baseCfg) getProduct() (string, error) {
	return b.getUSBString(Product)
}

func (b *baseCfg) setDeviceIDs(vid, pid int, customDescriptor bool) error {
	packet := append(MouseHead, MouseAddr, CmdGetParaCfg, LenGetParaCfg)
	if _, err := b.ser.Write(packet); err != nil {
		return err
	}
	receivedPacket := make([]byte, 128)
	if _, err := b.ser.Read(receivedPacket); err != nil {
		return err
	}

	vidBytes := make([]byte, 2)
	pidBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(vidBytes, uint16(vid))
	binary.LittleEndian.PutUint16(pidBytes, uint16(pid))

	modifiedData := append(receivedPacket[5:16], vidBytes...)
	modifiedData = append(modifiedData, pidBytes...)
	modifiedData = append(modifiedData, receivedPacket[20:55]...)

	if customDescriptor {
		modifiedData[35] = UsbStringEnableFlag
	}

	modifiedPacket := append(MouseHead, MouseAddr, CmdSetParaCfg, LenSetParaCfg)
	modifiedPacket = append(modifiedPacket, modifiedData...)
	if _, err := b.ser.Write(modifiedPacket); err != nil {
		return err
	}

	returnPacket := make([]byte, 7)
	if _, err := b.ser.Read(returnPacket); err != nil {
		return err
	}
	expectedPacket := []byte{0x57, 0xAB, 0x00, 0x89, 0x01, 0x00, 0x8C}
	if string(returnPacket) != string(expectedPacket) {
		return errors.New("unexpected response from device")
	}
	return nil
}
