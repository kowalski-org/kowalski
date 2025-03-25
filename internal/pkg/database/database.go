package database

import (
	"github.com/ostafen/clover"
)

type Knowledge struct {
	db *clover.DB
}

type KnowledgeOpts struct {
	filename string
}

func New(args ...KnowledgeArgs) *Knowledge {
	opts := KnowledgeOpts{
		filename: "cloverDB",
	}
	for _, arg := range args {
		arg(&opts)
	}
	db, _ := clover.Open(opts.filename)
	return &Knowledge{db: db}
}

type KnowledgeArgs func(*KnowledgeOpts)

func WithFile(filename string) KnowledgeArgs {
	return func(kn *KnowledgeOpts) {
		kn.filename = filename
	}

}

type Information struct {
	OS           string
	Title        string
	Sections     []string
	EmbeddingVec []float64
	Content      string
}
