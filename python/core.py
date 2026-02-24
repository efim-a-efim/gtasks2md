import os
import glob
from typing import Optional
from auth import authenticate
from google_tasks import GoogleTasksClient
from models import TaskList
from markdown import load_from_file, save_to_file

def _sync_tasklist(local_list: TaskList, remote_list_id: str, client: GoogleTasksClient):
    """Syncs a local TaskList to a remote Google Task list."""
    remote_tasks = client.get_tasks(remote_list_id)
    
    # Map remote tasks by title
    remote_map = {}
    for rt in remote_tasks:
        remote_map[rt.title] = rt
        for child in rt.children:
            remote_map[child.title] = child
            
    # Collect all local titles to find tasks to delete
    local_titles = set()
    def collect_titles(tasks):
        for t in tasks:
            local_titles.add(t.title)
            collect_titles(t.children)
    collect_titles(local_list.tasks)
    
    # Delete remote tasks that are not in local list
    # We delete children first to avoid API errors (though usually deleting a parent deletes children)
    for rt in remote_tasks:
        for child in rt.children:
            if child.title not in local_titles:
                client.delete_task(remote_list_id, child.id)
        if rt.title not in local_titles:
            client.delete_task(remote_list_id, rt.id)
            
    # Create or update tasks
    def sync_task(local_task, parent_id=None):
        if local_task.title in remote_map:
            # Update
            remote_task = remote_map[local_task.title]
            remote_task.status = local_task.status
            remote_task.notes = local_task.notes
            client.update_task(remote_list_id, remote_task)
            local_task.id = remote_task.id
        else:
            # Create
            created = client.create_task(remote_list_id, local_task, parent_id=parent_id)
            local_task.id = created.id

        # Sync children
        for child in local_task.children:
            sync_task(child, parent_id=local_task.id)

    for task in local_list.tasks:
        sync_task(task)


def export_tasks(output_path: str, list_name: Optional[str] = None, credentials_path: Optional[str] = None):
    """Exports task lists from Google Tasks to local Markdown files."""
    creds = authenticate(credentials_path)
    client = GoogleTasksClient(creds)
    
    remote_lists = client.get_tasklists()
    
    if os.path.isdir(output_path) or not output_path.endswith('.md'):
        # Directory export (many:many)
        if not os.path.exists(output_path):
            os.makedirs(output_path)
            
        for rl in remote_lists:
            # Check if list_name was specified, if so only export that one
            if list_name and rl.title != list_name:
                continue
                
            tasks = client.get_tasks(rl.id)
            rl.tasks = tasks
            
            filename = "".join([c for c in rl.title if c.isalnum() or c in (' ', '-', '_')]).rstrip()
            file_path = os.path.join(output_path, f"{filename}.md")
            save_to_file(rl, file_path)
            print(f"Exported '{rl.title}' to {file_path}")
    else:
        # File export (1:1)
        if not list_name:
            raise ValueError("list-name must be specified when exporting to a single file.")
            
        target_list = next((rl for rl in remote_lists if rl.title == list_name), None)
        if not target_list:
            raise ValueError(f"Task list '{list_name}' not found on Google Tasks.")
            
        tasks = client.get_tasks(target_list.id)
        target_list.tasks = tasks
        
        save_to_file(target_list, output_path)
        print(f"Exported '{target_list.title}' to {output_path}")


def import_tasks(input_path: str, list_name: Optional[str] = None, credentials_path: Optional[str] = None):
    """Imports task lists from local Markdown files to Google Tasks."""
    creds = authenticate(credentials_path)
    client = GoogleTasksClient(creds)
    
    remote_lists = client.get_tasklists()
    remote_lists_map = {rl.title: rl for rl in remote_lists}
    
    if os.path.isdir(input_path):
        # Directory import (many:many)
        md_files = glob.glob(os.path.join(input_path, '*.md'))
        for file_path in md_files:
            local_list = load_from_file(file_path)
            
            # The list title comes from the markdown # H1
            target_title = local_list.title
            
            if target_title in remote_lists_map:
                target_list = remote_lists_map[target_title]
                print(f"Syncing {file_path} to existing list '{target_title}'...")
            else:
                target_list = client.create_tasklist(target_title)
                remote_lists_map[target_title] = target_list
                print(f"Created new list '{target_title}' and syncing from {file_path}...")
                
            _sync_tasklist(local_list, target_list.id, client)
            print(f"Successfully imported {file_path}")
            
    else:
        # File import (1:1)
        if not os.path.exists(input_path):
            raise FileNotFoundError(f"File '{input_path}' not found.")
            
        local_list = load_from_file(input_path)
        
        # Override title if list_name provided, else use the title from file or filename
        target_title = list_name if list_name else local_list.title
        if not target_title or target_title == "Untitled List":
            target_title = os.path.splitext(os.path.basename(input_path))[0]
            
        if target_title in remote_lists_map:
            target_list = remote_lists_map[target_title]
            print(f"Syncing {input_path} to existing list '{target_title}'...")
        else:
            target_list = client.create_tasklist(target_title)
            remote_lists_map[target_title] = target_list
            print(f"Created new list '{target_title}' and syncing from {input_path}...")
            
        _sync_tasklist(local_list, target_list.id, client)
        print(f"Successfully imported {input_path}")
