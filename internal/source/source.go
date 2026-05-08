package source

// File 是源里的单个原始文件
type File struct {
	// Path 是相对源根的路径，如 "rules/coding-style.md"
	Path string
	Raw  []byte
}

// Source 是单个源（本地或 git）的抽象
type Source interface {
	Name() string
	Files() ([]File, error)
}
