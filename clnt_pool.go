// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import "log"

type Pool struct {
	low, high uint32
	id        chan uint32
}

func NewPool(low, high uint32) *Pool {
	id := make(chan uint32, high-low+1)

	for i := low; i < high; i++ {
		id <- i
	}
	return &Pool{id: id, low: low, high: high}
}

func (p *Pool) Get() uint32 {
	return <-p.id
}

func (p *Pool) Put(id uint32) {
	if id < p.low || id > p.high {
		log.Panicf("id out of range")
	}
	p.id <- id
}
