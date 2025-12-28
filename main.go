package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// Убираем периодическое обновление статуса и оставляем только по событиям

func main() {
	log.Println("=== Запуск LPF2 Visual Programmer ===")

	myApp := app.New()
	myApp.Settings().SetTheme(&CustomTheme{})

	window := myApp.NewWindow("Визуальный программист LPF2")
	window.SetMaster()
	window.Resize(fyne.NewSize(1200, 800))

	hubMgr, err := NewHubManager()
	if err != nil {
		log.Fatalf("Ошибка инициализации: %v", err)
	}

	programMgr := NewProgramManager(hubMgr)
	gui := NewGUI(window, hubMgr, programMgr)

	// Создаем монитор сенсоров
	sensorMonitor := NewSensorMonitor(hubMgr, programMgr.deviceMgr, gui)
	hubMgr.SetSensorMonitor(sensorMonitor)

	hubMgr.SetPortNotificationCallback(func(portID byte, deviceType byte, data []byte) {
		log.Printf("Порт %d: обнаружено устройство типа 0x%02x", portID, deviceType)

		if sensorMonitor != nil {
			go sensorMonitor.UpdateDevices()
		}
	})

	window.SetContent(gui.BuildUI())

	// Запускаем мониторинг только при подключении
	// (теперь это происходит в connectToHub)

	window.ShowAndRun()
	hubMgr.Disconnect()
}
