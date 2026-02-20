package models

// Review maps to the Review GraphQL type
type Review struct {
	ID        string `json:"id"`
	ProductID string `json:"productId"`
	UserID    string `json:"userId"`
	Body      string `json:"body"`
	Rating    int    `json:"rating"`
	CreatedAt string `json:"createdAt"` // In production, consider using time.Time
}

func (Review) IsEntity() {}
