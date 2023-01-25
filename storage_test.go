package certmagicsqlite3_test

import (
	"context"
	"testing"

	"github.com/crmejia/certmagic-sqlite3"
	"github.com/google/go-cmp/cmp"
)

func TestStorage_RoundtripStoreLoadExistsDelete(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "roundtrip.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	key := "certkey"
	want := []byte("certificate")

	exists := storage.Exists(ctx, key)
	if exists {
		t.Error("key should not exist")
	}

	err = storage.Delete(ctx, key)
	if err != nil {
		t.Error("want no error on delete non-existing key")
	}

	err = storage.Store(ctx, key, want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := storage.Load(ctx, key)
	if !cmp.Equal(want, got) {
		t.Errorf("want value to be %s, got %s", want, got)
	}

	exists = storage.Exists(ctx, key)
	if !exists {
		t.Error("want key to exist")
	}

	err = storage.Delete(ctx, key)
	if err != nil {
		t.Error("want no error on delete existing key")
	}
	
	exists = storage.Exists(ctx, key)
	if exists {
		t.Error("key should not exist")
	}
}

// func TestDeleteKey(t *testing.T){
// 	t.Parallel()
// 	tempDB := t.TempDir() + "delete.db"
// 	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	ctx := context.Background()
// 	key := "certkey"
// 	want := []byte("certificate")
// 	err = storage.Delete(ctx, key)
// 	if err != nil {
// 		t.Error("want no error on delete non-existing key")
// 	}
//
// 	err = storage.Store(ctx, key, want)
// 	if err != nil {
// 		t.Error("want no error on store")
// 	}
// }
