package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

// LPF2Parser парсер протокола LPF2/WeDo 2.0
type LPF2Parser struct{}

// ParsePortNotification парсит уведомление о порте
func (p *LPF2Parser) ParsePortNotification(data []byte) (PortNotification, error) {
	if len(data) < 12 {
		return PortNotification{}, fmt.Errorf("недостаточно данных для парсинга порта")
	}

	notification := PortNotification{
		MessageType: data[0],
		PortID:      data[2],
		DeviceType:  data[3],
		Data:        data,
	}

	// Парсинг значения в зависимости от типа устройства
	switch notification.DeviceType {
	case 0x01: // Мотор
		notification.Value = int(data[4])
		notification.Unit = "мощность"
	case 0x02: // Датчик наклона
		notification.Value = int(data[4])
		notification.Unit = "градус"
	case 0x08: // Светодиод
		notification.Value = int(data[4])
		notification.Unit = "яркость"
	case 0x14, 0x15, 0x16, 0x17: // Датчик расстояния
		notification.Value = int(data[4])
		notification.Unit = "см"
	default:
		// Пытаемся парсить как 32-битное целое
		if len(data) >= 8 {
			notification.Value = int(binary.LittleEndian.Uint32(data[4:8]))
		}
	}

	return notification, nil
}

// ParseSensorValue парсит значение датчика
func (p *LPF2Parser) ParseSensorValue(data []byte) (SensorValue, error) {
	if len(data) < 8 {
		return SensorValue{}, fmt.Errorf("недостаточно данных для парсинга датчика")
	}

	value := SensorValue{
		PortID:  data[1],
		Mode:    data[2],
		RawData: data,
	}

	// Парсинг в зависимости от режима
	switch value.Mode {
	case 0x00: // RAW
		if len(data) >= 8 {
			value.NumericValue = float64(binary.LittleEndian.Uint32(data[4:8]))
			value.DataType = "raw"
		}
	case 0x01: // Процент
		if len(data) >= 8 {
			value.NumericValue = float64(binary.LittleEndian.Uint32(data[4:8]))
			value.DataType = "percent"
			value.Unit = "%"
		}
	case 0x02: // SI
		if len(data) >= 8 {
			// Парсинг как float32
			bits := binary.LittleEndian.Uint32(data[4:8])
			value.NumericValue = float64(math.Float32frombits(bits))
			value.DataType = "si"
		}
	}

	return value, nil
}

// PortNotification структура уведомления о порте
type PortNotification struct {
	MessageType byte
	PortID      byte
	DeviceType  byte
	Value       int
	Unit        string
	Data        []byte
}

// SensorValue значение датчика
type SensorValue struct {
	PortID       byte
	Mode         byte
	NumericValue float64
	StringValue  string
	DataType     string
	Unit         string
	RawData      []byte
}

// MotorCommand команда для мотора
type MotorCommand struct {
	PortID   byte
	Power    int8   // -100..100
	Duration uint16 // в миллисекундах, 0 = бесконечно
}

// EncodeMotorCommand кодирует команду для мотора
func (p *LPF2Parser) EncodeMotorCommand(cmd MotorCommand) ([]byte, error) {
	// Формат команды для мотора WeDo 2.0: 04 01 [port] 01 [power]
	data := make([]byte, 5)

	// Заголовок команды
	data[0] = 0x04 // Длина команды
	data[1] = 0x01 // Команда: мотор
	data[2] = cmd.PortID

	// Режим: абсолютная мощность
	data[3] = 0x01

	// Мощность: преобразуем -100..100 в -127..127
	power := int8(float64(cmd.Power) * 1.27)
	if power < -127 {
		power = -127
	} else if power > 127 {
		power = 127
	}
	data[4] = byte(power)

	return data, nil
}

// LEDCommand команда для светодиода
type LEDCommand struct {
	PortID byte
	Red    byte
	Green  byte
	Blue   byte
}

// EncodeLEDCommand кодирует команду для светодиода
func (p *LPF2Parser) EncodeLEDCommand(cmd LEDCommand) ([]byte, error) {
	// Формат команды для RGB светодиода: 06 06 [port] 03 00 [R] [G] [B]
	data := make([]byte, 8)

	data[0] = 0x06 // Длина команды
	data[1] = 0x06 // Команда: светодиод
	data[2] = cmd.PortID
	data[3] = 0x03 // Режим: RGB
	data[4] = 0x00 // Reserved
	data[5] = cmd.Red
	data[6] = cmd.Green
	data[7] = cmd.Blue

	return data, nil
}

// SetSensorModeCommand команда установки режима датчика
type SetSensorModeCommand struct {
	PortID byte
	Mode   byte
}

// EncodeSetSensorMode кодирует команду установки режима датчика
func (p *LPF2Parser) EncodeSetSensorMode(cmd SetSensorModeCommand) ([]byte, error) {
	// Формат: зависит от типа датчика
	// Для простоты используем стандартный формат WeDo 2.0
	data := make([]byte, 4)

	data[0] = 0x01 // Subcommand
	data[1] = cmd.PortID
	data[2] = 0x22 // Mode change
	data[3] = cmd.Mode

	return data, nil
}
