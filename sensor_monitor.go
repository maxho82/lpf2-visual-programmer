package main

import (
	"fmt"
	"log"
	"time"
)

// SensorMonitor мониторит состояние устройств и батареи
type SensorMonitor struct {
	hubMgr      *HubManager
	deviceMgr   *DeviceManager
	gui         *GUI
	stopCh      chan struct{}
	initialScan bool
}

// NewSensorMonitor создает новый монитор
func NewSensorMonitor(hubMgr *HubManager, deviceMgr *DeviceManager, gui *GUI) *SensorMonitor {
	return &SensorMonitor{
		hubMgr:    hubMgr,
		deviceMgr: deviceMgr,
		gui:       gui,
		stopCh:    make(chan struct{}),
	}
}

// Start запускает мониторинг - делаем однократный опрос при подключении
func (sm *SensorMonitor) Start() {
	sm.initialScan = true
	go sm.initialDeviceScan()
	log.Println("Мониторинг устройств запущен (однократный опрос при подключении)")
}

// Stop останавливает мониторинг
func (sm *SensorMonitor) Stop() {
	if sm.stopCh != nil {
		close(sm.stopCh)
	}
	log.Println("Мониторинг устройств остановлен")
}

// initialDeviceScan выполняет однократный опрос всех устройств
func (sm *SensorMonitor) initialDeviceScan() {
	if !sm.hubMgr.IsConnected() {
		return
	}

	log.Println("Выполняем начальный опрос устройств...")

	// Опрашиваем батарею
	sm.updateBatteryLevel()

	// Опрашиваем основные порты
	ports := []byte{1, 2, 6}
	for _, port := range ports {
		sm.updatePortInfo(port)
	}

	sm.initialScan = false
	log.Println("Начальный опрос устройств завершен")
}

// updatePortInfo обновляет информацию о конкретном порте
func (sm *SensorMonitor) updatePortInfo(portID byte) {
	sm.hubMgr.connectionMutex.Lock()
	defer sm.hubMgr.connectionMutex.Unlock()

	if sm.hubMgr.hubInfo == nil {
		return
	}

	// Ищем или создаем информацию о порте
	var portInfo *PortInfo
	for i := range sm.hubMgr.hubInfo.Ports {
		if sm.hubMgr.hubInfo.Ports[i].PortID == portID {
			portInfo = &sm.hubMgr.hubInfo.Ports[i]
			break
		}
	}

	if portInfo == nil {
		// Добавляем новый порт
		portInfo = &PortInfo{
			PortID:      portID,
			DeviceType:  sm.getDeviceTypeForPort(portID),
			DeviceName:  sm.getDeviceNameForPort(portID),
			IsConnected: true,
			LastValue:   []byte{0},
			LastUpdate:  time.Now(),
		}
		sm.hubMgr.hubInfo.Ports = append(sm.hubMgr.hubInfo.Ports, *portInfo)
	} else {
		// Обновляем существующий
		portInfo.DeviceType = sm.getDeviceTypeForPort(portID)
		portInfo.DeviceName = sm.getDeviceNameForPort(portID)
		portInfo.IsConnected = true
		portInfo.LastUpdate = time.Now()

		// Устанавливаем начальные значения
		switch portInfo.DeviceType {
		case 0x01: // Мотор
			portInfo.LastValue = []byte{0} // Мощность 0%
		case 0x17: // RGB светодиод
			portInfo.LastValue = []byte{0, 0, 0} // Выключен
		}
	}

	// Синхронизируем с DeviceManager
	if sm.deviceMgr != nil {
		sm.deviceMgr.UpdateDevices([]PortInfo{*portInfo})
	}
}

/* func (sm *SensorMonitor) monitorLoop() {
	ticker := time.NewTicker(30 * time.Second) // Только для периодического опроса батареи
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopCh:
			return
		case <-ticker.C:
			if sm.hubMgr.IsConnected() {
				sm.updateBatteryLevel()
			}
		}
	}
} */

// updateBatteryLevel обновляет уровень батареи
func (sm *SensorMonitor) updateBatteryLevel() {
	// Чтение уровня батареи (стандартная BLE характеристика)
	// UUID для Battery Level: 00002a19-0000-1000-8000-00805f9b34fb
	if sm.hubMgr.IsConnected() { // ИСПРАВЛЕНО
		// В реальном приложении здесь будет чтение BLE характеристики
		// Для теста используем фиктивное значение
		batteryLevel := 85 // Тестовое значение 85%

		sm.hubMgr.connectionMutex.Lock()
		if sm.hubMgr.hubInfo != nil {
			sm.hubMgr.hubInfo.Battery = batteryLevel
		}
		sm.hubMgr.connectionMutex.Unlock()
	}
}

