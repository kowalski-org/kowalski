package database

import (
	"errors"
	"io/fs"
	"os"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

type Knowledge struct {
	db         *clover.DB
	faissIndex *faiss.IndexFlat
	faissId    []string
}

type KnowledgeOpts struct {
	filename string
}

func New(args ...KnowledgeArgs) (*Knowledge, error) {
	opts := KnowledgeOpts{
		filename: "cloverDB",
	}
	for _, arg := range args {
		arg(&opts)
	}
	_, err := os.Stat(opts.filename)
	if errors.Is(err, fs.ErrNotExist) {
		os.MkdirAll(opts.filename, 0755)
	}
	db, err := clover.Open(opts.filename)
	if err != nil {
		return nil, err
	}
	faissIndex, err := faiss.NewIndexFlat(ollamaconnector.DefaultEmbeddingDim, 1)
	if err != nil {
		return nil, err
	}
	return &Knowledge{db: db, faissIndex: faissIndex}, nil
}

type KnowledgeArgs func(*KnowledgeOpts)

func OptionWithFile(filename string) KnowledgeArgs {
	return func(kn *KnowledgeOpts) {
		kn.filename = filename
	}

}

func (kn *Knowledge) CreateIndex(collection string) (err error) {
	kn.db.ForEach(query.NewQuery(collection), func(doc *document.Document) bool {
		embFromDB := doc.Get("EmbeddingVec").([]interface{})
		emb := make([]float32, ollamaconnector.DefaultEmbeddingDim)
		if len(embFromDB) != len(emb) {
			panic("wrong embedding dimesions")
		}
		for i := range embFromDB {
			emb[i] = float32(embFromDB[i].(float64))
		}
		err := kn.faissIndex.Add(emb)
		if err != nil {
			panic("failed to add document to faiss index")
		}
		kn.faissId = append(kn.faissId, doc.ObjectId())
		return true
	})
	// \TODO close db
	// kn.db.Close()
	return
}
