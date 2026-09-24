package prompt

import "github.com/manifoldco/promptui"

func TextPrompt(msg string) string {
	response, responseErr := (&promptui.Prompt{Label: msg}).Run()
	if responseErr != nil {
		return ""
	}
	return response
}

func ConfirmPrompt(msg string) bool {
	_, responseErr := (&promptui.Prompt{IsConfirm: true, Label: msg}).Run()
	return responseErr == nil
}

func SelectPrompt[T any](msg string, choices []T, printFn func(T) string) (T, bool) {
	names := []string{}
	for _, dst := range choices {
		names = append(names, printFn(dst))
	}
	index, _, selectErr := (&promptui.Select{
		Label: msg,
		Items: names,
	}).Run()
	if selectErr != nil {
		return *new(T), false
	}
	return choices[index], true
}

func MultiselectPrompt[T any](msg string, choices []T, printFn func(T) string) []T {
	names := []string{}
	availableChoices := choices
	for _, dst := range choices {
		names = append(names, printFn(dst))
	}
	results := []T{}
	cursor := 0
	scroll := 0
	for {
		sel := &promptui.Select{
			Label: msg,
			Items: names,
		}
		// Run() always resets the scroll position to 0, which hides the cursor
		// when it is below the first page.
		index, _, selectErr := sel.RunCursorAt(cursor, scroll)
		if selectErr != nil {
			return results
		}
		cursor = index
		scroll = sel.ScrollPosition()
		results = append(results, availableChoices[index])
		names = append(names[:index], names[index+1:]...)
		availableChoices = append(availableChoices[:index], availableChoices[index+1:]...)
	}
}
