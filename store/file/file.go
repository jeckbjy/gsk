package file

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jeckbjy/gsk/store"
)

func New(opts ...Option) store.Store {
	o := Options{}
	for _, fn := range opts {
		fn(&o)
	}

	s := &fileStore{base: o.Base}
	return s
}

// 基于本地文件的存储系统
// 目前不支持watch
// Get,List不支持MVCC
type fileStore struct {
	base string
}

func (f *fileStore) normalize(key string) string {
	if f.base != "" {
		return path.Join(f.base, key)
	} else {
		return key
	}
}

func (f *fileStore) Name() string {
	return "file"
}

func (f *fileStore) List(ctx context.Context, key string, opts ...store.Option) ([]*store.KV, error) {
	o := store.Options{}
	o.Build(opts...)

	// 当前目录全部
	if key == "" {
		key = "."
	}

	name := f.normalize(key)
	s, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	results := make([]*store.KV, 0)

	if s.IsDir() {
		// 遍历所有子目录
		_ = filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// ignore dir
			if info.IsDir() {
				return nil
			}

			// ignore hide
			if strings.HasPrefix(path, ".") {
				return nil
			}

			fileKey := path
			if f.base != "" {
				// remove base and /?
				if len(path) > len(f.base)+1 {
					fileKey = path[len(f.base)+1:]
				}
			}
			if o.KeyOnly {
				results = append(results, &store.KV{Key: fileKey})
			} else {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return nil
				}
				results = append(results, &store.KV{Key: fileKey, Value: data, ModifyRevision: info.ModTime().UnixNano() / int64(time.Millisecond)})
			}

			return nil
		})
	} else {
		if o.KeyOnly {
			results = append(results, &store.KV{Key: key})
		} else {
			data, err := ioutil.ReadFile(name)
			if err != nil {
				return nil, err
			}
			results = append(results, &store.KV{Key: key, Value: data, ModifyRevision: s.ModTime().UnixNano() / int64(time.Millisecond)})
		}
	}

	if len(results) == 0 {
		return nil, store.ErrNotFound
	}

	return results, nil
}

func (f *fileStore) Get(ctx context.Context, key string, opts ...store.Option) (*store.KV, error) {
	o := store.Options{}
	o.Build(opts...)

	name := f.normalize(key)
	s, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	if s.IsDir() {
		return nil, store.ErrNotFound
	}
	if o.KeyOnly {
		return &store.KV{Key: key}, nil
	} else {
		data, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}

		kv := &store.KV{Key: key, Value: data, ModifyRevision: s.ModTime().UnixNano() / int64(time.Millisecond)}
		return kv, nil
	}
}

func (f *fileStore) Put(ctx context.Context, key string, value []byte) error {
	name := f.normalize(key)
	dir := path.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(name, value, os.ModePerm)
}

func (f *fileStore) Delete(ctx context.Context, key string, opts ...store.Option) error {
	o := store.Options{}
	o.Build(opts...)
	name := f.normalize(key)
	if !o.Prefix {
		return os.Remove(name)
	} else {
		return os.RemoveAll(name)
	}
}

func (f *fileStore) Exists(ctx context.Context, key string) (bool, error) {
	name := f.normalize(key)
	s, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	if s.IsDir() {
		return false, errors.New("is a dir,not file")
	}

	return true, nil
}

func (f *fileStore) Watch(ctx context.Context, key string, cb store.Callback, opts ...store.Option) error {
	// https://github.com/fsnotify/fsnotify
	// https://github.com/radovskyb/watcher
	// TODO: implement me
	return store.ErrNotSupport
}

func (f *fileStore) Close() error {
	return nil
}
