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
 	"bytes"
	"net/http"

	"github.com/pkg/errors"
)

type HTTPPublisher struct {
	Topics 		map[string]*Topic
}

func NewHTTPPublisher() (*HTTPPublisher){
	hPub := &HTTPPublisher{}
	hPub.Init()
	return hPub
}

func (p *HTTPPublisher) Init(){
	p.Topics = make(map[string]*Topic)
}

func (p *HTTPPublisher) CreateTopic(name string) (*Topic, error) {
	if p.TopicExists(name){
		return nil, errors.New(TOPIC_EXISTS)
	}
	topic := NewTopic(name)
	topic.Publisher = p
	p.Topics[topic.name] = topic

	return topic, nil
}

func (p *HTTPPublisher) TopicExists(name string) (bool){
	_, ok := p.Topics[name]
	return ok
}

func (p *HTTPPublisher) GetTopic(name string) (*Topic, error){
	if p.TopicExists(name){
		return p.Topics[name], nil
	}
	return nil, errors.New(TOPIC_NOTFOUND)
}

func (p *HTTPPublisher) Publish(pr PublishRequest) (*PublishResponse, error){
	topic, err := p.GetTopic(pr.Topic)
	if err != nil {
		return nil, err
	}
	var msg []byte
	msg, err = pr.GetMessagesBytes()
	if err != nil {
		return nil, err
	}
	res := &PublishResponse{make([]string, len(topic.Subscriptions))}
	client := &http.Client{}
	buf := bytes.NewBuffer(msg)
	for i, sub := range topic.Subscriptions {
		req, err := http.NewRequest("POST", sub.Endpoint, buf)
		for key, value := range sub.Headers {
			req.Header.Set(key, value)
		}
		_, err = client.Do(req)
    if err != nil {
      res.MessageResults[i] = err.Error()
      sub.ErrCount++
      if sub.ErrCount >= topic.ErrorThreshold {
      	err = topic.Unsubscribe(sub.hash)
      	if (err) {
      		utils.Error(err)
      	}
      }
    } else {
    	res.MessageResults[i] = "Ok"
    }
	}
	return res, nil
}