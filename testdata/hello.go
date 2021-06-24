package domain

import "time"

type Note struct {
	ID        string
	LessonID  string
	Heading   string
	CreatorID string
	Content   string
	Sticky    bool
	CreatedAt time.Time
}
