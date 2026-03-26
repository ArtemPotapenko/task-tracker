package kafka

type RegisterMessage struct {
	Email string `json:"email"`
}

type DailySummaryUser struct {
	UserID       int64 `json:"user_id"`
	Completed    int   `json:"completed"`
	NotCompleted int   `json:"not_completed"`
}

type DailySummaryMessage struct {
	Date  string             `json:"date"`
	Users []DailySummaryUser `json:"users"`
}
