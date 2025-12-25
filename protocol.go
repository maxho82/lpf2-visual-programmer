package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

// LPF2Parser парсер протокола LPF2/WeDo 2.0
type LPF2Parser struct{}

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
	case 0x08, 0x17: // Светодиод (0x08 - простой, 0x17 - RGB)
		notification.Value = int(data[4])
		notification.Unit = "яркость"
	case 0x14, 0x15, 0x16: // Датчик расстояния - убрали 0x17
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
	// Формат команды для мотора WeDo 2.0: 06 04 [port] 01 [power] 00
	data := make([]byte, 6)

	// Заголовок команды
	data[0] = 0x06 // Длина команды (6 байт)
	data[1] = 0x04 // Команда: вывод на порт (Output Command)
	data[2] = cmd.PortID
	data[3] = 0x01 // Режим: абсолютная мощность

	// Мощность: преобразуем -100..100 в -127..127
	powerScaled := int8(float64(cmd.Power) * 1.27)
	if powerScaled < -127 {
		powerScaled = -127
	}
	data[4] = byte(powerScaled)
	data[5] = 0x00 // Reserved

	return data, nil
}

// LEDCommand команда для светодиода
type LEDCommand struct {
	PortID byte
	Red    byte
	Green  byte
	Blue   byte
}

// EncodeLEDCommand кодирует команду для светодиода (WeDo 2.0/Boost формат)
func (p *LPF2Parser) EncodeLEDCommand(cmd LEDCommand) ([]byte, error) {
	// Формат команды для RGB светодиода: 08 04 [port] 03 [R] [G] [B] 00
	data := []byte{
		0x08,       // Длина пакета (8 байт)
		0x04,       // Команда: вывод на порт
		cmd.PortID, // Порт (6)
		0x03,       // Режим: RGB (0x03)
		cmd.Red,    // Красный
		cmd.Green,  // Зеленый
		cmd.Blue,   // Синий
		0x00,       // Reserved
	}
	return data, nil
}

// PortModeCommand команда установки режима порта
type PortModeCommand struct {
	PortID     byte
	DeviceType byte // 0x17 для RGB светодиода
	Mode       byte // 0x00 = absolute, 0x01 = relative
}

// EncodePortModeCommand кодирует команду установки режима порта
func (p *LPF2Parser) EncodePortModeCommand(cmd PortModeCommand) ([]byte, error) {
	// Формат: 0B 01 02 [port] [deviceType] [mode] 01 00 00 00 01
	data := []byte{
		0x0B,                         // Длина пакета (11 байт)
		0x01,                         // Команда: установка режима
		0x02,                         // Подкоманда: изменение режима
		cmd.PortID,                   // Порт
		cmd.DeviceType,               // Тип устройства
		cmd.Mode,                     // Режим: 0x00 = absolute
		0x01, 0x00, 0x00, 0x00, 0x01, // Зарезервированные байты
	}
	return data, nil
}
