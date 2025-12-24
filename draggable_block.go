package main

import (
	"log"

	"fyne.io/fyne/v2"
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
func NewDraggableBlock(block *ProgramBlock, pm *ProgramManager, content fyne.CanvasObject) *DraggableBlock {
	d := &DraggableBlock{
		block:   block,
		pm:      pm,
		content: content,
	}
	d.ExtendBaseWidget(d)
	return d
}

// CreateRenderer создает рендерер
func (d *DraggableBlock) CreateRenderer() fyne.WidgetRenderer {
	return &draggableBlockRenderer{
		widget:  d,
		objects: []fyne.CanvasObject{d.content},
	}
}

// Tapped обрабатывает клик - ВАЖНО: функция должна вызываться!
func (d *DraggableBlock) Tapped(e *fyne.PointEvent) {
	log.Printf("Клик по блоку: %s (ID: %d)", d.block.Title, d.block.ID)

	// Устанавливаем выбранный блок
	d.pm.selected = d.block
	d.pm.ShowBlockProperties(d.block)

	// Обновляем выделение (можно добавить визуальное выделение)
	d.Refresh()
}

// TappedSecondary обрабатывает правый клик
func (d *DraggableBlock) TappedSecondary(e *fyne.PointEvent) {
	// Можно добавить контекстное меню
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

	// Ограничиваем движение
	if newPos.X < 0 {
		newPos.X = 0
	}
	if newPos.Y < 0 {
		newPos.Y = 0
	}

	// Перемещаем виджет
	d.Move(newPos)

	// Обновляем позицию в данных блока
	d.block.X = float64(newPos.X)
	d.block.Y = float64(newPos.Y)

	// Обновляем виджет
	d.Refresh()
}

// DragEnd завершает перетаскивание
func (d *DraggableBlock) DragEnd() {
	d.isDragging = false
	log.Printf("Блок %s перемещен в позицию: %.0f, %.0f",
		d.block.Title, d.block.X, d.block.Y)
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
	// Обновляем содержимое
	r.objects[0].Refresh()
}

func (r *draggableBlockRenderer) Destroy() {}

func (r *draggableBlockRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
