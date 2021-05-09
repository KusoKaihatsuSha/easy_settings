// Package easy_settings
// standalone DB with settings and other information
//
package easy_settings

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// Patern include randomize function
type Patern struct {
	Generate func(int) string
}

// Item is elements of Data
type Item struct {
	Parent *Items
	Name   string
	ID     string
	Values []*Values
}

// Values consist Pair values of data element
type Values struct {
	Key   string
	Value string
}

// Items include elements of Data
type Items struct {
	ID    string
	Name  string
	Items []*Item
	Db    *DataBase
}

// DataBase working with Bolt DB
type DataBase struct {
	Db      *bolt.DB
	Err     error
	Backets *[]DataBaseBucket
	Lock    chan bool
	Timeout time.Duration
}

// DataBaseBucket working with Bolt DB childs - buckets
type DataBaseBucket struct {
	Parent  *bolt.DB
	Err     error
	Name    string
	Lock    chan bool
	Timeout time.Duration
	Inc     int
}

// NewDB(string) *DataBase
// consructor of DB Dolt function
func NewDB(name string) *DataBase {
	return new(DataBase).Open(name)
}

// (*DataBase) Open(string) *DataBase
// open DB file
func (o *DataBase) Open(name string) *DataBase {
	o.Db, _ = bolt.Open(name+".db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	return o
}

// (*DataBase) Close()
// close DB file
func (obj *DataBase) Close() {
	obj.Db.Close()
}

// (*DataBase) Bucket(string) *DataBaseBucket
// return or create bucket of db
func (obj *DataBase) Bucket(name string) *DataBaseBucket {
	objb := new(DataBaseBucket)
	objb.Name = name
	objb.Parent = obj.Db
	obj.Db.Update(func(tx *bolt.Tx) error {
		_, obj.Err = tx.CreateBucketIfNotExists([]byte(objb.Name))
		if obj.Err != nil {
			return fmt.Errorf("-")
		}
		return nil
	})
	return objb
}

// (*DataBaseBucket) Delete()
// delete bucket
func (obj *DataBaseBucket) Delete() {
	obj.Parent.Update(func(tx *bolt.Tx) error {
		obj.Err = tx.DeleteBucket([]byte(obj.Name))
		if obj.Err != nil {
			return fmt.Errorf("-")
		}
		return nil
	})
}

// (*DataBase) Delete(string)
// delete bucket by name
func (obj *DataBase) Delete(name string) {
	obj.Db.Update(func(tx *bolt.Tx) error {
		obj.Err = tx.DeleteBucket([]byte(name))
		if obj.Err != nil {
			return fmt.Errorf("-")
		}
		return nil
	})
}

// (*DataBaseBucket) Add(string, string, ...bool) string
// Add new elements for saving in db as settings
func (obj *DataBaseBucket) Add(key string, value string, uniq ...bool) string {
	uid := ""

	if len(uniq) > 0 {
		if uniq[0] {
			gen := new(Patern)
			gen.Generate = Generate
			uid = gen.Generate(24)
		}
	}

	tx, _ := obj.Parent.Begin(true)
	defer tx.Rollback()

	b := tx.Bucket([]byte(obj.Name))

	id, _ := b.NextSequence()
	obj.Err = b.Put([]byte(key+uid+strconv.FormatUint(id, 10)), []byte(value))

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	return key + uid
}

func hasPart(text []byte, part []byte) bool {
	return bytes.Contains(text, part)
}

// (*DataBaseBucket) Get(string, ...bool) map[string]string
// get elements from db
func (obj *DataBaseBucket) Get(key string, prefix ...bool) map[string]string {

	prfx := false
	if len(prefix) > 0 {
		if prefix[0] {
			prfx = true
		}
	}
	m := make(map[string]string)
	obj.Parent.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(obj.Name)).Cursor()
		for k, v := b.Seek([]byte(key)); k != nil; k, v = b.Next() {
			if hasPart(k, []byte(key)) {
				if key == string(k) || prfx {
					m[string(k)] = string(v)
				}
			}
		}
		return nil
	})
	return m
}