/* // updatePortStatus обновляет статус портов и синхронизирует с DeviceManager
func (sm *SensorMonitor) updatePortStatus() {
	if sm.deviceMgr == nil {
		return
	}

	// Получаем текущие порты из HubManager
	sm.hubMgr.connectionMutex.Lock()
	hubInfo := sm.hubMgr.hubInfo
	sm.hubMgr.connectionMutex.Unlock()

	if hubInfo == nil {
		return
	}

	// Создаем срез PortInfo для обновления DeviceManager
	var portInfos []PortInfo

	// Проверяем стандартные порты
	ports := []byte{1, 2, 6}
	for _, portID := range ports {
		portInfo := PortInfo{
			PortID:      portID,
			DeviceType:  sm.getDeviceTypeForPort(portID),
			DeviceName:  sm.getDeviceNameForPort(portID),
			IsConnected: true,
			LastValue:   []byte{0},
			LastUpdate:  time.Now(),
		}

		// Проверяем, есть ли данные в hubInfo
		for _, existingPort := range hubInfo.Ports {
			if existingPort.PortID == portID && existingPort.IsConnected {
				portInfo = existingPort
				break
			}
		}

		portInfos = append(portInfos, portInfo)
	}

	// Обновляем устройства в DeviceManager
	sm.deviceMgr.UpdateDevices(portInfos)
}

// updateSensorValues обновляет значения датчиков
func (sm *SensorMonitor) updateSensorValues() {
	// Обновляем данные в HubManager
	sm.hubMgr.connectionMutex.Lock()
	defer sm.hubMgr.connectionMutex.Unlock()

	if sm.hubMgr.hubInfo == nil {
		return
	}

	// Обновляем значения для каждого порта
	for i := range sm.hubMgr.hubInfo.Ports {
		port := &sm.hubMgr.hubInfo.Ports[i]

		// Генерируем тестовые значения в зависимости от типа устройства
		switch port.DeviceType {
		case 0x01: // Мотор
			// Значение мощности (0-100%)
			port.LastValue = []byte{byte(50)} // Фиксированное тестовое значение

		case 0x17: // RGB светодиод
			// Цвет (R,G,B) - начальное значение
			if len(port.LastValue) == 0 {
				port.LastValue = []byte{0, 0, 0} // Выключен
			}

		case 0x02: // Датчик наклона
			port.LastValue = []byte{byte(0)} // Нет наклона

		case 0x23: // Датчик расстояния
			port.LastValue = []byte{byte(10)} // 10 см
		}

		port.LastUpdate = time.Now()
	}
}
*/
// getDeviceTypeForPort возвращает тип устройства для порта
func (sm *SensorMonitor) getDeviceTypeForPort(port byte) byte {
	switch port {
	case 1, 2:
		return 0x01 // Мотор (предполагаем по умолчанию)
	case 6:
		return 0x17 // RGB светодиод
	default:
		return 0x00 // Неизвестное
	}
}

// getDeviceNameForPort возвращает имя устройства для порта
func (sm *SensorMonitor) getDeviceNameForPort(port byte) string {
	switch port {
	case 1:
		return "Мотор A"
	case 2:
		return "Мотор B"
	case 6:
		return "Светодиод"
	default:
		return fmt.Sprintf("Порт %d", port)
	}
}

// GetSensorValue возвращает значение датчика в удобном формате
func (sm *SensorMonitor) GetSensorValue(portID byte) (string, string) {
	sm.hubMgr.connectionMutex.RLock()
	defer sm.hubMgr.connectionMutex.RUnlock()

	if sm.hubMgr.hubInfo == nil {
		return "Нет данных", ""
	}

	for _, port := range sm.hubMgr.hubInfo.Ports {
		if port.PortID == portID {
			return sm.formatSensorValue(port), port.DeviceName
		}
	}

	return "Не подключен", ""
}

// formatSensorValue форматирует значение датчика
func (sm *SensorMonitor) formatSensorValue(port PortInfo) string {
	if len(port.LastValue) == 0 {
		return "Нет данных"
	}

	switch port.DeviceType {
	case 0x01: // Мотор
		return fmt.Sprintf("Мощность: %d%%", port.LastValue[0])

	case 0x02: // Датчик наклона
		return fmt.Sprintf("Угол: %d°", port.LastValue[0])

	case 0x17: // RGB светодиод
		if len(port.LastValue) >= 3 {
			return fmt.Sprintf("RGB(%d,%d,%d)",
				port.LastValue[0], port.LastValue[1], port.LastValue[2])
		}
		return "Цвет: неизвестен"

	case 0x23: // Датчик расстояния
		return fmt.Sprintf("Расстояние: %d см", port.LastValue[0])

	default:
		return fmt.Sprintf("Значение: %v", port.LastValue)
	}
}

// UpdateDevices уведомляет об обновлении устройств (для вызова извне)
func (sm *SensorMonitor) UpdateDevices() {
	sm.hubMgr.connectionMutex.Lock()
	defer sm.hubMgr.connectionMutex.Unlock()

	if sm.hubMgr.hubInfo == nil || sm.deviceMgr == nil {
		return
	}

	// Собираем актуальные порты
	var portInfos []PortInfo
	ports := []byte{1, 2, 6}

	for _, portID := range ports {
		for _, port := range sm.hubMgr.hubInfo.Ports {
			if port.PortID == portID && port.IsConnected {
				portInfos = append(portInfos, port)
				break
			}
		}
	}

	sm.deviceMgr.UpdateDevices(portInfos)
}

/* func (sm *SensorMonitor) getDeviceNameForType(deviceType byte) string {
	switch deviceType {
	case 0x00:
		return "Нет устройства"
	case 0x01:
		return "Мотор"
	case 0x02:
		return "Датчик наклона"
	case 0x08:
		return "Светодиод"
	case 0x17:
		return "RGB светодиод"
	case 0x14, 0x15, 0x16:
		return "Датчик расстояния"
	case 0x20:
		return "Внешний мотор"
	case 0x21:
		return "Датчик касания"
	case 0x22:
		return "Датчик тока"
	case 0x23:
		return "Датчик напряжения"
	case 0x25:
		return "Датчик цвета"
	default:
		return fmt.Sprintf("Неизвестное (0x%02x)", deviceType)
	}
} */
