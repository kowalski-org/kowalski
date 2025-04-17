package database

import "github.com/ostafen/clover/v2/query"

type DocumentInfo struct {
	Id     string
	Title  string
	Source string
}

func (kn *Knowledge) List(collection string) (docLst []DocumentInfo, err error) {
	docs, err := kn.db.FindAll(query.NewQuery(collection))
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		docLst = append(docLst, DocumentInfo{
			Id:     doc.ObjectId(),
			Title:  doc.Get("Title").(string),
			Source: doc.Get("Source").(string),
		})
	}
	return
}

func (kn *Knowledge) ListCollections() (collections []string, err error) {
	return kn.db.ListCollections()
}
