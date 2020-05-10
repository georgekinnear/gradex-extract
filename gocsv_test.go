package gradexextract

import (
	"bufio"
	"os"
	"testing"

	"github.com/gocarina/gocsv"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

type Foo struct {
	Foo string `csv:"foo"`
	Bar string `csv:"bar"`
}

type Dab struct {
	Far Foo
	Dab string `csv:"dab"`
	Dob string `csv:"dob"`
}

func TestMarhsallNestedStruct(t *testing.T) {

	a := Foo{Foo: "--", Bar: "xx"}
	b := []Dab{Dab{Far: a, Dab: "ff", Dob: "gg"}}

	file, err := os.OpenFile("./test/test.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		t.Error(err)
	}

	defer file.Close()

	err = gocsv.MarshalFile(&b, file)

	if err != nil {
		t.Error(err)
	}

	if _, err := file.Seek(0, 0); err != nil { // Go to the start of the file
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	scanner.Scan()
	assertEqual(t, scanner.Text(), "foo,bar,Far,dab,dob")
	scanner.Scan()
	assertEqual(t, scanner.Text(), "--,xx,,ff,gg")

}
