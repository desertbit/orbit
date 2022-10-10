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
	"github.com/desertbit/grumble"
	"github.com/desertbit/orbit/internal/codegen/gen"
)

const (
	argOrbitFiles = "orbit-files"

	flagForce = "force"
)

var cmdGen = &grumble.Command{
	Name: "gen",
	Help: "generate go code from .orbit file. Args: <files>",
	Run:  runGen,
	Flags: func(f *grumble.Flags) {
		f.Bool("f", flagForce, false, "generate all files, ignoring their last modification time")
	},
	Args: func(a *grumble.Args) {
		a.StringList(argOrbitFiles, "the paths to the orbit files", grumble.Min(1))
	},
}

func init() {
	App.AddCommand(cmdGen)
}

func runGen(ctx *grumble.Context) (err error) {
	// Iterate over each provided file path and generate the .orbit file.
	for _, fp := range ctx.Args.StringList(argOrbitFiles) {
		err = gen.Generate(fp, ctx.Flags.Bool(flagForce))
		if err != nil {
			return
		}
	}
	return
}
