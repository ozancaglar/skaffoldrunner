package prompts

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"slices"
)

type SelectPromptParams struct {
	Label string
	Items []string
}

const (
	SELECTED_ALL_ITEMS = "I'm done selecting items"
)

// SelectPrompt prompts the user with a selection of items and returns the string representing the item they selected
func SelectPrompt(params SelectPromptParams) (string, error) {
	selectPrompt := promptui.Select{
		Label: params.Label,
		Items: params.Items,
	}

	_, result, err := selectPrompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

// MultiSelectPrompt prompts the user with a selection of items and returns the strings representing the item they selected
func MultiSelectPrompt(params SelectPromptParams) ([]string, error) {
	params.Items = append(params.Items, SELECTED_ALL_ITEMS)
	var selectedItems []string

	for {
		if len(params.Items) == 1 || slices.Contains(selectedItems, SELECTED_ALL_ITEMS) {
			break
		}
		result, err := SelectPrompt(params)
		if err != nil {
			return nil, fmt.Errorf("error running SelectPrompt: %w", err)
		}
		selectedItems = append(selectedItems, result)
		params.Items = slices.DeleteFunc(params.Items, func(item string) bool { return item == result })
	}

	selectedItems = slices.DeleteFunc(selectedItems, func(item string) bool { return item == SELECTED_ALL_ITEMS })

	if len(selectedItems) == 0 {
		return nil, fmt.Errorf("no items selected in prompt")
	}

	return selectedItems, nil
}
