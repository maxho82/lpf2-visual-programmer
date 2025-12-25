package main

import (
	"log"
	"time"
)

// TestLPF2Protocol запускает тесты протокола LPF2
func TestLPF2Protocol(hubMgr *HubManager) {
	log.Println("=== Тестирование протокола LPF2 ===")

	if !hubMgr.IsConnected() {
		log.Println("Хаб не подключен")
		return
	}

	// Тест 1: Простая команда светодиода (красный)
	log.Println("Тест 1: Включение красного светодиода")
	data := []byte{0x08, 0x04, 0x06, 0x03, 0xFF, 0x00, 0x00, 0x00}
	err := hubMgr.WriteCharacteristic("00001565-1212-efde-1523-785feabcd123", data)
	if err != nil {
		log.Printf("Ошибка теста 1: %v", err)
	} else {
		log.Println("Тест 1 успешен")
		time.Sleep(1 * time.Second)

		// Тест 2: Выключение светодиода
		log.Println("Тест 2: Выключение светодиода")
		data2 := []byte{0x08, 0x04, 0x06, 0x03, 0x00, 0x00, 0x00, 0x00}
		err2 := hubMgr.WriteCharacteristic("00001565-1212-efde-1523-785feabcd123", data2)
		if err2 != nil {
			log.Printf("Ошибка теста 2: %v", err2)
		} else {
			log.Println("Тест 2 успешен")
		}
	}

	// Тест 3: Мотор
	log.Println("Тест 3: Запуск мотора на порту 1 (50% мощности, 1 секунда)")
	data3 := []byte{0x06, 0x04, 0x01, 0x01, 0x32, 0x00}
	err3 := hubMgr.WriteCharacteristic("00001565-1212-efde-1523-785feabcd123", data3)
	if err3 != nil {
		log.Printf("Ошибка теста 3: %v", err3)
	} else {
		log.Println("Тест 3 успешен")
	}
}
