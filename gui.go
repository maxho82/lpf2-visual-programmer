package main

import (
	"fmt"
	"image/color"
	"log"
	"strconv" // Добавляем для hexStringToBytes
	"strings" // Добавляем для hexStringToBytes
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
	//scrollContainer *container.Scroll // ДОБАВЛЕНО - прямой доступ к Scroll
	programPanel *container.Scroll
	blocksPanel  fyne.CanvasObject

	// Для обновления устройств
	deviceUpdateRequest chan bool
	//deviceUpdateTicker      *time.Ticker

	dynamicDevicesContainer *fyne.Container
	portWidgets             map[byte]*widget.Label // Для хранения меток статуса портов
}

// NewGUI создает новый GUI
func NewGUI(window fyne.Window, hubMgr *HubManager, programMgr *ProgramManager) *GUI {
	gui := &GUI{
		window:              window,
		hubMgr:              hubMgr,
		programMgr:          programMgr,
		deviceUpdateRequest: make(chan bool, 10), // Буферизованный канал
		portWidgets:         make(map[byte]*widget.Label),
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
	leftSplit.SetOffset(0.25)

	rightSplit := container.NewHSplit(leftSplit, gui.propertiesPanel)
	rightSplit.SetOffset(0.75)

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

	// Запускаем обработчик обновлений устройств
	go gui.deviceUpdateHandler()

	// Настраиваем callback в DeviceManager
	if gui.programMgr != nil && gui.programMgr.deviceMgr != nil {
		gui.programMgr.deviceMgr.SetDeviceChangedCallback(func(portID byte, device *Device) {
			// Отправляем запрос на обновление GUI
			select {
			case gui.deviceUpdateRequest <- true:
			default:
				// Канал полон, пропускаем
			}
		})
	}

	return fullLayout
}

// deviceUpdateHandler обрабатывает запросы на обновление устройств
func (gui *GUI) deviceUpdateHandler() {
	// Запускаем тикер для периодического обновления батареи (раз в 10 секунд)
	batteryTicker := time.NewTicker(10 * time.Second)
	defer batteryTicker.Stop()

	for {
		select {
		case <-gui.deviceUpdateRequest:
			// Обновляем устройства в GUI
			gui.updateDeviceDisplay()

		case <-batteryTicker.C:
			// Периодически обновляем только батарею
			if gui.hubMgr != nil && gui.hubMgr.IsConnected() {
				// Обновление батареи происходит автоматически в createBatteryWidget
				// Ничего дополнительно не делаем
			}

		case <-time.After(100 * time.Millisecond):
			// Неблокирующий цикл
		}
	}
}

// updateBatteryDisplay обновляет только отображение батареи
// (этот метод больше не нужен, т.к. батарея обновляется автоматически)
// Если хотим оставить, исправляем:
/* func (gui *GUI) updateBatteryDisplay() {
	// Этот метод теперь пустой, т.к. батарея обновляется автоматически
	// в createBatteryWidget через горутину
} */

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

	testProtocolButton := widget.NewButtonWithIcon("Тест протокола", theme.VisibilityIcon(), func() {
		gui.showProtocolTestDialog()
	})

	// Метка статусаЫ
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
		testProtocolButton, // Добавляем эту кнопку
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

	// Основной контейнер
	mainContainer := container.NewVBox(
		container.NewCenter(title),
		widget.NewSeparator(),
	)

	// --- Батарея ---
	batteryWidget := gui.createBatteryWidget()
	mainContainer.Add(batteryWidget)

	// --- Информация о хабе ---
	hubInfoWidget := gui.createHubInfoWidget()
	mainContainer.Add(hubInfoWidget)

	// --- Динамические устройства из DeviceManager ---
	devicesTitle := widget.NewLabel("Подключенные устройства:")
	devicesTitle.TextStyle.Bold = true
	mainContainer.Add(devicesTitle)

	// Контейнер для динамических устройств с фиксированной минимальной высотой
	//gui.dynamicDevicesContainer = container.NewVBox()
	gui.dynamicDevicesContainer = container.NewVBox()
	// Создаем контейнер с минимальной высотой (примерно 5 строк по 30px = 150px)

	ScrollContainer := container.NewVScroll(gui.dynamicDevicesContainer)
	ScrollContainer.SetMinSize(fyne.NewSize(250, 300))
	mainContainer.Add(ScrollContainer)
	mainContainer.Add(widget.NewSeparator())

	/* 	// --- Фиксированные порты (для удобства) ---
	   	portsTitle := widget.NewLabel("Порты хаба:")
	   	portsTitle.TextStyle.Bold = true
	   	mainContainer.Add(portsTitle)

	   	// Порт 1 - Мотор A
	   	port1 := gui.createPortWidget(1, "Порт A (Мотор)")
	   	mainContainer.Add(port1)

	   	// Порт 2 - Мотор B
	   	port2 := gui.createPortWidget(2, "Порт B (Мотор)")
	   	mainContainer.Add(port2)

	   	// Порт 6 - Светодиод
	   	port6 := gui.createPortWidget(6, "Встроенный светодиод")
	   	mainContainer.Add(port6)

	   	mainContainer.Add(widget.NewSeparator()) */

	return container.NewVScroll(container.NewPadded(mainContainer))
}

