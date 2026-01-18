package main

import (
    "fmt"
    "log"
)

// LPF2Manager управляет протоколом LPF2
type LPF2Manager struct {
    hubMgr *HubManager
}

// NewLPF2Manager создает новый менеджер LPF2
func NewLPF2Manager(hubMgr *HubManager) *LPF2Manager {
    return &LPF2Manager{
        hubMgr: hubMgr,
    }
}

// SetupLEDMode настраивает режим светодиода
func (lm *LPF2Manager) SetupLEDMode(port byte, mode byte) error {
    cmd := []byte{0x01, 0x02, port, 0x17, mode, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01}
    log.Printf("Настройка режима светодиода: порт=%d, режим=%d", port, mode)
    return lm.hubMgr.WriteCharacteristic(INPUT_COMMAND_UUID, cmd)
}

// SetLEDColorRGB устанавливает RGB цвет
func (lm *LPF2Manager) SetLEDColorRGB(port byte, red, green, blue byte) error {
    cmd := []byte{0x06, 0x04, 0x03, red, green, blue}
    log.Printf("Установка RGB цвета: порт=%d, R=%d, G=%d, B=%d", port, red, green, blue)
    return lm.hubMgr.WriteCharacteristic(OUTPUT_COMMAND_UUID, cmd)
}

// SetLEDColorIndex устанавливает индексный цвет
func (lm *LPF2Manager) SetLEDColorIndex(port byte, colorIndex byte) error {
    cmd := []byte{0x06, 0x04, 0x01, colorIndex}
    log.Printf("Установка индексного цвета: порт=%d, индекс=%d", port, colorIndex)
    return lm.hubMgr.WriteCharacteristic(OUTPUT_COMMAND_UUID, cmd)
}

// SetupTiltSensor настраивает датчик наклона
func (lm *LPF2Manager) SetupTiltSensor(port byte, mode byte) error {
    cmd := []byte{0x01, 0x02, port, 0x22, mode, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01}
    log.Printf("Настройка датчика наклона: порт=%d, режим=%d", port, mode)
    return lm.hubMgr.WriteCharacteristic(INPUT_COMMAND_UUID, cmd)
}

// SetupDistanceSensor настраивает датчик расстояния
func (lm *LPF2Manager) SetupDistanceSensor(port byte, mode byte) error {
    cmd := []byte{0x01, 0x02, port, 0x23, mode, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01}
    log.Printf("Настройка датчика расстояния: порт=%d, режим=%d", port, mode)
    return lm.hubMgr.WriteCharacteristic(INPUT_COMMAND_UUID, cmd)
}

// SetMotorSpeed устанавливает скорость мотора
func (lm *LPF2Manager) SetMotorSpeed(port byte, speed float64, factor float64) error {
    if speed < -1.0 || speed > 1.0 {
        return fmt.Errorf("скорость должна быть в диапазоне [-1, 1]")
    }
    
    var speedByte byte
    if speed < 0 {
        speedByte = byte((0x54 * maxFloat(speed, -1) * factor) + 0xF0)
    } else if speed > 0 {
        speedByte = byte((0x54 * minFloat(speed, 1) * factor) + 0x10)
    } else {
        speedByte = 0x00
    }
    
    cmd := []byte{port, 0x01, 0x01, speedByte}
    log.Printf("Установка скорости мотора: порт=%d, скорость=%.2f -> 0x%02x", port, speed, speedByte)
    return lm.hubMgr.WriteCharacteristic(OUTPUT_COMMAND_UUID, cmd)
}

// Вспомогательные функции
func maxFloat(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}

func minFloat(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}