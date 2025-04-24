package database

import (
	"fmt"

	"github.com/openSUSE/kowalski/internal/pkg/information"
	"github.com/ostafen/clover/v2/query"
)

type DocumentInfo struct {
	Id         string
	Title      string
	Source     string
	NrFiles    int
	NrCommands int
}

func (kn *Knowledge) List(collection string) (docLst []DocumentInfo, err error) {
	docs, err := kn.db.FindAll(query.NewQuery(collection))
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		docInfo := DocumentInfo{
			Id:         doc.ObjectId(),
			Title:      fmt.Sprintf("%v", doc.Get("Title")),
			Source:     fmt.Sprintf("%v", doc.Get("Source")),
			NrFiles:    len(doc.Get("Files").([]interface{})),
			NrCommands: len(doc.Get("Commands").([]interface{})),
		}
		docLst = append(docLst, docInfo)
	}
	return
}

func (kn *Knowledge) ListCollections() (collections []string, err error) {
	return kn.db.ListCollections()
}

func (kn *Knowledge) Get(id string) (information.Information, error) {
	var info information.Information
	collections, err := kn.db.ListCollections()
	if err != nil {
		return info, err
	}
	for _, coll := range collections {
		doc, err := kn.db.FindById(coll, id)
		if err != nil {
			return info, err
		}
		if doc == nil {
			continue
		}
		err = doc.Unmarshal(&info)
		if err != nil {
			return info, err
		}
		return info, nil
	}
	return info, fmt.Errorf("couldn't find document with id: %s", id)
}
