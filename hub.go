package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tinybluetooth "tinygo.org/x/bluetooth"
)

// HubManager управляет подключением к LPF2-хабу
type HubManager struct {
	adapter                  *tinybluetooth.Adapter
	device                   tinybluetooth.Device
	deviceAddress            string
	isConnected              bool
	connectionMutex          sync.RWMutex
	hubInfo                  *HubInfo
	stopScan                 context.CancelFunc
	services                 map[string]tinybluetooth.DeviceService
	characteristics          map[string]tinybluetooth.DeviceCharacteristic
	sensorMonitor            *SensorMonitor
	portNotificationCallback func(portID byte, deviceType byte, data []byte)
	stateChangedCallback     func() // Добавим callback для изменений состояния
	batteryUpdateCallback    func(batteryLevel int)
}

// HubInfo содержит информацию о подключенном хабе
type HubInfo struct {
	Name        string
	Address     string
	Battery     int
	Firmware    string
	Ports       []PortInfo
	LastUpdated time.Time
}

// PortInfo информация о порте хаба
type PortInfo struct {
	PortID      byte
	DeviceType  byte
	DeviceName  string
	IsConnected bool
	LastValue   []byte
	Mode        byte
	LastUpdate  time.Time // ДОБАВЛЕНО
}

// NewHubManager создает новый менеджер хаба
func NewHubManager() (*HubManager, error) {
	adapter := tinybluetooth.DefaultAdapter
	if adapter == nil {
		return nil, fmt.Errorf("BLE адаптер не найден")
	}

	// Включение адаптера
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("ошибка включения BLE адаптера: %v", err)
	}

	return &HubManager{
		adapter: adapter,
		hubInfo: &HubInfo{
			Ports: make([]PortInfo, 6),
		},
		services:        make(map[string]tinybluetooth.DeviceService),
		characteristics: make(map[string]tinybluetooth.DeviceCharacteristic),
	}, nil
}

// ScanForHubs сканирует LPF2-хабы
func (hm *HubManager) ScanForHubs(timeout time.Duration) ([]HubInfo, error) {
	var foundHubs []HubInfo
	var scanMutex sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	hm.stopScan = cancel

	log.Println("=== Начало сканирования LPF2 хабов ===")

	err := hm.adapter.Scan(func(adapter *tinybluetooth.Adapter, result tinybluetooth.ScanResult) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		name := result.LocalName()
		address := result.Address.String()

		// Логируем все устройства для отладки
		if name != "" {
			log.Printf("Устройство: %s [%s]", name, address)
		}

		// Ищем LPF2 хаб по имени
		if strings.Contains(strings.ToUpper(name), "LPF2") || strings.Contains(strings.ToUpper(name), "WEDO") || strings.Contains(strings.ToUpper(name), "LEGO") {
			log.Printf("!!! НАЙДЕН LPF2 ХАБ: %s [%s]", name, address)

			scanMutex.Lock()
			foundHubs = append(foundHubs, HubInfo{
				Name:    name,
				Address: address,
			})
			scanMutex.Unlock()

			// Останавливаем сканирование при нахождении
			adapter.StopScan()
			cancel()
			return
		}
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка сканирования: %v", err)
	}

	// Ждем завершения таймаута или отмены
	<-ctx.Done()
	hm.adapter.StopScan()

	log.Printf("Сканирование завершено. Найдено хабов: %d", len(foundHubs))
	return foundHubs, nil
}

