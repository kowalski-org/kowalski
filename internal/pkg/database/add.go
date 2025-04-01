package database

import (
	"log"

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
	info.CreateHash()
	qr := kn.db.Query(collection).Where(clover.Field("Hash").Eq(info.Hash))
	docs, _ := qr.FindAll()
	if len(docs) == 0 {
		doc := clover.NewDocumentOf(info)
		docId, _ := kn.db.InsertOne(collection, doc)
		log.Printf("Added id: %s sum: %s\n", docId, info.Hash)
	} else {
		log.Printf("Found document: %s\n", docs[0].ObjectId())
	}
	return nil

}
