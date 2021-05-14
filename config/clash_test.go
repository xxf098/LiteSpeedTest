package config

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParseClash(t *testing.T) {
	dat, err := ioutil.ReadFile("./clash.yaml")
	if err != nil {
		t.Error(err)
	}
	c, err := ParseClash(dat)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(c.Proxies)
}
