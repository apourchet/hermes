package hermes

import "encoding/json"

type Error struct {
	Message string
}

func (e *Error) Error() string {
	content, _ := json.Marshal(e)
	return string(content)
}
