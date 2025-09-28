package main

import (
	"fmt"
	"math"
)

//Дана последовательность температурных колебаний: -25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5. Объединить эти значения в группы с шагом 10 градусов.
//Пример: -20:{-25.4, -27.0, -21.0}, 10:{13.0, 19.0, 15.5}, 20:{24.5}, 30:{32.5}.
//Пояснение: диапазон -20 включает значения от -20 до -29.9, диапазон 10 – от 10 до 19.9, и т.д. Порядок в подмножествах не важен.

func getRange(temp float64) int {
	if temp >= 0 {
		return int(math.Floor(temp/10)) * 10
	} else {
		return int(math.Ceil(temp/10)) * 10
	}
}

func groupTemperatures(temps []float64) map[int][]float64 {
	groups := make(map[int][]float64)

	for _, temp := range temps {
		r := getRange(temp)
		groups[r] = append(groups[r], temp)
	}

	return groups
}

func main() {
	temperatures := []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}

	groups := groupTemperatures(temperatures)

	for r, temps := range groups {
		fmt.Printf("%d: %v\n", r, temps)
	}
}
