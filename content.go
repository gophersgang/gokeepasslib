//Package gokeepasslib is a library written in go which provides functionality to decrypt and parse keepass 2 files (kdbx)
package gokeepasslib

import (
	"encoding/xml"
	"time"
)
//Container for all elements of a keepass database
type DBContent struct {
	XMLName xml.Name  `xml:"KeePassFile"`
	Meta    *MetaData `xml:"Meta"`
	Root    *RootData `xml:"Root"`
}

//Creates a new DB content with some good defaults
func NewDBContent () (*DBContent) {
	content := new(DBContent)
	content.Meta = NewMetaData()
	content.Root = NewRootData()
	return content
}

//The metadata headers at the top of kdbx files, contains things like the name of the database
type MetaData struct {
	Generator                  string        `xml:"Generator"`
	HeaderHash                 string        `xml:"HeaderHash"`
	DatabaseName               string        `xml:"DatabaseName"`
	DatabaseNameChanged        *time.Time    `xml:"DatabaseNameChanged"`
	DatabaseDescription        string        `xml:"DatabaseDescription"`
	DatabaseDescriptionChanged *time.Time    `xml:"DatabaseDescriptionChanged"`
	DefaultUserName            string        `xml:"DefaultUserName"`
	DefaultUserNameChanged     *time.Time    `xml:"DefaultUserNameChanged"`
	MaintenanceHistoryDays     string        `xml:"MaintenanceHistoryDays"`
	Color                      string        `xml:"Color"`
	MasterKeyChanged           *time.Time    `xml:"MasterKeyChanged"`
	MasterKeyChangeRec         int64         `xml:"MasterKeyChangeRec"`
	MasterKeyChangeForce       int64         `xml:"MasterKeyChangeForce"`
	MemoryProtection           MemProtection `xml:"MemoryProtection"`
	RecycleBinEnabled          boolWrapper   `xml:"RecycleBinEnabled"`
	RecycleBinUUID             string        `xml:"RecycleBinUUID"`
	RecycleBinChanged          *time.Time    `xml:"RecycleBinChanged"`
	EntryTemplatesGroup        string        `xml:"EntryTemplatesGroup"`
	EntryTemplatesGroupChanged *time.Time    `xml:"EntryTemplatesGroupChanged"`
	HistoryMaxItems            int64         `xml:"HistoryMaxItems"`
	HistoryMaxSize             int64         `xml:"HistoryMaxSize"`
	LastSelectedGroup          string        `xml:"LastSelectedGroup"`
	LastTopVisibleGroup        string        `xml:"LastTopVisibleGroup"`
	Binaries                   Binaries      `xml:"Binaries>Binary"`
	CustomData                 string        `xml:"CustomData"`
}

//NewMetaData creates a MetaData struct with some defaults set
func NewMetaData () (*MetaData) {
	meta := new(MetaData)
	now := time.Now()
	*meta.MasterKeyChanged = now
	meta.MasterKeyChangeRec = -1
	meta.MasterKeyChangeForce = -1
	return meta
}

type MemProtection struct {
	ProtectTitle    boolWrapper `xml:"ProtectTitle"`
	ProtectUserName boolWrapper `xml:"ProtectUserName"`
	ProtectPassword boolWrapper `xml:"ProtectPassword"`
	ProtectURL      boolWrapper `xml:"ProtectURL"`
	ProtectNotes    boolWrapper `xml:"ProtectNotes"`
}

//Stores the actual content of a database (all enteries sorted into groups and the recycle bin)
type RootData struct {
	Groups         []Group             `xml:"Group"`
	DeletedObjects []DeletedObjectData `xml:"DeletedObjects>DeletedObject"`
}

//NewRootData returns a RootData struct with good defaults
func NewRootData () (*RootData) {
	root := new(RootData)
	return root
}

