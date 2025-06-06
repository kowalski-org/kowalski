package database

import (
	"errors"
	"fmt"
	"math/rand"
	"path"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/timshannon/bolthold"

	"github.com/openSUSE/kowalski/internal/app/ollamaconnector"
	"github.com/openSUSE/kowalski/internal/pkg/docbook"
	"github.com/openSUSE/kowalski/internal/pkg/information"
)

func (kn *Knowledge) AddFile(collection string, fileName string, embeddingSize uint) (err error) {
	info, err := docbook.ParseDocBook(fileName, embeddingSize)
	if err != nil {
		return
	}
	return kn.AddInformation(collection, info)
}

func (kn *Knowledge) AddInformation(collection string, info information.Information) (err error) {
	collectionSplit := strings.Split(collection, "@")
	if len(collectionSplit) != 2 {
		return errors.New("wrong collection format must be 'name@embeddingmodell'")
	}
	embeddingName := collectionSplit[1]
	if _, ok := kn.db[collection]; !ok {
		newStore, err := bolthold.Open(path.Join(kn.dbPath, collection+".md"), 0644, kn.boldOpts)
		if err != nil {
			return err
		}
		kn.db[collection] = newStore
	}

	docs := kn.db[collection].Find(information.Information{}, bolthold.Where("Hash").Eq(info.Hash))
	if docs == nil {
		err = info.CreateEmbedding(embeddingName)
		if err != nil {
			return err
		}
		err = kn.db[collection].Insert(info.Hash, info)
		/* Do not add to faiss right now, as the index isn't stored
		err := kn.faissIndex.Add(info.EmbeddingVec)
		if err != nil {
			return err
		}
		*/
		log.Infof("added '%s' with id: %s", info.Source, info.Hash)
	} else {
		log.Infof("found document '%s': %s ", info.Source, info.Hash)
	}
	return nil
}

// Get the infos out of the database for the given question. The returned documents only
// contain this section
func (kn *Knowledge) GetInfos(question string, collections []string, nrDocs int64) (documents []information.RetSection, err error) {
	embedding, err := GetEmbedding(collections)
	if err != nil {
		return documents, err
	}
	kn.CreateIndex()
	emb, err := ollamaconnector.Ollamasettings.GetEmbeddings([]string{question}, embedding)
	if err != nil {
		return nil, err
	}
	lengthVec, indexVec, err := kn.faissIndex.Search(emb.Embeddings[0], nrDocs)
	if err != nil {
		return nil, err
	}
	for i, indx := range indexVec {
		if indx >= 0 && indx < int64(len(kn.faissId)) {
			// the faiss index vector has following format "hash:index" where
			// index refers to the section, so we have to split up
			var id []string
			var sectIndex int
			var info information.Information
			if len(collections) == 0 {
				collections = kn.ListCollections()
			}
			found := false
			for _, collection := range collections {
				id = strings.Split(kn.faissId[indx], ":")
				if len(id) != 2 {
					return nil, errors.New("document id in faiss index has wrong format")
				}
				sectIndex, err = strconv.Atoi(id[1])
				if err != nil {
					return nil, errors.New("couldn't get index of section")

				}
				err = kn.db[collection].FindOne(&info, bolthold.Where("Hash").Eq(id[0]))
				if err != nil {
					return nil, err
				}
				log.Debugf("in collection %s, doc: %s", collection, id[0])
				if info.Hash == id[0] {
					found = true
					break
				}
			}
			if !found {
				return nil, nil
			}
			ret := information.RetSection{
				Section: info.Sections[sectIndex],
				Dist:    lengthVec[i],
				Hash:    info.Hash,
			}
			log.Debugf("Doc title: %s", ret.Title)
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

/*
Get the embdding fromt the collections as their should encode the embedding. Format
is collectionName/embeddingName
*/
func GetEmbedding(collections []string) (embedding string, err error) {
	for _, col := range collections {
		collSp := strings.Split(col, "/")
		if len(collSp) != 2 {
			return embedding, fmt.Errorf("invalid format for collection: %s", col)
		}
		if embedding == "" {
			embedding = collSp[1]
		}
		if collSp[1] != embedding {
			return "", fmt.Errorf("different embeddings in collections: %s != %s", embedding, collSp[1])
		}
	}
	if embedding == "" {
		return "", fmt.Errorf("couldn't get embedding modell from %v", collections)
	}
	return
}

/*
pass DropCollection function
*/
func (kn *Knowledge) DropCollection(collection string) error {
	if _, ok := kn.db[collection]; ok {
		delete(kn.db, collection)
		return nil
	} else {
		return fmt.Errorf("couldn't drop collection: %s", collection)
	}
}