// Connect подключается к хабу
/* func (hm *HubManager) Connect(address string) error {
	hm.connectionMutex.Lock()
	defer hm.connectionMutex.Unlock()

	if hm.isConnected {
		hm.Disconnect()
	}

	log.Printf("Попытка подключения к %s", address)

	// Находим устройство через сканирование
	var targetDevice tinybluetooth.ScanResult
	found := false

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Сканируем для получения устройства...")

	err := hm.adapter.Scan(func(adapter *tinybluetooth.Adapter, result tinybluetooth.ScanResult) {
		if result.Address.String() == address {
			log.Printf("Найдено устройство для подключения: %s", result.LocalName())
			adapter.StopScan()
			targetDevice = result
			found = true
			cancel()
		}
	})

	if err != nil {
		return fmt.Errorf("ошибка сканирования: %v", err)
	}

	<-ctx.Done()
	hm.adapter.StopScan()

	if !found {
		return fmt.Errorf("устройство с адресом %s не найдено", address)
	}

	// Подключаемся
	log.Printf("Устанавливаем соединение с %s...", address)
	device, err := hm.adapter.Connect(targetDevice.Address, tinybluetooth.ConnectionParams{})
	if err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}

	hm.device = device
	hm.deviceAddress = address
	hm.isConnected = true

	// ОБНАРУЖИВАЕМ СЛУЖБЫ И ХАРАКТЕРИСТИКИ
	log.Println("Обнаружение служб и характеристик...")
	services, err := device.DiscoverServices(nil)
	if err != nil {
		log.Printf("Ошибка обнаружения служб: %v", err)
	} else {
		for _, service := range services {
			uuid := service.UUID().String()
			log.Printf("Найдена служба: %s", uuid)
			hm.services[uuid] = service

			// Обнаруживаем характеристики
			chars, err := service.DiscoverCharacteristics(nil)
			if err != nil {
				log.Printf("Ошибка обнаружения характеристик: %v", err)
				continue
			}

			for _, char := range chars {
				charUUID := char.UUID().String()
				log.Printf("  Характеристика: %s", charUUID)
				hm.characteristics[charUUID] = char
			}
		}
	}
	// 1. Получаем версию прошивки
	if serviceUUID := tinybluetooth.NewUUID(parseUUID(FIRMWARE_SERVICE_UUID)); err == nil {
		services, err := hm.device.DiscoverServices([]tinybluetooth.UUID{serviceUUID})
		if err == nil && len(services) > 0 {
			charUUID := tinybluetooth.NewUUID(parseUUID(FIRMWARE_CHAR_UUID))
			chars, err := services[0].DiscoverCharacteristics([]tinybluetooth.UUID{charUUID})
			if err == nil && len(chars) > 0 {
				char := chars[0]

				// Читаем версию прошивки
				data := []byte{}
				_, err := char.Read(data)
				if err == nil && len(data) > 0 {
					firmware := string(data)
					log.Printf("Версия прошивки: %s", firmware)

					hm.connectionMutex.Lock()
					if hm.hubInfo != nil {
						hm.hubInfo.Firmware = firmware
					}
					hm.connectionMutex.Unlock()
				}

				// Подписываемся на обновления
				char.EnableNotifications(func(data []byte) {
					if len(data) > 0 {
						firmware := string(data)
						log.Printf("Обновление версии прошивки: %s", firmware)

						hm.connectionMutex.Lock()
						if hm.hubInfo != nil {
							hm.hubInfo.Firmware = firmware
						}
						hm.connectionMutex.Unlock()
					}
				})
			}
		}
	}
	// 2. Получаем уровень батареи с правильным чтением
	log.Println("Получение уровня батареи...")

	batteryServiceUUID := "0000180f-0000-1000-8000-00805f9b34fb"
	batteryCharUUID := "00002a19-0000-1000-8000-00805f9b34fb"

	// Находим службу батареи
	services, err = hm.device.DiscoverServices(nil)
	if err == nil {
		for _, service := range services {
			if service.UUID().String() == batteryServiceUUID {
				chars, err := service.DiscoverCharacteristics(nil)
				if err == nil {
					for _, char := range chars {
						if char.UUID().String() == batteryCharUUID {
							// Читаем начальное значение
							data := make([]byte, 1)
							n, err := char.Read(data)
							if err == nil && n > 0 {
								batteryLevel := int(data[0])
								log.Printf("Уровень батареи: %d%%", batteryLevel)

								hm.connectionMutex.Lock()
								hm.hubInfo.Battery = batteryLevel
								hm.connectionMutex.Unlock()

								if hm.batteryUpdateCallback != nil {
									hm.batteryUpdateCallback(batteryLevel)
								}
							}

							// Подписываемся на обновления
							char.EnableNotifications(func(data []byte) {
								if len(data) > 0 {
									batteryLevel := int(data[0])
									log.Printf("Обновление уровня батареи: %d%%", batteryLevel)

									hm.connectionMutex.Lock()
									hm.hubInfo.Battery = batteryLevel
									hm.connectionMutex.Unlock()

									if hm.batteryUpdateCallback != nil {
										hm.batteryUpdateCallback(batteryLevel)
									}

									if hm.stateChangedCallback != nil {
										hm.stateChangedCallback()
									}
								}
							})
							log.Println("Подписка на обновления батареи установлена")
							break
						}
					}
				}
			}
		}
	}

	// Подписка на уведомления портов
	if portChar, ok := hm.characteristics[PORT_NOTIF_UUID]; ok {
		err := portChar.EnableNotifications(func(data []byte) {
			if len(data) >= 3 {
				portID := data[1]
				deviceType := data[2]

				log.Printf("Уведомление порта %d, тип устройства: 0x%02x", portID, deviceType)

				if hm.portNotificationCallback != nil {
					hm.portNotificationCallback(portID, deviceType, data)
				}

				hm.connectionMutex.Lock()
				if hm.hubInfo != nil {
					var portInfo *PortInfo
					for i := range hm.hubInfo.Ports {
						if hm.hubInfo.Ports[i].PortID == portID {
							portInfo = &hm.hubInfo.Ports[i]
							break
						}
					}

					if portInfo == nil {
						portInfo = &PortInfo{
							PortID:      portID,
							DeviceType:  deviceType,
							IsConnected: true,
							LastUpdate:  time.Now(),
						}
						hm.hubInfo.Ports = append(hm.hubInfo.Ports, *portInfo)
					} else {
						portInfo.DeviceType = deviceType
						portInfo.IsConnected = true
						portInfo.LastUpdate = time.Now()
					}
				}
				hm.connectionMutex.Unlock()
			}
		})

		if err != nil {
			log.Printf("Ошибка подписки на порты: %v", err)
		} else {
			log.Println("Подписка на уведомления портов установлена")

			// Отправляем запрос на получение информации о портах
			if inputChar, ok := hm.characteristics[INPUT_COMMAND_UUID]; ok {
				_, err := inputChar.WriteWithoutResponse([]byte{0x01})
				if err != nil {
					log.Printf("Ошибка запроса информации о портах: %v", err)
				} else {
					log.Println("Запрос информации о портах отправлен")
				}
			}
		}
	}

	// Обновляем информацию о хабе
	hm.hubInfo.Name = targetDevice.LocalName()
	hm.hubInfo.Address = address
	hm.hubInfo.LastUpdated = time.Now()

	// Запускаем мониторинг сенсоров
	if hm.sensorMonitor != nil {
		hm.sensorMonitor.Start() // Теперь это однократный опрос
	}

	// Инициируем опрос портов
	go func() {
		time.Sleep(1 * time.Second) // Даем время на установку соединения
		if hm.sensorMonitor != nil {
			hm.sensorMonitor.UpdateDevices()
		}
	}()

	return nil
} */
//----------------------------------------------------------
func (hm *HubManager) Connect(address string) error {
	hm.connectionMutex.Lock()
	defer hm.connectionMutex.Unlock()

	if hm.isConnected {
		hm.Disconnect()
	}

	log.Printf("Попытка подключения к %s", address)

	// Запускаем сканирование в отдельной горутине
	//var targetDevice tinybluetooth.ScanResult
	//var found bool

	// Создаем канал для результата сканирования
	scanResult := make(chan struct {
		device tinybluetooth.ScanResult
		found  bool
		err    error
	}, 1)

	// Запускаем сканирование в отдельной горутине
	go func() {
		var targetDevice tinybluetooth.ScanResult
		found := false

		// Используем контекст с таймаутом для сканирования
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := hm.adapter.Scan(func(adapter *tinybluetooth.Adapter, result tinybluetooth.ScanResult) {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if result.Address.String() == address {
				log.Printf("Найдено устройство для подключения: %s", result.LocalName())
				adapter.StopScan()
				targetDevice = result
				found = true
				cancel()
			}
		})

		scanResult <- struct {
			device tinybluetooth.ScanResult
			found  bool
			err    error
		}{targetDevice, found, err}
	}()

	// Ждем результат сканирования
	result := <-scanResult

	if result.err != nil {
		return fmt.Errorf("ошибка сканирования: %v", result.err)
	}

	if !result.found {
		return fmt.Errorf("устройство с адресом %s не найдено", address)
	}

	// Подключаемся
	log.Printf("Устанавливаем соединение с %s...", address)
	device, err := hm.adapter.Connect(result.device.Address, tinybluetooth.ConnectionParams{})
	if err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}

	hm.device = device
	hm.deviceAddress = address
	hm.isConnected = true

	// Обнаружение служб и характеристик
	log.Println("Обнаружение служб и характеристик...")
	services, err := device.DiscoverServices(nil)
	if err != nil {
		log.Printf("Ошибка обнаружения служб: %v", err)
	} else {
		for _, service := range services {
			uuid := service.UUID().String()
			log.Printf("Найдена служба: %s", uuid)
			hm.services[uuid] = service

			// Обнаруживаем характеристики
			chars, err := service.DiscoverCharacteristics(nil)
			if err != nil {
				log.Printf("Ошибка обнаружения характеристик: %v", err)
				continue
			}

			for _, char := range chars {
				charUUID := char.UUID().String()
				log.Printf("  Характеристика: %s", charUUID)
				hm.characteristics[charUUID] = char
			}
		}
	}

	// Обновляем информацию о хабе

	if hm.hubInfo == nil {
		hm.hubInfo = &HubInfo{
			Ports: make([]PortInfo, 6),
		}
	}

	hm.hubInfo.Name = result.device.LocalName()
	hm.hubInfo.Address = address
	hm.hubInfo.LastUpdated = time.Now()

	return nil
}