// updateDeviceDisplay обновляет отображение устройств
func (gui *GUI) updateDeviceDisplay() {
	fyne.Do(func() {
		// Обновляем динамические устройства
		if gui.dynamicDevicesContainer != nil {
			gui.dynamicDevicesContainer.Objects = nil

			if gui.programMgr != nil && gui.programMgr.deviceMgr != nil {
				devices := gui.programMgr.deviceMgr.GetDevices()

				if len(devices) == 0 {
					noDevicesLabel := widget.NewLabel("Нет подключенных устройств")
					noDevicesLabel.TextStyle.Italic = true
					noDevicesLabel.Alignment = fyne.TextAlignCenter
					gui.dynamicDevicesContainer.Add(noDevicesLabel)
				} else {
					for _, device := range devices {
						if device.IsConnected {
							deviceCard := gui.createDeviceCard(device)
							gui.dynamicDevicesContainer.Add(deviceCard)
							gui.dynamicDevicesContainer.MinSize()
							gui.dynamicDevicesContainer.Refresh()
						}
					}
				}
			} else {
				noManagerLabel := widget.NewLabel("Менеджер устройств не инициализирован")
				noManagerLabel.TextStyle.Italic = true
				noManagerLabel.Alignment = fyne.TextAlignCenter
				gui.dynamicDevicesContainer.Add(noManagerLabel)
			}

			gui.dynamicDevicesContainer.Resize(fyne.NewSize(0, 300))
			gui.dynamicDevicesContainer.Refresh()
		}
	})
}

// createDeviceCard создает карточку устройства
func (gui *GUI) createDeviceCard(device Device) fyne.CanvasObject {
	// Определяем иконку по типу устройства
	var iconRes fyne.Resource
	switch device.DeviceType {
	case 0x01: // Мотор
		iconRes = theme.StorageIcon()
	case 0x17: // RGB светодиод
		iconRes = theme.VisibilityIcon()
	case 0x02: // Датчик наклона
		iconRes = theme.ViewRefreshIcon()
	default:
		iconRes = theme.ComputerIcon()
	}

	// Статус подключения
	statusText := "✓ Подключено"
	statusColor := color.NRGBA{R: 0, G: 200, B: 0, A: 255}
	if !device.IsConnected {
		statusText = "✗ Отключено"
		statusColor = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
	}

	statusLabel := canvas.NewText(statusText, statusColor)
	statusLabel.TextSize = 11

	// Основная информация
	mainInfo := container.NewHBox(
		widget.NewIcon(iconRes),
		widget.NewLabel(fmt.Sprintf("Порт %d: %s", device.PortID, device.Name)),
		layout.NewSpacer(),
		container.NewCenter(statusLabel),
	)

	// Дополнительные свойства
	propsContainer := container.NewVBox()
	// Показываем последние значения или свойства
	if device.LastValue != nil {
		valLabel := widget.NewLabel(fmt.Sprintf("Значение: %v", device.LastValue))
		valLabel.TextStyle.Italic = true
		propsContainer.Add(valLabel)
	}

	// Если есть свойства, показываем их
	if len(device.Properties) > 0 {
		for key, value := range device.Properties {
			propLabel := widget.NewLabel(fmt.Sprintf("  %s: %v", key, value))
			//propLabel.TextSize = 10
			propsContainer.Add(propLabel)
		}
	}

	// Время последнего обновления
	if !device.LastUpdate.IsZero() {
		updateLabel := widget.NewLabel(fmt.Sprintf("Обновлено: %s",
			device.LastUpdate.Format("15:04:05")))
		//updateLabel.TextSize = 9
		updateLabel.TextStyle.Italic = true
		propsContainer.Add(updateLabel)
	}
	propsContainer.Resize(fyne.NewSize(0, 150))
	return container.NewVBox(
		mainInfo,
		widget.NewSeparator(),
		propsContainer,
		widget.NewSeparator(),
	)
}

