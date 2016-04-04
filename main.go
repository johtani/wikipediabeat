package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/johtani/wikipediabeat/beater"
)

func main() {
	err := beat.Run("wikipediabeat", "", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
