package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

//Создать программу, печатающую точное текущее время с использованием NTP-сервера.
//Реализовать проект как модуль Go.
//Использовать библиотеку ntp для получения времени.
//Программа должна выводить текущее время, полученное через NTP (Network Time Protocol).
//Необходимо обрабатывать ошибки библиотеки: в случае ошибки вывести её текст в STDERR и вернуть ненулевой код выхода.
//Код должен проходить проверки (vet и golint), т.е. быть написан идиоматически корректно.

// ntpServer - адрес публичного NTP-сервера
const ntpServer = "0.beevik-ntp.pool.ntp.org"

func main() {
	// ntp.Time возвращает точное время или ошибку
	currentTime, err := ntp.Time(ntpServer)

	if err != nil {
		// os.Stderr - это стандартный поток ошибок
		_, printErr := fmt.Fprintf(os.Stderr, "Error getting NTP time: %v\n", err)
		if printErr != nil {
			fmt.Print("Couldn't print error message\n")
		}
		// Завершаем программу с ненулевым кодом, сигнализируя об ошибке
		os.Exit(1)
	}

	// time.RFC1123 - один из стандартных форматов времени
	fmt.Printf("Current NTP time: %s\n", currentTime.Format(time.RFC1123))
}
