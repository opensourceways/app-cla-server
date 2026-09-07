package repositoryimpl

import (
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BackfillAdminAddedDate populates the admin_added_date field for historical
// corp_signing documents that were completed before this field was introduced.
//
// It is idempotent: only documents where admin_added_date does not exist are
// updated, so subsequent runs scan zero rows. A failure is logged but does not
// return an error that would block service startup (the caller in initSigning
// ignores the returned error).
//
// Source of the date:
//   - primary: the user collection document whose cs_id matches the
//     corp_signing _id (hex) and account matches admin.id; the user _id
//     (ObjectId) embeds the account creation timestamp, which is the same
//     transactional flow as AddAdmin.
//   - fallback: the corp_signing _id (ObjectId) timestamp.
func BackfillAdminAddedDate(corpSigningDao, userDao dao) error {
	filter := bson.M{
		"admin.id":          bson.M{"$ne": ""},
		fieldAdminAddedDate: bson.M{"$exists": false},
	}
	project := bson.M{
		"_id":      1,
		"admin.id": 1,
	}

	var docs []corpSigningDO
	if err := corpSigningDao.GetDocs(filter, project, &docs); err != nil {
		log.Printf("admin_added_date backfill: query failed: %v", err)
		return nil
	}

	if len(docs) == 0 {
		return nil
	}

	count := 0
	for i := range docs {
		doc := &docs[i]
		date := deriveAdminAddedDate(doc, userDao)
		if date == "" {
			continue
		}

		updateFilter := bson.M{"_id": doc.Id}
		updateDoc := bson.M{fieldAdminAddedDate: date}
		if err := corpSigningDao.UpdateDocsWithoutVersion(updateFilter, updateDoc); err != nil {
			log.Printf("admin_added_date backfill: update failed for doc %s: %v", doc.Id.Hex(), err)
			continue
		}
		count++
	}

	log.Printf("admin_added_date backfill: %d documents updated", count)
	return nil
}

// deriveAdminAddedDate resolves the admin_added_date for a single corp_signing
// document. It tries the user collection first and falls back to the
// corp_signing _id timestamp.
func deriveAdminAddedDate(doc *corpSigningDO, userDao dao) string {
	hex := doc.Id.Hex()

	userFilter := bson.M{
		fieldCsId:    hex,
		fieldAccount: doc.Admin.Id,
	}

	var u userDO
	if err := userDao.GetDoc(userFilter, bson.M{"_id": 1}, &u); err == nil {
		if d := formatObjectIDDate(u.Id); d != "" {
			return d
		}
	}

	return formatObjectIDDate(doc.Id)
}

// formatObjectIDDate extracts the 4-byte timestamp prefix from a MongoDB
// ObjectId and formats it as YYYY-MM-DD.
func formatObjectIDDate(id primitive.ObjectID) string {
	return id.Timestamp().Format("2006-01-02")
}
