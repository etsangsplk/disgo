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
package types

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/patrickmn/go-cache"
)

// Gossip
type Gossip struct {
	Transaction Transaction
	Rumors      []Rumor
}

// Key
func (this Gossip) Key() string {
	return fmt.Sprintf("table-gossip-%s", this.Transaction.Hash)
}

func (this Gossip) RumorKey(txHash string) string {
	return fmt.Sprintf("cache-rumor-%s", txHash)
}

func (this *Gossip) HaveSent(cache *cache.Cache, txHash, delegateAddress string) bool {
	value, ok := cache.Get(this.RumorKey(txHash))
	if !ok{
		return false
	}
	addresses := value.([]string)
	for _, address := range addresses {
		if address == delegateAddress {
			return true
		}
	}
	return false
}

func (this *Gossip) CacheSentDelegate(cache *cache.Cache, txHash, nodeAddress string) {
	value, ok := cache.Get(this.RumorKey(txHash))
	array := make([]string, 0)
	if !ok {
		array = append(array, nodeAddress)
		cache.Set(this.RumorKey(txHash), array, GossipCacheTTL)
	} else {
		addresses := value.([]string)
		array = append(addresses, nodeAddress)
	}
	cache.Set(this.RumorKey(txHash), array, GossipCacheTTL)
}

// Cache
func (this *Gossip) Cache(cache *cache.Cache) {
	cache.Set(this.Key(), this, GossipCacheTTL)
}

// Persist
func (this *Gossip) Persist(txn *badger.Txn) error{
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	return nil
}

// PersistAndCache
func (this *Gossip) Set(txn *badger.Txn,cache *cache.Cache) error {
	this.Cache(cache)
	err := this.Persist(txn)
	if err != nil {
		return err
	}
	return nil
}

//Unset
func (this *Gossip) Unset(txn *badger.Txn,cache *cache.Cache) error {
	cache.Delete(this.Key())
	err := txn.Delete([]byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

// String
func (this Gossip) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal gossip", err)
		return ""
	}
	return string(bytes)
}

// ContainsRumor
func (this Gossip) ContainsRumor(address string) bool {
	for _, persistedRumor := range this.Rumors {
		if persistedRumor.Address == address {
			return true
		}
	}
	return false
}

// ValidRumors
func (this Gossip) ValidRumors() int {
	validRumors := 0
	for _, rumor := range this.Rumors {
		if !rumor.Verify() {
			continue
		}
		validRumors++
	}
	return validRumors
}

// Refresh
func (this *Gossip) Refresh(txn *badger.Txn) error {
	item, err := txn.Get([]byte(this.Key()))
	if err != nil {
		return err
	}
	value, err := item.Value()
	if err != nil {
		return err
	}
	gossip, err := ToGossipFromJson(value)
	if err != nil {
		return err
	}
	this = gossip
	return nil
}

// NewGossip -
func NewGossip(transaction Transaction) *Gossip {
	gossip := &Gossip{}
	gossip.Transaction = transaction
	gossip.Rumors = []Rumor{}
	return gossip
}

// ToGossipFromJson -
func ToGossipFromJson(payload []byte) (*Gossip, error) {
	gossip := &Gossip{}
	err := json.Unmarshal(payload, gossip)
	if err != nil {
		return nil, err
	}
	return gossip, nil
}

// ToJsonByGossip
func ToJsonByGossip(gossip Gossip) ([]byte, error) {
	bytes, err := json.Marshal(gossip)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ToGossipFromCache -
func ToGossipFromCache(cache *cache.Cache, txHash string) (*Gossip, error) {
	value, ok := cache.Get(fmt.Sprintf("table-gossip-%s", txHash))
	if !ok{
		return nil, ErrNotFound
	}
	gossip := value.(*Gossip)
	return gossip, nil
}

// ToGossipByKey
func ToGossipByKey(txn *badger.Txn, key []byte) (*Gossip, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	gossip, err := ToGossipFromJson(value)
	if err != nil {
		return nil, err
	}
	return gossip, err
}

// ToGossipByTransactionHash
func ToGossipByTransactionHash(txn *badger.Txn, transactionHash string) (*Gossip, error) {
	item, err := txn.Get([]byte(fmt.Sprintf("table-gossip-%s", transactionHash)))
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	gossip, err := ToGossipFromJson(value)
	if err != nil {
		return nil, err
	}
	return gossip, err
}


func GossipPaging(page int,txn *badger.Txn) ([]*Gossip, error){
	var iteratorCount = 0
	var firstItem int
	pageSize := 10
	if page <= 0 {
		return nil, ErrInvalidRequest
	}else{
		firstItem = (page * pageSize) - (pageSize - 1)
	}

	defer txn.Discard()
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	iterator := txn.NewIterator(opts)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-gossip-"))
	var Gossips = make([]*Gossip, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		iteratorCount++
		if iteratorCount >= firstItem && iteratorCount <= (firstItem+9) {
			item := iterator.Item()
			value, err := item.Value()
			if err != nil {
				utils.Error(err)
				continue
			}
			Gossip, err := ToGossipFromJson(value)
			if err != nil {
				utils.Error(err)
				continue
			}
			Gossips = append(Gossips, Gossip)
		}
		if iteratorCount > (firstItem+9){
			break
		}
	}
	return Gossips, nil //TODO: return error if empty?
}