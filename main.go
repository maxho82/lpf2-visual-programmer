package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Включаем подробное логирование
	log.Println("=== Запуск LPF2 Visual Programmer ===")

	// Создание приложения Fyne
	myApp := app.New()
	myApp.Settings().SetTheme(&CustomTheme{})

	// Создание главного окна
	window := myApp.NewWindow("Визуальный программист LPF2")
	window.SetMaster()
	window.Resize(fyne.NewSize(1200, 800))

	// Настройка обработки ошибок
	window.SetCloseIntercept(func() {
		log.Println("Закрытие приложения...")
		window.Close()
	})

	// Инициализация менеджера хаба
	log.Println("Инициализация менеджера хаба...")
	hubMgr, err := NewHubManager()
	if err != nil {
		log.Fatalf("Ошибка инициализации менеджера хаба: %v", err)
	}

	// Инициализация системы визуального программирования
	log.Println("Инициализация менеджера программ...")
	programMgr := NewProgramManager(hubMgr)

	// Инициализация GUI
	log.Println("Инициализация GUI...")
	gui := NewGUI(window, hubMgr, programMgr)

	// Создаем и настраиваем монитор сенсоров
	sensorMonitor := NewSensorMonitor(hubMgr, programMgr.deviceMgr, gui)
	hubMgr.SetSensorMonitor(sensorMonitor) // Устанавливаем монитор в hubMgr

	// Показ окна и запуск приложения
	log.Println("Запуск GUI...")
	window.SetContent(gui.BuildUI())

	// Запускаем периодическое обновление статуса подключения
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			gui.updateConnectionStatus()
		}
	}()

	// Запускаем мониторинг при подключении
	go func() {
		for {
			if hubMgr.IsConnected() && sensorMonitor != nil {
				sensorMonitor.Start()
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	window.ShowAndRun()

	// Очистка при закрытии
	log.Println("Очистка ресурсов...")
	hubMgr.Disconnect()
}
