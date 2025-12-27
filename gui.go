package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// GUI управляет графическим интерфейсом
type GUI struct {
	window     fyne.Window
	hubMgr     *HubManager
	programMgr *ProgramManager

	// Виджеты
	statusLabel      *widget.Label
	connectButton    *widget.Button
	disconnectButton *widget.Button
	runButton        *widget.Button
	stopButton       *widget.Button
	clearButton      *widget.Button

	// Панели
	devicePanel     fyne.CanvasObject
	propertiesPanel fyne.CanvasObject
	//programScroll   *container.Scroll // ДОБАВЛЕНО - прямой доступ к Scroll
	programPanel *container.Scroll
	blocksPanel  fyne.CanvasObject
}

// NewGUI создает новый GUI
func NewGUI(window fyne.Window, hubMgr *HubManager, programMgr *ProgramManager) *GUI {
	gui := &GUI{
		window:     window,
		hubMgr:     hubMgr,
		programMgr: programMgr,
	}

	return gui
}

// BuildUI строит интерфейс
func (gui *GUI) BuildUI() fyne.CanvasObject {
	// Создаем верхнюю панель инструментов
	toolbar := gui.createToolbar()

	// Создаем левую панель с устройствами
	gui.devicePanel = gui.createDevicePanel()

	// Создаем правую панель свойств
	gui.propertiesPanel = gui.createPropertiesPanel()

	// Создаем панель блоков
	gui.blocksPanel = gui.createBlocksPanel()

	// Получаем холст программы
	canvasObj := gui.programMgr.GetCanvas()
	if scroll, ok := canvasObj.(*container.Scroll); ok {
		gui.programPanel = scroll
	} else {
		// Если не Scroll, создаем новый
		content := container.NewWithoutLayout()

		// Создаем сетку
		grid := gui.programMgr.createGrid()
		content.Add(grid)

		gui.programPanel = container.NewScroll(content)
		gui.programPanel.SetMinSize(fyne.NewSize(800, 600))
	}

	// Устанавливаем содержимое programMgr.canvas
	gui.programMgr.canvas = gui.programPanel

	// Разделители
	leftSplit := container.NewHSplit(gui.devicePanel, gui.programPanel)
	leftSplit.SetOffset(0.2)

	rightSplit := container.NewHSplit(leftSplit, gui.propertiesPanel)
	rightSplit.SetOffset(0.8)

	// Основной макет
	mainContent := container.NewBorder(
		toolbar,               // Верх
		gui.createStatusBar(), // Низ
		nil,                   // Лево
		nil,                   // Право
		rightSplit,            // Центр
	)

	// Добавляем панель блоков слева
	fullLayout := container.NewBorder(
		nil, nil, gui.blocksPanel, nil, mainContent,
	)

	return fullLayout
}

