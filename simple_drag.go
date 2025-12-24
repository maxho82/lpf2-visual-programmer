package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// DragController - контроллер для управления перетаскиванием
type DragController struct {
	block          *ProgramBlock
	selected       bool
	dragOffsetX    float32
	dragOffsetY    float32
	isDragging     bool
}

// SimpleDraggableBlock - простой перетаскиваемый блок
type SimpleDraggableBlock struct {
	widget.BaseWidget
	controller *DragController
	content    fyne.CanvasObject
}

// NewSimpleDraggableBlock создает простой перетаскиваемый блок
func NewSimpleDraggableBlock(block *ProgramBlock, pm *ProgramManager) *SimpleDraggableBlock {
	s := &SimpleDraggableBlock{
		controller: &DragController{
			block: block,
		},
	}
	s.ExtendBaseWidget(s)
	s.createContent(pm)
	return s
}

func (s *SimpleDraggableBlock) createContent(pm *ProgramManager) {
	// Фон блока
	bg := canvas.NewRectangle(s.controller.block.Color)
	bg.SetMinSize(fyne.NewSize(float32(s.controller.block.Width), float32(s.controller.block.Height)))
	
	// Заголовок
	title := canvas.NewText(s.controller.block.Title, color.White)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 14
	
	// Описание
	desc := canvas.NewText(s.controller.block.Description, color.White)
	desc.Alignment = fyne.TextAlignCenter
	desc.TextSize = 10
	
	// Кнопка выбора
	selectBtn := widget.NewButton("Выбрать", func() {
		pm.selected = s.controller.block
		pm.ShowBlockProperties(s.controller.block)
	})
	
	// Кнопки перемещения
	moveUp := widget.NewButton("↑", func() {
		s.controller.block.Y -= 10
		s.Move(fyne.NewPos(float32(s.controller.block.X), float32(s.controller.block.Y)))
	})
	moveDown := widget.NewButton("↓", func() {
		s.controller.block.Y += 10
		s.Move(fyne.NewPos(float32(s.controller.block.X), float32(s.controller.block.Y)))
	})
	moveLeft := widget.NewButton("←", func() {
		s.controller.block.X -= 10
		s.Move(fyne.NewPos(float32(s.controller.block.X), float32(s.controller.block.Y)))
	})
	moveRight := widget.NewButton("→", func() {
		s.controller.block.X += 10
		s.Move(fyne.NewPos(float32(s.controller.block.X), float32(s.controller.block.Y)))
	})
	
	// Панель управления
	controls := container.NewHBox(
		moveUp,
		moveDown,
		moveLeft,
		moveRight,
		selectBtn,
	)
	
	// Основное содержимое
	mainContent := container.NewVBox(
		container.NewStack(
			bg,
			container.NewVBox(
				container.NewCenter(title),
				container.NewCenter(desc),
			),
		),
		controls,
	)
	
	s.content = mainContent
}

func (s *SimpleDraggableBlock) CreateRenderer() fyne.WidgetRenderer {
	return &simpleDraggableRenderer{
		widget:  s,
		objects: []fyne.CanvasObject{s.content},
	}
}

type simpleDraggableRenderer struct {
	widget  *SimpleDraggableBlock
	objects []fyne.CanvasObject
}

func (r *simpleDraggableRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *simpleDraggableRenderer) MinSize() fyne.Size {
	return r.objects[0].MinSize()
}

func (r *simpleDraggableRenderer) Refresh() {
	r.objects[0].Refresh()
}

func (r *simpleDraggableRenderer) Destroy() {}

func (r *simpleDraggableRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}