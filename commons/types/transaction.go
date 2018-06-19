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
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"time"

	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/dispatchlabs/disgo/commons/crypto"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/pkg/errors"
)

// Transaction - The transaction info
type Transaction struct {
	Hash      string // Hash = (Type + From + To + Value + Code + Abi + Method + Params + Time)
	Type      byte
	From      string
	To        string
	Value     int64
	Code      string
	Abi       string
	Method    string
	Params    []interface{}
	Time      int64 // Milliseconds
	Signature string
	Hertz     int64  //our version of Gas
	FromName  string // Transient
	ToName    string // Transient
}

// ToTransactionFromJson -
func ToTransactionFromJson(payload []byte) (*Transaction, error) {
	transaction := &Transaction{}
	err := json.Unmarshal(payload, transaction)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// ToTransactions
func ToTransactions(txn *badger.Txn) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("table-transaction-"))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionFromJson(value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionsByFromAddress
func ToTransactionsByFromAddress(txn *badger.Txn, address string) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-from-%s", address))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionsByToAddress
func ToTransactionsByToAddress(txn *badger.Txn, address string) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-to-%s", address))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionsByType
func ToTransactionsByType(txn *badger.Txn, tipe byte) ([]*Transaction, error) {
	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()
	prefix := []byte(fmt.Sprintf("key-transaction-type-%d", tipe))
	var transactions = make([]*Transaction, 0)
	for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
		item := iterator.Item()
		value, err := item.Value()
		if err != nil {
			utils.Error(err)
			continue
		}
		transaction, err := ToTransactionByKey(txn, value)
		if err != nil {
			utils.Error(err)
			continue
		}
		fromAccount, err := ToAccountByAddress(txn, transaction.From)
		if err == nil {
			transaction.FromName = fromAccount.Name
		}
		toAccount, err := ToAccountByAddress(txn, transaction.To)
		if err == nil {
			transaction.ToName = toAccount.Name
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

// ToTransactionByKey
func ToTransactionByKey(txn *badger.Txn, key []byte) (*Transaction, error) {
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	value, err := item.Value()
	if err != nil {
		return nil, err
	}
	transaction, err := ToTransactionFromJson(value)
	if err != nil {
		return nil, err
	}
	return transaction, err
}

// NewTransferTokensTransaction -
func NewTransferTokensTransaction(privateKey string, from, to string, value, hertz, timeInMiliseconds int64) (*Transaction, error) {
	var err error
	transaction := &Transaction{}
	transaction.Type = TypeTransferTokens
	transaction.From = from
	transaction.To = to
	transaction.Value = value
	transaction.Time = timeInMiliseconds
	transaction.Hash, err = transaction.NewHash()
	if err != nil {
		return nil, err
	}
	transaction.Signature, err = transaction.NewSignature(privateKey)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// NewDeployContractTransaction -
func NewDeployContractTransaction(privateKey string, from string, code string, timeInMiliseconds int64) (*Transaction, error) {
	var err error
	transaction := &Transaction{}
	transaction.Type = TypeDeploySmartContract
	transaction.From = from
	transaction.To = ""
	transaction.Code = code
	transaction.Time = timeInMiliseconds
	transaction.Hash, err = transaction.NewHash()
	if err != nil {
		return nil, err
	}
	transaction.Signature, err = transaction.NewSignature(privateKey)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// NewExecuteContractTransaction -
func NewExecuteContractTransaction(privateKey string, from string, to string, abi string, method string, params []interface{}, timeInMiliseconds int64) (*Transaction, error) {
	var err error
	transaction := &Transaction{}
	transaction.Type = TypeExecuteSmartContract
	transaction.From = from
	transaction.To = to
	transaction.Abi = abi
	transaction.Method = method
	transaction.Params = params
	transaction.Time = timeInMiliseconds
	transaction.Hash, err = transaction.NewHash()
	if err != nil {
		return nil, err
	}
	transaction.Signature, err = transaction.NewSignature(privateKey)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// NewHash
	func (this Transaction) NewHash() (string, error) {
		fromBytes, err := hex.DecodeString(this.From)
		if err != nil {
			utils.Error("unable decode from", err)
			return "", err
		}
		toBytes, err := hex.DecodeString(this.To)
		if err != nil {
			utils.Error("unable decode to", err)
			return "", err
		}
		codeBytes, err := hex.DecodeString(this.Code)
		if err != nil {
			utils.Error("unable decode code", err)
			return "", err
		}
	var values = []interface{}{
		this.Type,
		fromBytes,
		toBytes,
		this.Value,
		codeBytes,
		[]byte(this.Abi),
		[]byte(this.Method),
		// TODO: this.Params,
		this.Time,
	}
	buffer := new(bytes.Buffer)
	for _, value := range values {
		err := binary.Write(buffer, binary.LittleEndian, value)
		if err != nil {
			utils.Fatal("unable to write transaction bytes to buffer", err)
			return "", err
		}
	}
	hash := crypto.NewHash(buffer.Bytes())
	hashString := hex.EncodeToString(hash[:])
	fmt.Printf("\n%v\nHash: %v\n\n", this.String(), hashString)
	return hex.EncodeToString(hash[:]), nil
}

// NewSignature
func (this *Transaction) NewSignature(privateKey string) (string, error) {
	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return "", err
	}
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		utils.Error("unable to decode privateKey", err)
		return "", err
	}
	signatureBytes, err := crypto.NewSignature(privateKeyBytes, hashBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signatureBytes), nil
}

// Verify
func (this Transaction) Verify() bool {
	if len(this.Hash) != crypto.HashLength*2 {
		utils.Debug("Invalid Hash")
		return false
	}
	if len(this.From) != crypto.AddressLength*2 {
		utils.Debug("Invalid From Address")
		return false
	}
	if this.To != "" && len(this.To) != crypto.AddressLength*2 {
		utils.Debug("Invalid To Address")
		return false
	}
	if this.To != "" && this.Value < 0 {
		return false
	}
	if len(this.Signature) != crypto.SignatureLength*2 {
		utils.Debug("Invalid Signature")
		return false
	}

	// Hash ok?
	hash, err := this.NewHash()
	if err != nil {
		utils.Error(err)
		return false
	}
	if this.Hash != hash {
		return false
	}

	hashBytes, err := hex.DecodeString(this.Hash)
	if err != nil {
		utils.Error("unable to decode hash", err)
		return false
	}
	signatureBytes, err := hex.DecodeString(this.Signature)
	if err != nil {
		utils.Error("unable to decode signature", err)
		return false
	}
	publicKeyBytes, err := crypto.ToPublicKey(hashBytes, signatureBytes)
	if err != nil {
		return false
	}

	// Derived address from publicKeyBytes match from?
	address := hex.EncodeToString(crypto.ToAddress(publicKeyBytes))
	if address != this.From {
		return false
	}
	return crypto.VerifySignature(publicKeyBytes, hashBytes, signatureBytes)
}

// ToTime
func (this Transaction) ToTime() time.Time {
	return time.Unix(0, this.Time*int64(time.Millisecond))
}

// Key
func (this Transaction) Key() string {
	return fmt.Sprintf("table-transaction-%s", this.Hash)
}

// TypeKey
func (this Transaction) TypeKey() string {
	return fmt.Sprintf("key-transaction-type-%d-%d-%s", this.Type, this.Time, this.Hash)
}

// TimeKey
func (this Transaction) TimeKey() string {
	return fmt.Sprintf("key-transaction-time-%d-%s", this.Time, this.Hash)
}

// FromKey
func (this Transaction) FromKey() string {
	return fmt.Sprintf("key-transaction-from-%s-%d", this.From, this.Time)
}

// ToKey
func (this Transaction) ToKey() string {
	return fmt.Sprintf("key-transaction-to-%s-%d", this.To, this.Time)
}

// Set
func (this *Transaction) Set(txn *badger.Txn) error {
	err := txn.Set([]byte(this.Key()), []byte(this.String()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.TypeKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.TimeKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.FromKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	err = txn.Set([]byte(this.ToKey()), []byte(this.Key()))
	if err != nil {
		return err
	}
	return nil
}

// String
func (this Transaction) String() string {
	bytes, err := json.Marshal(this)
	if err != nil {
		utils.Error("unable to marshal transaction", err)
		return ""
	}
	return string(bytes)
}

// UnmarshalJSON
func (this *Transaction) UnmarshalJSON(bytes []byte) error {
	var jsonMap map[string]interface{}
	var ok bool
	error := json.Unmarshal(bytes, &jsonMap)
	if error != nil {
		return error
	}
	if jsonMap["hash"] != nil {
		this.Hash, ok = jsonMap["hash"].(string)
		if !ok {
			return errors.Errorf("value for field 'hash' must be a string")
		}
	}
	if jsonMap["type"] != nil {
		typ, ok := jsonMap["type"].(float64)
		if !ok {
			return errors.Errorf("value for field 'type' must be a number")
		}
		this.Type = byte(typ)
	}
	if jsonMap["from"] != nil {
		this.From, ok = jsonMap["from"].(string)
		if !ok {
			return errors.Errorf("value for field 'from' must be a string")
		}
	}
	if jsonMap["to"] != nil {
		this.To, ok = jsonMap["to"].(string)
		if !ok {
			return errors.Errorf("value for field 'to' must be a string")
		}
	}
	if jsonMap["value"] != nil {
		val, ok := jsonMap["value"].(float64)
		if !ok {
			return errors.Errorf("value for field 'value' must be a number")
		}
		this.Value = int64(val)
	}
	if jsonMap["code"] != nil {
		this.Code, ok = jsonMap["code"].(string)
		if !ok {
			return errors.Errorf("value for field 'code' must be a string")
		}
	}
	if jsonMap["abi"] != nil {
		this.Abi, ok = jsonMap["abi"].(string)
		if !ok {
			return errors.Errorf("value for field 'abi' must be a string")
		}
	}
	if jsonMap["method"] != nil {
		this.Method, ok = jsonMap["method"].(string)
		if !ok {
			return errors.Errorf("value for field 'method' must be a string")
		}
	}
	if jsonMap["params"] != nil {
		this.Params, ok = jsonMap["params"].([]interface{})
		if !ok {
			return errors.Errorf("value for field 'params' must be an array")
		}
	}
	if jsonMap["time"] != nil {
		time, ok := jsonMap["time"].(float64)
		if !ok {
			return errors.Errorf("value for field 'time' must be a number")
		}
		this.Time = int64(time)
	}
	if jsonMap["signature"] != nil {
		this.Signature, ok = jsonMap["signature"].(string)
		if !ok {
			return errors.Errorf("value for field 'signature' must be a string")
		}
	}
	if jsonMap["hertz"] != nil {
		hertz, ok := jsonMap["type"].(float64)
		if !ok {
			return errors.Errorf("value for field 'hertz' must be a number")
		}
		this.Hertz = int64(hertz)
	}
	if jsonMap["fromName"] != nil {
		this.FromName, ok = jsonMap["fromName"].(string)
		if !ok {
			return errors.Errorf("value for field 'fromName' must be a string")
		}
	}
	if jsonMap["toName"] != nil {
		this.ToName, ok = jsonMap["toName"].(string)
		if !ok {
			return errors.Errorf("value for field 'toName' must be a string")
		}
	}
	return nil
}

// MarshalJSON
func (this Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Hash      string        `json:"hash"`
		Type      byte          `json:"type"`
		From      string        `json:"from"`
		To        string        `json:"to"`
		Value     int64         `json:"value"`
		Code      string        `json:"code"`
		Abi       string        `json:"abi"`
		Method    string        `json:"method"`
		Params    []interface{} `json:"params"`
		Time      int64         `json:"time"`
		Signature string        `json:"signature"`
		Hertz     int64         `json:"hertz"`
		FromName  string        `json:"fromName"`
		ToName    string        `json:"toName"`
	}{
		Hash:      this.Hash,
		Type:      this.Type,
		From:      this.From,
		To:        this.To,
		Value:     this.Value,
		Code:      this.Code,
		Abi:       this.Abi,
		Method:    this.Method,
		Params:    this.Params,
		Time:      this.Time,
		Signature: this.Signature,
		Hertz:     this.Hertz,
		FromName:  this.FromName,
		ToName:    this.ToName,
	})
}

// Equals
func (this Transaction) Equals(other string) bool {
	return this.Hash == other
}