// createToolbar создает панель инструментов
func (gui *GUI) createToolbar() *fyne.Container {
	// Кнопка подключения с иконкой
	gui.connectButton = widget.NewButtonWithIcon("Поиск хаба", theme.SearchIcon(), func() {
		gui.showHubDiscoveryDialog()
	})
	gui.connectButton.Importance = widget.MediumImportance

	// Кнопка отключения
	gui.disconnectButton = widget.NewButtonWithIcon("Отключиться", theme.CancelIcon(), func() {
		gui.hubMgr.Disconnect()
		gui.updateConnectionStatus()
	})
	gui.disconnectButton.Importance = widget.MediumImportance
	gui.disconnectButton.Disable()

	// Кнопка запуска программы
	gui.runButton = widget.NewButtonWithIcon("Запуск", theme.MediaPlayIcon(), func() {
		if err := gui.programMgr.RunProgram(); err != nil {
			dialog.ShowError(err, gui.window)
		}
	})
	gui.runButton.Importance = widget.HighImportance
	gui.runButton.Disable()

	// Кнопка остановки
	gui.stopButton = widget.NewButtonWithIcon("Стоп", theme.MediaStopIcon(), func() {
		gui.programMgr.StopProgram()
	})
	gui.stopButton.Importance = widget.MediumImportance
	gui.stopButton.Disable()

	// Кнопка очистки
	gui.clearButton = widget.NewButtonWithIcon("Очистить", theme.DeleteIcon(), func() {
		gui.programMgr.ClearProgram()
		gui.updateProgramCanvas()
	})
	gui.clearButton.Importance = widget.MediumImportance

	// Кнопка создания программы мигания
	blinkButton := widget.NewButtonWithIcon("Создать мигание", theme.RadioButtonIcon(), func() {
		gui.createBlinkProgram()
	})

	testLEDButton := widget.NewButtonWithIcon("Тест светодиод", theme.RadioButtonIcon(), func() {
		if !gui.hubMgr.IsConnected() {
			dialog.ShowError(fmt.Errorf("не подключено к хабу"), gui.window)
			return
		}

		// Включить красный светодиод
		err := gui.programMgr.deviceMgr.SetLEDColor(6, 255, 0, 0)
		if err != nil {
			dialog.ShowError(err, gui.window)
		} else {
			dialog.ShowInformation("Успех", "Светодиод включен (красный)", gui.window)

			// Через 1 секунду выключить
			go func() {
				time.Sleep(1 * time.Second)
				fyne.Do(func() {
					gui.programMgr.deviceMgr.SetLEDColor(6, 0, 0, 0)
				})
			}()
		}
	})

	// В createToolbar добавьте:
	testWedoProtocolButton := widget.NewButtonWithIcon("Тест WeDo протокол", theme.MediaPlayIcon(), func() {
		if !gui.hubMgr.IsConnected() {
			dialog.ShowError(fmt.Errorf("не подключено к хабу"), gui.window)
			return
		}

		// Тест 1: RGB режим (дискретный) - красный
		go func() {
			err := gui.programMgr.deviceMgr.SetLEDColor(6, 255, 0, 0)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(fmt.Errorf("RGB ошибка: %v", err), gui.window)
				} else {
					dialog.ShowInformation("Успех", "RGB красный отправлен", gui.window)
				}
			})

			time.Sleep(2 * time.Second)

			// Тест 2: Выключить
			err = gui.programMgr.deviceMgr.SetLEDColor(6, 0, 0, 0)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(fmt.Errorf("Выкл ошибка: %v", err), gui.window)
				}
			})

			time.Sleep(1 * time.Second)

			// Тест 3: Абсолютный режим (индекс цвета)
			err = gui.programMgr.deviceMgr.SetLEDColorIndex(6, 0x01) // Розовый
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(fmt.Errorf("Индекс ошибка: %v", err), gui.window)
				} else {
					dialog.ShowInformation("Успех", "Цвет по индексу отправлен", gui.window)
				}
			})
		}()
	})

	// Кнопка теста протокола
	testProtocolButton := widget.NewButtonWithIcon("Тест протокола", theme.VisibilityIcon(), func() {
		if !gui.hubMgr.IsConnected() {
			dialog.ShowError(fmt.Errorf("не подключено к хабу"), gui.window)
			return
		}

		// Запускаем тест в отдельной горутине
		go func() {
			TestLPF2Protocol(gui.hubMgr)
		}()

		dialog.ShowInformation("Тест запущен", "Проверяем протокол LPF2. Смотрите логи в консоли.", gui.window)
	})

	// Метка статуса
	gui.statusLabel = widget.NewLabel("Не подключено")
	gui.statusLabel.Alignment = fyne.TextAlignCenter
	gui.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	toolbar := container.NewHBox(
		gui.connectButton,
		gui.disconnectButton,
		widget.NewSeparator(),
		gui.runButton,
		gui.stopButton,
		widget.NewSeparator(),
		gui.clearButton,
		blinkButton,
		testLEDButton,          // ДОБАВЛЕНО
		testWedoProtocolButton, // Добавляем эту кнопку
		testProtocolButton,     // Добавляем эту кнопку
		layout.NewSpacer(),
		gui.statusLabel,
		layout.NewSpacer(),
	)

	return toolbar
}