// (*DataBase) PrintAll(string) map[string]([]byte)
// Print all elements by bucket name
func (obj *DataBase) PrintAll(name string) map[string]([]byte) {
	m := make(map[string]([]byte))
	i := 0
	obj.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				m[string(k)] = v
				i++
				return nil
			})
		}
		return nil
	})
	return m
}

// (*DataBaseBucket) PrintAllPrefix(string) map[string]string
// Print all elements in bucket by prefix in key
func (obj *DataBaseBucket) PrintAllPrefix(prx string) map[string]string {
	m := make(map[string]string)
	obj.Parent.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(obj.Name)).Cursor()
		prefix := []byte(prx)
		for k, v := b.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = b.Next() {
			m[string(k)] = string(v)
		}
		return nil
	})
	return m
}

// (*DataBaseBucket) Print(string) string
// Print elements by key
func (obj *DataBaseBucket) Print(key string) string {
	value := ""
	obj.Parent.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(obj.Name))
		v := b.Get([]byte(key))
		if v != nil {
			value = string(v)
		}
		return nil
	})
	return value
}

// Generate(int) string
// return random value
func Generate(i int) string {
	return randomString(i)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		switch randInt(1, 7) {
		case 1:
			bytes[i] = byte(randInt(48, 57))
		case 2:
			bytes[i] = byte(randInt(65, 90))
		case 3:
			bytes[i] = byte(randInt(97, 122))
		default:
			bytes[i] = byte(randInt(97, 122))
		}
	}
	return string(bytes)
}

func randInt(min int64, max int64) int {
	return_, _ := rand.Int(rand.Reader, big.NewInt(max-min))
	var return__ int64
	if return_.Int64() < min {
		return__ = return_.Int64() + min
	} else {
		return__ = return_.Int64()
	}
	return int(return__)
}

func (o *Item) _init() *Item {
	patGen := new(Patern)
	patGen.Generate = Generate
	o.ID = patGen.Generate(16)
	return o
}

// NewPack(string) *Item
// constructor for settings values
func NewPack(elName string) *Items {
	o := new(Items)
	return o._init(elName)
}

func (o *Items) _init(elName string) *Items {
	patGen := new(Patern)
	patGen.Generate = Generate
	o.ID = patGen.Generate(16)
	o.Db = NewDB("Data")
	o.Db.Close()
	o.Name = elName
	o.LoadDb()
	return o
}

// (*Items) Save(...bool)
// save Object to serialize value
func (o *Items) Save(flagdb ...bool) {
	prfx := false
	if len(flagdb) > 0 {
		if flagdb[0] {
			prfx = true
		}
	}
	if prfx {
		o.SaveDb()
	}

	o.SaveJson()
}

