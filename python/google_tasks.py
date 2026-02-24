import logging
from typing import List, Optional, Dict
from googleapiclient.discovery import build
from models import Task, TaskList

logger = logging.getLogger(__name__)

class GoogleTasksClient:
    def __init__(self, credentials):
        self.service = build('tasks', 'v1', credentials=credentials)

    def get_tasklists(self) -> List[TaskList]:
        """Fetches all task lists and returns them as a list of TaskList models."""
        results = self.service.tasklists().list(maxResults=100).execute()
        raw_lists = results.get('items', [])
        
        tasklists = []
        for rl in raw_lists:
            tasklists.append(TaskList(id=rl['id'], title=rl['title']))
        return tasklists

    def create_tasklist(self, title: str) -> TaskList:
        """Creates a new task list."""
        result = self.service.tasklists().insert(body={'title': title}).execute()
        return TaskList(id=result['id'], title=result['title'])

    def get_tasks(self, tasklist_id: str) -> List[Task]:
        """Fetches all tasks in a task list and structures them into a hierarchy."""
        raw_tasks = []
        page_token = None
        while True:
            result = self.service.tasks().list(
                tasklist=tasklist_id,
                showHidden=True,
                maxResults=100,
                pageToken=page_token
            ).execute()
            raw_tasks.extend(result.get('items', []))
            page_token = result.get('nextPageToken')
            if not page_token:
                break
        
        # Build Task models map
        task_dict: Dict[str, Task] = {}
        # Temporarily store position to sort them
        positions: Dict[str, str] = {}
        
        for rt in raw_tasks:
            # Skip deleted tasks
            if rt.get('deleted'):
                continue
            
            task = Task(
                id=rt['id'],
                title=rt.get('title', ''),
                status=rt.get('status', 'needsAction'),
                notes=rt.get('notes'),
                parent=rt.get('parent')
            )
            task_dict[task.id] = task
            positions[task.id] = rt.get('position', '')

        # Build hierarchy
        root_tasks: List[Task] = []
        for task in task_dict.values():
            if task.parent and task.parent in task_dict:
                task_dict[task.parent].children.append(task)
            else:
                root_tasks.append(task)
                
        # Sort root tasks and children by position to maintain order
        root_tasks.sort(key=lambda t: positions.get(t.id, ''))
        for task in task_dict.values():
            task.children.sort(key=lambda t: positions.get(t.id, ''))
            
        return root_tasks
        
    def create_task(self, tasklist_id: str, task: Task, parent_id: Optional[str] = None) -> Task:
        """Creates a new task in the specified list, optionally as a child of parent_id."""
        body = {
            'title': task.title,
            'status': task.status,
            'notes': task.notes
        }
        
        result = self.service.tasks().insert(
            tasklist=tasklist_id,
            parent=parent_id,
            body=body
        ).execute()
        
        task.id = result['id']
        task.parent = result.get('parent')
        return task

    def update_task(self, tasklist_id: str, task: Task) -> Task:
        """Updates an existing task."""
        if not task.id:
            raise ValueError("Task ID is required for updating")
            
        body = {
            'id': task.id,
            'title': task.title,
            'status': task.status,
            'notes': task.notes
        }
        
        result = self.service.tasks().update(
            tasklist=tasklist_id,
            task=task.id,
            body=body
        ).execute()
        return task

    def delete_task(self, tasklist_id: str, task_id: str):
        """Deletes a task."""
        self.service.tasks().delete(tasklist=tasklist_id, task=task_id).execute()
