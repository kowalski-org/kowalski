package database

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/openSUSE/kowalski/internal/pkg/information"
	"github.com/timshannon/bolthold"
)

type DocumentInfo struct {
	Id         string
	Source     string
	NrFiles    int
	NrCommands int
}

// return all documents of given collection
func (kn *Knowledge) List(collection string) (docLst []DocumentInfo, err error) {
	if collStor, ok := kn.db[collection]; ok {
		collStor.ForEach(&bolthold.Query{}, func(info *information.Information) error {
			docLst = append(docLst, DocumentInfo{
				Id:         info.Hash,
				Source:     info.Source,
				NrFiles:    len(info.Files),
				NrCommands: len(info.Commands),
			})
			return nil
		})
		return
	}
	return docLst, fmt.Errorf("collection %s not found", collection)
}

// return the whole informaton assosciated with document, either by the file hash
func (kn *Knowledge) Get(id string) (information.Information, error) {
	var info information.Information
	found := false
	for collName, coll := range kn.db {
		err := coll.Get(id, &info)
		if err == nil {
			found = true
			log.Debugf("found in coll %s doc: %s", collName, id)
			break
		}
	}
	if !found {
		return info, fmt.Errorf("couldn't find document with id: %s", id)
	}
	return info, nil
}

/*
return a list of all colletions in the database
*/
func (kn *Knowledge) ListCollections() (collections []string) {
	for key := range kn.db {
		collections = append(collections, key)
	}
	return collections
}
