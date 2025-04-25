package rhilex

import (
	"testing"
)

type TestStruct struct {
	Name  string
	Value int
}

func TestKVSetGet(t *testing.T) {
	store, err := NewSqliteCacheStore("./sql-kv-test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	key := "k1"
	slot := "s1"
	obj := TestStruct{Name: "hello", Value: 42}
	data, _ := EncodeStruct(obj)

	err = store.Set(slot, key, data)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := store.Get(slot, key)
	if err != nil {
		t.Fatal(err)
	}

	var decoded TestStruct
	err = DecodeStruct(raw, &decoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded != obj {
		t.Fatalf("Expected %v, got %v", obj, decoded)
	}
}

func TestStackPushPop(t *testing.T) {
	store, _ := NewSqliteCacheStore("./sql-kv-test.db")
	defer store.Close()

	slot := "stackslot"
	key := "mystack"

	for i := 0; i < 3; i++ {
		obj := TestStruct{Name: "item", Value: i}
		data, _ := EncodeStruct(obj)
		store.Push(slot, key, data)
	}

	for i := 2; i >= 0; i-- {
		data, err := store.Pop(slot, key)
		if err != nil {
			t.Fatal(err)
		}
		var out TestStruct
		_ = DecodeStruct(data, &out)
		if out.Value != i {
			t.Fatalf("Expected Value %d, got %d", i, out.Value)
		}
	}
}

func BenchmarkStackPush(b *testing.B) {
	store, _ := NewSqliteCacheStore("./sql-kv-test.db")
	defer store.Close()
	slot := "benchslot"
	key := "stack"

	obj := TestStruct{Name: "bench", Value: 123}
	data, _ := EncodeStruct(obj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Push(slot, key, data)
	}
}

func BenchmarkStackPop(b *testing.B) {
	store, _ := NewSqliteCacheStore("./sql-kv-test.db")
	defer store.Close()
	slot := "benchslot"
	key := "stack"

	obj := TestStruct{Name: "bench", Value: 123}
	data, _ := EncodeStruct(obj)
	for i := 0; i < b.N; i++ {
		store.Push(slot, key, data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Pop(slot, key)
	}
}

func BenchmarkKVSetGet(b *testing.B) {
	store, _ := NewSqliteCacheStore("./sql-kv-test.db")
	defer store.Close()

	slot := "kv"
	key := "bench"
	obj := TestStruct{Name: "kvbench", Value: 99}
	data, _ := EncodeStruct(obj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(slot, key, data)
		store.Get(slot, key)
	}
}
func TestStoreDelete(t *testing.T) {
	store, err := NewSqliteCacheStore("./sql-kv-test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	key := "k1"
	slot := "s1"
	obj := TestStruct{Name: "hello", Value: 42}
	data, _ := EncodeStruct(obj)

	err = store.Set(slot, key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Delete(slot, key)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := store.Get(slot, key)
	if err != nil {
		t.Fatal(err)
	}

	if raw != nil {
		t.Fatalf("Expected nil, got %v", raw)
	}
}

func BenchmarkStoreExists(b *testing.B) {
	store, err := NewSqliteCacheStore("./sql-kv-test.db")
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	key := "k1"
	slot := "s1"
	obj := TestStruct{Name: "hello", Value: 42}
	data, _ := EncodeStruct(obj)

	err = store.Set(slot, key, data)
	if err != nil {
		b.Fatal(err)
	}

	exists, err := store.Exists(slot, key)
	if err != nil {
		b.Fatal(err)
	}

	if !exists {
		b.Fatalf("Expected true, got false")
	}

	err = store.Delete(slot, key)
	if err != nil {
		b.Fatal(err)
	}

	exists, err = store.Exists(slot, key)
	if err != nil {
		b.Fatal(err)
	}

	if exists {
		b.Fatalf("Expected false, got true after deletion")
	}
}
func BenchmarkStoreCodec(b *testing.B) {
	store, err := NewSqliteCacheStore("./sql-kv-test.db")
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	key := "k1"
	slot := "s1"
	obj := TestStruct{Name: "hello", Value: 42}
	data, _ := EncodeStruct(obj)

	err = store.Set(slot, key, data)
	if err != nil {
		b.Fatal(err)
	}

	raw, err := store.Get(slot, key)
	if err != nil {
		b.Fatal(err)
	}

	var decoded TestStruct
	err = DecodeStruct(raw, &decoded)
	if err != nil {
		b.Fatal(err)
	}

	if decoded != obj {
		b.Fatalf("Expected %v, got %v", obj, decoded)
	}
	err = store.Delete(slot, key)
	if err != nil {
		b.Fatal(err)
	}
	raw, err = store.Get(slot, key)
	if err != nil {
		b.Fatal(err)
	}
	if raw != nil {
		b.Fatalf("Expected nil, got %v", raw)
	}
	err = store.Set(slot, key, data)
	if err != nil {
		b.Fatal(err)
	}
	raw, err = store.Get(slot, key)
	if err != nil {
		b.Fatal(err)
	}
	var decoded2 TestStruct
	err = DecodeStruct(raw, &decoded2)
	if err != nil {
		b.Fatal(err)
	}
	if decoded2 != obj {
		b.Fatalf("Expected %v, got %v", obj, decoded2)
	}
	err = store.Delete(slot, key)
	if err != nil {
		b.Fatal(err)
	}
	raw, err = store.Get(slot, key)
	if err != nil {
		b.Fatal(err)
	}
	if raw != nil {
		b.Fatalf("Expected nil, got %v", raw)
	}
}
