package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// DeviceManager управляет устройствами хаба
type DeviceManager struct {
	hubMgr    *HubManager
	parser    *LPF2Parser
	devices   map[byte]*Device
	devicesMu sync.RWMutex
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
		device.Name = getDeviceName(portInfo.DeviceType) // Исправлено!
		device.IsConnected = portInfo.IsConnected
		device.LastUpdate = time.Now()
	}
}

// SetMotorPower устанавливает мощность мотора
func (dm *DeviceManager) SetMotorPower(portID byte, power int8, duration uint16) error {
	if !dm.hubMgr.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	log.Printf("Установка мощности мотора на порту %d: %d%% на %d мс", portID, power, duration)

	// Создание команды
	cmd := MotorCommand{
		PortID:   portID,
		Power:    power,
		Duration: duration,
	}

	data, err := dm.parser.EncodeMotorCommand(cmd)
	if err != nil {
		return fmt.Errorf("ошибка кодирования команды: %v", err)
	}

	// Отправка команды
	err = dm.hubMgr.WriteCharacteristic("00001565-1212-efde-1523-785feabcd123", data)
	if err != nil {
		return fmt.Errorf("ошибка отправки команды: %v", err)
	}

	// Обновление состояния устройства
	dm.devicesMu.Lock()
	if device, exists := dm.devices[portID]; exists {
		device.Properties["power"] = power
		device.Properties["duration"] = duration
		device.Properties["is_running"] = true
		device.LastUpdate = time.Now()
	}
	dm.devicesMu.Unlock()

	// Если указана длительность, через указанное время останавливаем мотор
	if duration > 0 {
		go func() {
			time.Sleep(time.Duration(duration) * time.Millisecond)
			dm.StopMotor(portID)
		}()
	}

	return nil
}

// StopMotor останавливает мотор
func (dm *DeviceManager) StopMotor(portID byte) error {
	return dm.SetMotorPower(portID, 0, 0)
}

// SetLEDColor устанавливает цвет светодиода
func (dm *DeviceManager) SetLEDColor(portID byte, red, green, blue byte) error {
	if !dm.hubMgr.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	// 1. Устанавливаем режим датчика в "absolute" (0x00)
	modeCmd := SetSensorModeCommand{
		PortID: portID,
		Mode:   0x00, // absolute
	}
	modeData, err := dm.parser.EncodeSetSensorMode(modeCmd)
	if err != nil {
		return fmt.Errorf("ошибка кодирования режима: %v", err)
	}
	err = dm.hubMgr.WriteCharacteristic("00001563-1212-efde-1523-785feabcd123", modeData)
	if err != nil {
		return fmt.Errorf("ошибка установки режима: %v", err)
	}

	// 2. Отправляем команду цвета
	cmd := LEDCommand{
		PortID: portID,
		Red:    red,
		Green:  green,
		Blue:   blue,
	}
	data, err := dm.parser.EncodeLEDCommand(cmd)
	if err != nil {
		return fmt.Errorf("ошибка кодирования цвета: %v", err)
	}
	err = dm.hubMgr.WriteCharacteristic("00001565-1212-efde-1523-785feabcd123", data)
	if err != nil {
		return fmt.Errorf("ошибка отправки цвета: %v", err)
	}

	// Обновляем состояние устройства
	dm.devicesMu.Lock()
	if device, exists := dm.devices[portID]; exists {
		device.Properties["red"] = red
		device.Properties["green"] = green
		device.Properties["blue"] = blue
		device.LastUpdate = time.Now()
	}
	dm.devicesMu.Unlock()

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

// getDeviceName возвращает имя устройства по типу
func getDeviceName(deviceType byte) string {
	switch deviceType {
	case 0x00:
		return "Нет устройства"
	case 0x01:
		return "Основной мотор"
	case 0x02:
		return "Трехосный датчик наклона"
	case 0x08:
		return "Светодиод"
	case 0x14, 0x15, 0x16, 0x17:
		return "Датчик расстояния"
	case 0x20:
		return "Внешний мотор"
	default:
		return fmt.Sprintf("Неизвестное (0x%02x)", deviceType)
	}
}
