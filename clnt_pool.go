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

// Get gets an id. It will block until there are some.
func (p *Pool) Get() uint32 {
	return <-p.id
}

// Put puts an ID. Because of the way this code
// was written, it can only panic if the ID is out of
// range. Tags are 16 bits, so in principle we can
// just change this to be uint16. Maybe that's the ticket.
func (p *Pool) Put(id uint32) {
	if id < p.low || id > p.high {
		log.Panicf("id out of range")
	}
	p.id <- id
}
