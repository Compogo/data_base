package data_base

import (
	"github.com/Compogo/compogo"
	dbClient "github.com/Compogo/db-client"
	dbMigrator "github.com/Compogo/db-migrator"
	sqlGenerator "github.com/Compogo/db-sql-generator"
)

var Component = compogo.Component{
	Dependencies: compogo.Components{
		&dbClient.Component,
		&sqlGenerator.Component,
		&dbMigrator.Component,
	},
}
