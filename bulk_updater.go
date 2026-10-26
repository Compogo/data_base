package data_base

import (
	"context"

	dbClient "github.com/Compogo/db-client"
	"github.com/Compogo/db-client/repository"
	"github.com/doug-martin/goqu/v9"
)

type BulkUpdater struct {
	tableName string
	db        dbClient.Client
	gen       *goqu.DialectWrapper
}

func NewBulkUpdater(db dbClient.Client, gen *goqu.DialectWrapper, tableName string) *BulkUpdater {
	return &BulkUpdater{tableName: tableName, db: db, gen: gen}
}

func (b *BulkUpdater) BulkUpdate(ctx context.Context, pairs map[string]any, filters ...*repository.Filter) error {
	record := goqu.Record{}
	for columnName, value := range pairs {
		record[columnName] = value
	}

	updateDataset := b.gen.Update(record).Table(b.tableName).Prepared(true)

	var err error
	var exp goqu.Expression
	for _, filter := range filters {
		exp, err = NewFilter(filter)
		if err != nil {
			return err
		}

		updateDataset = updateDataset.Where(exp)
	}

	query, args, err := updateDataset.ToSQL()
	if err != nil {
		return err
	}

	_, err = b.db.ExecContext(ctx, query, args...)
	return err
}
