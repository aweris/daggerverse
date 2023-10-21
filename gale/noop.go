package main

// dagger expects a struct with at least on public method to be used in a module. This is a all dummy methods to make
// dagger happy. FIXME: remove this file when dagger supports modules without public methods.

func (_ *RepoOpts) Noop() {}

func (_ *WorkflowOpts) Noop() {}

func (_ *WorkflowRunOpts) Noop() {}

func (_ *WorkflowRunDirectoryOpts) Noop() {}

func (_ GithubRepository) Noop() {}

func (_ RepositoryRef) Noop() {}

func (_ *RepoConfig) Noop() {}

func (_ *WorkflowsConfig) Noop() {}

func (_ *WorkflowRunConfig) Noop() {}