//Structure to store entries in their named groups for organization
type Group struct {
	UUID                    string      `xml:"UUID"`
	Name                    string      `xml:"Name"`
	Notes                   string      `xml:"Notes"`
	IconID                  int64       `xml:"IconID"`
	Times                   TimeData    `xml:"Times"`
	IsExpanded              boolWrapper `xml:"IsExpanded"`
	DefaultAutoTypeSequence string      `xml:"DefaultAutoTypeSequence"`
	EnableAutoType          string      `xml:"EnableAutoType"`
	EnableSearching         string      `xml:"EnableSearching"`
	LastTopVisibleEntry     string      `xml:"LastTopVisibleEntry"`
	Groups                  []Group     `xml:"Group,omitempty"`
	Entries                 []Entry     `xml:"Entry,omitempty"`
}

//All metadata relating to times for groups and entries, such as last modification time
type TimeData struct {
	CreationTime         *time.Time  `xml:"CreationTime"`
	LastModificationTime *time.Time  `xml:"LastModificationTime"`
	LastAcessTime        *time.Time  `xml:"LastAcessTime"`
	ExpiryTime           *time.Time  `xml:"ExpiryTime"`
	Expires              boolWrapper `xml:"Expires"`
	UsageCount           int64       `xml:"UsageCount"`
	LocationChanged      *time.Time  `xml:"LocationChanged"`
}

//NewTimeData returns a TimeData struct with good defaults (no expire time, all times set to now)
func NewTimeData () (TimeData) {
	data := TimeData{}
	now := time.Now()
	*data.CreationTime = now
	*data.LastModificationTime = now
	*data.LastAcessTime = now
	data.Expires = false
	data.UsageCount = 0
	*data.LocationChanged = now
	return data
}

//structure for each parsed entry in a keepass database
type Entry struct {
	UUID            string            `xml:"UUID"`
	IconID          int64             `xml:"IconID"`
	ForegroundColor string            `xml:"ForegroundColor"`
	BackgroundColor string            `xml:"BackgroundColor"`
	OverrideURL     string            `xml:"OverrideURL"`
	Tags            string            `xml:"Tags"`
	Times           TimeData          `xml:"Times"`
	Values          []ValueData       `xml:"String,omitempty"`
	AutoType        AutoTypeData      `xml:"AutoType"`
	Histories       []History         `xml:"History"`
	Password        []byte            `xml:"-"`
	Binaries        []BinaryReference `xml:"Binary,omitempty"`
}

//Returns true if the e's password has in-memory protection
func (e *Entry) protected() bool {
	for _, v := range e.Values {
		if v.Key == "Password" && bool(v.Value.Protected) {
			return true
		}
	}
	return false
}

//Gets the value in e corresponding with key k, or an empty string otherwise
func (e *Entry) Get (key string) (*ValueData) {
	for i,_ := range e.Values {
		if e.Values[i].Key == key {
			return &e.Values[i]
		}
	}
	return nil
}
func (e *Entry) GetContent(key string) string {
	val := e.Get(key)
	if val == nil {
		return ""
	}
	return val.Value.Content
}
func (e *Entry) GetIndex(key string) int {
	for i,_ := range e.Values {
		if e.Values[i].Key == key {
			return i
		}
	}
	return -1
}
func (e *Entry) GetPassword() string {
	return e.GetContent("Password")
}

func (e *Entry) GetPasswordIndex() int {
	return e.GetIndex("Password")
}

func (e *Entry) GetTitle() string {
	return e.GetContent("Title")
}
//Stores the history (changes) of an entry, a list of previous versions of that entry
type History struct {
	Entries []Entry `xml:"Entry"`
}

type ValueData struct {
	Key   string `xml:"Key"`
	Value V      `xml:"Value"`
}

//Wraper for the content of a value, so that it can store whether it is protected
type V struct {
	Content   string      `xml:",innerxml"`
	Protected boolWrapper `xml:"Protected,attr,omitempty"`
}

type AutoTypeData struct {
	Enabled                 boolWrapper         `xml:"Enabled"`
	DataTransferObfuscation int64               `xml:"DataTransferObfuscation"`
	Association             AutoTypeAssociation `xml:"Association"`
}

type AutoTypeAssociation struct {
	Window            string `xml:"Window"`
	KeystrokeSequence string `xml:"KeystrokeSequence"`
}

type DeletedObjectData struct {
	XMLName      xml.Name   `xml:"DeletedObject"`
	UUID         string     `xml:"UUID"`
	DeletionTime *time.Time `xml:"DeletionTime"`
}
