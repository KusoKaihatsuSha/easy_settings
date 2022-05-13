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
	"sync"
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
	Uniq  bool
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
	//<-o.Lock

	//fmt.Println(Writable())
	var err error
	o.Db, err = bolt.Open(name+".db", 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return o.Open(name)
	}
	return o
}

// (*DataBase) Close()
// close DB file
func (o *DataBase) Close() {
	o.Db.Close()
	//o.Lock <- true
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
// func (obj *DataBaseBucket) PrintAllPrefix(prx string) map[string]string {
// 	m := make(map[string]string)
// 	obj.Parent.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(obj.Name)).Cursor()
// 		prefix := []byte(prx)
// 		for k, v := b.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = b.Next() {
// 			m[string(k)] = string(v)
// 		}
// 		return nil
// 	})
// 	return m
// }

func (obj *DataBaseBucket) PrintAllPrefix(prx string) map[string]([]byte) {
	m := make(map[string]([]byte))
	obj.Parent.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(obj.Name)).Cursor()
		prefix := []byte(prx)
		for k, v := b.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = b.Next() {
			m[string(k)] = v
		}
		return nil
	})
	return m
}

// (*DataBaseBucket) Print(string) string
// Print elements by key
// func (obj *DataBaseBucket) Print(key string) string {
// 	value := ""
// 	obj.Parent.View(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(obj.Name))
// 		v := b.Get([]byte(key))
// 		if v != nil {
// 			value = string(v)
// 		}
// 		return nil
// 	})
// 	return value
// }

func (obj *DataBaseBucket) Print(key string) []byte {
	value := []byte{}
	obj.Parent.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(obj.Name))
		v := b.Get([]byte(key))
		if v != nil {
			value = v
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
	//o := new(Items)
	//return o._init(elName)
	return nil
}

//------------

// Pack
//
func Pack(name string) *Items {
	o := new(Items)
	return o.init(name)
}

// init
//
func (o *Items) init(name string) *Items {
	patGen := new(Patern)
	patGen.Generate = Generate
	o.ID = patGen.Generate(16)
	o.Db = NewDB("data")
	o.Db.Close()
	o.Name = name
	return o
}

// Exist
//
func (o *Items) Exist(element *Item) *Item {
	temp := Item{}
	o.Db = NewDB("Data")
	json.Unmarshal(o.Db.Bucket(o.Name).Print(element.Name), &temp)
	o.Db.Close()
	for _, v := range o.Items {
		if v.Name == element.Name {
			if temp.Name != "" {
				*v = temp
			}
			for _, vv := range element.Values {
				v.Add(vv.Key, vv.Value)
			}
			return v
		}
	}
	if temp.Name == element.Name {
		if temp.Name != "" {
			for _, vv := range element.Values {
				temp.Add(vv.Key, vv.Value)
			}
			o.Items = append(o.Items, &temp)
			return &temp
		}
	}
	o.Items = append(o.Items, element)
	return element
}

// Test
//
func (o *Items) Test() {
	for _, v := range o.Items {
		for _, vv := range v.Values {
			fmt.Println(o.Name, v.Name, vv.Key, vv.Value)
		}
	}
}

// End
//
func (o *Items) End() {
	o.Db.Close()
}

// Item
//
func (o *Items) Item(name interface{}, flagFindAsPrefix ...bool) *Items {
	prfx := false
	if len(flagFindAsPrefix) > 0 {
		if flagFindAsPrefix[0] {
			prfx = true
		}
	}
	tmp := New("empty")
	save := false
	switch val := name.(type) {
	case *Item:
		tmp = val
		save = true
	case string:
		tmp.Name = val
	case int:
		tmp.Name = strconv.Itoa(val)
	case int64:
		tmp.Name = strconv.Itoa(int(val))
	default:
		return o
	}
	if !prfx {
		o.Exist(tmp)
	} else {
		o.Db = NewDB("Data")
		els := o.Db.Bucket(o.Name).PrintAllPrefix(tmp.Name)
		defer o.Db.Close()
		var itm []*Item
		for _, v := range els {
			n := new(Item)
			json.Unmarshal(v, n)
			itm = append(itm, n)
		}
		o.Db.Close()
		for _, v := range itm {
			o.Item(v, false)
		}
	}
	if save {
		o.Save()
	}
	return o
}

func (o *Items) Load() {
	o.Db = NewDB("Data")
	for _, v := range o.Items {
		json.Unmarshal(o.Db.Bucket(o.Name).Print(v.Name), &v)
	}
	o.Db.Close()
}

func (o *Items) Add(key string, val string) *Items {
	oo := new(Values)
	oo.Key = key
	oo.Value = val
	for _, v := range o.Items {
		find := false
		for _, vv := range v.Values {
			if vv.Key == oo.Key {
				vv.Value = oo.Value
				find = true
			}
		}
		if !find {
			v.Values = append(v.Values, oo)
		}
	}
	o.Save()
	return o
}

func (o *Items) Get(key string, flagFindAsPrefix ...bool) []string {
	prfx := false
	if len(flagFindAsPrefix) > 0 {
		if flagFindAsPrefix[0] {
			prfx = true
		}
	}
	var ret []string
	for _, v := range o.Items {
		for _, vv := range v.Values {
			if !prfx {
				if vv.Key == key {
					ret = append(ret, vv.Value)
				}
			} else {
				if strings.HasPrefix(vv.Key, key) {
					ret = append(ret, vv.Value)
				}
			}
		}
	}
	return ret
}