func (hm *HubManager) discoverServicesAndCharacteristics(device tinybluetooth.Device) {
	log.Println("Обнаружение служб и характеристик...")
	services, err := device.DiscoverServices(nil)
	if err != nil {
		log.Printf("Ошибка обнаружения служб: %v", err)
		return
	}

	for _, service := range services {
		uuid := service.UUID().String()
		log.Printf("Найдена служба: %s", uuid)
		hm.services[uuid] = service

		chars, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			log.Printf("Ошибка обнаружения характеристик: %v", err)
			continue
		}

		for _, char := range chars {
			charUUID := char.UUID().String()
			log.Printf("  Характеристика: %s", charUUID)
			hm.characteristics[charUUID] = char
		}
	}

	log.Println("Обнаружение служб и характеристик завершено")
}

//-------------------------------------------------------------------

func (hm *HubManager) SetBatteryUpdateCallback(callback func(batteryLevel int)) {
	hm.batteryUpdateCallback = callback
}
func parseUUID(uuidStr string) [16]byte {
	// Упрощенный парсинг UUID - в реальном коде нужно обрабатывать дефисы
	var uuid [16]byte
	// Преобразуем строку UUID в байты
	// Пример: "00004f0e-1212-efde-1523-785feabcd123"
	hexStr := strings.ReplaceAll(uuidStr, "-", "")
	for i := 0; i < 32; i += 2 {
		b, _ := strconv.ParseUint(hexStr[i:i+2], 16, 8)
		uuid[i/2] = byte(b)
	}
	return uuid
}

