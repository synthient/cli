package feed

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/synthient/go-synthient/v2"
)

func SnapshotMeta(client synthient.Client, stream Stream, snapshot string) (synthient.FeedSnapshotMeta, error) {
	parts, err := SnapshotParts(snapshot)
	if err != nil {
		return synthient.FeedSnapshotMeta{}, err
	}
	parts = append(parts, "meta")
	path, err := ExportPath(client.BaseAPI, stream, parts...)
	if err != nil {
		return synthient.FeedSnapshotMeta{}, err
	}
	return GetJSON[synthient.FeedSnapshotMeta](client, path)
}

func DownloadSnapshot(client synthient.Client, stream Stream, snapshot string, filename string, force bool) (int64, error) {
	if !strings.HasSuffix(filename, ".parquet") {
		return 0, fmt.Errorf("file must have a .parquet extension")
	}
	if _, err := os.Stat(filename); err == nil && !force {
		return 0, fmt.Errorf("file exists already")
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return 0, err
	}

	parts, err := SnapshotParts(snapshot)
	if err != nil {
		return 0, err
	}
	path, err := ExportPath(client.BaseAPI, stream, parts...)
	if err != nil {
		return 0, err
	}
	req, err := NewRequest(client, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := Do(client, req, http.StatusOK)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	tmp := filename + ".part"
	_ = os.Remove(tmp)
	file, err := os.Create(tmp)
	if err != nil {
		return 0, fmt.Errorf("creating file: %w", err)
	}

	size, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return 0, fmt.Errorf("writing snapshot: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return 0, fmt.Errorf("closing snapshot file: %w", closeErr)
	}
	if force {
		_ = os.Remove(filename)
	}
	err = os.Rename(tmp, filename)
	if err != nil {
		_ = os.Remove(tmp)
		return 0, fmt.Errorf("moving snapshot into place: %w", err)
	}
	return size, nil
}

func FileChecksum(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
