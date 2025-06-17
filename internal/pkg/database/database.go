package database

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/information"
	"github.com/timshannon/bolthold"
)

const dbSuffix = ".md"

type Knowledge struct {
	db         map[string]*bolthold.Store
	faissIndex *faiss.IndexFlat
	faissId    []string
	dbPath     string
	boldOpts   *bolthold.Options
}

type KnowledgeOpts struct {
	dbPath      string
	BoltOptions *bolthold.Options
}

type KnowledgeArgs func(*KnowledgeOpts)

func OptionWithFile(filename string) KnowledgeArgs {
	return func(kn *KnowledgeOpts) {
		kn.dbPath = filename
	}
}

var DBLocation string

func New(args ...KnowledgeArgs) (*Knowledge, error) {
	dbopts := KnowledgeOpts{
		dbPath: DBLocation,
	}
	for _, arg := range args {
		arg(&dbopts)
	}

	dbBackends, err := fs.Glob(os.DirFS(dbopts.dbPath), "*"+dbSuffix)
	if err != nil {
		return nil, err
	}
	kn := Knowledge{
		db:       make(map[string]*bolthold.Store),
		dbPath:   dbopts.dbPath,
		boldOpts: dbopts.BoltOptions,
	}
	for _, dbFilename := range dbBackends {
		store, err := bolthold.Open(path.Join(dbopts.dbPath, dbFilename), 0644, dbopts.BoltOptions)
		if err != nil {
			return nil, err
		}
		dbName := strings.TrimSuffix(dbFilename, dbSuffix)
		log.Debugf("opened db: %s file: %s", dbName, dbFilename)
		kn.db[dbName] = store
	}
	return &kn, nil
}

func (kn *Knowledge) Close() {
	for _, dbName := range kn.db {
		dbName.Close()
	}
}

func (kn *Knowledge) CreateIndex() (err error) {
	if kn.faissIndex == nil {
		collections := kn.ListCollections()
		embedding, err := GetEmbedding(collections)
		if err != nil {
			return err
		}
		embeddingDim := ollamaconnector.Ollamasettings.GetEmbeddingDimension(embedding)
		if embeddingDim <= 0 {
			return errors.New("invalid embedding dimension. Is ollama running?")
		}
		kn.faissIndex, err = faiss.NewIndexFlat(embeddingDim, 1)
		if err != nil {
			return err
		}

	}
	for collectionKey, collection := range kn.db {
		collection.ForEach(&bolthold.Query{}, func(info *information.Information) error {
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
				index := i
				if sec.IsAlias {
					index = 0
				}
				kn.faissId = append(kn.faissId, info.Hash+fmt.Sprintf(":%d", index))
			}
			return nil
		})
		log.Debugf("indexed: %s", collectionKey)
	}
	// \TODO close db
	// kn.db.Close()

	return
}

// drop the information from the database. As well the clover document id is matched
// as the hash of the file which was used to add the documentation
func (kn *Knowledge) DropInformation(docId string) (err error) {
	for _, coll := range kn.db {
		count, err := coll.Count(&information.Information{}, bolthold.Where("Hash").Eq(docId))
		if err != nil {
			return err
		}
		if count == 0 {
			continue
		}
		log.Infof("deleted document: %s", docId)
		return coll.DeleteMatching(information.Information{}, bolthold.Where("Hash").Eq(docId))
	}
	return fmt.Errorf("document wasn't found in db: %s", docId)
}