func (hm *HubManager) SetPortNotificationCallback(callback func(portID byte, deviceType byte, data []byte)) {
	hm.portNotificationCallback = callback
}

// Disconnect отключается от хаба
func (hm *HubManager) Disconnect() {
	hm.connectionMutex.Lock()
	defer hm.connectionMutex.Unlock()

	if hm.isConnected {
		log.Println("Отключение от хаба...")

		// Останавливаем мониторинг сенсоров
		if hm.sensorMonitor != nil {
			hm.sensorMonitor.Stop()
		}

		hm.device.Disconnect()
		hm.isConnected = false
		hm.hubInfo = &HubInfo{
			Ports: make([]PortInfo, 6),
		}
		log.Println("Отключено")
	}
}

// IsConnected возвращает статус подключения
func (hm *HubManager) IsConnected() bool {
	hm.connectionMutex.RLock()
	defer hm.connectionMutex.RUnlock()
	return hm.isConnected
}

// GetHubInfo возвращает информацию о хабе
func (hm *HubManager) GetHubInfo() HubInfo {
	hm.connectionMutex.RLock()
	defer hm.connectionMutex.RUnlock()

	if hm.hubInfo == nil {
		return HubInfo{}
	}
	return *hm.hubInfo
}

// SetSensorMonitor устанавливает монитор сенсоров
func (hm *HubManager) SetSensorMonitor(monitor *SensorMonitor) {
	hm.sensorMonitor = monitor
}

// GetSensorMonitor возвращает монитор сенсоров
func (hm *HubManager) GetSensorMonitor() *SensorMonitor {
	return hm.sensorMonitor
}

// WriteCharacteristic записывает данные в характеристику
func (hm *HubManager) WriteCharacteristic(uuid string, data []byte) error {
	hm.connectionMutex.RLock()
	defer hm.connectionMutex.RUnlock()

	if !hm.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	// Находим характеристику по UUID
	char, exists := hm.characteristics[uuid]
	if !exists {
		return fmt.Errorf("характеристика %s не найдена", uuid)
	}

	// Отправляем данные
	_, err := char.WriteWithoutResponse(data)
	if err != nil {
		return fmt.Errorf("ошибка отправки данных: %v", err)
	}

	log.Printf("Данные успешно отправлены на хаб")
	return nil
}

// SetStateChangedCallback устанавливает callback для изменений состояния
func (hm *HubManager) SetStateChangedCallback(callback func()) {
	hm.stateChangedCallback = callback
}
