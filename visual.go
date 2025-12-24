package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// ProgramManager управляет визуальным программированием
type ProgramManager struct {
	hubMgr      *HubManager
	deviceMgr   *DeviceManager
	blocks      []*ProgramBlock
	connections []*Connection
	canvas      *container.Scroll
	selected    *ProgramBlock
	programName string
	isRunning   bool
}

// ProgramBlock представляет блок программы
type ProgramBlock struct {
	ID          int
	Type        BlockType
	Title       string
	Description string
	X, Y        float64
	Width       float64
	Height      float64
	Parameters  map[string]interface{}
	NextBlockID int // ID следующего блока
	IsStart     bool
	Color       color.Color
	OnExecute   func() error // Функция выполнения
}

// Connection соединяет два блока
type Connection struct {
	FromBlockID int
	FromPort    int
	ToBlockID   int
	ToPort      int
}

// BlockType тип блока
type BlockType int

const (
	BlockTypeStart BlockType = iota
	BlockTypeMotor
	BlockTypeLED
	BlockTypeWait
	BlockTypeLoop
	BlockTypeCondition
	BlockTypeSensor
	BlockTypeSound
	BlockTypeStop
)

// NewProgramManager создает менеджер программ
func NewProgramManager(hubMgr *HubManager) *ProgramManager {
	deviceMgr := NewDeviceManager(hubMgr)

	pm := &ProgramManager{
		hubMgr:      hubMgr,
		deviceMgr:   deviceMgr,
		blocks:      make([]*ProgramBlock, 0),
		programName: "Новая программа",
	}

	// Создание холста
	pm.canvas = pm.createCanvas()

	return pm
}

/* // createCanvas создает холст для программирования
func (pm *ProgramManager) createCanvas() *container.Scroll {
	// Создаем контейнер с прокруткой
	content := container.NewWithoutLayout()

	// Добавляем сетку для привязки
	grid := canvas.NewRectangle(color.Transparent)
	grid.StrokeColor = color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	grid.StrokeWidth = 1
	grid.SetMinSize(fyne.NewSize(2000, 2000))
	content.Add(grid)

	return container.NewScroll(content)
} */

// createCanvas создает холст для программирования
func (pm *ProgramManager) createCanvas() *container.Scroll {
	// Создаем контейнер без компоновки (для ручного позиционирования)
	content := container.NewWithoutLayout()

	// Создаем и добавляем сетку
	grid := pm.createGrid()
	content.Add(grid)

	// Создаем прокручиваемый контейнер
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(800, 600))

	return scroll
}

// createGrid создает сетку для холста
func (pm *ProgramManager) createGrid() fyne.CanvasObject {
	// Создаем прямоугольник для фона сетки
	grid := canvas.NewRectangle(color.NRGBA{R: 245, G: 245, B: 245, A: 255})
	grid.SetMinSize(fyne.NewSize(2000, 2000)) // Устанавливаем размер для фона

	// Линии сетки
	lines := container.NewWithoutLayout()

	// Вертикальные линии
	for x := 0; x <= 2000; x += 20 {
		line := canvas.NewLine(color.NRGBA{R: 220, G: 220, B: 220, A: 255})
		line.Position1 = fyne.NewPos(float32(x), 0)
		line.Position2 = fyne.NewPos(float32(x), 2000)
		line.StrokeWidth = 1
		lines.Add(line)
	}

	// Горизонтальные линии
	for y := 0; y <= 2000; y += 20 {
		line := canvas.NewLine(color.NRGBA{R: 220, G: 220, B: 220, A: 255})
		line.Position1 = fyne.NewPos(0, float32(y))
		line.Position2 = fyne.NewPos(2000, float32(y))
		line.StrokeWidth = 1
		lines.Add(line)
	}

	// Возвращаем стек с фоном и линиями
	return container.NewStack(grid, lines)
}

