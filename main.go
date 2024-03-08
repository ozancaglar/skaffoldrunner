package main

import (
	"fmt"
	"log"

	"github.com/ozancaglar/skaffoldrunner/prompts"
)

func main() {
	multiSelectPromptResult, err := prompts.MultiSelectPrompt(prompts.SelectPromptParams{Label: "blank", Items: []string{"some items", "some more items", "even more items"}})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(multiSelectPromptResult)
}
