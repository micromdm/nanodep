package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/micromdm/nanodep/client"
	"github.com/micromdm/nanodep/storage"
)

const defaultFileMode = 0644

// FileStorage implements filesystem-based storage for DEP services.
type FileStorage struct {
	path string
}

// New creates a new FileStorage backend.
func New(path string) (*FileStorage, error) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			f, err := os.Stat(path)
			if err != nil {
				return nil, err
			}
			if !f.IsDir() {
				return nil, errors.New("path is not a directory")
			}
		} else {
			return nil, err
		}
	}
	return &FileStorage{path: path}, nil
}

func (s *FileStorage) tokensFilename(name string) string {
	return path.Join(s.path, name+".tokens.json")
}

func (s *FileStorage) configFilename(name string) string {
	return path.Join(s.path, name+".config.json")
}

func (s *FileStorage) profileFilename(name string) string {
	return path.Join(s.path, name+".profile.txt")
}

func (s *FileStorage) cursorFilename(name string) string {
	return path.Join(s.path, name+".cursor.txt")
}

func (s *FileStorage) tokenpkiFilename(name, kind string) string {
	return path.Join(s.path, name+".tokenpki."+kind+".txt")
}

// RetrieveAuthTokens reads the JSON DEP OAuth tokens from disk for name (DEP name).
func (s *FileStorage) RetrieveAuthTokens(_ context.Context, name string) (*client.OAuth1Tokens, error) {
	tokens := new(client.OAuth1Tokens)
	err := decodeJSONfile(s.tokensFilename(name), tokens)
	if errors.Is(err, os.ErrNotExist) {
		err = fmt.Errorf("%v: %w", err, storage.ErrNotFound)
	}
	return tokens, err
}

// StoreAuthTokens saves the DEP OAuth tokens to disk as JSON for name (DEP name).
func (s *FileStorage) StoreAuthTokens(_ context.Context, name string, tokens *client.OAuth1Tokens) error {
	f, err := os.Create(s.tokensFilename(name))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tokens)
}

func decodeJSONfile(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

// RetrieveConfig reads the JSON DEP config of a DEP name.
//
// Returns (nil, nil) if the DEP name does not exist, or if the config
// for the DEP name does not exist.
func (s *FileStorage) RetrieveConfig(_ context.Context, name string) (*client.Config, error) {
	config := new(client.Config)
	err := decodeJSONfile(s.configFilename(name), config)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// DEP name does not exist, or config for such DEP name does not exist.
		return nil, nil
	}
	return config, err
}

// StoreConfig saves the DEP config to disk as JSON for name (DEP name).
func (s *FileStorage) StoreConfig(_ context.Context, name string, config *client.Config) error {
	f, err := os.Create(s.configFilename(name))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(config)
}

// RetrieveAssignerProfile reads the assigner profile UUID and its configured
// timestamp from disk for name (DEP name).
//
// Returns an empty profile if it does not exist.
func (s *FileStorage) RetrieveAssignerProfile(_ context.Context, name string) (string, time.Time, error) {
	profileBytes, err := os.ReadFile(s.profileFilename(name))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// an 'empty' profile is valid
		return "", time.Time{}, nil
	}
	modTime := time.Time{}
	if err == nil {
		var stat fs.FileInfo
		stat, err = os.Stat(s.profileFilename(name))
		if err == nil {
			modTime = stat.ModTime()
		}
	}
	return strings.TrimSpace(string(profileBytes)), modTime, err
}

// StoreAssignerProfile saves the assigner profile UUID to disk for name (DEP name).
func (s *FileStorage) StoreAssignerProfile(_ context.Context, name string, profileUUID string) error {
	return os.WriteFile(s.profileFilename(name), []byte(profileUUID+"\n"), defaultFileMode)
}

// RetrieveCursor reads the reads the DEP fetch and sync cursor from disk
// for name (DEP name). We return an empty cursor if the cursor does not exist
// on disk.
func (s *FileStorage) RetrieveCursor(_ context.Context, name string) (string, error) {
	cursorBytes, err := os.ReadFile(s.cursorFilename(name))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// an 'empty' cursor is valid
		return "", nil
	}
	return strings.TrimSpace(string(cursorBytes)), err
}

// StoreCursor saves the DEP fetch and sync cursor to disk for name (DEP name).
func (s *FileStorage) StoreCursor(_ context.Context, name, cursor string) error {
	return os.WriteFile(s.cursorFilename(name), []byte(cursor+"\n"), defaultFileMode)
}

// StoreTokenPKI stores the PEM bytes in pemCert and pemKey to disk for name (DEP name).
func (s *FileStorage) StoreTokenPKI(_ context.Context, name string, pemCert []byte, pemKey []byte) error {
	if err := os.WriteFile(s.tokenpkiFilename(name, "staging.cert"), pemCert, 0664); err != nil {
		return err
	}
	if err := os.WriteFile(s.tokenpkiFilename(name, "staging.key"), pemKey, 0664); err != nil {
		return err
	}
	return nil
}

// copyFile non-atomically copies file at path src to file at path dst.
func copyFile(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// UpstageTokenPKI copies the staging PKI certificate and key to the current PKI certificate and key.
// Warning: this operation is not atomic.
func (s *FileStorage) UpstageTokenPKI(ctx context.Context, name string) error {
	err := copyFile(
		s.tokenpkiFilename(name, "cert"),
		s.tokenpkiFilename(name, "staging.cert"),
	)
	if err != nil {
		return err
	}
	return copyFile(
		s.tokenpkiFilename(name, "key"),
		s.tokenpkiFilename(name, "staging.key"),
	)
}

// RetrieveStagingTokenPKI reads and returns the PEM bytes for the staged
// DEP token exchange certificate and private key from disk using name (DEP name).
func (s *FileStorage) RetrieveStagingTokenPKI(ctx context.Context, name string) ([]byte, []byte, error) {
	return s.retrieveTokenPKIExtn(name, "staging.")
}

// RetrieveCurrentTokenPKI reads and returns the PEM bytes for the previously-
// upstaged DEP token exchange certificate and private key from disk using
// name (DEP name).
func (s *FileStorage) RetrieveCurrentTokenPKI(_ context.Context, name string) ([]byte, []byte, error) {
	return s.retrieveTokenPKIExtn(name, "")
}

// retrieveTokenPKIExtn reads and returns the PEM bytes for the DEP token exchange
// certificate and private key from disk using name (DEP name) and extn type.
func (s *FileStorage) retrieveTokenPKIExtn(name, extn string) ([]byte, []byte, error) {
	certBytes, err := os.ReadFile(s.tokenpkiFilename(name, extn+"cert"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("%v: %w", err, storage.ErrNotFound)
		}
		return nil, nil, err
	}
	keyBytes, err := os.ReadFile(s.tokenpkiFilename(name, extn+"key"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("%v: %w", err, storage.ErrNotFound)
		}
		return nil, nil, err
	}
	return certBytes, keyBytes, err
}
