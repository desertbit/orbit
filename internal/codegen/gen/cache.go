/*
 * ORBIT - Interlink Remote Applications
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2020 Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (c) 2020 Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package gen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/desertbit/orbit/internal/codegen"
	yaml "gopkg.in/yaml.v3"
)

const (
	cacheDirName = "orbit"
	modTimesFile = "mod_times"
)

var (
	errCacheNotFound = errors.New("gen cache not found")
	errCacheInvalid  = errors.New("gen cache is invalid")
)

type cacheEntry struct {
	LastModified time.Time `yaml:"last-modified"`
	Version      int       `yaml:"version"`
}

// compareWithGenCache compares the file at the given path against the gen cache
// and returns true, if the file is newer.
// Returns errCacheNotFound, if no cache could be found.
// Returns errCacheInvalid, if the cache found was invalid.
func compareWithGenCache(orbitFile string, force bool) (modified bool, err error) {
	gc, _, err := loadGenCache()
	if err != nil {
		return
	}
	gcEntry, foundInCache := gc[orbitFile]

	// Retrieve the file's info.
	fi, err := os.Lstat(orbitFile)
	if err != nil {
		return
	}

	// Check, if the file has been generated before.
	genFileExists, err := fileExists(strings.Replace(orbitFile, orbitSuffix, genOrbitSuffix, 1))
	if err != nil {
		return
	}

	// The file counts as modified:
	modified = !foundInCache || // if it is not found in the cache or,
		!gcEntry.LastModified.Equal(fi.ModTime()) || // if its last modification timestamp does not match the cached modification time or,
		!genFileExists || // if its generated file does not exist or,
		gcEntry.Version != codegen.CacheVersion || // if its version does not match the current cache version or,
		force // if force is enabled.
	return
}

// updateGenCache updates the gen cache on disk for the given file.
func updateGenCache(orbitFile string) (err error) {
	// Load the current gen cache.
	gc, cacheDir, err := loadGenCache()
	if err != nil {
		if errors.Is(err, errCacheInvalid) || errors.Is(err, errCacheNotFound) {
			// Create/overwrite the cache.
			gc = make(map[string]cacheEntry)
			err = nil
		} else {
			return
		}
	}

	// Retrieve the file's info.
	fi, err := os.Lstat(orbitFile)
	if err != nil {
		return
	}

	// Update the cache.
	gc[orbitFile] = cacheEntry{LastModified: fi.ModTime(), Version: codegen.CacheVersion}

	data, err := yaml.Marshal(gc)
	if err != nil {
		return
	}

	return os.WriteFile(filepath.Join(cacheDir, modTimesFile), data, filePerm)
}

// If one of the predefined errors is returned, the cacheDir is guaranteed to
// be set to a valid value.
// Returns errCacheNotFound, if no cache could be found.
// Returns errCacheInvalid, if the cache found was invalid.
func loadGenCache() (gc map[string]cacheEntry, cacheDir string, err error) {
	// Get the cache dir.
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		return
	}

	// Ensure, our directory exists.
	cacheDir = filepath.Join(cacheDir, cacheDirName)
	err = os.MkdirAll(cacheDir, dirPerm)
	if err != nil {
		return
	}

	// Read the data from the cache file.
	data, err := os.ReadFile(filepath.Join(cacheDir, modTimesFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = errCacheNotFound
		}
		return
	}

	// Parse it to our struct using yaml.
	gc = make(map[string]cacheEntry)
	err = yaml.Unmarshal(data, &gc)
	if err != nil {
		err = fmt.Errorf("%w: %v", errCacheInvalid, err)
		return
	}

	return
}
