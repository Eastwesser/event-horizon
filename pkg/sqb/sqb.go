// Package sqb is a minimal Squirrel-like SQL builder (Week 3) with zero deps.
// Prefer github.com/Masterminds/squirrel when you can pull modules; this keeps
// Event Horizon buildable offline while teaching the same pattern.
package sqb

import (
	"fmt"
	"strings"
)

type SelectBuilder struct {
	cols  []string
	from  string
	where []string
	args  []any
	group string
}

func Select(cols ...string) SelectBuilder {
	return SelectBuilder{cols: cols}
}

func (b SelectBuilder) From(table string) SelectBuilder {
	b.from = table
	return b
}

func (b SelectBuilder) Where(cond string, args ...any) SelectBuilder {
	b.where = append(b.where, cond)
	b.args = append(b.args, args...)
	return b
}

func (b SelectBuilder) GroupBy(expr string) SelectBuilder {
	b.group = expr
	return b
}

func (b SelectBuilder) ToSql() (string, []any, error) {
	if b.from == "" {
		return "", nil, fmt.Errorf("sqb: From is required")
	}
	cols := "*"
	if len(b.cols) > 0 {
		cols = strings.Join(b.cols, ", ")
	}
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(cols)
	sb.WriteString(" FROM ")
	sb.WriteString(b.from)
	if len(b.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.where, " AND "))
	}
	if b.group != "" {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(b.group)
	}
	return sb.String(), b.args, nil
}

type InsertBuilder struct {
	table  string
	cols   []string
	values []any
	ret    string
}

func Insert(table string) InsertBuilder {
	return InsertBuilder{table: table}
}

func (b InsertBuilder) Columns(cols ...string) InsertBuilder {
	b.cols = cols
	return b
}

func (b InsertBuilder) Values(vals ...any) InsertBuilder {
	b.values = vals
	return b
}

func (b InsertBuilder) Suffix(s string) InsertBuilder {
	b.ret = s
	return b
}

func (b InsertBuilder) ToSql() (string, []any, error) {
	if b.table == "" || len(b.cols) == 0 {
		return "", nil, fmt.Errorf("sqb: table and columns required")
	}
	if len(b.cols) != len(b.values) {
		return "", nil, fmt.Errorf("sqb: columns/values length mismatch")
	}
	ph := make([]string, len(b.values))
	for i := range b.values {
		ph[i] = fmt.Sprintf("$%d", i+1)
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		b.table, strings.Join(b.cols, ", "), strings.Join(ph, ", "))
	if b.ret != "" {
		sql += " " + b.ret
	}
	return sql, b.values, nil
}

type UpdateBuilder struct {
	table string
	sets  []string
	args  []any
	where []string
}

func Update(table string) UpdateBuilder {
	return UpdateBuilder{table: table}
}

func (b UpdateBuilder) Set(col string, val any) UpdateBuilder {
	b.args = append(b.args, val)
	b.sets = append(b.sets, fmt.Sprintf("%s = $%d", col, len(b.args)))
	return b
}

func (b UpdateBuilder) SetRaw(expr string) UpdateBuilder {
	b.sets = append(b.sets, expr)
	return b
}

func (b UpdateBuilder) Where(cond string, args ...any) UpdateBuilder {
	// re-number placeholders in cond relative to current args len
	offset := len(b.args)
	renumbered := cond
	for i := len(args); i >= 1; i-- {
		old := fmt.Sprintf("$%d", i)
		neu := fmt.Sprintf("$%d", offset+i)
		renumbered = strings.ReplaceAll(renumbered, old, neu)
	}
	b.where = append(b.where, renumbered)
	b.args = append(b.args, args...)
	return b
}

func (b UpdateBuilder) ToSql() (string, []any, error) {
	if b.table == "" || len(b.sets) == 0 {
		return "", nil, fmt.Errorf("sqb: table and Set required")
	}
	sql := fmt.Sprintf("UPDATE %s SET %s", b.table, strings.Join(b.sets, ", "))
	if len(b.where) > 0 {
		sql += " WHERE " + strings.Join(b.where, " AND ")
	}
	return sql, b.args, nil
}
