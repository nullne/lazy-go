package domain

import "time"

type Note struct {
	ID        string
	LessonID  string
	Heading   string
	CreatorID string
	Content   string
	Ints      []int
	Hi        Foo
	Sticky    Fuck
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// go:generate enumer -type=VerificationType
type Foo int

type Fuck struct {
}
