package main

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// DraggableBlock - перетаскиваемый блок программы
type DraggableBlock struct {
	widget.BaseWidget
	block      *ProgramBlock
	pm         *ProgramManager
	content    fyne.CanvasObject
	isDragging bool
	dragStart  fyne.Position
	blockStart fyne.Position
}

// NewDraggableBlock создает новый перетаскиваемый блок
func NewDraggableBlock(block *ProgramBlock, pm *ProgramManager) *DraggableBlock {
	d := &DraggableBlock{
		block: block,
		pm:    pm,
	}
	d.ExtendBaseWidget(d)

	// Создаем содержимое блока
	d.createContent()

	return d
}

// createContent создает графическое представление блока
func (d *DraggableBlock) createContent() {
	// Заголовок блока
	title := canvas.NewText(d.block.Title, color.White)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 14

	// Описание блока
	description := canvas.NewText(d.block.Description, color.White)
	description.Alignment = fyne.TextAlignCenter
	description.TextSize = 10

	// Фон блока
	bg := canvas.NewRectangle(d.block.Color)
	bg.SetMinSize(fyne.NewSize(float32(d.block.Width), float32(d.block.Height)))

	// Контейнер с фоном и текстом
	bgContainer := container.NewStack(
		bg,
		container.NewVBox(
			container.NewCenter(title),
			container.NewCenter(description),
		),
	)

	// Добавляем порты для соединений
	ports := d.createPorts()

	// Объединяем все элементы
	d.content = container.NewStack(
		bgContainer,
		ports,
	)
}

// createPorts создает порты для соединения блоков
func (d *DraggableBlock) createPorts() fyne.CanvasObject {
	ports := container.NewWithoutLayout()

	// Входной порт (сверху)
	if !d.block.IsStart { // Стартовый блок не имеет входа
		inputPort := canvas.NewCircle(color.White)
		inputPort.StrokeColor = color.Black
		inputPort.StrokeWidth = 2
		inputPort.Resize(fyne.NewSize(15, 15))
		inputPort.Move(fyne.NewPos(
			float32(d.block.Width/2-7.5),
			0,
		))
		ports.Add(inputPort)
	}

	// Выходной порт (снизу)
	outputPort := canvas.NewCircle(color.White)
	outputPort.StrokeColor = color.Black
	outputPort.StrokeWidth = 2
	outputPort.Resize(fyne.NewSize(15, 15))
	outputPort.Move(fyne.NewPos(
		float32(d.block.Width/2-7.5),
		float32(d.block.Height-15),
	))
	ports.Add(outputPort)

	return ports
}

// CreateRenderer создает рендерер для виджета
func (d *DraggableBlock) CreateRenderer() fyne.WidgetRenderer {
	return &draggableBlockRenderer{
		widget:  d,
		objects: []fyne.CanvasObject{d.content},
	}
}

// Tapped обрабатывает клик на блок
func (d *DraggableBlock) Tapped(e *fyne.PointEvent) {
	d.pm.selected = d.block
	log.Printf("Выбран блок: %s (ID: %d)", d.block.Title, d.block.ID)
	// Временная реализация - просто логируем
	// TODO: Показать свойства блока в GUI
}

// Dragged обрабатывает перетаскивание
func (d *DraggableBlock) Dragged(e *fyne.DragEvent) {
	if !d.isDragging {
		d.isDragging = true
		d.dragStart = e.Position
		d.blockStart = d.Position()
	}

	// Вычисляем новую позицию
	deltaX := e.Position.X - d.dragStart.X
	deltaY := e.Position.Y - d.dragStart.Y
	newPos := fyne.NewPos(
		d.blockStart.X+deltaX,
		d.blockStart.Y+deltaY,
	)

	// Ограничиваем движение в пределах положительных координат
	if newPos.X < 0 {
		newPos.X = 0
	}
	if newPos.Y < 0 {
		newPos.Y = 0
	}

	d.Move(newPos)

	// Обновляем позицию в данных блока
	d.block.X = float64(newPos.X)
	d.block.Y = float64(newPos.Y)
}

// DragEnd завершает перетаскивание
func (d *DraggableBlock) DragEnd() {
	d.isDragging = false
}

// Cursor возвращает курсор для наведения
func (d *DraggableBlock) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// draggableBlockRenderer - рендерер для DraggableBlock
type draggableBlockRenderer struct {
	widget  *DraggableBlock
	objects []fyne.CanvasObject
}

func (r *draggableBlockRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *draggableBlockRenderer) MinSize() fyne.Size {
	return r.objects[0].MinSize()
}

func (r *draggableBlockRenderer) Refresh() {
	r.objects[0].Refresh()
}

func (r *draggableBlockRenderer) Destroy() {}

func (r *draggableBlockRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
