package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BlockPropertyEditor редактор свойств блока
type BlockPropertyEditor struct {
	block     *ProgramBlock
	container *fyne.Container
	deviceMgr *DeviceManager
}

// NewBlockPropertyEditor создает редактор свойств блока
func NewBlockPropertyEditor(block *ProgramBlock, deviceMgr *DeviceManager) *BlockPropertyEditor {
	editor := &BlockPropertyEditor{
		block:     block,
		deviceMgr: deviceMgr,
	}

	editor.container = editor.buildUI()
	return editor
}

// buildUI строит интерфейс редактора
func (editor *BlockPropertyEditor) buildUI() *fyne.Container {
	mainContainer := container.NewVBox()

	title := widget.NewLabelWithStyle(
		"Настройки: "+editor.block.Title,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	mainContainer.Add(title)
	mainContainer.Add(widget.NewSeparator())

	switch editor.block.Type {
	case BlockTypeMotor:
		editor.addMotorControls(mainContainer)
	case BlockTypeLED:
		editor.addLEDControls(mainContainer)
	case BlockTypeWait:
		editor.addWaitControls(mainContainer)
	case BlockTypeLoop:
		editor.addLoopControls(mainContainer)
	case BlockTypeTiltSensor:
		editor.addTiltSensorControls(mainContainer)
	case BlockTypeDistanceSensor:
		editor.addDistanceSensorControls(mainContainer)
	case BlockTypePiezoTone:
		editor.addPiezoToneControls(mainContainer)
	case BlockTypeVoltageSensor, BlockTypeCurrentSensor:
		// Простые датчики
		editor.addSimpleSensorControls(mainContainer, editor.block.Type)
	}

	return mainContainer
}

// addSimpleSensorControls для простых датчиков
func (editor *BlockPropertyEditor) addSimpleSensorControls(mainContainer *fyne.Container, sensorType BlockType) {
	portSelect := widget.NewSelect([]string{"Порт A (1)", "Порт B (2)", "Порт C (6)"}, func(selected string) {
		switch selected {
		case "Порт A (1)":
			editor.block.Parameters["port"] = byte(1)
		case "Порт B (2)":
			editor.block.Parameters["port"] = byte(2)
		case "Порт C (6)":
			editor.block.Parameters["port"] = byte(6)
		}
	})
	portSelect.SetSelected("Порт A (1)")
	editor.block.Parameters["port"] = byte(1)

	mainContainer.Add(widget.NewLabel("Порт:"))
	mainContainer.Add(portSelect)
}

// addMotorControls добавляет элементы управления для мотора
func (editor *BlockPropertyEditor) addMotorControls(mainContainer *fyne.Container) {
	// Выбор порта
	portSelect := widget.NewSelect([]string{"Порт A (1)", "Порт B (2)"}, func(selected string) {
		if selected == "Порт A (1)" {
			editor.block.Parameters["port"] = byte(1)
		} else {
			editor.block.Parameters["port"] = byte(2)
		}
	})

	// Устанавливаем начальное значение
	if port, ok := editor.block.Parameters["port"].(byte); ok && port == 2 {
		portSelect.SetSelected("Порт B (2)")
	} else {
		portSelect.SetSelected("Порт A (1)")
		editor.block.Parameters["port"] = byte(1)
	}

	mainContainer.Add(widget.NewLabel("Порт:"))
	mainContainer.Add(portSelect)

	// Мощность (-100 до 100)
	powerLabel := widget.NewLabel("Мощность: 50%")
	powerSlider := widget.NewSlider(-100, 100)

	// Устанавливаем начальное значение
	if power, ok := editor.block.Parameters["power"].(int8); ok {
		powerSlider.Value = float64(power)
		powerLabel.SetText(fmt.Sprintf("Мощность: %d%%", power))
	} else {
		powerSlider.Value = 50
		editor.block.Parameters["power"] = int8(50)
	}

	powerSlider.OnChanged = func(value float64) {
		editor.block.Parameters["power"] = int8(value)
		powerLabel.SetText(fmt.Sprintf("Мощность: %d%%", int(value)))
	}
	mainContainer.Add(powerLabel)
	mainContainer.Add(powerSlider)

	// Длительность
	durationEntry := widget.NewEntry()
	if duration, ok := editor.block.Parameters["duration"].(uint16); ok {
		durationEntry.SetText(fmt.Sprintf("%d", duration))
	} else {
		durationEntry.SetText("1000")
		editor.block.Parameters["duration"] = uint16(1000)
	}

	durationEntry.OnChanged = func(text string) {
		if dur, err := strconv.ParseUint(text, 10, 16); err == nil {
			editor.block.Parameters["duration"] = uint16(dur)
		}
	}
	mainContainer.Add(widget.NewLabel("Длительность (мс):"))
	mainContainer.Add(durationEntry)

	// Кнопка теста
	testButton := widget.NewButton("Тест", func() {
		port := editor.block.Parameters["port"].(byte)
		power := editor.block.Parameters["power"].(int8)
		duration := editor.block.Parameters["duration"].(uint16)

		// Используем переменные, чтобы избежать ошибки "declared and not used"
		if editor.deviceMgr != nil {
			_ = editor.deviceMgr.SetMotorPower(port, power, duration)
		}
	})
	mainContainer.Add(testButton)
}

// addLEDControls добавляет элементы управления для светодиода
func (editor *BlockPropertyEditor) addLEDControls(mainContainer *fyne.Container) {
	// Выбор порта (обычно порт 6 для встроенного светодиода)
	portSelect := widget.NewSelect([]string{"Встроенный (6)"}, func(selected string) {
		editor.block.Parameters["port"] = byte(6)
	})
	portSelect.SetSelected("Встроенный (6)")
	mainContainer.Add(widget.NewLabel("Порт:"))
	mainContainer.Add(portSelect)

	// Выбор цвета
	colorLabel := widget.NewLabel("Цвет:")
	mainContainer.Add(colorLabel)

	// Создаем контейнер для палитры цветов
	colorContainer := container.NewVBox()
	colors := []struct {
		name    string
		r, g, b byte
	}{
		{"Красный", 255, 0, 0},
		{"Зеленый", 0, 255, 0},
		{"Синий", 0, 0, 255},
		{"Желтый", 255, 255, 0},
		{"Фиолетовый", 255, 0, 255},
		{"Голубой", 0, 255, 255},
		{"Белый", 255, 255, 255},
	}

	// Создаем строки для палитры
	for i := 0; i < len(colors); i += 3 {
		row := container.NewHBox()
		for j := i; j < i+3 && j < len(colors); j++ {
			color := colors[j]
			btn := widget.NewButton("", func(r, g, b byte) func() {
				return func() {
					editor.block.Parameters["red"] = r
					editor.block.Parameters["green"] = g
					editor.block.Parameters["blue"] = b
				}
			}(color.r, color.g, color.b))
			btn.Importance = widget.LowImportance
			row.Add(btn)
		}
		colorContainer.Add(row)
	}

	mainContainer.Add(colorContainer)

	// Индивидуальные слайдеры RGB
	redLabel := widget.NewLabel("Красный: 255")
	redSlider := widget.NewSlider(0, 255)
	if red, ok := editor.block.Parameters["red"].(byte); ok {
		redSlider.Value = float64(red)
		redLabel.SetText(fmt.Sprintf("Красный: %d", red))
	} else {
		redSlider.Value = 255
		editor.block.Parameters["red"] = byte(255)
	}
	redSlider.OnChanged = func(value float64) {
		editor.block.Parameters["red"] = byte(value)
		redLabel.SetText(fmt.Sprintf("Красный: %d", int(value)))
	}
	mainContainer.Add(redLabel)
	mainContainer.Add(redSlider)

	greenLabel := widget.NewLabel("Зеленый: 0")
	greenSlider := widget.NewSlider(0, 255)
	if green, ok := editor.block.Parameters["green"].(byte); ok {
		greenSlider.Value = float64(green)
		greenLabel.SetText(fmt.Sprintf("Зеленый: %d", green))
	} else {
		greenSlider.Value = 0
		editor.block.Parameters["green"] = byte(0)
	}
	greenSlider.OnChanged = func(value float64) {
		editor.block.Parameters["green"] = byte(value)
		greenLabel.SetText(fmt.Sprintf("Зеленый: %d", int(value)))
	}
	mainContainer.Add(greenLabel)
	mainContainer.Add(greenSlider)

	blueLabel := widget.NewLabel("Синий: 0")
	blueSlider := widget.NewSlider(0, 255)
	if blue, ok := editor.block.Parameters["blue"].(byte); ok {
		blueSlider.Value = float64(blue)
		blueLabel.SetText(fmt.Sprintf("Синий: %d", blue))
	} else {
		blueSlider.Value = 0
		editor.block.Parameters["blue"] = byte(0)
	}
	blueSlider.OnChanged = func(value float64) {
		editor.block.Parameters["blue"] = byte(value)
		blueLabel.SetText(fmt.Sprintf("Синий: %d", int(value)))
	}
	mainContainer.Add(blueLabel)
	mainContainer.Add(blueSlider)

	// Кнопка теста
	testButton := widget.NewButton("Тест", func() {
		port := editor.block.Parameters["port"].(byte)
		red := editor.block.Parameters["red"].(byte)
		green := editor.block.Parameters["green"].(byte)
		blue := editor.block.Parameters["blue"].(byte)

		// Используем переменные, чтобы избежать ошибки "declared and not used"
		if editor.deviceMgr != nil {
			_ = editor.deviceMgr.SetLEDColor(port, red, green, blue)
		}
	})
	mainContainer.Add(testButton)
}

// addWaitControls добавляет элементы управления для блока ожидания
func (editor *BlockPropertyEditor) addWaitControls(mainContainer *fyne.Container) {
	// Длительность ожидания
	durationEntry := widget.NewEntry()
	if duration, ok := editor.block.Parameters["duration"].(float64); ok {
		durationEntry.SetText(fmt.Sprintf("%.1f", duration))
	} else {
		durationEntry.SetText("1.0")
		editor.block.Parameters["duration"] = 1.0
	}

	durationEntry.OnChanged = func(text string) {
		if dur, err := strconv.ParseFloat(text, 64); err == nil {
			editor.block.Parameters["duration"] = dur
		}
	}
	mainContainer.Add(widget.NewLabel("Длительность (сек):"))
	mainContainer.Add(durationEntry)
}

// addLoopControls добавляет элементы управления для блока цикла
func (editor *BlockPropertyEditor) addLoopControls(mainContainer *fyne.Container) {
	// Выбор типа цикла
	loopType := widget.NewRadioGroup([]string{"Определенное число раз", "Бесконечно"}, func(selected string) {
		editor.block.Parameters["forever"] = (selected == "Бесконечно")
	})

	if forever, ok := editor.block.Parameters["forever"].(bool); ok && forever {
		loopType.SetSelected("Бесконечно")
	} else {
		loopType.SetSelected("Определенное число раз")
		editor.block.Parameters["forever"] = false
	}

	mainContainer.Add(widget.NewLabel("Тип цикла:"))
	mainContainer.Add(loopType)

	// Количество повторений (если не бесконечно)
	countEntry := widget.NewEntry()
	if count, ok := editor.block.Parameters["count"].(int); ok {
		countEntry.SetText(fmt.Sprintf("%d", count))
	} else {
		countEntry.SetText("5")
		editor.block.Parameters["count"] = 5
	}

	countEntry.OnChanged = func(text string) {
		if count, err := strconv.Atoi(text); err == nil {
			editor.block.Parameters["count"] = count
		}
	}
	mainContainer.Add(widget.NewLabel("Количество повторений:"))
	mainContainer.Add(countEntry)
}

// GetContainer возвращает контейнер редактора
func (editor *BlockPropertyEditor) GetContainer() *fyne.Container {
	return editor.container
}

// addTiltSensorControls добавляет элементы управления для датчика наклона
func (editor *BlockPropertyEditor) addTiltSensorControls(mainContainer *fyne.Container) {
	// Выбор порта
	portSelect := widget.NewSelect([]string{"Порт A (1)", "Порт B (2)", "Порт C (6)"}, func(selected string) {
		switch selected {
		case "Порт A (1)":
			editor.block.Parameters["port"] = byte(1)
		case "Порт B (2)":
			editor.block.Parameters["port"] = byte(2)
		case "Порт C (6)":
			editor.block.Parameters["port"] = byte(6)
		}
	})
	portSelect.SetSelected("Порт A (1)")
	editor.block.Parameters["port"] = byte(1)

	// Выбор режима
	modeSelect := widget.NewSelect([]string{
		"Угол наклона (0)",
		"Определение наклона (1)",
		"Определение удара (2)",
	}, func(selected string) {
		switch selected {
		case "Угол наклона (0)":
			editor.block.Parameters["mode"] = byte(0)
		case "Определение наклона (1)":
			editor.block.Parameters["mode"] = byte(1)
		case "Определение удара (2)":
			editor.block.Parameters["mode"] = byte(2)
		}
	})
	modeSelect.SetSelected("Определение наклона (1)")
	editor.block.Parameters["mode"] = byte(1)

	mainContainer.Add(widget.NewLabel("Порт:"))
	mainContainer.Add(portSelect)
	mainContainer.Add(widget.NewLabel("Режим:"))
	mainContainer.Add(modeSelect)
}

// addDistanceSensorControls добавляет элементы управления для датчика расстояния
func (editor *BlockPropertyEditor) addDistanceSensorControls(mainContainer *fyne.Container) {
	portSelect := widget.NewSelect([]string{"Порт A (1)", "Порт B (2)", "Порт C (6)"}, func(selected string) {
		switch selected {
		case "Порт A (1)":
			editor.block.Parameters["port"] = byte(1)
		case "Порт B (2)":
			editor.block.Parameters["port"] = byte(2)
		case "Порт C (6)":
			editor.block.Parameters["port"] = byte(6)
		}
	})
	portSelect.SetSelected("Порт A (1)")
	editor.block.Parameters["port"] = byte(1)

	// Выбор режима датчика расстояния
	modeSelect := widget.NewSelect([]string{
		"Измерение расстояния (0)",
		"Подсчет объектов (1)",
	}, func(selected string) {
		switch selected {
		case "Измерение расстояния (0)":
			editor.block.Parameters["mode"] = byte(0)
		case "Подсчет объектов (1)":
			editor.block.Parameters["mode"] = byte(1)
		}
	})
	modeSelect.SetSelected("Измерение расстояния (0)")
	editor.block.Parameters["mode"] = byte(0)

	mainContainer.Add(widget.NewLabel("Порт:"))
	mainContainer.Add(portSelect)
	mainContainer.Add(widget.NewLabel("Режим:"))
	mainContainer.Add(modeSelect)
}

// addPiezoToneControls добавляет элементы управления для пищалки
func (editor *BlockPropertyEditor) addPiezoToneControls(mainContainer *fyne.Container) {
	portSelect := widget.NewSelect([]string{"Порт A (1)", "Порт B (2)", "Порт C (6)"}, func(selected string) {
		switch selected {
		case "Порт A (1)":
			editor.block.Parameters["port"] = byte(1)
		case "Порт B (2)":
			editor.block.Parameters["port"] = byte(2)
		case "Порт C (6)":
			editor.block.Parameters["port"] = byte(6)
		}
	})
	portSelect.SetSelected("Порт A (1)")
	editor.block.Parameters["port"] = byte(1)

	// Частота
	freqEntry := widget.NewEntry()
	freqEntry.SetText("440")
	freqEntry.OnChanged = func(text string) {
		if freq, err := strconv.ParseUint(text, 10, 16); err == nil {
			editor.block.Parameters["frequency"] = uint16(freq)
		}
	}
	editor.block.Parameters["frequency"] = uint16(440)

	// Длительность
	durationEntry := widget.NewEntry()
	durationEntry.SetText("1000")
	durationEntry.OnChanged = func(text string) {
		if dur, err := strconv.ParseUint(text, 10, 16); err == nil {
			editor.block.Parameters["duration"] = uint16(dur)
		}
	}
	editor.block.Parameters["duration"] = uint16(1000)

	// Кнопка теста
	testButton := widget.NewButton("Тест звука", func() {
		port := editor.block.Parameters["port"].(byte)
		freq := editor.block.Parameters["frequency"].(uint16)
		dur := editor.block.Parameters["duration"].(uint16)

		if editor.deviceMgr != nil {
			_ = editor.deviceMgr.PlayTone(port, freq, dur)
		}
	})

	mainContainer.Add(widget.NewLabel("Порт:"))
	mainContainer.Add(portSelect)
	mainContainer.Add(widget.NewLabel("Частота (Гц):"))
	mainContainer.Add(freqEntry)
	mainContainer.Add(widget.NewLabel("Длительность (мс):"))
	mainContainer.Add(durationEntry)
	mainContainer.Add(testButton)
}