// AddBlock добавляет новый блок на холст
func (pm *ProgramManager) AddBlock(blockType BlockType, x, y float64) *ProgramBlock {
	block := &ProgramBlock{
		ID:          len(pm.blocks) + 1,
		Type:        blockType,
		X:           x,
		Y:           y,
		Width:       150,
		Height:      80,
		Parameters:  make(map[string]interface{}),
		NextBlockID: -1,
	}

	// Настройка блока в зависимости от типа
	switch blockType {
	case BlockTypeStart:
		block.Title = "Начать"
		block.Description = "Начало программы"
		block.Color = color.NRGBA{R: 76, G: 175, B: 80, A: 255}
		block.IsStart = true
		block.OnExecute = func() error {
			// Начало программы
			return nil
		}

	case BlockTypeMotor:
		block.Title = "Мотор"
		block.Description = "Управление мотором"
		block.Color = color.NRGBA{R: 33, G: 150, B: 243, A: 255}
		block.Parameters["port"] = byte(1)
		block.Parameters["power"] = int8(50)
		block.Parameters["duration"] = uint16(1000)
		block.OnExecute = func() error {
			port := block.Parameters["port"].(byte)
			power := block.Parameters["power"].(int8)
			duration := block.Parameters["duration"].(uint16)
			return pm.deviceMgr.SetMotorPower(port, power, duration)
		}

	case BlockTypeLED:
		block.Title = "Светодиод"
		block.Description = "Управление RGB светодиодом"
		block.Color = color.NRGBA{R: 255, G: 193, B: 7, A: 255}
		block.Parameters["port"] = byte(6)
		block.Parameters["red"] = byte(255)
		block.Parameters["green"] = byte(0)
		block.Parameters["blue"] = byte(0)
		block.OnExecute = func() error {
			port := block.Parameters["port"].(byte)
			red := block.Parameters["red"].(byte)
			green := block.Parameters["green"].(byte)
			blue := block.Parameters["blue"].(byte)
			return pm.deviceMgr.SetLEDColor(port, red, green, blue)
		}

	case BlockTypeWait:
		block.Title = "Ждать"
		block.Description = "Пауза в программе"
		block.Color = color.NRGBA{R: 158, G: 158, B: 158, A: 255}
		block.Parameters["duration"] = 1.0 // секунды
		block.OnExecute = func() error {
			duration := block.Parameters["duration"].(float64)
			time.Sleep(time.Duration(duration*1000) * time.Millisecond)
			return nil
		}

	case BlockTypeLoop:
		block.Title = "Повторять"
		block.Description = "Цикл повторений"
		block.Color = color.NRGBA{R: 156, G: 39, B: 176, A: 255}
		block.Parameters["count"] = 5
		block.Parameters["forever"] = false

	case BlockTypeStop:
		block.Title = "Стоп"
		block.Description = "Остановка программы"
		block.Color = color.NRGBA{R: 244, G: 67, B: 54, A: 255}
		block.OnExecute = func() error {
			pm.StopProgram()
			return nil
		}
	}

	pm.blocks = append(pm.blocks, block)
	pm.updateCanvas()

	return block
}

// updateCanvas обновляет отображение холста
func (pm *ProgramManager) updateCanvas() {
	log.Println("Обновление холста...")

	// Получаем контейнер холста
	content := pm.canvas.Content.(*fyne.Container)

	// Удаляем все старые виджеты блоков (кроме фона)
	// Находим и сохраняем фон (первый элемент)
	var background fyne.CanvasObject
	if len(content.Objects) > 0 {
		background = content.Objects[0]
	}

	// Очищаем контейнер
	content.RemoveAll()

	// Добавляем фон обратно
	if background != nil {
		content.Add(background)
	}

	// Добавляем все блоки
	for _, block := range pm.blocks {
		blockWidget := pm.CreateBlockWidget(block)
		content.Add(blockWidget)
	}

	// Обновляем отображение
	content.Refresh()
	pm.canvas.Refresh()
}

