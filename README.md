# Task - a simple task management utility

Task is a simply task management utility written in Go. It stores its state in a JSON file that's local to each project.

Because the tasks are represented in JSON on disk, they're version controlled along with the code.

## Usage

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

When no flags are provided, `task new` opens `$EDITOR` with YAML frontmatter for the task fields and the description below it. Avoid using the bare `task new` form in non-interactive shells or automation, since it will block waiting for an editor.

### `task edit`

Edit a task in `$EDITOR`. The editor opens with YAML frontmatter containing the task fields and the description below it. Avoid using `task edit` in non-interactive shells or automation, since it will block waiting for an editor.

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

## Development

To build locally with existing tags:
```bash
make build             # Builds with version from git tag
```

To tag and create a release:
```bash
git tag -a v1.0.0 -m "Release name"
git push origin v1.0.0
goreleaser release    # Requires goreleaser and GITHUB_TOKEN
```
