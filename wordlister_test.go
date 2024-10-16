package wordlister

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestFromHTML(t *testing.T) {
	file, err := ioutil.ReadFile("./testdata/links1.html")
	if err != nil {
		t.Fatal(err)
	}
	wlister := NewWordlist()
	wlister.FromHTML(file)
	list := wlister.List()
	fmt.Printf("list (%d): %+v\n\n", len(list), list)

	file, err = ioutil.ReadFile("./testdata/json1.json")
	if err != nil {
		t.Fatal(err)
	}
	wlister.FromHTML(file)
	list = wlister.List()
	fmt.Printf("list 2 (%d): %+v\n", len(list), list)
}
