package database

func (kn *Knowledge) List(collection string) (docLst []string, err error) {
	qr := kn.db.Query(collection)
	docs, err := qr.FindAll()
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
