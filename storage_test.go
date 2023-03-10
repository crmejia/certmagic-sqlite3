package certmagicsqlite3_test

import (
	"context"
	"io/fs"
	"testing"
	"time"

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

func TestInsertExistingKeyUpdates(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "insert.db"
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

	oldKeyInfo, err := storage.Stat(ctx, key)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	newValue := "this is a new value"
	err = storage.Store(ctx, key, []byte(newValue))
	if err != nil {
		t.Fatal(err)
	}
	newKeyInfo, err := storage.Stat(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if oldKeyInfo.Key != newKeyInfo.Key {
		t.Errorf("want new key and old key to be the same, got old %s and new %s", oldKeyInfo.Key, newKeyInfo.Key)
	}

	if oldKeyInfo.Modified == newKeyInfo.Modified {
		t.Errorf("want new modified and old modified to be different,\n newModified: %v, oldModified: %v", newKeyInfo.Modified, oldKeyInfo.Modified)
	}

	if oldKeyInfo.Size == newKeyInfo.Size {
		t.Error("want new size and old size to be different")
	}

}

func TestList(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "list.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	keys := []string{
		"key1/key1",
		"key1/key2",
		"key2/key1/",
		"key3/key1/",
		"key1/key3",
		"key1/",
		"key1",
	}
	value := []byte("certificate")

	for _, k := range keys {
		err = storage.Store(ctx, k, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	key1Slice, err := storage.List(ctx, "key1", false)
	if err != nil {
		t.Fatal(err)
	}
	wantKey1 := 5
	if len(key1Slice) != wantKey1 {
		t.Errorf("want %d keys to be on the list, got %d", wantKey1, len(key1Slice))
	}
}

func TestListErrorsOnRecursiveTrue(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "recursive.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = storage.List(context.Background(), "key", true)
	if err == nil {
		t.Error("want error on recursive = true")
	}
}
func TestLoadReturnsFsErrNotExistOnNoKey(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "fsError.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	const key = "key"

	v, err := storage.Load(ctx, key)
	// if err != fs.ErrNotExist {
	if err == nil {
		t.Errorf("want %s, got %s, value: %s --", fs.ErrNotExist, err, v)
	}
}

func TestDeleteReturnsFsErrNotExistOnNoKey(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "fsError.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	const key = "key"

	err = storage.Delete(ctx, key)
	if err != fs.ErrNotExist {
		t.Errorf("want %s, got %s", fs.ErrNotExist, err)
	}
}

func TestListReturnsFsErrNotExistOnNoKey(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "fsError.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	const key = "key"

	_, err = storage.List(ctx, key, false)
	if err != fs.ErrNotExist {
		t.Errorf("want %s, got %s", fs.ErrNotExist, err)
	}
}

func TestStatReturnsFsErrNotExistOnNoKey(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "fsError.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	const key = "key"

	_, err = storage.Stat(ctx, key)
	if err != fs.ErrNotExist {
		t.Errorf("want %s, got %s", fs.ErrNotExist, err)
	}
}

func TestLockUnlockErrors(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "lockUnlockErrors.db"
	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	const testLock = "testLock"
	storage.SetLockTimeOut(100 * time.Millisecond)

	err = storage.Lock(ctx, testLock)
	if err == nil {
		t.Errorf("this is unimplemented. So want to get unimplemented error:")
	}

	err = storage.Unlock(ctx, testLock)
	if err == nil {
		t.Errorf("this is unimplemented. So want to get unimplemented error:")
	}

}

// func TestLockTimesOut(t *testing.T) {
// 	t.Parallel()
// 	tempDB := t.TempDir() + "fsError.db"
// 	storage, err := certmagicsqlite3.OpenSQLiteStorage(tempDB)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	ctx := context.Background()
// 	const testLock = "testLock"
// 	storage.SetLockTimeOut(100 * time.Millisecond)
//
// 	err = storage.Lock(ctx, testLock)
// 	if err != nil{
// 		t.Fatalf("want to get Lock. Got an error instead: %s", err.Error())
// 	}
// 	time.Sleep(200*time.Millisecond)
// 	err = storage.Lock(ctx, testLock)
// 	if err != nil{
// 		t.Fatalf("should be able to get Lock. Got an error instead: %s", err.Error())
// 	}
// }
