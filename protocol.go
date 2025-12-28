package main

// LPF2Parser парсер протокола LPF2/WeDo 2.0
type LPF2Parser struct{}

// --- UUID характеристик (как в WeDo2.py) ---
const (
	SENSOR_VAL_UUID     = "00001560-1212-efde-1523-785feabcd123"
	PORT_INFO_UUID      = "00001527-1212-efde-1523-785feabcd123"
	INPUT_COMMAND_UUID  = "00001563-1212-efde-1523-785feabcd123" // Для настройки режима
	OUTPUT_COMMAND_UUID = "00001565-1212-efde-1523-785feabcd123" // Для отправки команд
	PORT_NOTIF_UUID     = "00001524-1212-efde-1523-785feabcd123"

	FIRMWARE_SERVICE_UUID = "00004f0e-1212-efde-1523-785feabcd123"
	FIRMWARE_CHAR_UUID    = "00004f01-1212-efde-1523-785feabcd123"
)

// --- Константы для режимов (как в WeDo2.py) ---
const (
	LED_ABSOLUTE_MODE = 0
	LED_DISCRETE_MODE = 1
)

// --- Команды для настройки режима устройства (Input Commands) ---

// EncodeLEDModeCommand кодирует команду установки режима светодиода
// Формат: [0x01, 0x02, port, 0x17, mode, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01]
func (p *LPF2Parser) EncodeLEDModeCommand(portID byte, mode byte) ([]byte, error) {
	data := []byte{
		0x01, 0x02, // Заголовок подкоманды
		portID,                             // Порт (6 для встроенного светодиода)
		0x17,                               // Тип устройства: RGB светодиод
		mode,                               // Режим: 0=absolute, 1=discrete
		0x01, 0x00, 0x00, 0x00, 0x02, 0x01, // Фиксированные байты
	}
	return data, nil
}

// --- Команды для управления устройствами (Output Commands) ---

// LEDCommand команда для светодиода
type LEDCommand struct {
	PortID byte
	Red    byte
	Green  byte
	Blue   byte
}

// EncodeLEDCommand кодирует команду цвета для светодиода (дискретный режим)
// Формат для RGB: [0x06, 0x04, 0x03, R, G, B]
func (p *LPF2Parser) EncodeLEDCommand(cmd LEDCommand) ([]byte, error) {
	data := []byte{
		0x06,      // Длина пакета (6 байт)
		0x04,      // Команда: вывод на порт
		0x03,      // Режим: RGB (0x03)
		cmd.Red,   // Красный
		cmd.Green, // Зеленый
		cmd.Blue,  // Синий
	}
	return data, nil
}

// EncodeLEDIndexCommand кодирует команду для светодиода (абсолютный режим - индекс цвета)
// Формат: [0x06, 0x04, 0x01, color_index]
func (p *LPF2Parser) EncodeLEDIndexCommand(portID byte, colorIndex byte) ([]byte, error) {
	data := []byte{
		0x06,       // Длина пакета (6 байт)
		0x04,       // Команда: вывод на порт
		0x01,       // Режим: индекс цвета (0x01)
		colorIndex, // Индекс цвета (0x01 = розовый, 0x02 = фиолетовый и т.д.)
	}
	return data, nil
}

// MotorCommand команда для мотора
type MotorCommand struct {
	PortID   byte
	Power    int8 // -100..100
	Duration uint16
}

// EncodeMotorCommand кодирует команду для мотора (используем перевод скорости из WeDo2.py)
func (p *LPF2Parser) EncodeMotorCommand(cmd MotorCommand) ([]byte, error) {
	// Преобразуем мощность (-100..100) в значение байта
	var speedByte byte
	powerFloat := float64(cmd.Power) / 100.0 // Нормализуем к -1.0..1.0

	// Логика перевода из WeDo2.py: translate_speed
	if powerFloat < 0 {
		// Для отрицательной скорости: (0x54 * speed) + 0xF0
		speedByte = byte(int(0x54*powerFloat) + 0xF0)
	} else if powerFloat > 0 {
		// Для положительной скорости: (0x54 * speed) + 0x10
		speedByte = byte(int(0x54*powerFloat) + 0x10)
	} else {
		// Стоп
		speedByte = 0x00
	}

	data := []byte{
		cmd.PortID, // Порт
		0x01,       // Подкоманда?
		0x01,       // ???
		speedByte,  // Скорость
	}
	return data, nil
}

// EncodePortModeRequest кодирует запрос режима порта
func (p *LPF2Parser) EncodePortModeRequest(portID byte, mode byte) ([]byte, error) {
	// Формат: [0x01, 0x00, portID, mode]
	return []byte{0x01, 0x00, portID, mode}, nil
}

// EncodePortValueRequest кодирует запрос значения порта
func (p *LPF2Parser) EncodePortValueRequest(portID byte) ([]byte, error) {
	// Формат: [0x00, 0x21, portID]
	return []byte{0x00, 0x21, portID}, nil
}
