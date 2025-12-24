package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// DraggableBlockWidget - полноценный перетаскиваемый виджет блока
type DraggableBlockWidget struct {
	widget.BaseWidget
	block       *ProgramBlock
	pm          *ProgramManager
	content     fyne.CanvasObject
	dragStart   fyne.Position
	widgetStart fyne.Position
}

// NewDraggableBlockWidget создает новый перетаскиваемый виджет блока
func NewDraggableBlockWidget(block *ProgramBlock, pm *ProgramManager) *DraggableBlockWidget {
	d := &DraggableBlockWidget{
		block: block,
		pm:    pm,
	}
	d.ExtendBaseWidget(d)
	d.createContent()
	return d
}

func (d *DraggableBlockWidget) createContent() {
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

// CreateRenderer создает рендерер виджета
func (d *DraggableBlockWidget) CreateRenderer() fyne.WidgetRenderer {
	return &draggableBlockWidgetRenderer{
		widget:  d,
		objects: []fyne.CanvasObject{d.content},
	}
}

// Tapped обрабатывает клик на виджет
func (d *DraggableBlockWidget) Tapped(ev *fyne.PointEvent) {
	// При клике выбираем блок
	d.pm.selected = d.block
	d.pm.ShowBlockProperties(d.block)
}

// DoubleTapped обрабатывает двойной клик
func (d *DraggableBlockWidget) DoubleTapped(ev *fyne.PointEvent) {
	// Можно добавить дополнительные действия при двойном клике
}

// Dragged обрабатывает перетаскивание виджета
func (d *DraggableBlockWidget) Dragged(ev *fyne.DragEvent) {
	if d.dragStart.IsZero() {
		d.dragStart = ev.Position
		d.widgetStart = d.Position()
		return
	}

	// Вычисляем новую позицию
	deltaX := ev.Position.X - d.dragStart.X
	deltaY := ev.Position.Y - d.dragStart.Y
	newPos := fyne.NewPos(
		d.widgetStart.X+deltaX,
		d.widgetStart.Y+deltaY,
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

	// Обновляем позицию в данных блока
	d.block.X = float64(newPos.X)
	d.block.Y = float64(newPos.Y)

	// Обновляем отображение
	d.Refresh()
}

// DragEnd завершает перетаскивание
func (d *DraggableBlockWidget) DragEnd() {
	d.dragStart = fyne.NewPos(0, 0)
	d.widgetStart = fyne.NewPos(0, 0)
}

// MouseIn изменяет курсор при наведении
func (d *DraggableBlockWidget) MouseIn(ev *desktop.MouseEvent) {
	// Курсор изменяется при наведении
}

// MouseOut сбрасывает курсор при уходе мыши
func (d *DraggableBlockWidget) MouseOut() {
	// Курсор возвращается к стандартному
}

// MouseMoved отслеживает движение мыши
func (d *DraggableBlockWidget) MouseMoved(ev *desktop.MouseEvent) {
	// Можно добавить логику при движении мыши
}

// Cursor возвращает курсор для виджета
func (d *DraggableBlockWidget) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// Тип для рендерера
type draggableBlockWidgetRenderer struct {
	widget  *DraggableBlockWidget
	objects []fyne.CanvasObject
}

func (r *draggableBlockWidgetRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *draggableBlockWidgetRenderer) MinSize() fyne.Size {
	return r.objects[0].MinSize()
}

func (r *draggableBlockWidgetRenderer) Refresh() {
	r.objects[0].Refresh()
}

func (r *draggableBlockWidgetRenderer) Destroy() {}

func (r *draggableBlockWidgetRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