// createDevicePanel создает панель устройств
func (gui *GUI) createDevicePanel() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Устройства и датчики", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 16
	title.TextStyle.Bold = true

	// Контейнер для устройств
	devicesContainer := container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
	)

	// Порт 1 - Мотор A
	port1 := gui.createPortWidget(1, "Порт A (Мотор)")
	devicesContainer.Add(port1)

	// Порт 2 - Мотор B
	port2 := gui.createPortWidget(2, "Порт B (Мотор)")
	devicesContainer.Add(port2)

	// Порт 6 - Светодиод
	port6 := gui.createPortWidget(6, "Встроенный светодиод")
	devicesContainer.Add(port6)

	// Датчики
	sensorsWidget := gui.createSensorsWidget()
	devicesContainer.Add(sensorsWidget)

	// Батарея
	batteryWidget := gui.createBatteryWidget()
	devicesContainer.Add(batteryWidget)

	// Информация о хабе
	hubInfoWidget := gui.createHubInfoWidget()
	devicesContainer.Add(hubInfoWidget)

	return container.NewVScroll(container.NewPadded(devicesContainer))
}

// createPortWidget создает виджет порта с динамическими данными
func (gui *GUI) createPortWidget(portID byte, label string) fyne.CanvasObject {
	// Иконка порта
	icon := widget.NewIcon(theme.StorageIcon())

	// Метка порта
	portLabel := widget.NewLabel(label)
	portLabel.Alignment = fyne.TextAlignCenter
	portLabel.TextStyle.Bold = true

	// Метка значения
	valueLabel := widget.NewLabel("Нет данных")
	valueLabel.Alignment = fyne.TextAlignCenter
	valueLabel.TextStyle.Italic = true

	// Метка статуса
	statusLabel := widget.NewLabel("Не подключено")
	statusLabel.Alignment = fyne.TextAlignCenter
	//statusLabel.TextSize = 10

	// Контейнер порта
	portContainer := container.NewVBox(
		container.NewCenter(icon),
		portLabel,
		valueLabel,
		statusLabel,
		widget.NewSeparator(),
	)

	// Обновляем виджет каждые 500мс
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			// Получаем данные из монитора сенсоров
			var valueStr, deviceName string
			if gui.hubMgr != nil && gui.hubMgr.GetSensorMonitor() != nil { // ИСПРАВЛЕНО
				valueStr, deviceName = gui.hubMgr.GetSensorMonitor().GetSensorValue(portID) // ИСПРАВЛЕНО
			}
			fyne.Do(func() {
				// Обновляем отображение
				if deviceName != "" && deviceName != "Не подключено" {
					statusLabel.SetText(fmt.Sprintf("✓ %s", deviceName))
					statusLabel.TextStyle.Bold = true
					icon.SetResource(theme.ConfirmIcon())
				} else {
					statusLabel.SetText("Не подключено")
					statusLabel.TextStyle.Bold = false
					icon.SetResource(theme.StorageIcon())
				}

				valueLabel.SetText(valueStr)

				// Обновляем виджеты
				statusLabel.Refresh()
				valueLabel.Refresh()
				icon.Refresh()
			})
		}
	}()

	return portContainer
}

// formatPortValue форматирует значение порта
func formatPortValue(port PortInfo) string {
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
			return fmt.Sprintf("RGB(%d,%d,%d)", port.LastValue[0], port.LastValue[1], port.LastValue[2])
		}
		return "Светодиод"
	default:
		return fmt.Sprintf("Значение: %v", port.LastValue)
	}
}

// createBatteryWidget создает виджет батареи
func (gui *GUI) createBatteryWidget() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Батарея", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 14
	title.TextStyle.Bold = true

	// Прогресс-бар
	progress := widget.NewProgressBar()

	// Метка процентов
	percentLabel := widget.NewLabel("--%")
	percentLabel.Alignment = fyne.TextAlignCenter

	// Контейнер
	batteryContainer := container.NewVBox(
		container.NewCenter(title),
		progress,
		percentLabel,
		widget.NewSeparator(),
	)

	// Обновление состояния батареи
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			hubInfo := gui.hubMgr.GetHubInfo()
			batteryLevel := hubInfo.Battery

			fyne.Do(func() {
				progress.SetValue(float64(batteryLevel) / 100)
				percentLabel.SetText(fmt.Sprintf("%d%%", batteryLevel))
				progress.Refresh()
				percentLabel.Refresh()
			})
		}
	}()

	return batteryContainer
}

