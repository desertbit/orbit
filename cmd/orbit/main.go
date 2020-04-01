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
	"fmt"

	"github.com/desertbit/grumble"
	"github.com/desertbit/orbit/internal/codegen"
)

const (
	flagVersion = "version"
)

// Create the grumble app.
var App = grumble.New(&grumble.Config{
	Name:        "orbit",
	Description: "orbit's helper application",

	Flags: func(f *grumble.Flags) {
		f.Bool("v", flagVersion, false, "print the version and exit")
	},
})

func main() {
	App.OnInit(func(a *grumble.App, flags grumble.FlagMap) error {
		if flags.Bool(flagVersion) {
			fmt.Printf("version: %d.%d\n", codegen.OrbitFileVersion, codegen.CacheVersion)
		}

		return errors.New("done")
	})
	grumble.Main(App)
}
