package main

import (
	"fmt"
	"sort"
	"strings"
)

//Напишите функцию, которая находит все множества анаграмм по заданному словарю.
//
//Требования
//На вход подается срез строк (слов на русском языке в Unicode).
//
//На выходе: map-множество -> список, где ключом является первое встреченное слово множества, а значением — срез из всех слов,
//принадлежащих этому множеству анаграмм, отсортированных по возрастанию.
//
//Множества из одного слова не должны выводиться (т.е. если нет анаграмм, слово игнорируется).
//
//Все слова нужно привести к нижнему регистру.
//
//Пример:
//
//Вход: ["пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"]
//Результат (ключи в примере могут быть в другом порядке):
//– "пятак": ["пятак", "пятка", "тяпка"]
//– "листок": ["листок", "слиток", "столик"]
//
//Слово «стол» отсутствует в результатах, так как не имеет анаграмм.
//
//Для решения задачи потребуется умение работать со строками, сортировать
//и использовать структуры данных (map).
//
//Оценим эффективность: решение должно работать за линейно-логарифмическое время относительно количества слов
//(допустимо n * m log m, где m — средняя длина слова для сортировки букв).

func main() {
	words := []string{"тяпка", "пятак", "пятка", "листок", "слиток", "столик", "стол"}
	anagramSets := FindAnagramSets(words)

	for k, v := range anagramSets {
		fmt.Printf("%s: %v\n", k, v)
	}
}

// FindAnagramSets принимает список слов и возвращает map:
// ключ — первое встретившееся слово множества,
// значение — отсортированный список всех анаграмм
func FindAnagramSets(words []string) map[string][]string {
	// Вспомогательная структура: ключ по буквам -> список слов
	anagramGroups := make(map[string][]string)

	for _, w := range words {
		wLow := strings.ToLower(w)
		key := sortRunesInWord(wLow)
		anagramGroups[key] = append(anagramGroups[key], wLow)
	}

	result := make(map[string][]string)

	for _, group := range anagramGroups {
		if len(group) > 1 {
			// Убираем дубликаты
			uniqueWordsMap := make(map[string]struct{})
			for _, w := range group {
				uniqueWordsMap[w] = struct{}{}
			}
			// Преобразуем в срез
			final := make([]string, 0, len(uniqueWordsMap))
			for w := range uniqueWordsMap {
				final = append(final, w)
			}
			sort.Strings(final)
			result[group[0]] = final // ключом берем первое слово в группе
		}
	}

	return result
}

// sortRunesInWord строит ключ для анаграмм: сортированные руны в слове
func sortRunesInWord(word string) string {
	runes := []rune(word)

	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}