// CreateBlockWidget создает виджет для блока
func (pm *ProgramManager) CreateBlockWidget(block *ProgramBlock) fyne.CanvasObject {
	// Создаем содержимое блока
	content := pm.createBlockContent(block)

	// Обертываем в перетаскиваемый контейнер
	draggable := NewDraggableBlock(block, pm, content)

	// Устанавливаем размер и позицию
	draggable.Resize(fyne.NewSize(float32(block.Width), float32(block.Height)))
	draggable.Move(fyne.NewPos(float32(block.X), float32(block.Y)))

	return draggable
}

// createBlockContent создает содержимое блока
func (pm *ProgramManager) createBlockContent(block *ProgramBlock) fyne.CanvasObject {
	// Фон блока
	bg := canvas.NewRectangle(block.Color)
	bg.SetMinSize(fyne.NewSize(float32(block.Width), float32(block.Height)))

	// Заголовок
	title := canvas.NewText(block.Title, color.White)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 14

	// Описание
	desc := canvas.NewText(block.Description, color.White)
	desc.Alignment = fyne.TextAlignCenter
	desc.TextSize = 10

	// Создаем контейнер
	return container.NewStack(
		bg,
		container.NewVBox(
			container.NewCenter(title),
			container.NewCenter(desc),
		),
	)
}

// GetCanvas возвращает холст
func (pm *ProgramManager) GetCanvas() fyne.CanvasObject {
	return pm.canvas
}

// RunProgram запускает выполнение программы
func (pm *ProgramManager) RunProgram() error {
	if pm.isRunning {
		return fmt.Errorf("программа уже выполняется")
	}

	if !pm.hubMgr.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	pm.isRunning = true

	// Находим стартовый блок
	var startBlock *ProgramBlock
	for _, block := range pm.blocks {
		if block.IsStart {
			startBlock = block
			break
		}
	}

	if startBlock == nil {
		return fmt.Errorf("стартовый блок не найден")
	}

	// Запускаем выполнение в отдельной горутине
	go pm.executeBlock(startBlock)

	return nil
}

// executeBlock выполняет блок и переходит к следующему
func (pm *ProgramManager) executeBlock(block *ProgramBlock) {
	if !pm.isRunning || block == nil {
		return
	}

	// Выполняем блок
	if block.OnExecute != nil {
		if err := block.OnExecute(); err != nil {
			log.Printf("Ошибка выполнения блока %d: %v", block.ID, err)
		}
	}

	// Ищем следующий блок
	var nextBlock *ProgramBlock
	if block.NextBlockID > 0 {
		for _, b := range pm.blocks {
			if b.ID == block.NextBlockID {
				nextBlock = b
				break
			}
		}
	}

	// Рекурсивно выполняем следующий блок
	if nextBlock != nil {
		time.Sleep(100 * time.Millisecond) // Небольшая задержка между блоками
		pm.executeBlock(nextBlock)
	} else {
		pm.isRunning = false
	}
}

// StopProgram останавливает выполнение программы
func (pm *ProgramManager) StopProgram() {
	pm.isRunning = false

	// Останавливаем все моторы
	for _, device := range pm.deviceMgr.GetDevices() {
		if device.DeviceType == 0x01 { // Мотор
			pm.deviceMgr.StopMotor(device.PortID)
		}
	}
}

// ClearProgram очищает программу
func (pm *ProgramManager) ClearProgram() {
	pm.blocks = make([]*ProgramBlock, 0)
	pm.connections = make([]*Connection, 0)
	pm.updateCanvas()
}

// ShowBlockProperties показывает свойства выбранного блока
func (pm *ProgramManager) ShowBlockProperties(block *ProgramBlock) {
	// TODO: Реализовать панель свойств блока
	// Временно: просто устанавливаем выбранный блок
	pm.selected = block
}

// SaveProgram сохраняет программу в файл
func (pm *ProgramManager) SaveProgram(filename string) error {
	// TODO: Реализовать сохранение программы
	return nil
}

// LoadProgram загружает программу из файла
func (pm *ProgramManager) LoadProgram(filename string) error {
	// TODO: Реализовать загрузку программы
	return nil
}
