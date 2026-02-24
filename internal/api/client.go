package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"

	"gtasks2md/internal/models"
)

// GoogleTasksClient holds the Google Tasks service client.
type GoogleTasksClient struct {
	service *tasks.Service
}

// NewClient initializes a new GoogleTasksClient.
func NewClient(ctx context.Context, client *http.Client) (*GoogleTasksClient, error) {
	service, err := tasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create tasks service: %v", err)
	}
	return &GoogleTasksClient{service: service}, nil
}

// GetTasklists fetches all task lists and returns them as a slice of TaskList models.
func (c *GoogleTasksClient) GetTasklists() ([]*models.TaskList, error) {
	var tasklists []*models.TaskList
	pageToken := ""

	for {
		req := c.service.Tasklists.List().MaxResults(100)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		results, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to fetch tasklists: %v", err)
		}

		for _, tl := range results.Items {
			id := tl.Id
			tasklists = append(tasklists, &models.TaskList{
				ID:    &id,
				Title: tl.Title,
			})
		}

		if results.NextPageToken == "" {
			break
		}
		pageToken = results.NextPageToken
	}

	return tasklists, nil
}

// CreateTasklist creates a new task list.
func (c *GoogleTasksClient) CreateTasklist(title string) (*models.TaskList, error) {
	tl := &tasks.TaskList{
		Title: title,
	}
	result, err := c.service.Tasklists.Insert(tl).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create tasklist: %v", err)
	}

	id := result.Id
	return &models.TaskList{
		ID:    &id,
		Title: result.Title,
	}, nil
}

// GetTasks fetches all tasks in a task list and structures them into a hierarchy.
func (c *GoogleTasksClient) GetTasks(tasklistID string) ([]*models.Task, error) {
	var rawTasks []*tasks.Task
	pageToken := ""

	for {
		req := c.service.Tasks.List(tasklistID).ShowHidden(true).MaxResults(100)
		if pageToken != "" {
			req.PageToken(pageToken)
		}

		result, err := req.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to fetch tasks: %v", err)
		}

		rawTasks = append(rawTasks, result.Items...)

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	// Build Task models map
	taskDict := make(map[string]*models.Task)
	// Temporarily store position to sort them
	positions := make(map[string]string)

	for _, rt := range rawTasks {
		// Skip deleted tasks
		if rt.Deleted {
			continue
		}

		status := "needsAction"
		if rt.Status != "" {
			status = rt.Status
		}

		var notes *string
		if rt.Notes != "" {
			n := rt.Notes
			notes = &n
		}

		var parent *string
		if rt.Parent != "" {
			p := rt.Parent
			parent = &p
		}

		id := rt.Id
		task := &models.Task{
			ID:     &id,
			Title:  rt.Title,
			Status: status,
			Notes:  notes,
			Parent: parent,
		}
		taskDict[id] = task
		positions[id] = rt.Position
	}

	// Build hierarchy
	var rootTasks []*models.Task
	for _, task := range taskDict {
		if task.Parent != nil {
			if parentTask, ok := taskDict[*task.Parent]; ok {
				parentTask.Children = append(parentTask.Children, task)
			} else {
				// Parent not found, add to root
				rootTasks = append(rootTasks, task)
			}
		} else {
			rootTasks = append(rootTasks, task)
		}
	}

	// Sort root tasks and children by position to maintain order
	sort.Slice(rootTasks, func(i, j int) bool {
		return positions[*rootTasks[i].ID] < positions[*rootTasks[j].ID]
	})

	for _, task := range taskDict {
		sort.Slice(task.Children, func(i, j int) bool {
			return positions[*task.Children[i].ID] < positions[*task.Children[j].ID]
		})
	}

	return rootTasks, nil
}

// CreateTask creates a new task in the specified list, optionally as a child of parentID.
func (c *GoogleTasksClient) CreateTask(tasklistID string, task *models.Task, parentID string) (*models.Task, error) {
	t := &tasks.Task{
		Title:  task.Title,
		Status: task.Status,
	}
	if task.Notes != nil {
		t.Notes = *task.Notes
	}

	req := c.service.Tasks.Insert(tasklistID, t)
	if parentID != "" {
		req.Parent(parentID)
	}

	result, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create task: %v", err)
	}

	id := result.Id
	task.ID = &id

	if result.Parent != "" {
		p := result.Parent
		task.Parent = &p
	}

	return task, nil
}

// UpdateTask updates an existing task.
func (c *GoogleTasksClient) UpdateTask(tasklistID string, task *models.Task) (*models.Task, error) {
	if task.ID == nil || *task.ID == "" {
		return nil, fmt.Errorf("Task ID is required for updating")
	}

	t := &tasks.Task{
		Id:     *task.ID,
		Title:  task.Title,
		Status: task.Status,
	}
	if task.Notes != nil {
		t.Notes = *task.Notes
	} else {
		t.NullFields = []string{"Notes"}
	}

	// Use Patch instead of Update to preserve unspecified fields
	result, err := c.service.Tasks.Patch(tasklistID, *task.ID, t).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to update task: %v", err)
	}

	id := result.Id
	task.ID = &id

	return task, nil
}

// DeleteTask deletes a task.
func (c *GoogleTasksClient) DeleteTask(tasklistID string, taskID string) error {
	err := c.service.Tasks.Delete(tasklistID, taskID).Do()
	if err != nil {
		return fmt.Errorf("unable to delete task: %v", err)
	}
	return nil
}
