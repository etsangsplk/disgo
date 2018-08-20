/*
 *    This file is part of DAPoS library.
 *
 *    The DAPoS library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DAPoS library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DAPoS library.  If not, see <http://www.gnu.org/licenses/>.
 */
package pubsub

import (
	"encoding/json"
)

type Publisher interface {
	CreateTopic(name string) (*Topic, error)
	TopicExists(name string) (bool)
	GetTopic(name string) (*Topic, error)
	Publish(pr PublishRequest) (*PublishResponse, error)
}

type Message struct {
	Body			[]byte `json:"body"`
}

type PublishResult struct {
	ready    						chan struct{}
	result 							string
	err      						error
}
func (r *PublishResult) Ready() <-chan struct{} { return r.ready }
func (r *PublishResult) Get() (result string, err error) {
	return r.result, r.err
}
func (r *PublishResult) Set(result string, err error) {
	r.result = result
	r.err = err
	close(r.ready)
}

type PublishRequest struct {
	Topic 							string
	Messages            []*Message
}
func (m *PublishRequest) GetMessagesBytes() ([]byte, error) {
	return json.Marshal(m.Messages)
}

// Response for the `Publish` method.
type PublishResponse struct {
	MessageResults           []string `json:"message_results"`
}