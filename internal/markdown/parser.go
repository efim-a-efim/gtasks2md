package markdown

import (
	"os"
	"regexp"
	"strings"

	"gtasks2md/internal/models"
)

type Parser struct {
	content string
}

func NewParser(content string) *Parser {
	return &Parser{content: content}
}

func (p *Parser) Parse() *models.TaskList {
	lines := strings.Split(p.content, "\n")

	listTitle := "Untitled List"
	var tasks []*models.Task

	var currentTask *models.Task
	var currentSubtask *models.Task

	titleSet := false

	titlePattern := regexp.MustCompile(`^# (.*)`)
	taskPattern := regexp.MustCompile(`^- \[( |x|X)\] (.*)`)
	subtaskPattern := regexp.MustCompile(`^(    |\t)- \[( |x|X)\] (.*)`)
	notePrefixPattern := regexp.MustCompile(`^(    |\t)(.*)`)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if match := titlePattern.FindStringSubmatch(line); match != nil {
			if !titleSet {
				listTitle = strings.TrimSpace(match[1])
				titleSet = true
			}
			continue
		}

		if match := taskPattern.FindStringSubmatch(line); match != nil {
			statusChar := strings.ToLower(match[1])
			status := "needsAction"
			if statusChar == "x" {
				status = "completed"
			}
			title := strings.TrimSpace(match[2])

			currentTask = &models.Task{
				Title:  title,
				Status: status,
			}
			currentSubtask = nil
			tasks = append(tasks, currentTask)
			continue
		}

		if match := subtaskPattern.FindStringSubmatch(line); match != nil {
			if currentTask == nil {
				continue
			}

			statusChar := strings.ToLower(match[2])
			status := "needsAction"
			if statusChar == "x" {
				status = "completed"
			}
			title := strings.TrimSpace(match[3])

			currentSubtask = &models.Task{
				Title:  title,
				Status: status,
			}
			currentTask.Children = append(currentTask.Children, currentSubtask)
			continue
		}

		if match := notePrefixPattern.FindStringSubmatch(line); match != nil {
			rest := match[2]
			if !strings.HasPrefix(rest, "- [") {
				noteText := rest
				targetTask := currentTask
				if currentSubtask != nil {
					targetTask = currentSubtask
				}

				if targetTask != nil {
					if targetTask.Notes != nil {
						newNotes := *targetTask.Notes + "\n" + noteText
						targetTask.Notes = &newNotes
					} else {
						targetTask.Notes = &noteText
					}
				}
				continue
			}
		}
	}

	return &models.TaskList{
		Title: listTitle,
		Tasks: tasks,
	}
}

func LoadFromFile(filePath string) (*models.TaskList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	parser := NewParser(string(data))
	return parser.Parse(), nil
}