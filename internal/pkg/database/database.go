package database

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/information"
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

var DBLocation string = "cloverDB"

func New(args ...KnowledgeArgs) (*Knowledge, error) {
	opts := KnowledgeOpts{
		filename: DBLocation,
	}
	for _, arg := range args {
		arg(&opts)
	}
	_, err := os.Stat(opts.filename)
	if errors.Is(err, fs.ErrNotExist) {
		os.MkdirAll(opts.filename, 0755)
	}
	log.Debugf("opening database: %s", opts.filename)
	db, err := clover.Open(opts.filename)
	if err != nil {
		return nil, err
	}
	faissIndex, err := faiss.NewIndexFlat(ollamaconnector.Ollamasettings.GetEmbeddingSize(), 1)
	if err != nil {
		return nil, err
	}
	return &Knowledge{db: db, faissIndex: faissIndex}, nil
}

func (kn *Knowledge) Close() {
	kn.db.Close()
}

type KnowledgeArgs func(*KnowledgeOpts)

func OptionWithFile(filename string) KnowledgeArgs {
	return func(kn *KnowledgeOpts) {
		kn.filename = filename
	}

}

func (kn *Knowledge) CreateIndex(collections []string) (err error) {
	if len(collections) == 0 {
		collections, err = kn.db.ListCollections()
	}
	if err != nil {
		return err
	}
	for _, collection := range collections {
		kn.db.ForEach(query.NewQuery(collection), func(doc *document.Document) bool {
			var info information.Information
			err := doc.Unmarshal(&info)
			if err != nil {
				return false
			}
			for i, sec := range info.Sections {
				// will have to convert from float64 to float32
				/*
					emb := make([]float32, ollamaconnector.Ollamasettings.GetEmbeddingSize())
					if len(sec.EmbeddingVec) != len(emb) {
						panic(fmt.Sprintf("wrong embedding dimensions faiss: %d emb: %d", len(sec.EmbeddingVec), len(emb)))
					}
					for j := range sec.EmbeddingVec {
						emb[j] = float32(sec.EmbeddingVec[j])
					}
				*/
				if len(sec.EmbeddingVec) == 0 {
					log.Debugf("couldn't add %s %d\n", sec.Title, len(sec.EmbeddingVec))
					continue
				}
				err := kn.faissIndex.Add(sec.EmbeddingVec)
				if err != nil {
					panic("failed to add document to faiss index")
				}
				kn.faissId = append(kn.faissId, doc.ObjectId()+fmt.Sprintf(":%d", i))
			}
			return true
		})
	}
	// \TODO close db
	// kn.db.Close()
	return
}
