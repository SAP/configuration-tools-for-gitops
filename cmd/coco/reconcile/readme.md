# Reconcile Package

The package in this folder helps to reconcile a target branch with a source branch.

## Usage
```shell
coco reconcile --source <source_branch> --target <target_branch> --ownerName <owner_name> --repoName <repo_name> [--dry-run]
```

For the reconcile command usage please run
```shell
coco reconcile --help
```

## Command Details
```flow
st=>start: Start
e=>end: End
op1=>operation: Set sourceBranch, targetBranch, owner and repo
op2=>operation: Authenticate with Github
op3=>operation: Attempt to merge branches
cond3=>condition: Merge conflict detected?
op5=>operation: Check if reconcileBranch exists
cond5=>condition: Yes or No?
op6=>operation: Check if targetBranch has new commits
cond6=>condition: Yes or No?
cond6_1=>condition: Delete the reconcileBranch?
io6_1=>inputoutput: Reconcile branch deleted
io6_2=>inputoutput: Manually rebase reconcileBranch with targetBranch and re-try
op7=>operation: Check mergability of the draft pull request
cond7=>condition: Mergeable?
io7_1=>inputoutput: Fast-forward merge of reconcileBranch into targetBranch
io7_2=>inputoutput: Resolve merge conflicts and re-try
op8=>operation: Create a new reconcileBranch
op9=>operation: Create a new draft pull request
io9=>inputoutput: Resolve merge conflicts in the pull request and re-try

st->op1->op2->op3->cond3->e
op8->op9->io9->e
cond3(yes)->op5->cond5
cond3(no)->e
cond5(yes)->op6->cond6
cond5(no)->op8
cond6(yes)->cond6_1
cond6(no)->op7->cond7
cond6_1(yes)->io6_1->op8
cond6_1(no)->io6_2-e
cond7(yes)->io7_1->e
cond7(no)->io7_2->e
```

## Authentication
This command requires access to a GitHub personal access token. The token must be stored in the GITHUB_TOKEN environment variable.

## Example
```shell
coco reconcile --source main --target dev --ownerName myorg --repoName myrepo
```
This will reconcile the `dev` branch with the `main` branch in the `myorg/myrepo` repository. If there are merge conflicts, it will create a new branch named `reconcile/dev` from the `dev` branch, and then attempt to merge the `main` branch into it. If there are no merge conflicts, it will merge the `main` branch into the `dev` branch directly.