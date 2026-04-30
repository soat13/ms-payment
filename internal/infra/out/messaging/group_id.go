package messaging

import "github.com/soat13/ms-payment/internal/domain"

func ExtractGroupID(event domain.Event) *string {
	if g, ok := event.(domain.GroupedEvent); ok {
		return g.GroupID()
	}

	return nil
}
