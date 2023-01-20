package certmagicsqlite3_test

import (
	"context"
	"testing"

	"github.com/crmejia/certmagic-sqlite3"
	"github.com/google/go-cmp/cmp"
)	

func TestStorage_RoundtripStoreLoad(t *testing.T){
	t.Parallel()
	tempDB := t.TempDir() + "roundtrip.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	key := "certkey"
	want := []byte("certificate")
	err = storage.Store(ctx, key, want)
	if err != nil {
		t.Fatal(err)
	}
	
	got, err := storage.Load(ctx,key)
	if !cmp.Equal(want, got) {
		t.Errorf("want value to be %s, got %s", want, got)
	}
}
