package domain

import "time"

type Answer struct {
	ID          string
	QuestionID  string
	Content     string
	ContentType string
	CreatorID   string
	Sticky      bool
	KudosCount  int
	CreatedAt   time.Time
	LastUpdated time.Time
}
