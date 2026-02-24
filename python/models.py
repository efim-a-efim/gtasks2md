from typing import List, Optional
from pydantic import BaseModel, Field

class Task(BaseModel):
    id: Optional[str] = None
    title: str
    status: str = "needsAction" # "needsAction" or "completed"
    notes: Optional[str] = None
    parent: Optional[str] = None
    children: List['Task'] = Field(default_factory=list)

class TaskList(BaseModel):
    id: Optional[str] = None
    title: str
    tasks: List[Task] = Field(default_factory=list)