// createHubInfoWidget создает виджет информации о хабе
func (gui *GUI) createHubInfoWidget() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Информация о хабе", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 14
	title.TextStyle.Bold = true

	// Поля информации
	nameLabel := widget.NewLabel("Имя: --")
	addressLabel := widget.NewLabel("Адрес: --")
	firmwareLabel := widget.NewLabel("Прошивка: --")

	// Контейнер
	infoContainer := container.NewVBox(
		container.NewCenter(title),
		nameLabel,
		addressLabel,
		firmwareLabel,
		widget.NewSeparator(),
	)

	// Обновление информации
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			hubInfo := gui.hubMgr.GetHubInfo()

			fyne.Do(func() {
				nameLabel.SetText(fmt.Sprintf("Имя: %s", hubInfo.Name))
				addressLabel.SetText(fmt.Sprintf("Адрес: %s", hubInfo.Address))
				firmwareLabel.SetText(fmt.Sprintf("Прошивка: %s", hubInfo.Firmware))

				nameLabel.Refresh()
				addressLabel.Refresh()
				firmwareLabel.Refresh()
			})
		}
	}()

	return infoContainer
}

// createPropertiesPanel создает панель свойств
func (gui *GUI) createPropertiesPanel() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Свойства блока", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 16
	title.TextStyle.Bold = true

	// Контейнер свойств
	propsContainer := container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
		widget.NewLabel("Выберите блок для настройки"),
	)

	return container.NewVScroll(container.NewPadded(propsContainer))
}

// createBlocksPanel создает панель блоков
func (gui *GUI) createBlocksPanel() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Блоки", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 16
	title.TextStyle.Bold = true

	// Список доступных блоков
	blocksList := container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
		gui.createBlockButton("Начать", BlockTypeStart),
		gui.createBlockButton("Мотор", BlockTypeMotor),
		gui.createBlockButton("Светодиод", BlockTypeLED),
		gui.createBlockButton("Ждать", BlockTypeWait),
		gui.createBlockButton("Повторять", BlockTypeLoop),
		gui.createBlockButton("Стоп", BlockTypeStop),
	)

	return container.NewVScroll(container.NewPadded(blocksList))
}

// createBlockButton создает кнопку для добавления блока
func (gui *GUI) createBlockButton(text string, blockType BlockType) *widget.Button {
	btn := widget.NewButton(text, func() {
		// Добавляем блок в позицию с небольшим смещением
		x := 50.0 + float64(len(gui.programMgr.blocks))*180
		y := 50.0

		log.Printf("Создание блока: %s в позиции (%.0f, %.0f)", text, x, y)

		// Добавляем блок в менеджер
		block := gui.programMgr.AddBlock(blockType, x, y)

		// Добавляем на холст
		gui.addBlockToCanvas(block)

		// Показываем сообщение
		log.Printf("Блок '%s' создан успешно!", text)
	})

	btn.Importance = widget.HighImportance
	return btn
}

// addBlockToCanvas добавляет блок на холст
func (gui *GUI) addBlockToCanvas(block *ProgramBlock) {
	log.Printf("Добавление блока на холст: %s (ID: %d)", block.Title, block.ID)

	// Создаем виджет блока
	blockWidget := gui.programMgr.CreateBlockWidget(block)

	// Добавляем на холст
	if gui.programPanel != nil {
		if content, ok := gui.programPanel.Content.(*fyne.Container); ok {
			// Проверяем, нет ли уже такого блока
			for _, obj := range content.Objects {
				if db, ok := obj.(*DraggableBlock); ok && db.block.ID == block.ID {
					log.Printf("Блок ID %d уже на холсте, пропускаем", block.ID)
					return
				}
			}

			content.Add(blockWidget)
			content.Refresh()
			gui.programPanel.Refresh()

			log.Printf("Блок %s добавлен на холст. Всего блоков: %d",
				block.Title, len(content.Objects))
		} else {
			log.Println("Ошибка: контент programPanel не является *fyne.Container")
		}
	} else {
		log.Println("Ошибка: programPanel равен nil")
	}
}