/* // createPortWidget создает виджет порта с динамическими данными
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
} */

/* // createPortWidget создает виджет порта (только статическая информация)
func (gui *GUI) createPortWidget(portID byte, label string) fyne.CanvasObject {
	// Иконка порта
	icon := widget.NewIcon(theme.StorageIcon())

	// Метка порта
	portLabel := widget.NewLabel(label)
	portLabel.Alignment = fyne.TextAlignCenter
	portLabel.TextStyle.Bold = true

	// Метка статуса
	statusLabel := widget.NewLabel("Готов к подключению")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Сохраняем ссылку на виджет статуса
	gui.portWidgets[portID] = statusLabel

	// Контейнер порта
	portContainer := container.NewVBox(
		container.NewCenter(icon),
		portLabel,
		statusLabel,
		widget.NewSeparator(),
	)

	// Запускаем обновление статуса только при изменениях
	go func(portID byte, statusLabel *widget.Label, icon *widget.Icon) {
		// Инициализируем начальное состояние
		var lastState string = ""

		for {
			time.Sleep(500 * time.Millisecond) // Проверяем раз в 500 мс

			currentState := ""
			if gui.programMgr != nil && gui.programMgr.deviceMgr != nil {
				if device, exists := gui.programMgr.deviceMgr.GetDevice(portID); exists {
					if device.IsConnected {
						currentState = fmt.Sprintf("✓ %s", device.Name)
					} else {
						currentState = "Не подключено"
					}
				}
			}

			// Обновляем только если состояние изменилось
			if currentState != lastState {
				lastState = currentState

				fyne.Do(func() {
					if currentState != "" {
						statusLabel.SetText(currentState)
						statusLabel.TextStyle.Bold = true
						icon.SetResource(theme.ConfirmIcon())
					} else {
						statusLabel.SetText("Готов к подключению")
						statusLabel.TextStyle.Bold = false
						icon.SetResource(theme.StorageIcon())
					}
					statusLabel.Refresh()
					icon.Refresh()
				})
			}
		}
	}(portID, statusLabel, icon)
	portContainer.Resize(fyne.NewSize(0, 150))
	return portContainer
} */

/* // formatPortValue форматирует значение порта
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
*/
/* // updateHubInfoDisplay обновляет отображение информации о хабе
func (gui *GUI) updateHubInfoDisplay() {
	// Этот метод может использоваться для принудительного обновления
	// информации о хабе, если потребуется
} */

// createBatteryWidget создает виджет батареи
func (gui *GUI) createBatteryWidget() fyne.CanvasObject {
	// Заголовок
	title := canvas.NewText("Батарея", color.NRGBA{R: 240, G: 240, B: 240, A: 255})
	title.TextSize = 14
	title.TextStyle.Bold = true

	// Прогресс-бар
	progress := widget.NewProgressBar()

	/* 	// Метка процентов
	   	percentLabel := widget.NewLabel("--%")
	   	percentLabel.Alignment = fyne.TextAlignCenter */

	// Контейнер
	batteryContainer := container.NewVBox(
		container.NewCenter(title),
		progress,
		//percentLabel,
		widget.NewSeparator(),
	)

	// Обновление состояния батареи (раз в 10 секунд)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if gui.hubMgr != nil && gui.hubMgr.IsConnected() {
				hubInfo := gui.hubMgr.GetHubInfo()
				batteryLevel := hubInfo.Battery

				fyne.Do(func() {
					progress.SetValue(float64(batteryLevel) / 100)
					//percentLabel.SetText(fmt.Sprintf("%d%%", batteryLevel))
					progress.Refresh()
					//percentLabel.Refresh()
				})
			}
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

				// Запускаем начальный опрос устройств
				if gui.hubMgr.GetSensorMonitor() != nil {
					go func() {
						time.Sleep(500 * time.Millisecond) // Даем время на установку соединения
						sensorMonitor := gui.hubMgr.GetSensorMonitor()
						if sensorMonitor != nil {
							// Вызываем начальный опрос
							sensorMonitor.UpdateDevices()

							// Запрашиваем обновление GUI
							select {
							case gui.deviceUpdateRequest <- true:
							default:
							}
						}
					}()
				}

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

