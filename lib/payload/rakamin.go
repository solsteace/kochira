package payload

import "fmt"

type Rakamin struct{}

func (r Rakamin) Ok(method string, data any) map[string]any {
	return map[string]any{
		"status":  true,
		"message": fmt.Sprintf("Succeed to %s data", method),
		"errors":  nil,
		"data":    data}
}

func (r Rakamin) Err(method string, err []error) map[string]any {
	errorMsg := []string{}
	for _, e := range err {
		errorMsg = append(errorMsg, e.Error())
	}

	return map[string]any{
		"status":  false,
		"message": fmt.Sprintf("Failed to %s data", method),
		"errors":  errorMsg,
		"data":    ""}
}