// createBlinkProgram создает программу мигания светодиода
func (gui *GUI) createBlinkProgram() {
	// Очищаем программу
	gui.programMgr.ClearProgram()

	// Очищаем холст
	gui.updateProgramCanvas()

	// Создаем блоки для мигания
	blocks := []struct {
		blockType BlockType
		x, y      float64
		config    func(*ProgramBlock)
	}{
		{BlockTypeStart, 200, 100, nil},
		{BlockTypeLoop, 200, 200, func(b *ProgramBlock) {
			b.Parameters["forever"] = true
		}},
		{BlockTypeLED, 200, 300, func(b *ProgramBlock) {
			b.Parameters["port"] = byte(6)
			b.Parameters["red"] = byte(255)
			b.Parameters["green"] = byte(0)
			b.Parameters["blue"] = byte(0)
		}},
		{BlockTypeWait, 200, 400, func(b *ProgramBlock) {
			b.Parameters["duration"] = 0.5
		}},
		{BlockTypeLED, 200, 500, func(b *ProgramBlock) {
			b.Parameters["port"] = byte(6)
			b.Parameters["red"] = byte(0)
			b.Parameters["green"] = byte(0)
			b.Parameters["blue"] = byte(0)
		}},
		{BlockTypeWait, 200, 600, func(b *ProgramBlock) {
			b.Parameters["duration"] = 0.5
		}},
	}

	// Добавляем и настраиваем блоки
	var prevBlock *ProgramBlock
	for _, item := range blocks {
		block := gui.programMgr.AddBlock(item.blockType, item.x, item.y)

		// Применяем конфигурацию
		if item.config != nil {
			item.config(block)
		}

		// Создаем связи между блоками
		if prevBlock != nil {
			prevBlock.NextBlockID = block.ID
		}
		prevBlock = block

		// Добавляем на холст
		gui.addBlockToCanvas(block)
	}

	dialog.ShowInformation("Готово",
		"Программа мигания светодиода создана!\n\n"+
			"Действия:\n"+
			"1. Подключитесь к хабу\n"+
			"2. Нажмите 'Запуск'\n"+
			"3. Светодиод на хабе начнет мигать\n"+
			"4. Нажмите 'Стоп' для остановки",
		gui.window)
}

// createStatusBar создает строку состояния
func (gui *GUI) createStatusBar() *fyne.Container {
	statusText := widget.NewLabel("Готово к работе")
	statusText.Alignment = fyne.TextAlignCenter
	statusText.TextStyle.Bold = true

	return container.NewCenter(statusText)
}

// showHubDiscoveryDialog показывает диалог поиска хаба
func (gui *GUI) showHubDiscoveryDialog() {
	progress := dialog.NewProgress("Поиск LPF2-хабов", "Сканирование...", gui.window)
	progress.Show()

	go func() {
		hubs, err := gui.hubMgr.ScanForHubs(15 * time.Second)

		fyne.Do(func() {
			progress.Hide()

			if err != nil {
				dialog.ShowError(err, gui.window)
				return
			}

			if len(hubs) == 0 {
				dialog.ShowInformation("Хабы не найдены",
					"Убедитесь, что:\n1. Хаб включен\n2. Хаб в режиме подключения (мигает)\n3. Bluetooth адаптер активен",
					gui.window)
				return
			}

			items := make([]string, len(hubs))
			for i, hub := range hubs {
				items[i] = fmt.Sprintf("%s (%s)", hub.Name, hub.Address)
			}

			list := widget.NewSelect(items, func(selected string) {
				if selected != "" {
					for _, hub := range hubs {
						fullName := fmt.Sprintf("%s (%s)", hub.Name, hub.Address)
						if fullName == selected {
							gui.connectToHub(hub.Address)
							break
						}
					}
				}
			})

			content := container.NewVBox(
				widget.NewLabel("Выберите хаб для подключения:"),
				list,
			)

			dialog.ShowCustom("Выбор хаба", "Подключиться", content, gui.window)
		})
	}()
}

