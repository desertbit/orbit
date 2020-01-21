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
	"path/filepath"

	"github.com/desertbit/grumble"
	"github.com/desertbit/orbit/internal/gen"
)

const (
	flagForce          = "force"
	flagNoMerge        = "no-merge"
	flagStreamChanSize = "stream-chan-size"
)

var cmdGen = &grumble.Command{
	Name:      "gen",
	Help:      "generate go code from .orbit file. Args: <dirs>...",
	AllowArgs: true,
	Run:       runGen,
	Flags: func(f *grumble.Flags) {
		f.Bool("f", flagForce, false, "generate all found files, ignoring their last modification time")
		f.Bool("n", flagNoMerge, false, "generate an individual output file for each encountered .orbit file, instead of a single one")

		f.UintL(flagStreamChanSize, 3, "size of channels used as stream argument and return values")
	},
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
		var absDir string
		absDir, err = filepath.Abs(dir)
		if err != nil {
			return
		}

		err = gen.Generate(
			absDir,
			ctx.Flags.Uint(flagStreamChanSize),
			ctx.Flags.Bool(flagNoMerge),
			ctx.Flags.Bool(flagForce),
		)
		if err != nil {
			return
		}
	}
	return
}
