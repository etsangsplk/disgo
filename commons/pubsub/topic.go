/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package pubsub

import (
  "time"
  "sync"
  "runtime"

  "google.golang.org/api/support/bundler"
)

const (
	TOPIC_EXISTS = "TOPIC_EXISTS"
	TOPIC_NOTFOUND = "TOPIC_NOTFOUND"
)

// Topic
type Topic struct {
	name								string
	PublishSettings			PublishSettings

	Subscriptions				map[string]*Subscription
	Publisher 					Publisher
	mu      						sync.RWMutex
	wg									sync.WaitGroup
	bundler							*bundler.Bundler
}

func NewTopic(name string) (*Topic){
	return &Topic{
		name: 							name,
		PublishSettings: 		DefaultPublishSettings,
		Subscriptions:			make([]*Subscription, 0),
	}
}

type PublishSettings struct {
	DelayThreshold      time.Duration
	CountThreshold      int
	// Timeout             time.Duration
	NumGoroutines				int
	ErrorThreshold			int
}

// DefaultPublishSettings holds the default values for topics' PublishSettings.
var DefaultPublishSettings = PublishSettings{
	DelayThreshold: 10 * time.Millisecond,
	CountThreshold: 100,
	// Timeout:        60 * time.Second,
	NumGoroutines: 0,
	ErrorThreshold: 5
}

type bundledMessage struct {
	msg							*Message
	res							*PublishResult
}

func (t *Topic) Publish(msg *Message) *PublishResult {
	r := &PublishResult{ready: make(chan struct{})}

	t.initBundler()
	t.mu.RLock()
	defer t.mu.RUnlock()

	err := t.bundler.Add(&bundledMessage{msg, r}, len(msg.Body))
	if err != nil {
		r.Set("", err)
	}
	return r
}

func (t *Topic) initBundler() {
	t.mu.RLock()
	noop := t.bundler != nil
	t.mu.RUnlock()
	if noop {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	// Must re-check, since we released the lock.
	if t.bundler != nil {
		return
	}

	t.bundler = bundler.NewBundler(&bundledMessage{}, func(items interface{}) {
		t.publishMessageBundle(items.([]*bundledMessage))
	})
	t.bundler.DelayThreshold = t.PublishSettings.DelayThreshold
	t.bundler.BundleCountThreshold = t.PublishSettings.CountThreshold
	if t.PublishSettings.NumGoroutines > 0 {
		t.bundler.HandlerLimit = t.PublishSettings.NumGoroutines
	} else {
		t.bundler.HandlerLimit = 25 * runtime.GOMAXPROCS(0)
	}
}

func (t *Topic) publishMessageBundle(bms []*bundledMessage) {
	pbMsgs := make([]*Message, len(bms))
	for i, bm := range bms {
		pbMsgs[i] = &Message{bm.msg.Body}
		bm.msg = nil // release bm.msg for GC
	}
	res, err := t.Publisher.Publish(PublishRequest{
		Topic:    t.name,
		Messages: pbMsgs,
	})
	for i, bm := range bms {
		if err != nil {
			bm.res.Set("", err)
		} else {
			bm.res.Set(res.MessageResults[i], nil)
		}
	}
}

func (t *Topic) Subscribe(sub *Subscription) (string, error) {
	if sub.Hash == "" {
		err := sub.MakeHash()
		if err != nil {
			return nil, err
		}
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.Subscriptions[sub.Hash] = sub
	return sub.Hash, nil
}

func (t *Topic) Unsubscribe(subHash string) (error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.Subscriptions[subHash] = nil
	return nil
}
