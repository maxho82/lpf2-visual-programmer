package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// DraggableContainer добавляет перетаскивание к любому виджету
type DraggableContainer struct {
	widget.BaseWidget
	content    fyne.CanvasObject
	block      *ProgramBlock
	isDragging bool
	dragStart  fyne.Position
	blockStart fyne.Position
}

// NewDraggableContainer создает перетаскиваемый контейнер
func NewDraggableContainer(content fyne.CanvasObject, block *ProgramBlock) *DraggableContainer {
	d := &DraggableContainer{
		content: content,
		block:   block,
	}
	d.ExtendBaseWidget(d)
	return d
}

// CreateRenderer создает рендерер
func (d *DraggableContainer) CreateRenderer() fyne.WidgetRenderer {
	return &draggableRenderer{
		widget:  d,
		objects: []fyne.CanvasObject{d.content},
	}
}

// Tapped обрабатывает клик
func (d *DraggableContainer) Tapped(e *fyne.PointEvent) {
	// При клике выбираем блок (можно добавить логику позже)
}

// Dragged обрабатывает перетаскивание
func (d *DraggableContainer) Dragged(e *fyne.DragEvent) {
	if !d.isDragging {
		d.isDragging = true
		d.dragStart = e.Position
		d.blockStart = d.Position()
	}

	deltaX := e.Position.X - d.dragStart.X
	deltaY := e.Position.Y - d.dragStart.Y
	newPos := fyne.NewPos(d.blockStart.X+deltaX, d.blockStart.Y+deltaY)

	// Ограничиваем движение
	if newPos.X < 0 {
		newPos.X = 0
	}
	if newPos.Y < 0 {
		newPos.Y = 0
	}

	d.Move(newPos)

	// Обновляем позицию блока
	d.block.X = float64(newPos.X)
	d.block.Y = float64(newPos.Y)
}

// DragEnd завершает перетаскивание
func (d *DraggableContainer) DragEnd() {
	d.isDragging = false
}

// Cursor возвращает курсор
func (d *DraggableContainer) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// draggableRenderer - рендерер для DraggableContainer
type draggableRenderer struct {
	widget  *DraggableContainer
	objects []fyne.CanvasObject
}

func (r *draggableRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *draggableRenderer) MinSize() fyne.Size {
	return r.objects[0].MinSize()
}

func (r *draggableRenderer) Refresh() {
	r.objects[0].Refresh()
}

func (r *draggableRenderer) Destroy() {}

func (r *draggableRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
