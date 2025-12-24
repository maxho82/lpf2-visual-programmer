package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme кастомная тема для приложения
type CustomTheme struct{}

var (
	// Основные цвета темы
	backgroundColor = color.NRGBA{R: 45, G: 45, B: 48, A: 255}    // Темно-серый фон
	foregroundColor = color.NRGBA{R: 240, G: 240, B: 240, A: 255} // Светлый текст
	primaryColor    = color.NRGBA{R: 0, G: 122, B: 204, A: 255}   // Синий акцент (как в VS Code)
	buttonTextColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // Белый текст на кнопках
	secondaryColor  = color.NRGBA{R: 63, G: 63, B: 70, A: 255}    // Цвет для второстепенных элементов
	disabledColor   = color.NRGBA{R: 104, G: 104, B: 104, A: 255} // Цвет для неактивных элементов
	hoverColor      = color.NRGBA{R: 28, G: 151, B: 234, A: 255}  // Цвет при наведении
	pressedColor    = color.NRGBA{R: 0, G: 97, B: 163, A: 255}    // Цвет при нажатии
	scrollBarColor  = color.NRGBA{R: 90, G: 90, B: 90, A: 255}    // Цвет скроллбара
	selectionColor  = color.NRGBA{R: 38, G: 79, B: 120, A: 255}   // Цвет выделения
)

func (c *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return backgroundColor
	case theme.ColorNameForeground:
		return foregroundColor
	case theme.ColorNamePrimary:
		return primaryColor
	case theme.ColorNameFocus:
		return hoverColor
	case theme.ColorNameDisabled:
		return disabledColor
	case theme.ColorNameDisabledButton:
		return disabledColor
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 187, G: 187, B: 187, A: 255}
	case theme.ColorNamePressed:
		return pressedColor
	case theme.ColorNameSelection:
		return selectionColor
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 80}
	case theme.ColorNameInputBackground:
		return secondaryColor
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 90, G: 90, B: 90, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 30, G: 30, B: 30, A: 230}
	case theme.ColorNameMenuBackground:
		return backgroundColor
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
	case theme.ColorNameHover:
		return hoverColor
	case theme.ColorNameScrollBar:
		return scrollBarColor
	case theme.ColorNameButton:
		return secondaryColor
	default:
		// Для всех остальных случаев возвращаем цвета темной темы
		return theme.DarkTheme().Color(name, variant)
	}
}

func (c *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	// Используем стандартный шрифт, но можно заменить на кастомный
	if style.Bold {
		return theme.DarkTheme().Font(style)
	}
	return theme.DarkTheme().Font(style)
}

func (c *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	// Возвращаем иконки из темной темы (они светлые)
	return theme.DarkTheme().Icon(name)
}

func (c *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 6
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameSeparatorThickness:
		return 1
	default:
		return theme.DarkTheme().Size(name)
	}
}
