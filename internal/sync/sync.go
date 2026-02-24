package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gtasks2md/internal/api"
	"gtasks2md/internal/markdown"
	"gtasks2md/internal/models"
)

// SyncTasklist Syncs a local TaskList to a remote Google Task list.
func SyncTasklist(localList *models.TaskList, remoteListID string, client *api.GoogleTasksClient) error {
	remoteTasks, err := client.GetTasks(remoteListID)
	if err != nil {
		return err
	}

	// Map remote tasks by title
	remoteMap := make(map[string]*models.Task)
	for _, rt := range remoteTasks {
		remoteMap[rt.Title] = rt
		for _, child := range rt.Children {
			remoteMap[child.Title] = child
		}
	}

	// Collect all local titles to find tasks to delete
	localTitles := make(map[string]bool)
	var collectTitles func(tasks []*models.Task)
	collectTitles = func(tasks []*models.Task) {
		for _, t := range tasks {
			localTitles[t.Title] = true
			collectTitles(t.Children)
		}
	}
	collectTitles(localList.Tasks)

	// Delete remote tasks that are not in local list
	// Delete children first
	for _, rt := range remoteTasks {
		for _, child := range rt.Children {
			if !localTitles[child.Title] {
				if err := client.DeleteTask(remoteListID, *child.ID); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to delete subtask '%s': %v\n", child.Title, err)
				}
			}
		}
		if !localTitles[rt.Title] {
			if err := client.DeleteTask(remoteListID, *rt.ID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete task '%s': %v\n", rt.Title, err)
			}
		}
	}

	// Create or update tasks
	var syncTask func(localTask *models.Task, parentID string) error
	syncTask = func(localTask *models.Task, parentID string) error {
		if remoteTask, exists := remoteMap[localTask.Title]; exists {
			// Update
			remoteTask.Status = localTask.Status
			remoteTask.Notes = localTask.Notes
			updated, err := client.UpdateTask(remoteListID, remoteTask)
			if err != nil {
				return err
			}
			localTask.ID = updated.ID
		} else {
			// Create
			created, err := client.CreateTask(remoteListID, localTask, parentID)
			if err != nil {
				return err
			}
			localTask.ID = created.ID
		}

		// Sync children
		for _, child := range localTask.Children {
			if err := syncTask(child, *localTask.ID); err != nil {
				return err
			}
		}
		return nil
	}

	for _, task := range localList.Tasks {
		if err := syncTask(task, ""); err != nil {
			return err
		}
	}

	return nil
}

// ExportTasks Exports task lists from Google Tasks to local Markdown files.
func ExportTasks(outputPath string, listName string, credentialsPath string) error {
	ctx := context.Background()
	authClient, err := api.Authenticate(ctx, credentialsPath)
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	client, err := api.NewClient(ctx, authClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	remoteLists, err := client.GetTasklists()
	if err != nil {
		return fmt.Errorf("failed to get tasklists: %v", err)
	}

	fileInfo, err := os.Stat(outputPath)
	isDir := err == nil && fileInfo.IsDir()
	if isDir || !strings.HasSuffix(outputPath, ".md") {
		// Directory export (many:many)
		if err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		}

		for _, rl := range remoteLists {
			if listName != "" && rl.Title != listName {
				continue
			}

			tasks, err := client.GetTasks(*rl.ID)
			if err != nil {
				return fmt.Errorf("failed to get tasks for list %s: %v", rl.Title, err)
			}
			rl.Tasks = tasks

			var builder strings.Builder
			for _, c := range rl.Title {
				if unicode.IsLetter(c) || unicode.IsDigit(c) || c == ' ' || c == '-' || c == '_' {
					builder.WriteRune(c)
				}
			}
			filename := strings.TrimRight(builder.String(), " \t\n\r")
			if filename == "" {
				filename = "untitled-list"
			}
			filePath := filepath.Join(outputPath, fmt.Sprintf("%s.md", filename))

			if err := markdown.SaveToFile(rl, filePath); err != nil {
				return fmt.Errorf("failed to save to file: %v", err)
			}
			fmt.Printf("Exported '%s' to %s\n", rl.Title, filePath)
		}
	} else {
		// File export (1:1)
		if listName == "" {
			return fmt.Errorf("list-name must be specified when exporting to a single file")
		}

		var targetList *models.TaskList
		for _, rl := range remoteLists {
			if rl.Title == listName {
				targetList = rl
				break
			}
		}

		if targetList == nil {
			return fmt.Errorf("task list '%s' not found on Google Tasks", listName)
		}

		tasks, err := client.GetTasks(*targetList.ID)
		if err != nil {
			return fmt.Errorf("failed to get tasks: %v", err)
		}
		targetList.Tasks = tasks

		if err := markdown.SaveToFile(targetList, outputPath); err != nil {
			return fmt.Errorf("failed to save to file: %v", err)
		}
		fmt.Printf("Exported '%s' to %s\n", targetList.Title, outputPath)
	}

	return nil
}

