package crypto

import (
	"bytes"
	"crypto/hmac"
	"github.com/tvdburgt/passman/store"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

var testStore *store.Store

func init() {
	testStore = store.NewStore()
	testStore.Entries = store.EntryMap{
		"foo": &store.Entry{
			Name:     "user",
			Password: []byte("tvsUzrGhzwTB9jF2"),
			Metadata: make(store.Metadata),
			Ctime:    time.Now(),
			Mtime:    time.Now(),
		},
		"bar": &store.Entry{
			Name:     "user",
			Password: []byte("GwRT7rcHFm2HfVU4"),
			Metadata: make(store.Metadata),
			Ctime:    time.Now(),
			Mtime:    time.Now(),
		},
		"baz": &store.Entry{
			Name:     "user",
			Password: []byte("QASRa4tzDSwxjzan"),
			Metadata: make(store.Metadata),
			Ctime:    time.Now(),
			Mtime:    time.Now(),
		},
	}

	err := ReadRand(testStore.Header.Salt[:])
	if err != nil {
		panic("failed to generate random salt: " + err.Error())
	}
}

func getStoreBuffer(tb testing.TB, s *store.Store, passphrase []byte) *bytes.Buffer {
	buf := new(bytes.Buffer)
	err := WriteStore(buf, s, passphrase)
	if err != nil {
		tb.Fatal(err)
	}
	return buf
}

func getRandSlice(tb testing.TB, n int) []byte {
	b := make([]byte, n)
	if err := ReadRand(b); err != nil {
		tb.Fatal(err)
	}
	return b
}

func TestWrite(t *testing.T) {
	buf := new(bytes.Buffer)
	err := WriteStore(buf, testStore, []byte("hunter2"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadWrongPass(t *testing.T) {
	buf := getStoreBuffer(t, testStore, []byte("hunter2"))

	_, err := ReadStore(buf, []byte("Hunter2"))
	if err != ErrWrongPass {
		t.Errorf("ReadStore: expected ErrWrongPass (received %v)", err)
	}
}

func TestRead(t *testing.T) {
	buf := getStoreBuffer(t, testStore, []byte("hunter2"))

	// Test with correct passphrase
	s, err := ReadStore(buf, []byte("hunter2"))
	if err != nil {
		t.Fatal(err)
	}

	// Test if store data has changed
	if !reflect.DeepEqual(testStore, s) {
		t.Error("ReadStore: deserialized store does not equal original store")
	}
}

// Test if header authentication follows the Encrypt-then-MAC scheme
func TestHeaderAuthentication(t *testing.T) {
	buf := getStoreBuffer(t, testStore, []byte("hunter2"))

	// Initialize HMAC
	_, mac := initStream([]byte("hunter2"), &testStore.Header)
	macBytes := make([]byte, mac.Size())

	// Read header
	r := io.TeeReader(buf, mac)
	if err := testStore.Header.Unmarshal(r); err != nil {
		t.Fatal(err)
	}

	// Read and compare HMAC
	if _, err := buf.Read(macBytes); err != nil {
		t.Fatal(err)
	}
	if !hmac.Equal(macBytes, mac.Sum(nil)) {
		t.Fatal("Header HMAC did not match expected value")
	}
}

// Test if store authentication follows the Encrypt-then-MAC scheme
func TestStoreAuthentication(t *testing.T) {
	buf := getStoreBuffer(t, testStore, []byte("hunter2"))

	// Initialize HMAC
	_, mac := initStream([]byte("hunter2"), &testStore.Header)
	macBytes := make([]byte, mac.Size())

	// Read store
	n := int64(buf.Len() - mac.Size())
	if _, err := io.CopyN(mac, buf, n); err != nil {
		t.Fatal(err)
	}

	// Read and compare HMAC
	if _, err := buf.Read(macBytes); err != nil {
		t.Fatal(err)
	}
	if !hmac.Equal(macBytes, mac.Sum(nil)) {
		t.Fatal("Store HMAC did not match expected value")
	}
}

func BenchmarkRead(b *testing.B) {
	buffer := getStoreBuffer(b, testStore, []byte("hunter2"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buf := bytes.NewBuffer(buffer.Bytes())
		b.StartTimer()
		ReadStore(buf, []byte("hunter2"))
	}
}

func BenchmarkReadWrongPass(b *testing.B) {
	buffer := getStoreBuffer(b, testStore, []byte("hunter2"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buf := bytes.NewBuffer(buffer.Bytes())
		b.StartTimer()
		ReadStore(buf, []byte("Hunter2"))
	}
}

func BenchmarkWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteStore(ioutil.Discard, testStore, []byte("hunter2"))
	}
}

func BenchmarkScrypt_14_8_1(b *testing.B) {
	phrase := getRandSlice(b, 64)
	salt := getRandSlice(b, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deriveKeys(phrase, salt, 14, 8, 1)
	}
}

func BenchmarkScrypt_20_8_1(b *testing.B) {
	phrase := getRandSlice(b, 64)
	salt := getRandSlice(b, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deriveKeys(phrase, salt, 20, 8, 1)
	}
}
