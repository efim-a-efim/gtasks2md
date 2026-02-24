package markdown

import (
	"strings"
	"testing"
)

func TestParserAndSerializer(t *testing.T) {
	testContent := `# My Google Tasks

- [ ] Buy groceries
    Milk, Eggs, Bread
    - [x] Pay at checkout
        Use new credit card
- [x] Clean the house
    Focus on living room
`

	// Test Parsing
	parser := NewParser(testContent)
	taskList := parser.Parse()

	if taskList.Title != "My Google Tasks" {
		t.Errorf("Expected title 'My Google Tasks', got '%s'", taskList.Title)
	}

	if len(taskList.Tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(taskList.Tasks))
	}

	task1 := taskList.Tasks[0]
	if task1.Title != "Buy groceries" || task1.Status != "needsAction" {
		t.Errorf("Task 1 parsed incorrectly: %+v", task1)
	}
	if task1.Notes == nil || *task1.Notes != "Milk, Eggs, Bread" {
		if task1.Notes == nil {
			t.Errorf("Task 1 notes parsed incorrectly: nil")
		} else {
			t.Errorf("Task 1 notes parsed incorrectly: '%s'", *task1.Notes)
		}
	}
	if len(task1.Children) != 1 {
		t.Fatalf("Expected 1 subtask for task 1, got %d", len(task1.Children))
	}

	subtask1 := task1.Children[0]
	if subtask1.Title != "Pay at checkout" || subtask1.Status != "completed" {
		t.Errorf("Subtask 1 parsed incorrectly: %+v", subtask1)
	}
	if subtask1.Notes == nil || *subtask1.Notes != "    Use new credit card" {
		if subtask1.Notes == nil {
			t.Errorf("Subtask 1 notes parsed incorrectly: nil")
		} else {
			t.Errorf("Subtask 1 notes parsed incorrectly: '%s'", *subtask1.Notes)
		}
	}

	task2 := taskList.Tasks[1]
	if task2.Title != "Clean the house" || task2.Status != "completed" {
		t.Errorf("Task 2 parsed incorrectly: %+v", task2)
	}
	if task2.Notes == nil || *task2.Notes != "Focus on living room" {
		if task2.Notes == nil {
			t.Errorf("Task 2 notes parsed incorrectly: nil")
		} else {
			t.Errorf("Task 2 notes parsed incorrectly: '%s'", *task2.Notes)
		}
	}

	// Test Serialization
	serializer := NewSerializer(taskList)
	serializedContent := serializer.Serialize()

	expectedContent := strings.ReplaceAll(testContent, "\r\n", "\n")
	serializedContent = strings.ReplaceAll(serializedContent, "\r\n", "\n")

	if serializedContent != expectedContent {
		t.Errorf("Serialized content does not match original.\nExpected:\n%s\nGot:\n%s", expectedContent, serializedContent)
	}
}

func TestParserEmptyLinesAndTabs(t *testing.T) {
	testContent := "# List with tabs\n\n" +
		"- [ ] Task 1\n" +
		"\tNote with tab\n" +
		"\t- [x] Subtask 1\n" +
		"\t\tNote for subtask 1\n"

	parser := NewParser(testContent)
	taskList := parser.Parse()

	if len(taskList.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(taskList.Tasks))
	}
	task1 := taskList.Tasks[0]
	if task1.Notes == nil || *task1.Notes != "Note with tab" {
		t.Errorf("Task 1 notes parsed incorrectly: %v", task1.Notes)
	}
	if len(task1.Children) != 1 {
		t.Fatalf("Expected 1 subtask, got %d", len(task1.Children))
	}
	subtask1 := task1.Children[0]
	if subtask1.Notes == nil || *subtask1.Notes != "\tNote for subtask 1" {
		t.Errorf("Subtask 1 notes parsed incorrectly: %v", subtask1.Notes)
	}
}
