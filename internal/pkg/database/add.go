package database

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/docbook"
	"github.com/openSUSE/kowalski/internal/pkg/information"
	"github.com/ostafen/clover/v2"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

var idLen = len(clover.NewObjectId())

func (kn *Knowledge) AddFile(collection string, fileName string) (err error) {
	info, err := docbook.ParseDocBook(fileName)
	if err != nil {
		return
	}
	return kn.AddInformation(collection, info)
}

func (kn *Knowledge) AddInformation(collection string, info information.Information) (err error) {
	if ok, err := kn.db.HasCollection(collection); !ok {
		err = kn.db.CreateCollection(collection)
		if err != nil {
			return err
		}
	}
	// qr := kn.db.Query(collection).Where(clover.Field("Hash").Eq(info.Hash))
	// docs, _ := qr.FindAll()

	docs, _ := kn.db.FindAll(query.NewQuery(collection).Where(query.Field("Hash").Eq(info.Hash)))
	if len(docs) == 0 {
		err = info.CreateEmbedding()
		if err != nil {
			return err
		}
		doc := document.NewDocumentOf(info)
		docId, _ := kn.db.InsertOne(collection, doc)
		/* Do not add to faiss right now, as the index isn't stored
		err := kn.faissIndex.Add(info.EmbeddingVec)
		if err != nil {
			return err
		}
		*/
		log.Infof("added '%s' with id: %s sum: %s", info.Source, docId, info.Hash)
	} else {
		log.Infof("found document '%s': %s %s", info.Source, docs[0].ObjectId(), info.Hash)
	}
	return nil

}

// Get the infos out of the database for the given question. The returned documents only
// contain this section
func (kn *Knowledge) GetInfos(question string, collections []string, nrDocs int64) (documents []information.RetSection, err error) {
	kn.CreateIndex(collections)
	emb, err := ollamaconnector.Ollamasettings.GetEmbeddings([]string{question})
	if err != nil {
		return nil, err
	}
	lengthVec, indexVec, err := kn.faissIndex.Search(emb.Embeddings[0], nrDocs)
	if err != nil {
		return nil, err
	}
	for i, indx := range indexVec {
		if indx >= 0 && indx < int64(len(kn.faissId)) {
			// the faiss index vector has following format "clover-id:index" where
			// index refers to the section, so we have to split up
			var id []string
			var sectIndex int
			var dbdoc *document.Document
			if len(collections) == 0 {
				collections, err = kn.db.ListCollections()
				if err != nil {
					return
				}
			}
			found := false
			for _, collection := range collections {
				id = strings.Split(kn.faissId[indx], ":")
				if len(id) != 2 {
					return nil, errors.New("document id in faiss index has wrong format")
				}
				sectIndex, err = strconv.Atoi(id[1])
				if err != nil {
					return nil, errors.New("couln't get index of section")

				}
				dbdoc, err = kn.db.FindById(collection, id[0])
				log.Debugf("int collection %s, getting doc: %s", collection, id[0])
				if err != nil {
					return nil, err
				}
				if dbdoc == nil {
					return nil, fmt.Errorf("couldn't find any document")
				}
				if dbdoc.ObjectId() != "" {
					found = true
					break
				}
			}
			if !found {
				return nil, nil
			}
			baseInfo := information.Information{}
			err = dbdoc.Unmarshal(&baseInfo)
			if err != nil {
				return nil, err
			}
			ret := information.RetSection{
				Section: baseInfo.Sections[sectIndex],
				Dist:    lengthVec[i],
			}
			documents = append(documents, ret)
		}
	}
	return
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
