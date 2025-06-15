package handlers

import (
	"encoding/json"
	"fmt"
)

func (h *Handler) HandleEvent(event *Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	// ... rest of the function ...
	return nil
}
