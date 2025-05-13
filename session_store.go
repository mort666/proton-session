/**
 * Copyright © 2020-2025 Stephen Kapp and Reaper Technologies Limited.
 * All Rights Reserved.
 *
 * @Author: Stephen Kapp
 * @Date: 2025-5-12 23:12:55
 * @Last Modified by: Stephen Kapp
 * @Last Modified time: 2025-5-12 23:12:55
 */

package protonsession

import (
	"encoding/base64"
	"encoding/json"

	"github.com/adrg/xdg"
	tkv "github.com/miteshbsjat/textfilekv"
)

type FileStore struct {
	accountName string
	fileName    string
	CacheDir    bool
}

// Implements basic file based key value store session DB, this is not a secure storage method, the stored information is
// stored within the filesystem of the user's home directory in a cleartext file. This is more for demonstration and if a secure
// store is required it is recommended a more robust implementation is used.

func NewFileStore(filename string, account string) *FileStore {
	return &FileStore{
		fileName:    filename,
		accountName: account,
		CacheDir:    true,
	}
}

func (fs *FileStore) Load() (*SessionConfig, error) {
	var sessionCachePath string
	var kvs *tkv.KeyValueStore
	var err error

	if fs.CacheDir {
		sessionCachePath, err = xdg.CacheFile(fs.fileName)
		if err != nil {
			return nil, err
		}
		if sessionCachePath == "" {
			return nil, nil
		}
	} else {
		sessionCachePath = fs.fileName
	}

	kvs, err = tkv.NewKeyValueStore(sessionCachePath)
	if err != nil {
		return nil, err
	}

	val, ok := kvs.Get(fs.accountName)
	if !ok {
		return nil, ErrKeyNotFound
	}

	if val == "" {
		return nil, nil
	}

	config := SessionConfig{}
	err = json.Unmarshal([]byte(val), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (fs *FileStore) Save(session *SessionConfig) error {
	var sessionCachePath string

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	if fs.CacheDir {
		sessionCachePath, err = xdg.CacheFile(fs.fileName)
		if err != nil {
			return err
		}
	} else {
		sessionCachePath = fs.fileName
	}

	kvs, err := tkv.NewKeyValueStore(sessionCachePath)
	if err != nil {
		return err
	}

	err = kvs.Set(fs.accountName, string(data))
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileStore) Delete() error {
	var sessionCachePath string
	var err error

	if fs.CacheDir {
		sessionCachePath, err = xdg.CacheFile(fs.fileName)
		if err != nil {
			return err
		}
	} else {
		sessionCachePath = fs.fileName
	}

	kvs, err := tkv.NewKeyValueStore(sessionCachePath)
	if err != nil {
		return err
	}

	return kvs.Delete(fs.accountName)
}

func (fs *FileStore) List() ([]string, error) {
	var sessionCachePath string
	var err error

	if fs.CacheDir {
		sessionCachePath, err = xdg.CacheFile(fs.fileName)
		if err != nil {
			return nil, err
		}
	} else {
		sessionCachePath = fs.fileName
	}

	kvs, err := tkv.NewKeyValueStore(sessionCachePath)
	if err != nil {
		return nil, err
	}

	return kvs.Keys(), nil
}

func (fs *FileStore) Switch(account string) error {
	// FIXME validate that the account exists in the store
	fs.accountName = account
	return nil
}

func Base64Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
