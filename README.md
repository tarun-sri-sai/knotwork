# Knotwork

Progress tracking made faster, simpler, and better.

## Prerequisites

### Software

- [Docker](https://www.docker.com)
- [Docker Compose](https://docs.docker.com/compose)
- [Git](https://git-scm.com)\*

> Note:
>
> 1. \* Required based on repo type selected

### Repositories

#### Git

##### Setup

To set up the `git` repo type for knotwork, initialize a Git repo that contains a file called `to-do.txt` at the root of the repo. Setup a workflow that makes a commit to this repo daily with the commit message being the date in `YYYY-MM-dd` format of the day before. You can also manually do it. It is not required that a snapshot is present for every day, but it is required that all commits have unique dates.

##### Syntax

The `to-do.txt` contents must follow the syntax rules mentioned below. Any commit where the contents do not match the syntax will be ignored.

1. **General Structure:**
   - The file is divided into _blocks_, separated by <ins>blank lines</ins>.
   - A _block_ must follow <ins>consistent indentation</ins>.
   - Each _block_ is either:
     - A _Category_ block, or
     - A _Task_ block

   Examples:

   ```plaintext
   ********************************
   Category 1
   ********************************

                                                                                (Multiple blank lines are fine)
   Task 1
                                                                                (Blank line indicates that the Task 2 block is different from the Task 1 block) 
       Task 2


   ********************************
   Category 2
   ********************************
   ```

1. **Category Block:**
   - A _Category block_ must be exactly <ins>3 lines</ins>.
   - Line 1 and 3 must be the exact string `********************************`.
   - Line 2 contains the _category name_.
   - _Category blocks_ must always be at <ins>root level indentation</ins> (0-indent).
  
   Example:

   ```plaintext
   ********************************
   Category Name
   ********************************
   ```

1. **Task Block:**
   - A _task block_ is a <ins>group of lines</ins> representing a task and its _updates_.
   - The first line is automatically considered the _task title_.
   - _Updates_ start from <ins>line 2</ins>.
   - A _finished task_ must be atleast <ins>two lines long</ins>, where the last _update_ contains the `[DONE]` marker.
   - A _task_ is uniquely identified by its title and the titles of its ancestors from the root task up to its parent task.
   - A _sub-task_ can be made by indenting <ins>4 spaces</ins> away from its parent task.
   - If a _task_ is finished, all its descendant _tasks_ are automatically considered finished, even if not explicitly marked.
  
   Examples:

   ```plaintext
   ********************************
   Important
   ********************************
   
   Find my phone
   Looked under the seat
   Checked under the table
   Went to the coffee shop
   
       Check the "Find my Phone" locator                                        (4-space indent)
       [DONE] Location is shown as the coffee shop                              (This sub-task finished)

           Find a way to login without the phone                                (This sub-task is indirectly considered "finished" because its ancestor is)
                                                                                
       Ask the staff if they've seen it

           Ask at the counter                                                   (Sub-tasks can be nested upto infinite depth)

   Get a new SIM                                                                (This is a new root task, because it's not indented)
   ```

## Usage

### MCP Server

#### Git Repository

1. You can edit the repo path in the volume bind for `/data` in `compose.mcp.yaml` under the `core` service to point to your Git repo.

2. Run this command to start the MCP server:

   ```bash
   docker compose --project-directory /abs/path/to/repo -f /abs/path/to/compose.mcp.yaml run --rm -T mcp
   ```
