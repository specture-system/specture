// Package spec provides shared spec parsing, discovery, and querying.
package spec

// SpecInfo represents a parsed spec file with all extracted metadata.
type SpecInfo struct {
	Path               string
	Name               string
	Number             int
	Status             string
	CurrentTask        string
	CurrentTaskSection string
	CompleteTasks      []Task
	IncompleteTasks    []Task
}

// Task represents a single task item from a spec's task list.
type Task struct {
	Text     string
	Complete bool
	Section  string
}
