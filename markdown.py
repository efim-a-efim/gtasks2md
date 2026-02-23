import re
from typing import List, Optional
from models import Task, TaskList

class MarkdownParser:
    """Parses custom Markdown format into TaskList and Task objects."""
    
    def __init__(self, content: str):
        self.content = content
        
    def parse(self) -> TaskList:
        lines = self.content.split('\n')
        
        list_title = "Untitled List"
        tasks: List[Task] = []
        
        current_task: Optional[Task] = None
        current_subtask: Optional[Task] = None
        
        title_set = False
        
        title_pattern = re.compile(r'^# (.*)')
        task_pattern = re.compile(r'^- \[( |x|X)\] (.*)')
        subtask_pattern = re.compile(r'^(    |\t)- \[( |x|X)\] (.*)')
        note_pattern = re.compile(r'^(    |\t)(?!- \[)(.*)')
        
        for line in lines:
            if not line.strip():
                # Ignore empty lines
                continue
                
            title_match = title_pattern.match(line)
            if title_match:
                if not title_set:
                    list_title = title_match.group(1).strip()
                    title_set = True
                continue
                
            task_match = task_pattern.match(line)
            if task_match:
                status_char = task_match.group(1).lower()
                status = 'completed' if status_char == 'x' else 'needsAction'
                title = task_match.group(2).strip()
                
                current_task = Task(title=title, status=status)
                current_subtask = None
                tasks.append(current_task)
                continue
                
            subtask_match = subtask_pattern.match(line)
            if subtask_match:
                if not current_task:
                    continue
                    
                status_char = subtask_match.group(2).lower()
                status = 'completed' if status_char == 'x' else 'needsAction'
                title = subtask_match.group(3).strip()
                
                current_subtask = Task(title=title, status=status)
                current_task.children.append(current_subtask)
                continue
                
            note_match = note_pattern.match(line)
            if note_match:
                note_text = note_match.group(2)
                target_task = current_subtask if current_subtask else current_task
                if target_task is not None:
                    if target_task.notes:
                        target_task.notes += "\n" + note_text
                    else:
                        target_task.notes = note_text
                continue
                
        return TaskList(title=list_title, tasks=tasks)


class MarkdownSerializer:
    """Serializes TaskList and Task objects back into the custom Markdown format."""
    
    def __init__(self, tasklist: TaskList):
        self.tasklist = tasklist
        
    def serialize(self) -> str:
        lines = []
        lines.append(f"# {self.tasklist.title}")
        lines.append("")
        
        for task in self.tasklist.tasks:
            status_char = 'x' if task.status == 'completed' else ' '
            lines.append(f"- [{status_char}] {task.title}")
            
            if task.notes:
                for note_line in task.notes.split('\n'):
                    lines.append(f"    {note_line}")
                    
            for subtask in task.children:
                sub_status_char = 'x' if subtask.status == 'completed' else ' '
                lines.append(f"    - [{sub_status_char}] {subtask.title}")
                
                if subtask.notes:
                    for note_line in subtask.notes.split('\n'):
                        lines.append(f"    {note_line}")
                        
        return "\n".join(lines) + "\n"

def load_from_file(file_path: str) -> TaskList:
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    parser = MarkdownParser(content)
    return parser.parse()

def save_to_file(tasklist: TaskList, file_path: str):
    serializer = MarkdownSerializer(tasklist)
    content = serializer.serialize()
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
