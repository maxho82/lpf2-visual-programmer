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

// DraggableBlock - перетаскиваемый блок
type DraggableBlock struct {
	widget.BaseWidget
	block      *ProgramBlock
	pm         *ProgramManager
	content    fyne.CanvasObject
	dragStart  fyne.Position
	blockStart fyne.Position
	isDragging bool
}

// NewDraggableBlock создает перетаскиваемый блок
func NewDraggableBlock(block *ProgramBlock, pm *ProgramManager) *DraggableBlock {
	d := &DraggableBlock{
		block: block,
		pm:    pm,
	}
	d.ExtendBaseWidget(d)
	d.createContent()
	return d
}

func (d *DraggableBlock) createContent() {
	// Фон блока
	bg := canvas.NewRectangle(d.block.Color)
	bg.SetMinSize(fyne.NewSize(float32(d.block.Width), float32(d.block.Height)))

	// Заголовок
	title := canvas.NewText(d.block.Title, color.White)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 14

	// Описание
	desc := canvas.NewText(d.block.Description, color.White)
	desc.Alignment = fyne.TextAlignCenter
	desc.TextSize = 10

	// Контейнер содержимого
	content := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(desc),
	)

	// Объединяем фон и содержимое
	d.content = container.NewStack(
		bg,
		container.NewPadded(content),
	)
}

// CreateRenderer создает рендерер
func (d *DraggableBlock) CreateRenderer() fyne.WidgetRenderer {
	return &draggableBlockRenderer{
		widget:  d,
		objects: []fyne.CanvasObject{d.content},
	}
}

// Tapped обрабатывает клик
func (d *DraggableBlock) Tapped(e *fyne.PointEvent) {
	log.Printf("Клик по блоку: %s (ID: %d)", d.block.Title, d.block.ID)
	d.pm.selected = d.block
	d.pm.ShowBlockProperties(d.block)
}

// DoubleTapped обрабатывает двойной клик
func (d *DraggableBlock) DoubleTapped(e *fyne.PointEvent) {
	// Можно добавить дополнительные действия
}

// Dragged обрабатывает перетаскивание – используем //Delta для плавного движения
func (d *DraggableBlock) Dragged(e *fyne.DragEvent) {
	if !d.isDragging {
		d.isDragging = true
		d.dragStart = d.Position()
		//e.Position = d.Position()
		return
	}

	// Вычисляем новую позицию, добавляя смещение мыши (Delta)
	newPos := fyne.NewPos(
		d.dragStart.X+e.Dragged.DX,
		//e.Position.X,
		d.dragStart.Y+e.Dragged.DY,
		//e.Position.Y,
	)

	// Ограничиваем движение в пределах положительных координат
	if newPos.X < 0 {
		newPos.X = 0
	}
	if newPos.Y < 0 {
		newPos.Y = 0
	}

	// Перемещаем виджет
	d.Move(newPos)
	d.dragStart = d.Position()
}

// DragEnd завершает перетаскивание
func (d *DraggableBlock) DragEnd() {
	if d.isDragging {
		// Обновляем позицию в данных блока
		currentPos := d.Position()
		d.block.X = float64(currentPos.X)
		d.block.Y = float64(currentPos.Y)
		d.isDragging = false

		log.Printf("Блок %s перемещен в позицию: %.0f, %.0f",
			d.block.Title, d.block.X, d.block.Y)
	}
}

// MouseIn изменяет курсор при наведении
func (d *DraggableBlock) MouseIn(e *desktop.MouseEvent) {
	// Курсор изменяется при наведении
}

// MouseOut сбрасывает курсор при уходе мыши
func (d *DraggableBlock) MouseOut() {
	// Курсор возвращается к стандартному
}

// Cursor возвращает курсор для виджета
func (d *DraggableBlock) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

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
