// Package pg contains the types for schema 'information_schema'.
package pg

// GENERATED BY XO. DO NOT EDIT.

import (
	"github.com/pkg/errors"
)

// ViewTableUsageTable is the database name for the table.
const ViewTableUsageTable = "information_schema.view_table_usage"

// ViewTableUsage represents a row from 'information_schema.view_table_usage'.
type ViewTableUsage struct {
	ViewCatalog  SQLIdentifier `json:"view_catalog"`  // view_catalog
	ViewSchema   SQLIdentifier `json:"view_schema"`   // view_schema
	ViewName     SQLIdentifier `json:"view_name"`     // view_name
	TableCatalog SQLIdentifier `json:"table_catalog"` // table_catalog
	TableSchema  SQLIdentifier `json:"table_schema"`  // table_schema
	TableName    SQLIdentifier `json:"table_name"`    // table_name
}

// Constants defining each column in the table.
const (
	ViewTableUsageViewCatalogField  = "view_catalog"
	ViewTableUsageViewSchemaField   = "view_schema"
	ViewTableUsageViewNameField     = "view_name"
	ViewTableUsageTableCatalogField = "table_catalog"
	ViewTableUsageTableSchemaField  = "table_schema"
	ViewTableUsageTableNameField    = "table_name"
)

// WhereClauses for every type in ViewTableUsage.
var (
	ViewTableUsageViewCatalogWhere  SQLIdentifierField = "view_catalog"
	ViewTableUsageViewSchemaWhere   SQLIdentifierField = "view_schema"
	ViewTableUsageViewNameWhere     SQLIdentifierField = "view_name"
	ViewTableUsageTableCatalogWhere SQLIdentifierField = "table_catalog"
	ViewTableUsageTableSchemaWhere  SQLIdentifierField = "table_schema"
	ViewTableUsageTableNameWhere    SQLIdentifierField = "table_name"
)

// QueryOneViewTableUsage retrieves a row from 'information_schema.view_table_usage' as a ViewTableUsage.
func QueryOneViewTableUsage(db XODB, where WhereClause, order OrderBy) (*ViewTableUsage, error) {
	const origsqlstr = `SELECT ` +
		`view_catalog, view_schema, view_name, table_catalog, table_schema, table_name ` +
		`FROM information_schema.view_table_usage WHERE (`

	idx := 1
	sqlstr := origsqlstr + where.String(&idx) + ") " + order.String() + " LIMIT 1"

	vtu := &ViewTableUsage{}
	err := db.QueryRow(sqlstr, where.Values()...).Scan(&vtu.ViewCatalog, &vtu.ViewSchema, &vtu.ViewName, &vtu.TableCatalog, &vtu.TableSchema, &vtu.TableName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return vtu, nil
}

// QueryViewTableUsage retrieves rows from 'information_schema.view_table_usage' as a slice of ViewTableUsage.
func QueryViewTableUsage(db XODB, where WhereClause, order OrderBy) ([]*ViewTableUsage, error) {
	const origsqlstr = `SELECT ` +
		`view_catalog, view_schema, view_name, table_catalog, table_schema, table_name ` +
		`FROM information_schema.view_table_usage WHERE (`

	idx := 1
	sqlstr := origsqlstr + where.String(&idx) + ") " + order.String()

	var vals []*ViewTableUsage
	q, err := db.Query(sqlstr, where.Values()...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for q.Next() {
		vtu := ViewTableUsage{}

		err = q.Scan(&vtu.ViewCatalog, &vtu.ViewSchema, &vtu.ViewName, &vtu.TableCatalog, &vtu.TableSchema, &vtu.TableName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		vals = append(vals, &vtu)
	}
	return vals, nil
}

// AllViewTableUsage retrieves all rows from 'information_schema.view_table_usage' as a slice of ViewTableUsage.
func AllViewTableUsage(db XODB, order OrderBy) ([]*ViewTableUsage, error) {
	const origsqlstr = `SELECT ` +
		`view_catalog, view_schema, view_name, table_catalog, table_schema, table_name ` +
		`FROM information_schema.view_table_usage`

	sqlstr := origsqlstr + order.String()

	var vals []*ViewTableUsage
	q, err := db.Query(sqlstr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for q.Next() {
		vtu := ViewTableUsage{}

		err = q.Scan(&vtu.ViewCatalog, &vtu.ViewSchema, &vtu.ViewName, &vtu.TableCatalog, &vtu.TableSchema, &vtu.TableName)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		vals = append(vals, &vtu)
	}
	return vals, nil
}