/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package flange

type seQueue struct {
	ch chan *Span
}

func newSeQueue(capacity int32) *seQueue {
	return &seQueue{
		ch: make(chan *Span, capacity),
	}
}

// put : call multi-goroutine, will be thread-safety
func (q *seQueue) put(span *Span) bool {
	select {
	case q.ch <- span:
		return true
	default:
		return false
	}
}

// get : must call in single-goroutine, cause its not concurrent-safety
func (q *seQueue) get() *Span {
	select {
	case span := <-q.ch:
		return span
	default:
		return nil
	}
}

// len : get current length for normal
func (q *seQueue) len() int {
	return len(q.ch)
}

// cap: get current capacity
func (q *seQueue) cap() int {
	return cap(q.ch)
}
