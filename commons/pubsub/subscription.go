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
  "bytes"
  "encoding/hex"
  "encoding/binary"

  "github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
)

// Subscription
type Subscription struct {
	Hash					string `json:"hash,omitempty"`
	Endpoint			string `json:"endpoint"`
	Headers				map[string]string `json:"headers,omitempty"`
	Address				string `json:"address"`

	Created       int64 `json:"created"`
	ErrCount			int64 `json:"errCount"`
}

func (this *Subscription) MakeHash() (error) {
	addressBytes, err := hex.DecodeString(this.Address)
	if err != nil {
		utils.Error("unable to decode address", err)
		return err
	}
	endpointBytes, err := hex.DecodeString(this.Endpoint)
	if err != nil {
		utils.Error("unable to decode endpoint", err)
		return err
	}
	values := []interface{}{
		addressBytes,
		endpointBytes,
		this.Created
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Error("unable to write sunscription bytes to buffer", err)
			return err
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	this.Hash = hex.EncodeToString(hash[:])
	return nil
}

func NewSubscription(endpoint string, headers map[string]string, address string) (*Subscription, error){
	s := &Subscription{
		Endpoint: 		endpoint,
		Headers:			headers,
		Address:			address,
		Created:			utils.ToMilliSeconds(time.Now()),
		ErrCount:			0
	}
	err := s.MakeHash()
	if err != nil {
		return nil, err
	}
	return s, nil
}

type SubscriptionRequest struct {
	Topic					string `json:"topic"`
	Endpoint			string `json:"endpoint"`
	Headers				map[string]string `json:"headers,omitempty"`
	Address				string `json:"address"`
}
