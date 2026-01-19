---
name: task
description: "Structured task planning and tracking using the `task` program"
---

# Task Skill
**IMPORTANT:** If you don't have `task` in the $PATH, you can't use this workflow.

## Primary workflow
1. **Orient tasks**: restate the goal and translate it into a small set of tasks in the `task` system, initializing the project if needed and aligning titles, types, and labels.
2. **Break down work when needed**: split larger tasks into smaller tasks only when the work is ambiguous, multi-step, or likely to block progress. Use labels to group work into themes or areas.
3. **Capture findings as notes**: add notes to tasks when you discover decisions, blockers, or context that should be preserved with the task.
4. **Track progress**: move tasks through `todo`, `progress`, `blocked`, `abandon`, or `done` as work advances and call out blockers explicitly as notes.
5. **Close the loop**: summarize what changed in `task` and list the next actions so the user can continue.

## Output expectations
- Start with a one-sentence recap of what you updated or created in `task`.
- Provide a task list with IDs, titles, and statuses; include labels and types when they help orient the work.
- When adding findings, include the exact note content you plan to append and the target task ID.
- Call out any blockers and the action needed to unblock them.
- End with explicit next actions tied to specific task IDs.

## `task` Usage guide
Task is a simply task management app written in Go that stores its state in a JSON file that's local to each project. It should be available in the $PATH as it's installed to the system with `go install`.

Because the tasks are represented in JSON on disk, they're version controlled along with the code.

### Commands

#### `task help`/`task -h`

Show the standard help command listing all subcommands and global arguments.

#### `task init`

Initialise the directory to use `task` by creating the `.task/` directory and the `.task/task.json`.

#### `task list`

Display a list of tasks in the project. Optional arguments:

- `--json` to show the full JSON structure rather than the pretty print
- `-l/--label` to filter the list by tasks with this label
- `-t/--type` to filter the list by tasks with this type
- `-s/--status` to filter the by tasks with this status

#### `task new`

Create a new task. First positional argument is the task name. All tasks start as todo. New tasks are given an automatically generated 3 character hash as an ID. Optional arguments:

- `-d/--description` taking a string for the description
- `-l/--label` taking a list of strings to add as labels to the task
- `-t/--type` taking `task`, `bug`, or `feature`

#### `task update`

Update an existing task. The first positional argument is the task ID. Optional arguments:

- `-n/--name` taking a string for the task name
- `-d/--description` taking a string for the description
- `-l/--label` taking a list of strings to add as labels to the task
- `-t/--type` taking `task`, `bug`, or `feature`
- `-s/--status` taking `todo`, `progress`, `blocked`, `abandon`, or `done`

#### `task show`

Show a task in full with all of its fields and notes. First positional argument is task ID. Can be run with `--json` to show the full JSON structure rather than the pretty print.

#### `task note`

Append a note to the task with ID passed as the first positional argument. The second positional argument is a string that is the content of the note. Also accepts stdin for the note content. In such cases, the first positional argument is still the task ID.

#### Aliases

- `task ready` -> `task list -s todo`
- `task take $id` -> `task update $id -s progress`
- `task complete $id` -> `task update $id -s done`
- `task block $id` -> `task update $id -s blocked`
- `task abandon $id` -> `task update $id -s abandon`

### Schema

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
