package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// DeviceManager управляет устройствами хаба
type DeviceManager struct {
	hubMgr          *HubManager
	parser          *LPF2Parser
	devices         map[byte]*Device
	devicesMu       sync.RWMutex
	onDeviceChanged func(portID byte, device *Device) // Callback при изменении устройства
}

// SetDeviceChangedCallback устанавливает callback для уведомлений об изменениях
func (dm *DeviceManager) SetDeviceChangedCallback(callback func(portID byte, device *Device)) {
	dm.devicesMu.Lock()
	defer dm.devicesMu.Unlock()
	dm.onDeviceChanged = callback
}

// notifyDeviceChanged уведомляет об изменении устройства
func (dm *DeviceManager) notifyDeviceChanged(portID byte, device *Device) {
	dm.devicesMu.RLock()
	callback := dm.onDeviceChanged
	dm.devicesMu.RUnlock()

	if callback != nil {
		callback(portID, device)
	}
}

// Device представляет подключенное устройство
type Device struct {
	PortID      byte
	DeviceType  byte
	Name        string
	IsConnected bool
	LastValue   interface{}
	LastUpdate  time.Time
	Properties  map[string]interface{}
}

// NewDeviceManager создает менеджер устройств
func NewDeviceManager(hubMgr *HubManager) *DeviceManager {
	return &DeviceManager{
		hubMgr:  hubMgr,
		parser:  &LPF2Parser{},
		devices: make(map[byte]*Device),
	}
}

// UpdateDevices обновляет информацию об устройствах
func (dm *DeviceManager) UpdateDevices(portInfos []PortInfo) {
	dm.devicesMu.Lock()
	defer dm.devicesMu.Unlock()

	// Обновляем или создаем устройства
	for _, portInfo := range portInfos {
		if portInfo.PortID == 0 {
			continue
		}

		device, exists := dm.devices[portInfo.PortID]
		if !exists {
			device = &Device{
				PortID:     portInfo.PortID,
				Properties: make(map[string]interface{}),
			}
			dm.devices[portInfo.PortID] = device
		}

		device.DeviceType = portInfo.DeviceType
		device.Name = dm.getDeviceName(portInfo.DeviceType)
		device.IsConnected = portInfo.IsConnected
		device.LastUpdate = time.Now()

		// Сохраняем последнее значение
		if len(portInfo.LastValue) > 0 {
			device.LastValue = portInfo.LastValue

			// Записываем значения в свойства в зависимости от типа устройства
			switch portInfo.DeviceType {
			case 0x01: // Мотор
				if len(portInfo.LastValue) >= 1 {
					device.Properties["power"] = portInfo.LastValue[0]
				}
			case 0x17: // RGB светодиод
				if len(portInfo.LastValue) >= 3 {
					device.Properties["red"] = portInfo.LastValue[0]
					device.Properties["green"] = portInfo.LastValue[1]
					device.Properties["blue"] = portInfo.LastValue[2]
				}
			case 0x02: // Датчик наклона
				if len(portInfo.LastValue) >= 1 {
					device.Properties["angle"] = portInfo.LastValue[0]
				}
			}
		}
	}
}

// getDeviceName возвращает имя устройства по типу
func (dm *DeviceManager) getDeviceName(deviceType byte) string {
	switch deviceType {
	case 0x00:
		return "Нет устройства"
	case 0x01:
		return "Основной мотор"
	case 0x02:
		return "Трехосный датчик наклона"
	case 0x08:
		return "Простой светодиод"
	case 0x17:
		return "RGB светодиод"
	case 0x14, 0x15, 0x16:
		return "Датчик расстояния"
	case 0x20:
		return "Внешний мотор"
	default:
		return fmt.Sprintf("Неизвестное (0x%02x)", deviceType)
	}
}

// StopMotor останавливает мотор
func (dm *DeviceManager) StopMotor(portID byte) error {
	return dm.SetMotorPower(portID, 0, 0)
}

