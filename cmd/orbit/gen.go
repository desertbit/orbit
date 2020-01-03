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

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/desertbit/grumble"
	"github.com/desertbit/orbit/internal/gen"
)

var cmdGen = &grumble.Command{
	Name:      "gen",
	Help:      "generate go code from .orbit file. Args: <dirs>...",
	AllowArgs: true,
	Run:       runGen,
}

func init() {
	App.AddCommand(cmdGen)
}

func runGen(ctx *grumble.Context) (err error) {
	if len(ctx.Args) == 0 {
		return errors.New("no directory specified")
	}

	// Iterate over each provided directory and generate the .orbit files in them.
	for _, dir := range ctx.Args {
		// Search the orbit file in the dir.
		var info []os.FileInfo
		info, err = ioutil.ReadDir(dir)
		if err != nil {
			return
		}

		var fileName string
		for _, fi := range info {
			if !fi.Mode().IsRegular() || !strings.HasSuffix(fi.Name(), gen.OrbitSuffix) {
				continue
			}

			fileName = fi.Name()
			break
		}

		if fileName == "" {
			continue
		}

		// Generate the go code for this file.
		err = gen.Generate(filepath.Join(dir, fileName))
		if err != nil {
			return
		}
	}
	return
}