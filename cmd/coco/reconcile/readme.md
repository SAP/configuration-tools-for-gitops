# Reconcile Package

The package in this folder helps to reconcile a target branch with a source
branch.

## Usage

```shell
coco reconcile \
  --source <source_branch> \
  --target <target_branch> \
  --owner <owner_name> \
  --repo <repo_name> \
  [--dry-run]
```

For the reconcile command usage please run

```shell
coco reconcile --help
```

## Command Details

```mermaid
flowchart TB
    st[Start]
    e[End]
    op1[[Set sourceBranch, targetBranch, owner and repo]]
    op2[[Authenticate with Github]]
    op3[[Attempt to merge branches]]
    cond3{Merge conflict detected?}
    cond5{reconcileBranch exists?}
    cond6{targetBranch has new commits?}
    cond6_1{Delete the reconcileBranch or <br> manually rebase with targetBranch?}
    io6_1[/reconcileBranch deleted/]
    cond7{Is the draft pull request mergeable?}
    io7_1[/Fast-forward merge of reconcileBranch into targetBranch/]
    io7_2[/Resolve merge conflicts and re-try/]
    op8[[Create a new reconcileBranch]]
    op9[[Create a new draft pull request]]
    io9[/Resolve merge conflicts in the pull request and re-try/]

    st-->op1-->op2-->op3-->cond3
    op8-->op9-->io9-->e
    cond3-->|YES|cond5
    cond3-->|NO|e
    cond5-->|YES|cond6
    cond5-->|NO|op8
    cond6-->|YES|cond6_1
    cond6-->|NO|cond7
    cond6_1-->|DELETE|io6_1-->op8
    cond6_1-->|REBASE MANUALLY|e
    cond7-->|YES|io7_1-->e
    cond7-->|NO|io7_2-->e
```

## Authentication

For pushing to Github, this command requires access to a GitHub personal access
token. The token must be stored in the GITHUB_TOKEN environment variable.

## Example

```shell
coco reconcile --source main --target dev --owner myorg --repo myrepo
```

This will reconcile the `origin/dev` branch with the `origin/main` branch in the
`myorg/myrepo` repository. If there are merge conflicts, it will create a new
branch named `reconcile/main-dev` from the `origin/dev` branch, and then attempt
to merge the `origin/main` branch into it. If there are no merge conflicts, it
will merge the `origin/main` branch into the `origin/dev` branch directly.

```mermaid
sequenceDiagram
    User->>Coco: Run reconcile command
    Coco->>Github: Switch to origin/dev branch
    Coco->>Github: Attempt merging origin/main
    Github-->>Coco: Return merge result
    Coco-->>Coco: Handle merge result
    Coco-->>User: Coco exits gracefully
```
