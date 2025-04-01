package database

import (
	"fmt"

	"github.com/mslacken/kowalski/internal/pkg/docbook"
	"github.com/mslacken/kowalski/internal/pkg/information"
	"github.com/ostafen/clover"
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
	if err != nil {
		return err
	}
	doc := clover.NewDocumentOf(info)
	docId, _ := kn.db.InsertOne(collection, doc)
	fmt.Printf("Add doc: %s\n", docId)
	return nil

}