// (*Items) SaveDb()
// save Object to serialize value to DB
func (o *Items) SaveDb(flagUniq ...bool) {
	uniq := false
	if len(flagUniq) > 0 {
		if flagUniq[0] {
			uniq = true
		}
	}

	o.Db = NewDB("Data")
	tx, _ := o.Db.Db.Begin(true)
	defer tx.Rollback()
	b, _ := tx.CreateBucketIfNotExists([]byte(o.Name))

	for _, one := range o.Items {
		id, _ := b.NextSequence()
		one.Parent = nil
		json_, _ := json.MarshalIndent(one, "", "  ")
		one.Parent = o
		if uniq {
			b.Put([]byte(one.Name+strconv.FormatUint(id, 10)), json_)
		} else {
			b.Put([]byte(one.Name), json_)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	o.Db.Close()
}

// (*Items) SaveJson()
// save Object to serialize value to Json
func (o *Items) SaveJson() {
	for _, one := range o.Items {
		one.Parent = nil
	}
	Dbtmp := o.Db
	o.Db = nil
	all, _ := json.MarshalIndent(o, "", "  ")
	o.Db = Dbtmp
	for _, one := range o.Items {
		one.Parent = o
	}
	ioutil.WriteFile("Data_"+o.Name+".json", all, 0775)
}

// (*Items) LoadDb()
// Load Object from serialize value in DB
func (o *Items) LoadDb() {
	o.Db = NewDB("Data")
	for _, v := range o.Db.PrintAll(o.Name) {
		n := new(Item)
		json.Unmarshal(v, &n)
		o.Add(n)
	}
	o.Db.Close()
}

// (*Items) UnmarshalJSON([]byte) error
// wrapper around unmarshaling data
func (o *Items) UnmarshalJSON(data []byte) error {
	type It Items
	t := new(It)
	t.Items = o.Items
	err := json.Unmarshal(data, t)
	o.Items = t.Items
	o.ID = t.ID
	return err
}

// (*Items) LoadJson()
// Load Object from serialize value
func (o *Items) LoadJson() {
	b, _ := ioutil.ReadFile("Data" + ".json")
	json.Unmarshal(b, &o)
}

// (*Items) Delete()
// delete items
func (o *Items) Delete() {
	o.Db = NewDB("Data")
	o.Db.Delete(o.Name)
	o.Db.Close()
}

// (*Item) Add(string, string) *Item
// add elements of data
func (o *Item) Add(key string, val string) *Item {
	oo := new(Values)
	oo.Key = key
	oo.Value = val
	o.Values = append(o.Values, oo)
	return o
}

// (*Item) GetValuesJson() string
// json marshaling data
func (o *Item) GetValuesJson() string {
	v, _ := json.MarshalIndent(o.Values, "", "  ")
	return string(v)
}

// (*Item) Find(string) string
// find elements by key
func (o *Item) Find(key string) string {
	if o == nil {
		return ""
	}
	for _, v := range o.Values {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

// New(string) *Item
// constructor of new item of settings
func New(type_ string) *Item {
	o := new(Item)
	o.Name = type_
	o._init()
	o.Add("name", type_)
	return o
}

// New(string) *Item
// constructor of new item of settings
func (o *Item) New(type_ string) *Item {
	oo := new(Item)
	oo.Name = type_
	oo._init()
	oo.ID = o.ID + oo.ID
	oo.Add("name", type_)
	return oo
}

// (*Items) Add(*Item) *Item
// add elements of data to Pack
func (o *Items) Add(item *Item) *Items {
	flag_ := true
	for _, v := range o.Items {
		if v.Name == item.Name {
			tmp1, _ := json.Marshal(item.Values)
			tmp2, _ := json.Marshal(v.Values)
			if string(tmp1) == string(tmp2) {
				flag_ = false
			}
		}
	}
	if flag_ {
		item.Parent = o
		o.Items = append(o.Items, item)
	}

	return o
}

// (*Items) Filter(string) *Items
// filtering information as original kind // edited
func (o *Items) Filter(key string) *Items {
	return_ := new(Items)
	*return_ = *o
	return_.Items = nil
	for _, v := range o.Items {
		if strings.HasPrefix(v.Name, key) {
			return_.Items = append(return_.Items, v)
		}
	}
	if len(return_.Items) > 0 {
		return return_
	}
	for _, v := range o.Items {
		for _, vv := range v.Values {
			if strings.HasPrefix(vv.Value, key) || strings.HasPrefix(vv.Key, key) {
				return_.Items = append(return_.Items, v)
			}
		}
	}
	return return_
}

// (*Items) Find(string) []string
// find information by key
func (o *Items) Find(key string) []string {
	ret := []string{}
	for _, v := range o.Items {
		for _, vv := range v.Values {
			if strings.Contains(vv.Key, key) {
				ret = append(ret, vv.Value)
			}
			if strings.Contains(vv.Value, key) {
				ret = append(ret, vv.Value)
			}
		}
	}
	return ret
}
