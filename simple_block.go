package main

import (
	"image/color"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SimpleBlock - простой блок без перетаскивания
type SimpleBlock struct {
	widget.BaseWidget
	block   *ProgramBlock
	content fyne.CanvasObject
}

// NewSimpleBlock создает простой блок
func NewSimpleBlock(block *ProgramBlock) *SimpleBlock {
	s := &SimpleBlock{block: block}
	s.ExtendBaseWidget(s)
	s.createContent()
	return s
}

func (s *SimpleBlock) createContent() {
	// Фон блока
	bg := canvas.NewRectangle(s.block.Color)
	bg.SetMinSize(fyne.NewSize(float32(s.block.Width), float32(s.block.Height)))
	
	// Заголовок
	title := canvas.NewText(s.block.Title, color.White)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter
	title.TextSize = 14
	
	// Описание
	desc := canvas.NewText(s.block.Description, color.White)
	desc.Alignment = fyne.TextAlignCenter
	desc.TextSize = 10
	
	// Основное содержимое
	s.content = container.NewStack(
		bg,
		container.NewVBox(
			container.NewCenter(title),
			container.NewCenter(desc),
		),
	)
}

func (s *SimpleBlock) CreateRenderer() fyne.WidgetRenderer {
	return &simpleBlockRenderer{
		widget:  s,
		objects: []fyne.CanvasObject{s.content},
	}
}

type simpleBlockRenderer struct {
	widget  *SimpleBlock
	objects []fyne.CanvasObject
}

func (r *simpleBlockRenderer) Layout(size fyne.Size) {
	r.objects[0].Resize(size)
}

func (r *simpleBlockRenderer) MinSize() fyne.Size {
	return r.objects[0].MinSize()
}

func (r *simpleBlockRenderer) Refresh() {
	r.objects[0].Refresh()
}

func (r *simpleBlockRenderer) Destroy() {}

func (r *simpleBlockRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}