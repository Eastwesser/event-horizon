package sqb_test

import (
	"testing"

	"github.com/Eastwesser/event-horizon/pkg/sqb"
)

func TestSelectWhere(t *testing.T) {
	sql, args, err := sqb.Select("id", "email").From("users").Where("email = $1", "a@b.c").ToSql()
	if err != nil {
		t.Fatal(err)
	}
	want := "SELECT id, email FROM users WHERE email = $1"
	if sql != want || len(args) != 1 {
		t.Fatalf("got %q %v", sql, args)
	}
}

func TestInsertReturning(t *testing.T) {
	sql, args, err := sqb.Insert("users").
		Columns("email", "password_hash").
		Values("a@b.c", "hash").
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || sql == "" {
		t.Fatalf("%s %v", sql, args)
	}
}
