// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

type StatsOps interface {
	statsRegister()
	statsUnregister()
}
