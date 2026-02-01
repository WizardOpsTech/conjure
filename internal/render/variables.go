package render

import (
	"fmt"
	"strings"
)

func ParseVariables(varsList []string) (map[string]interface{}, error) {
	variables := make(map[string]interface{})

	for _, v := range varsList {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format '%s', expected key=value", v)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("variable key cannot be empty")
		}

		variables[key] = value
	}

	return variables, nil
}