/* // getSensorValue возвращает значение для конкретного датчика
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
} */

// showProtocolTestDialog показывает диалог тестирования протокола
func (gui *GUI) showProtocolTestDialog() {
	// Поле для UUID
	uuidEntry := widget.NewEntry()
	uuidEntry.SetPlaceHolder("UUID характеристики (например: 00001565-1212-efde-1523-785feabcd123)")
	uuidEntry.SetText("00001565-1212-efde-1523-785feabcd123") // Значение по умолчанию

	// Поле для данных (в hex)
	dataEntry := widget.NewEntry()
	dataEntry.SetPlaceHolder("Данные в hex (например: 08040603FF000000)")
	dataEntry.SetText("0102061700010000000201") // Пример: включить красный светодиод

	// Поле для результата
	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	// Кнопка отправки
	sendButton := widget.NewButton("Отправить", func() {
		uuid := uuidEntry.Text
		hexData := dataEntry.Text

		if uuid == "" || hexData == "" {
			resultLabel.SetText("Ошибка: заполните оба поля")
			return
		}

		// Преобразуем hex строку в байты
		data, err := hexStringToBytes(hexData)
		if err != nil {
			resultLabel.SetText(fmt.Sprintf("Ошибка преобразования данных: %v", err))
			return
		}

		//data1 := []byte{0x01, 0x02, 0x06, 0x17, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01}

		// Отправляем данные
		err = gui.hubMgr.WriteCharacteristic(uuid, data)
		if err != nil {
			resultLabel.SetText(fmt.Sprintf("Ошибка отправки: %v", err))
		} else {
			resultLabel.SetText(fmt.Sprintf("Успешно отправлено %d байт:\n%v", len(data), data))
		}
	})

	// Кнопка примеров
	examplesButton := widget.NewButton("Примеры", func() {
		dialog.ShowCustom("Примеры команд", "Закрыть",
			container.NewVBox(
				widget.NewLabel("Часто используемые UUID:"),
				widget.NewLabel("• 00001565-1212-efde-1523-785feabcd123 - Команды (Output)"),
				widget.NewLabel("• 00001563-1212-efde-1523-785feabcd123 - Настройка (Input)"),
				widget.NewSeparator(),
				widget.NewLabel("Примеры данных:"),
				widget.NewLabel("• 0102061700010000000201 - Включить режим светодиода ЛЕГО"),
				widget.NewLabel("• 0102061701010000000201 - Включить режим светодиода RGB"),
				widget.NewLabel("• 08040603FF000000 - Красный светодиод"),
				widget.NewLabel("• 0604040101 - Мотор A (50%)"),
				widget.NewLabel("• 0804060300000000 - Выключить светодиод"),
			), gui.window)
	})

	// Основной контейнер диалога
	content := container.NewVBox(
		widget.NewLabel("Тест отправки данных по протоколу LPF2"),
		widget.NewSeparator(),
		widget.NewLabel("UUID характеристики:"),
		uuidEntry,
		widget.NewLabel("Данные (hex):"),
		dataEntry,
		container.NewHBox(sendButton, examplesButton),
		widget.NewSeparator(),
		resultLabel,
	)

	dialog.ShowCustom("Тест протокола LPF2", "Закрыть", content, gui.window)
}

// Функция для преобразования hex строки в байты
func hexStringToBytes(hexStr string) ([]byte, error) {
	// Убираем пробелы
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	// Проверяем чётность длины
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("нечётная длина hex строки")
	}

	data := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		b, err := strconv.ParseUint(hexStr[i:i+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("неверный hex формат: %v", err)
		}
		data[i/2] = byte(b)
	}
	return data, nil
}
