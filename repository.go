package main

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/zanz1n/gilmar-bot/logger"
	"go.mongodb.org/mongo-driver/bson"
)

var bsonEmpty = []byte{5, 0, 0, 0, 0}

type Repository[T any] struct {
	data    map[string]T
	fileDir string
	dataMu  sync.RWMutex
}

func NewRepository[T any](fileDir string) *Repository[T] {
	file, err := os.Open(fileDir)

	if err != nil {
		file, err = os.Create(fileDir)

		if err != nil {
			logger.Fatal(err)
		}

		file.Write(bsonEmpty)
		file.Close()

		file, err = os.Open(fileDir)

		if err != nil {
			logger.Fatal(err)
		}
	}

	buf, err := io.ReadAll(file)

	defer file.Close()

	if err != nil {
		logger.Fatal(err)
	}

	data := make(map[string]T)

	if err = bson.Unmarshal(buf, &data); err != nil {
		logger.Fatal(err)
	}

	return &Repository[T]{
		data:    data,
		fileDir: fileDir,
		dataMu:  sync.RWMutex{},
	}
}

func (r *Repository[T]) Get(key string) (T, bool) {
	r.dataMu.RLock()
	defer r.dataMu.RUnlock()

	value, ok := r.data[key]

	return value, ok
}

func (r *Repository[T]) Set(key string, value T) {
	r.dataMu.Lock()
	defer r.dataMu.Unlock()

	r.data[key] = value
}

func (r *Repository[T]) NotOverwriteSet(key string, value T) bool {
	r.dataMu.Lock()
	defer r.dataMu.Unlock()

	_, ok := r.data[key]

	if !ok {
		r.data[key] = value
	}

	return !ok
}

func (r *Repository[T]) Remove(key string) {
	r.dataMu.Lock()
	defer r.dataMu.Unlock()
	delete(r.data, key)
}

// It's heavily recommended to run in another goroutine
func (r *Repository[T]) Save() error {
	file, err := os.OpenFile(r.fileDir, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		logger.Error("Could not open '%s', trying to create", r.fileDir)

		if file, err = os.Create(r.fileDir); err != nil {
			logger.Error("Failed to open and create '%s', outputting content to stdout, '%s'", r.fileDir, err.Error())
			file = os.Stdout
			file.WriteString("\n")
		} else {
			logger.Info("File '%s' recreated at runtime", r.fileDir)
			file.Truncate(0)
			defer file.Close()
		}
	} else {
		file.Truncate(0)
		defer file.Close()
	}

	r.dataMu.Lock()
	defer r.dataMu.Unlock()

	buf, err := bson.Marshal(r.data)

	if err != nil {
		logger.Error("Failed to encode cached data, '%s'", err.Error())
		return err
	}

	_, err = file.Write(buf)

	if err != nil {
		logger.Error("Failed to write to file '%s', '%s'", r.fileDir, err.Error())
		return err
	}

	return nil
}

func (r *Repository[T]) GetValues() []T {
	r.dataMu.RLock()
	defer r.dataMu.RUnlock()
	arr := make([]T, len(r.data))

	i := 0
	for _, v := range r.data {
		arr[i] = v
		i++
	}

	return arr
}

func (r *Repository[T]) Transaction(key string, callback func(T) T) bool {
	r.dataMu.Lock()
	defer r.dataMu.Unlock()

	item, ok := r.data[key]

	if !ok {
		return false
	}

	newi := callback(item)

	r.data[key] = newi

	return true
}

// Blocks the current goroutine, use:
//
//	go func() {
//			r.BackgroundSave()
//	}
func (r *Repository[T]) BackgroundSave() {
	ch := time.Tick(time.Second * 30)

	for {
		<-ch
		r.Save()
	}
}
