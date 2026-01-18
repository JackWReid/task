# Task - a simple task management app
Task is a simply task management app written in Go that stores its state in a JSON file that's local to each project.

Because the tasks are represented in JSON on disk, they're version controlled along with the code.

## Commands

### `task help`/`task -h`
Show the standard help command listing all subcommands and global arguments.

### `task init`
Initialise the directory to use `task` by creating the `.task/` directory and the `.task/task.json`.

### `task list`
Display a list of tasks in the project. Optional arguments:
- `--json` to show the full JSON structure rather than the pretty print
- `-l/--label` to filter the list by tasks with this label
- `-t/--type` to filter the list by tasks with this type
- `-s/--status` to filter the by tasks with this status

### `task new`
Create a new task. First positional argument is the task name. All tasks start as todo. New tasks are given an automatically generated 3 character hash as an ID. Optional arguments:
- `-d/--description` taking a string for the description
- `-l/--label` taking a list of strings to add as labels to the task
- `-t/--type` taking `task`, `bug`, or `feature`

### `task update`
Update an existing task. The first positional argument is the task ID. Optional arguments:
- `-n/--name` taking a string for the task name
- `-d/--description` taking a string for the description
- `-l/--label` taking a list of strings to add as labels to the task
- `-t/--type` taking `task`, `bug`, or `feature`
- `-s/--status` taking `todo`, `progress`, `blocked`, `abandon`, or `done`

### `task show`
Show a task in full with all of its fields and notes. First positional argument is task ID. Can be run with `--json` to show the full JSON structure rather than the pretty print.

### `task note`
Append a note to the task with ID passed as the first positional argument. The second positional argument is a string that is the content of the note. Also accepts stdin for the note content. In such cases, the first positional argument is still the task ID.

### Aliases
- `task ready` -> `task list -s todo`
- `task take $id` -> `task update $id -s progress`
- `task complete $id` -> `task update $id -s done`
- `task block $id` -> `task update $id -s blocked`
- `task abandon $id` -> `task update $id -s abandon`

## Schema
The schema for a task is as follows. The `tasks.json` file is just an array of them until we feel the need to optimise.

```json
{
    "id": "9nk",                      // Generated on creation
    "created_at": "ISO datetime",
    "updated_at": "ISO datetime",     // Updated on every mutation, default DESC sort order for list
    "title": "Task title",
    "description": "Task description",// Optional, initialised as null
    "status": "progress",             // Required, initialised as todo
    "labels": ["label1", "label2"],   // Required, initialised as []
    "notes": [                        // Required, initialised at []
        {
            "id": "9nk-81f",          // Required, initialised with task ID plus new note key
            "created_at": "ISO datetime",
            "content": "Note content"
        }
    ]
}
```