func (o *Items) Clear(key string, flagFindAsPrefix ...bool) {
	prfx := false
	if len(flagFindAsPrefix) > 0 {
		if flagFindAsPrefix[0] {
			prfx = true
		}
	}
	newValues := []*Values{}
	for _, v := range o.Items {
		for _, vv := range v.Values {
			if !prfx {
				if vv.Key != key {
					newValues = append(newValues, vv)
				}
			} else {
				if !strings.HasPrefix(vv.Key, key) {
					newValues = append(newValues, vv)
				}
			}
		}
		v.Values = newValues
	}
	o.Save()
}

func (o *Item) Add(key string, val string) *Item {
	if o == nil {
		return nil
	}
	oo := new(Values)
	oo.Key = key
	oo.Value = val
	find := false
	for _, vv := range o.Values {
		if vv.Key == oo.Key {
			vv.Value = oo.Value
			find = true
		}
	}
	if !find {
		o.Values = append(o.Values, oo)
	}
	return o
}

func (o *Items) Delete() {
	var wg sync.WaitGroup
	wg.Add(1)

	o.Db = NewDB("Data")
	tx, _ := o.Db.Db.Begin(true)
	defer tx.Rollback()
	b, _ := tx.CreateBucketIfNotExists([]byte(o.Name))

	for _, one := range o.Items {
		b.Delete([]byte(one.Name))
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	wg.Done()
	wg.Wait()
	o.Db.Close()
}

func (o *Items) Save(flagUniq ...bool) {
	var wg sync.WaitGroup
	wg.Add(1)
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
	wg.Done()
	wg.Wait()
	o.Db.Close()
}

//-----------
// func (o *Items) _init(elName string) *Items {
// 	patGen := new(Patern)
// 	patGen.Generate = Generate
// 	o.ID = patGen.Generate(16)
// 	o.Db = NewDB("Data")
// 	//o.Db.Close()
// 	o.Name = elName
// 	o.LoadDb()
// 	return o
// }

// (*Items) Save(...bool)
// save Object to serialize value
// func (o *Items) Save(flagdb ...bool) {
// 	prfx := false
// 	if len(flagdb) > 0 {
// 		if flagdb[0] {
// 			prfx = true
// 		}
// 	}
// 	if prfx {
// 		o.SaveDb()
// 	}

// 	o.SaveJson()
// }

// (*Items) SaveDb()
// save Object to serialize value to DB
// func (o *Items) SaveDb(flagUniq ...bool) {
// 	uniq := false
// 	if len(flagUniq) > 0 {
// 		if flagUniq[0] {
// 			uniq = true
// 		}
// 	}

// 	o.Db = NewDB("Data")
// 	tx, _ := o.Db.Db.Begin(true)
// 	defer tx.Rollback()
// 	b, _ := tx.CreateBucketIfNotExists([]byte(o.Name))

// 	for _, one := range o.Items {
// 		id, _ := b.NextSequence()
// 		one.Parent = nil
// 		json_, _ := json.MarshalIndent(one, "", "  ")
// 		one.Parent = o
// 		if uniq {
// 			b.Put([]byte(one.Name+strconv.FormatUint(id, 10)), json_)
// 		} else {
// 			b.Put([]byte(one.Name), json_)
// 		}
// 	}

// 	if err := tx.Commit(); err != nil {
// 		log.Fatal(err)
// 	}
// 	o.Db.Close()
// }

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
	ioutil.WriteFile("data_"+o.Name+".json", all, 0775)
}

// (*Items) SaveJson()
// save Object to serialize value to Json
func (o *Items) Json() string {
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
	return string(all)
}

// (*Items) LoadDb()
// Load Object from serialize value in DB
// func (o *Items) LoadDb() {
// 	o.Db = NewDB("Data")
// 	for _, v := range o.Db.PrintAll(o.Name) {
// 		n := new(Item)
// 		json.Unmarshal(v, &n)
// 		o.Add(n)
// 	}
// 	o.Db.Close()
// }

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
// func (o *Items) Delete() {
// 	o.Db = NewDB("Data")
// 	o.Db.Delete(o.Name)
// 	o.Db.Close()
// }

// (*Item) Add(string, string) *Item
// add elements of data
// func (o *Item) Add(key string, val string) *Item {
// 	oo := new(Values)
// 	oo.Key = key
// 	oo.Value = val
// 	o.Values = append(o.Values, oo)
// 	return o
// }

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
	//o.Add("name", type_)
	return o
}

// New(string) *Item
// constructor of new item of settings
func (o *Item) New(type_ string) *Item {
	oo := new(Item)
	oo.Name = type_
	oo._init()
	oo.ID = o.ID + oo.ID
	//oo.Add("name", type_)
	return oo
}

// (*Items) Add(*Item) *Item
// add elements of data to Pack
// func (o *Items) Add(item *Item) *Items {
// 	flag_ := true
// 	for _, v := range o.Items {
// 		if v.Name == item.Name {
// 			tmp1, _ := json.Marshal(item.Values)
// 			tmp2, _ := json.Marshal(v.Values)
// 			if string(tmp1) == string(tmp2) {
// 				flag_ = false
// 			}
// 		}
// 	}
// 	if flag_ {
// 		item.Parent = o
// 		o.Items = append(o.Items, item)
// 	}

// 	return o
// }

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
