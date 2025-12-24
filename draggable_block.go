package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// DraggableBlock - контейнер для перетаскивания блока
type DraggableBlock struct {
	widget.BaseWidget
	block      *ProgramBlock
	content    fyne.CanvasObject
	dragStartX float32
	dragStartY float32
	isDragging bool
}

// NewDraggableBlock создает перетаскиваемый блок
func NewDraggableBlock(block *ProgramBlock, content fyne.CanvasObject) *DraggableBlock {
	d := &DraggableBlock{
		block:   block,
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

// Tapped обрабатывает клик
func (d *DraggableBlock) Tapped(e *fyne.PointEvent) {
	// Выбираем блок при клике
	// Можно добавить подсветку выбранного блока
}

// Dragged обрабатывает перетаскивание
func (d *DraggableBlock) Dragged(e *fyne.DragEvent) {
	if !d.isDragging {
		d.isDragging = true
		d.dragStartX = d.Position().X - e.Position.X
		d.dragStartY = d.Position().Y - e.Position.Y
	}

	newX := e.Position.X + d.dragStartX
	newY := e.Position.Y + d.dragStartY

	// Ограничиваем движение в пределах положительных координат
	if newX < 0 {
		newX = 0
	}
	if newY < 0 {
		newY = 0
	}

	d.Move(fyne.NewPos(newX, newY))

	// Обновляем позицию блока
	d.block.X = float64(newX)
	d.block.Y = float64(newY)
}

// DragEnd завершает перетаскивание
func (d *DraggableBlock) DragEnd() {
	d.isDragging = false
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

func (r *draggableBlockRenderer) Refresh() {}

func (r *draggableBlockRenderer) Destroy() {}

func (r *draggableBlockRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
