package gifting

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type History struct {
	Recipient string `yaml:"recipient"  xml:"recipient,attr"`
	Year      int    `yaml:"year" xml:"year,attr"`
}

type Person struct {
	Name      string     `yaml:"name" xml:"name,attr"`
	Histories []*History `yaml:"history" xml:"history"`

	// additional, non persistent, fields
	// ids are flags (we are limited to 64 people)
	id     uint64
	family int
}

type Family struct {
	Members []*Person `yaml:"person" xml:"person"`
}

type DataStore struct {
	Families []*Family `yaml:"family" xml:"family"`
}

func ReadAsML(fileName string) (*DataStore, error) {
	data := DataStore{}
	bytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		return nil, err
	}
	//fmt.Printf("read %d bytes\n", len(bytes))

	err = xml.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	// assign family and person ids
	// person ids are flags (we are limited to 64 people)
	cnt := uint64(0)
	for idxF, fam := range data.Families {
		for _, p := range fam.Members {
			p.id = uint64(1) << cnt
			p.family = idxF
			cnt++
		}
	}
	return &data, nil
}

func (ds *DataStore) SaveAsXML(fileName string) error {
	file, _ := os.Create(fileName)

	xmlWriter := io.Writer(file)

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("  ", "    ")
	if err := enc.Encode(ds); err != nil {
		return err
	}
	return nil
}

func (p *Person) Recipient(year int) (string, error) {
	for _, h := range p.Histories {
		if h.Year == year {
			return h.Recipient, nil
		}
	}
	return "", errors.New("No recipient for requested year " + string(year))
}

func (ds *DataStore) GiftReport(year int) {

	fmt.Println(year)
	fmt.Printf("%-15s  %s\n", "giver", "recipient")

	for _, fam := range ds.Families {
		for _, p := range fam.Members {
			r, err := p.Recipient(year)

			if err == nil {
				fmt.Printf("%-15s  %s\n", p.Name, r)
			}
		}
	}
}
