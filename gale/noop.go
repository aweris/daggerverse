package main

// dagger expects a struct with at least on public method to be used in a module. This is a all dummy methods to make
// dagger happy. FIXME: remove this file when dagger supports modules without public methods.

func (r *RepoOpts) Noop() {}

func (r *WorkflowOpts) Noop() {}

func (r *WorkflowRunOpts) Noop() {}

func (r *WorkflowRunExportOpts) Noop() {}

func (_ GithubRepository) Noop() {}

func (_ GithubRepositoryOwner) Noop() {}

func (_ GithubRepositoryBranchRef) Noop() {}

func (_ AuthConfig) Noop() {}

func (_ *RepoConfig) Noop() {}

func (_ *WorkflowsConfig) Noop() {}

func (_ *WorkflowRunConfig) Noop() {}
