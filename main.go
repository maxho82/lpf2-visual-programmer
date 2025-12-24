package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Создание приложения Fyne
	myApp := app.New()
	myApp.Settings().SetTheme(&CustomTheme{})

	// Создание главного окна
	window := myApp.NewWindow("Визуальный программист LPF2")
	window.SetMaster()
	window.Resize(fyne.NewSize(1200, 800))

	// Инициализация менеджера хаба
	log.Println("=== Инициализация LPF2 Visual Programmer ===")

	hubMgr, err := NewHubManager()
	if err != nil {
		log.Fatalf("Ошибка инициализации менеджера хаба: %v", err)
	}

	// Инициализация системы визуального программирования
	programMgr := NewProgramManager(hubMgr)

	// Инициализация GUI
	gui := NewGUI(window, hubMgr, programMgr)

	// Показ окна и запуск приложения
	window.SetContent(gui.BuildUI())

	// Запускаем периодическое обновление статуса подключения
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			gui.updateConnectionStatus()
		}
	}()

	window.ShowAndRun()

	// Очистка при закрытии
	hubMgr.Disconnect()
}