// ImportTasks Imports task lists from local Markdown files to Google Tasks.
func ImportTasks(inputPath string, listName string, credentialsPath string) error {
	ctx := context.Background()
	authClient, err := api.Authenticate(ctx, credentialsPath)
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	client, err := api.NewClient(ctx, authClient)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	remoteLists, err := client.GetTasklists()
	if err != nil {
		return fmt.Errorf("failed to get tasklists: %v", err)
	}

	remoteListsMap := make(map[string]*models.TaskList)
	for _, rl := range remoteLists {
		remoteListsMap[rl.Title] = rl
	}

	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("file or directory not found: %v", err)
	}

	if fileInfo.IsDir() {
		// Directory import (many:many)
		entries, err := os.ReadDir(inputPath)
		if err != nil {
			return fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				filePath := filepath.Join(inputPath, entry.Name())
				localList, err := markdown.LoadFromFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to load from file: %v", err)
				}

				targetTitle := localList.Title
				var targetList *models.TaskList

				if existing, ok := remoteListsMap[targetTitle]; ok {
					targetList = existing
					fmt.Printf("Syncing %s to existing list '%s'...\n", filePath, targetTitle)
				} else {
					created, err := client.CreateTasklist(targetTitle)
					if err != nil {
						return fmt.Errorf("failed to create tasklist: %v", err)
					}
					targetList = created
					remoteListsMap[targetTitle] = targetList
					fmt.Printf("Created new list '%s' and syncing from %s...\n", targetTitle, filePath)
				}

				if err := SyncTasklist(localList, *targetList.ID, client); err != nil {
					return fmt.Errorf("failed to sync tasklist: %v", err)
				}
				fmt.Printf("Successfully imported %s\n", filePath)
			}
		}
	} else {
		// File import (1:1)
		localList, err := markdown.LoadFromFile(inputPath)
		if err != nil {
			return fmt.Errorf("failed to load from file: %v", err)
		}

		targetTitle := localList.Title
		if listName != "" {
			targetTitle = listName
		}

		if targetTitle == "" || targetTitle == "Untitled List" {
			base := filepath.Base(inputPath)
			targetTitle = strings.TrimSuffix(base, filepath.Ext(base))
		}

		var targetList *models.TaskList
		if existing, ok := remoteListsMap[targetTitle]; ok {
			targetList = existing
			fmt.Printf("Syncing %s to existing list '%s'...\n", inputPath, targetTitle)
		} else {
			created, err := client.CreateTasklist(targetTitle)
			if err != nil {
				return fmt.Errorf("failed to create tasklist: %v", err)
			}
			targetList = created
			remoteListsMap[targetTitle] = targetList
			fmt.Printf("Created new list '%s' and syncing from %s...\n", targetTitle, inputPath)
		}

		if err := SyncTasklist(localList, *targetList.ID, client); err != nil {
			return fmt.Errorf("failed to sync tasklist: %v", err)
		}
		fmt.Printf("Successfully imported %s\n", inputPath)
	}

	return nil
}
