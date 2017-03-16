package hermes

import (
	"encoding/json"
	"fmt"
)

func ReadErrorMessage(content []byte) (string, error) {
	tmp := map[string]string{}
	err := json.Unmarshal(content, &tmp)
	if err != nil {
		return "", fmt.Errorf("Failed to read error from json blob: %v", err)
	}

	if message, found := tmp["message"]; found {
		return message, nil
	}

	return "", fmt.Errorf("Failed to read error from json blob")
}
