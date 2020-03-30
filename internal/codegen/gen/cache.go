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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/desertbit/orbit/internal/codegen"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v3"
)

type cacheEntry struct {
	LastModified time.Time `yaml:"last-modified"`
	Version      int       `yaml:"version"`
}

func checkIfModified(orbitFile string, force bool) (modified bool, err error) {
	genCache := make(map[string]cacheEntry)

	// Retrieve our cache dir, but only if force is not enabled.
	var ucd string
	if !force {
		ucd, err = os.UserCacheDir()
		if err != nil {
			log.Warn().
				Err(err).
				Msg("unable to retrieve cache dir, all files will be generated")
			err = nil
		} else {
			// Ensure, our directory exists.
			err = os.MkdirAll(filepath.Join(ucd, cacheDir), dirPerm)
			if err != nil {
				err = fmt.Errorf("failed to create orbit cache dir: %v", err)
				return
			}

			// Read the data from the cache file.
			var data []byte
			data, err = ioutil.ReadFile(filepath.Join(ucd, cacheDir, modTimesFile))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return
			}

			// Parse it to our struct using yaml.
			err = yaml.Unmarshal(data, &genCache)
			if err != nil {
				log.Warn().Msg("found invalid old cache, generating all files and overwriting cache")
				err = nil
			}
		}
	}

	// Retrieve the file's info.
	fi, err := os.Lstat(orbitFile)
	if err != nil {
		return
	}

	// The file counts as modified, if its last modification timestamp
	// does not match the cached modification time, if its generated file
	// does not exist, if its version does not match the current cache version,
	// or if force is enabled.
	var exists bool
	exists, err = fileExists(strings.Replace(orbitFile, orbitSuffix, genOrbitSuffix, 1))
	if err != nil {
		return
	}
	entry, ok := genCache[orbitFile]

	modified = !ok || !entry.LastModified.Equal(fi.ModTime()) || entry.Version != codegen.CacheVersion || force || !exists
	if modified {
		genCache[orbitFile] = cacheEntry{LastModified: fi.ModTime(), Version: codegen.CacheVersion}
	}

	// Save the updated gen cache to the cache dir.
	if !force && ucd != "" {
		var data []byte
		data, err = yaml.Marshal(genCache)
		if err != nil {
			return
		}

		err = ioutil.WriteFile(filepath.Join(ucd, cacheDir, modTimesFile), data, filePerm)
		if err != nil {
			return
		}
	}

	return
}
