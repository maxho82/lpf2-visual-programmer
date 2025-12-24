package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

// LPF2Parser парсер протокола LPF2/WeDo 2.0
type LPF2Parser struct{}

// SetSensorModeCommand команда установки режима датчика
type SetSensorModeCommand struct {
	PortID byte
	Mode   byte // 0x00 = absolute, 0x01 = discrete
}

// EncodeSetSensorMode кодирует команду установки режима датчика (для RGB‑светодиода)
func (p *LPF2Parser) EncodeSetSensorMode(cmd SetSensorModeCommand) ([]byte, error) {
	// Формат: [длина] [тип команды] [порт] [режим]
	data := []byte{
		0x0b,                         // Длина пакета (11 байт)
		0x01,                         // Команда: установка режима
		0x02,                         // Подкоманда: изменение режима
		cmd.PortID,                   // Порт (6 для встроенного светодиода)
		0x17,                         // Тип устройства: RGB (0x17)
		cmd.Mode,                     // Режим: 0x00 = absolute, 0x01 = discrete
		0x01, 0x00, 0x00, 0x00, 0x01, // Зарезервированные байты
	}
	return data, nil
}

// EncodeLEDCommand кодирует команду для светодиода (корректный формат LPF2)
func (p *LPF2Parser) EncodeLEDCommand(cmd LEDCommand) ([]byte, error) {
	// Формат: [длина] [тип команды] [порт] [режим] [R] [G] [B]
	data := []byte{
		0x06,       // Длина пакета (6 байт)
		0x06,       // Команда: вывод на порт
		cmd.PortID, // Порт (6)
		0x03,       // Режим: RGB (0x03)
		cmd.Red,    // Красный
		cmd.Green,  // Зеленый
		cmd.Blue,   // Синий
	}
	return data, nil
}

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
	powerScaled := int8(float64(cmd.Power) * 1.27)
	if powerScaled < -127 {
		powerScaled = -127
	}
	// Убрано условие > 127, т.к. int8 не может быть больше 127
	data[4] = byte(powerScaled)

	return data, nil
}

// LEDCommand команда для светодиода
type LEDCommand struct {
	PortID byte
	Red    byte
	Green  byte
	Blue   byte
}