// connectToHub подключается к указанному хабу
func (gui *GUI) connectToHub(address string) {
	progress := dialog.NewProgress("Подключение", "Подключение к хабу...", gui.window)
	progress.Show()

	go func() {
		err := gui.hubMgr.Connect(address)

		fyne.Do(func() {
			progress.Hide()

			if err != nil {
				dialog.ShowError(err, gui.window)
			} else {
				gui.updateConnectionStatus()
				dialog.ShowInformation("Успешно", "Подключение установлено!", gui.window)
			}
		})
	}()
}

// updateConnectionStatus обновляет статус подключения
func (gui *GUI) updateConnectionStatus() {
	fyne.Do(func() {
		isConnected := gui.hubMgr.IsConnected()

		if isConnected {
			gui.statusLabel.SetText("Подключено ✓")
			gui.statusLabel.Importance = widget.HighImportance
			gui.connectButton.Disable()
			gui.disconnectButton.Enable()
			gui.runButton.Enable()
			gui.stopButton.Enable()
		} else {
			gui.statusLabel.SetText("Не подключено")
			gui.statusLabel.Importance = widget.MediumImportance
			gui.connectButton.Enable()
			gui.disconnectButton.Disable()
			gui.runButton.Disable()
			gui.stopButton.Disable()
		}

		gui.statusLabel.Refresh()
		gui.connectButton.Refresh()
		gui.disconnectButton.Refresh()
		gui.runButton.Refresh()
		gui.stopButton.Refresh()
	})
}

func (gui *GUI) updateProgramCanvas() {
	log.Println("Полное обновление холста...")

	// Вызываем обновление в ProgramManager
	gui.programMgr.updateCanvas()

	// Обновляем programPanel, если он существует
	if gui.programPanel != nil {
		gui.programPanel.Refresh()
	}
}

// createSensorsWidget создает виджет для отображения всех датчиков
func (gui *GUI) createSensorsWidget() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Датчики", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 14
	title.TextStyle.Bold = true

	// Таблица датчиков
	sensorTable := widget.NewTable(
		func() (int, int) {
			return 5, 2 // 5 строк, 2 колонки
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Данные")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			// Определяем строки
			sensorNames := []string{
				"Батарея:",
				"Мотор A:",
				"Мотор B:",
				"Светодиод:",
				"Температура:",
			}

			if id.Row < len(sensorNames) {
				if id.Col == 0 {
					label.SetText(sensorNames[id.Row])
					label.TextStyle.Bold = true
				} else {
					// Динамические значения
					value := gui.getSensorValue(id.Row)
					label.SetText(value)
				}
			}
		},
	)

	sensorTable.SetColumnWidth(0, 100)
	sensorTable.SetColumnWidth(1, 100)

	// Контейнер
	sensorContainer := container.NewVBox(
		container.NewCenter(title),
		sensorTable,
		widget.NewSeparator(),
	)

	// Обновление таблицы
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			fyne.Do(func() {
				sensorTable.Refresh()
			})
		}
	}()

	return sensorContainer
}

// getSensorValue возвращает значение для конкретного датчика
func (gui *GUI) getSensorValue(sensorID int) string {
	if gui.hubMgr == nil || gui.hubMgr.GetSensorMonitor() == nil {
		return "--"
	}

	switch sensorID {
	case 0: // Батарея
		hubInfo := gui.hubMgr.GetHubInfo()
		return fmt.Sprintf("%d%%", hubInfo.Battery)

	case 1: // Мотор A
		value, _ := gui.hubMgr.GetSensorMonitor().GetSensorValue(1)
		return value

	case 2: // Мотор B
		value, _ := gui.hubMgr.GetSensorMonitor().GetSensorValue(2)
		return value

	case 3: // Светодиод
		value, _ := gui.hubMgr.GetSensorMonitor().GetSensorValue(6)
		return value

	case 4: // Температура (тест)
		return "23°C" // Заглушка

	default:
		return "--"
	}
}
