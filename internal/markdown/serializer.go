package markdown

import (
	"fmt"
	"os"
	"strings"

	"gtasks2md/internal/models"
)

type Serializer struct {
	tasklist *models.TaskList
}

func NewSerializer(tasklist *models.TaskList) *Serializer {
	return &Serializer{tasklist: tasklist}
}

func (s *Serializer) Serialize() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("# %s", s.tasklist.Title))
	lines = append(lines, "")

	for _, task := range s.tasklist.Tasks {
		statusChar := " "
		if task.Status == "completed" {
			statusChar = "x"
		}
		lines = append(lines, fmt.Sprintf("- [%s] %s", statusChar, task.Title))

		if task.Notes != nil && *task.Notes != "" {
			for _, noteLine := range strings.Split(*task.Notes, "\n") {
				lines = append(lines, fmt.Sprintf("    %s", noteLine))
			}
		}

		for _, subtask := range task.Children {
			subStatusChar := " "
			if subtask.Status == "completed" {
				subStatusChar = "x"
			}
			lines = append(lines, fmt.Sprintf("    - [%s] %s", subStatusChar, subtask.Title))

			if subtask.Notes != nil && *subtask.Notes != "" {
				for _, noteLine := range strings.Split(*subtask.Notes, "\n") {
					lines = append(lines, fmt.Sprintf("    %s", noteLine))
				}
			}
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

func SaveToFile(tasklist *models.TaskList, filePath string) error {
	serializer := NewSerializer(tasklist)
	content := serializer.Serialize()
	return os.WriteFile(filePath, []byte(content), 0644)
}