// SetLEDColor устанавливает цвет светодиода (ДИСКРЕТНЫЙ режим RGB)
func (dm *DeviceManager) SetLEDColor(portID byte, red, green, blue byte) error {
	if !dm.hubMgr.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	log.Printf("Установка RGB цвета светодиода на порту %d: (%d,%d,%d)", portID, red, green, blue)

	// 1. Устанавливаем режим светодиода в DISCRETE (для RGB)
	modeData, err := dm.parser.EncodeLEDModeCommand(portID, LED_DISCRETE_MODE)
	if err != nil {
		return fmt.Errorf("ошибка кодирования режима: %v", err)
	}

	err = dm.hubMgr.WriteCharacteristic(INPUT_COMMAND_UUID, modeData)
	if err != nil {
		log.Printf("Предупреждение при установке режима: %v", err)
		// Не прерываем выполнение, продолжаем
	}

	// 2. Отправляем команду RGB цвета
	cmd := LEDCommand{
		PortID: portID,
		Red:    red,
		Green:  green,
		Blue:   blue,
	}

	colorData, err := dm.parser.EncodeLEDCommand(cmd)
	if err != nil {
		return fmt.Errorf("ошибка кодирования цвета: %v", err)
	}

	err = dm.hubMgr.WriteCharacteristic(OUTPUT_COMMAND_UUID, colorData)
	if err != nil {
		return fmt.Errorf("ошибка отправки цвета: %v", err)
	}

	// Обновляем состояние
	dm.devicesMu.Lock()
	if device, exists := dm.devices[portID]; exists {
		device.Properties["red"] = red
		device.Properties["green"] = green
		device.Properties["blue"] = blue
		device.Properties["mode"] = "discrete"
		device.LastUpdate = time.Now()
		dm.devicesMu.Unlock()
		dm.notifyDeviceChanged(portID, device) // Уведомляем об изменении
	} else {
		dm.devicesMu.Unlock()
	}

	return nil
}

// SetLEDColorIndex устанавливает цвет светодиода по индексу (АБСОЛЮТНЫЙ режим)
func (dm *DeviceManager) SetLEDColorIndex(portID byte, colorIndex byte) error {
	if !dm.hubMgr.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	log.Printf("Установка цвета светодиода (индекс) на порту %d: %d", portID, colorIndex)

	// 1. Устанавливаем режим светодиода в ABSOLUTE
	modeData, err := dm.parser.EncodeLEDModeCommand(portID, LED_ABSOLUTE_MODE)
	if err != nil {
		return fmt.Errorf("ошибка кодирования режима: %v", err)
	}

	err = dm.hubMgr.WriteCharacteristic(INPUT_COMMAND_UUID, modeData)
	if err != nil {
		log.Printf("Предупреждение при установке режима: %v", err)
	}

	// 2. Отправляем команду индекса цвета
	colorData, err := dm.parser.EncodeLEDIndexCommand(portID, colorIndex)
	if err != nil {
		return fmt.Errorf("ошибка кодирования индекса: %v", err)
	}

	err = dm.hubMgr.WriteCharacteristic(OUTPUT_COMMAND_UUID, colorData)
	if err != nil {
		return fmt.Errorf("ошибка отправки индекса: %v", err)
	}

	// Обновляем состояние
	dm.devicesMu.Lock()
	if device, exists := dm.devices[portID]; exists {
		device.Properties["color_index"] = colorIndex
		device.Properties["mode"] = "absolute"
		device.LastUpdate = time.Now()
		dm.devicesMu.Unlock()

		dm.notifyDeviceChanged(portID, device) // Уведомляем об изменении
	} else {
		dm.devicesMu.Unlock()
	}

	return nil
}

// SetMotorPower устанавливает мощность мотора
func (dm *DeviceManager) SetMotorPower(portID byte, power int8, duration uint16) error {
	if !dm.hubMgr.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	log.Printf("Установка мощности мотора на порту %d: %d%%", portID, power)

	cmd := MotorCommand{
		PortID:   portID,
		Power:    power,
		Duration: duration,
	}

	motorData, err := dm.parser.EncodeMotorCommand(cmd)
	if err != nil {
		return fmt.Errorf("ошибка кодирования команды мотора: %v", err)
	}

	err = dm.hubMgr.WriteCharacteristic(OUTPUT_COMMAND_UUID, motorData)
	if err != nil {
		return fmt.Errorf("ошибка отправки команды мотора: %v", err)
	}

	// Обновление состояния и таймер для остановки (если нужно)
	dm.devicesMu.Lock()
	if device, exists := dm.devices[portID]; exists {
		device.Properties["power"] = power
		device.Properties["is_running"] = true
		device.LastUpdate = time.Now()
		dm.devicesMu.Unlock()

		dm.notifyDeviceChanged(portID, device) // Уведомляем об изменении
	} else {
		dm.devicesMu.Unlock()
	}

	if duration > 0 {
		go func() {
			time.Sleep(time.Duration(duration) * time.Millisecond)
			dm.StopMotor(portID)
		}()
	}

	return nil
}

// GetDevices возвращает список устройств
func (dm *DeviceManager) GetDevices() []Device {
	dm.devicesMu.RLock()
	defer dm.devicesMu.RUnlock()

	devices := make([]Device, 0, len(dm.devices))
	for _, device := range dm.devices {
		if device.IsConnected {
			devices = append(devices, *device)
		}
	}

	return devices
}

// GetDevice возвращает устройство по порту
func (dm *DeviceManager) GetDevice(portID byte) (*Device, bool) {
	dm.devicesMu.RLock()
	defer dm.devicesMu.RUnlock()

	device, exists := dm.devices[portID]
	if !exists || !device.IsConnected {
		return nil, false
	}

	return device, true
}
