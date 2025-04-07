package database

import "github.com/ostafen/clover/v2/query"

func (kn *Knowledge) List(collection string) (docLst []string, err error) {
	docs, err := kn.db.FindAll(query.NewQuery(collection))
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		docLst = append(docLst, doc.ObjectId())
	}
	return
}

func (kn *Knowledge) ListCollections() (collections []string, err error) {
	return kn.db.ListCollections()
}
