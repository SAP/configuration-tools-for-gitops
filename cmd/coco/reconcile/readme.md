# Reconcile Package

The package in this folder helps to reconcile a target branch with a source branch.

## Usage
```shell
coco reconcile --source <source_branch> --target <target_branch> --ownerName <owner_name> --repoName <repo_name> [--dry-run]
```

## Required Flags

- `--source` - The source branch to reconcile from.
- `--target` - The target branch to reconcile to.
- `--ownerName` - The name of the owner of the repository.
- `--repoName` - The name of the repository.

## Optional Flags
- `--dry-run` - Perform a dry-run to check for merge conflicts without making any changes.

## Command Details
This command reconciles a target branch with a source branch. If merge conflicts occur, it will attempt to reconcile the targetBranch by creating a new branch, `reconcile/targetBranch`, that will be used to merge the `sourceBranch` into the `targetBranch`. If the `reconcile/targetBranch` already exists, it will try to merge the latest changes from the `targetBranch` into the `reconcile/targetBranch`. If no changes exist, it will create a draft pull request that will merge the `sourceBranch` into the `reconcile/targetBranch`. If there are merge conflicts, it will create a new branch named `reconcile/targetBranch>` from the target branch, and then attempt to merge the source branch into it. If there are no merge conflicts, it will merge the source branch into the target branch directly.

## Authentication
This command requires access to a GitHub personal access token. The token must be stored in the GITHUB_TOKEN environment variable.

## Example
```shell
coco reconcile --source main --target dev --ownerName myorg --repoName myrepo
```
This will reconcile the `dev` branch with the `main` branch in the `myorg/myrepo` repository. If there are merge conflicts, it will create a new branch named `reconcile/dev` from the `dev` branch, and then attempt to merge the `main` branch into it. If there are no merge conflicts, it will merge the `main` branch into the `dev` branch directly.

