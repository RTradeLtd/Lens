package planetary_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens/xtractor/planetary"
)

const (
	testHash = "QmSi9TLyzTXmrLMXDvhztDoX3jghoG3vcRrnPkLvGgfpdW"
)

func TestPlanetaryExtractor(t *testing.T) {
	px, err := planetary.NewPlanetaryExtractor()
	if err != nil {
		t.Fatal(err)
	}
	var out interface{}
	if err = px.ExtractObject(testHash, &out); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", out)
	cidObj, err := px.DecodeStringToCID(testHash)
	if err != nil {
		t.Fatal(err)
	}
	expectedCodecUint := uint64(112) // hex 70, aka dag-protobuf codec
	if cidObj.Prefix().Codec != expectedCodecUint {
		t.Fatal("unexpected codec returned")
	}
	fmt.Println("code is ", cidObj.Prefix().Codec)
	marshaled, err := cidObj.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(marshaled))
	contents, err := px.ExtractContents(testHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(contents))
}
