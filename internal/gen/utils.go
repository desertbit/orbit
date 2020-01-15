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
	"os/exec"
	"unicode"
)

// execCmd executes the given command with the given arguments.
// It checks for exec specific ExitErrors and annotates them.
func execCmd(name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	err = cmd.Run()
	if err != nil {
		var eErr *exec.ExitError
		if errors.As(err, &eErr) {
			return fmt.Errorf("%s: %v", name, string(eErr.Stderr))
		}
		return
	}
	return
}

// fileExists checks, if the file at the given path exists.
// Returns true, if the file exists, else false and optionally an error.
func fileExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// strExplode splits up s, so that every uppercase rune is converted to lowercase
// and gets prepended a space.
func strExplode(s string) string {
	n := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsUpper(r) {
			if len(n) > 0 {
				n = append(n, ' ')
			}
			n = append(n, unicode.ToLower(r))
		} else {
			n = append(n, r)
		}
	}
	return string(n)
}
