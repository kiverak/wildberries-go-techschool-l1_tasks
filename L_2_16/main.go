package main

// Реализовать утилиту загрузки веб-страниц вместе со всем вложенным контентом (ресурсы, ссылки), аналогичную wget -m (мирроринг сайта).

// Требования
// Программа должна принимать URL и, возможно, глубину рекурсии (количество уровней ссылок, которые нужно скачать).
// Должна уметь скачивать HTML-страницы, сохранять их локально, а также рекурсивно скачивать ресурсы: CSS, JS, изображения и т.д., а так же страницы,
// на которые есть ссылки (в рамках того же домена).
// На выходе должен получиться локальный каталог, содержащий копию сайта (или его части), чтобы страницу можно было открыть офлайн.
// Необходимо обрабатывать различные нюансы: относительные и абсолютные ссылки, дублирование (не скачивать один и тот же ресурс несколько раз),
// корректно формировать локальные пути для сохранения, избегать зацикливания по ссылкам.
// Опционально: поддержать параллельное скачивание (например, ограничить до N одновременных загрузок), управлять robots.txt и пр.

// Эта задача проверяет навыки сетевого программирования (HTTP-запросы), работы с файлами и строками, а также проектирования (нужно спланировать структуру,
// как хранить информацию о посещенных URL, как сохранять файлы и менять ссылки внутри HTML на локальные и т.д.).

// Постарайтесь разбить программу на функции и пакеты: например, парсер HTML, загрузчик и т.п.

// Обязательно учтите обработку ошибок (сетевых, файловых) и время выполнения (можно добавить таймауты на запросы).

import (
	"flag"
	"fmt"
	"my-wget/mirror"
	"os"

	"net/url"
)

func main() {
	urlFlag := flag.String("url", "", "Root URL to mirror")
	outFlag := flag.String("out", "./mirror", "Output directory")
	concurrency := flag.Int("concurrency", 8, "Max concurrent downloads")
	maxDepth := flag.Int("depth", 2, "Max recursion depth for links")
	flag.Parse()

	if *urlFlag == "" {
		fmt.Fprintln(os.Stderr, "-url is required")
		os.Exit(2)
	}

	_, err := url.Parse(*urlFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid url:", err)
		os.Exit(2)
	}

	crawler, err := mirror.NewCrawler(*urlFlag, *outFlag, *concurrency, *maxDepth)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating crawler:", err)
		os.Exit(1)
	}

	fmt.Println("Starting mirror of", *urlFlag)
	if err := crawler.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println("Done. Output in", *outFlag)
}
