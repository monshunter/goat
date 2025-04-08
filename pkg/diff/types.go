package diff

// FileChange represents file change information
type FileChange struct {
	Path        string       `json:"path"` // file path
	LineChanges []LineChange `json:"line_changes"`
}

// LineChange represents line-level change information
type LineChange struct {
	Start int `json:"start"` // starting line number of new code
	Lines int `json:"lines"` // number of lines of new code
}
