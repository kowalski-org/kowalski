package database

import (
	"fmt"

	"github.com/ostafen/clover"
)

func (kn *Knowledge) AddFile(collection string, fileName string) (err error) {
	if ok, err := kn.db.HasCollection(collection); !ok {
		err = kn.db.CreateCollection(collection)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (kn *Knowledge) AddInformation(collection string, info Information) (err error) {
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
