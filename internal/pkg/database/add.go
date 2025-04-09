package database

import (
	"log"
	"math/rand"

	"github.com/mslacken/kowalski/internal/app/ollamaconnector"
	"github.com/mslacken/kowalski/internal/pkg/docbook"
	"github.com/mslacken/kowalski/internal/pkg/information"
	"github.com/ostafen/clover/v2/document"
	"github.com/ostafen/clover/v2/query"
)

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
	info.CreateHash()
	// qr := kn.db.Query(collection).Where(clover.Field("Hash").Eq(info.Hash))
	// docs, _ := qr.FindAll()

	docs, _ := kn.db.FindAll(query.NewQuery(collection).Where(query.Field("Hash").Eq(info.Hash)))
	if len(docs) == 0 {
		_, err = info.CreateEmbedding()
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
		log.Printf("Added '%s' with id: %s sum: %s\n", info.Title, docId, info.Hash)
	} else {
		log.Printf("Found document '%s': %s %s\n", info.Title, docs[0].ObjectId(), info.Hash)
	}
	return nil

}

func (kn *Knowledge) GetInfos(question string, collections []string) (documents []information.Information, err error) {
	kn.CreateIndex(collections)
	emb, err := ollamaconnector.Ollama().GetEmbeddings([]string{question})
	if err != nil {
		return nil, err
	}
	lengthVec, indexVec, err := kn.faissIndex.Search(emb.Embeddings[0], 5)
	if err != nil {
		return nil, err
	}
	for i, indx := range indexVec {
		if indx >= 0 && indx < int64(len(kn.faissId)) {
			var dbdoc *document.Document
			if len(collections) == 0 {
				collections, err = kn.db.ListCollections()
				if err != nil {
					return
				}
			}
			for _, collection := range collections {
				dbdoc, err = kn.db.FindById(collection, kn.faissId[indx])
				if err != nil {
					return nil, err
				}
				if dbdoc.ObjectId() != "" {
					break
				}
			}
			var info information.Information
			err = dbdoc.Unmarshal(&info)
			if err != nil {
				return nil, err
			}
			info.Dist = lengthVec[i]
			documents = append(documents, info)
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
