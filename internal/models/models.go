package models

type Task struct {
	ID       *string
	Title    string
	Status   string // "needsAction" or "completed"
	Notes    *string
	Parent   *string
	Children []*Task
}

type TaskList struct {
	ID    *string
	Title string
	Tasks []*Task
}