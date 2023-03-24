## todo

- define parameters for `reconcie`
- behaviour in merge-conflicts
  - defaults
  - configurable
- behaviour without merge-conflicts
  - defaults
  - configurable
- do we want to create Draft PRs in the remote?
  - not
  - only in conflicts
  - indicated by parameter

## command interface

option 1:

- `coco reconcile source->target`

  - translates to (in spirit)

    ```bash
      git checkout target; git merge source
    ```

- (NOT IN VERSION 1) `coco reconcile source->target_1->target_2->...->target_N`

  - translates to (in spirit)

    ```bash
      git checkout target_1; git merge source
      git checkout target_2; git merge target_1
    ```

option 2:

- `coco reconcile --source source_branch --target target_branch`

additional parameters? :

- `--dry-run`: validate if the merge is conflict free
- `--debug`: present the diff between source and target
- `--local-only` (if default pushes to remote): no remote update
- `--push` (if default does not push to remote): remote update
- (NOT IN VERSION 1) `--resolve-pattern`: defines a pattern that is used to
  resolve conflicts (if they match the pattern)
- `--draft-pr`: creates a draft pull-request

## tasks:

- describe the happy path:
  - starting situation (source- and target-branch not connected in git graph (2
    HEAD leaf nodes))
  - desired target situation (source-branch is connected to the target-branch
    (flows into the target branch) in a new commit)
- describe what the command does (happy path)
  - what is the entire flow from start situation to target situation?
- describe failure paths
  - what does the command do in merge conflicts?
- in general: where does the command manipulate files (local or remote or both)
  in what situation?

  - for failure mode
  - for happy path

(-) describe parameters that can be given to command
