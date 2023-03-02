package pngtext_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jkawamoto/go-pngtext"
)

const (
	expectTextData         = "text data"
	expectZTxtData         = "compressed text data                              end"
	expectITxtData         = "🦄"
	expectITxtCompressData = "🦄                                      end"
)

func TestTextualDataList_Find(t *testing.T) {
	testList := pngtext.TextualDataList{
		{
			Keyword: "keyword-1",
			Text:    "text-1",
		},
		{
			Keyword: "keyword-2",
			Text:    "text-2",
		},
		{
			Keyword: "keyword-3",
			Text:    "text-3",
		},
	}

	res := testList.Find("keyword-2")
	if !reflect.DeepEqual(res, testList[1]) {
		t.Errorf("expect %v, got %v", testList[1], res)
	}

	res = testList.Find("keyword-10")
	if res != nil {
		t.Errorf("expect %v, got %v", nil, res)
	}
}

func TestTextualDataList_Len(t *testing.T) {
	testList := pngtext.TextualDataList{
		{
			Keyword: "keyword-1",
			Text:    "text-1",
		},
		{
			Keyword: "keyword-2",
			Text:    "text-2",
		},
		{
			Keyword: "keyword-3",
			Text:    "text-3",
		},
	}
	if res := testList.Len(); res != len(testList) {
		t.Errorf("expect %v, got %v", len(testList), res)
	}
}

func TestTextualDataList_Less(t *testing.T) {
	testList := pngtext.TextualDataList{
		{
			Keyword: "keyword-1",
			Text:    "text-1",
		},
		{
			Keyword: "keyword-2",
			Text:    "text-2",
		},
		{
			Keyword: "keyword-3",
			Text:    "text-3",
		},
	}
	for i := 0; i != testList.Len(); i++ {
		for j := 0; j != testList.Len(); j++ {
			res := testList.Less(i, j)
			expect := strings.Compare(testList[i].Keyword, testList[j].Keyword) < 0
			if res != expect {
				t.Errorf("expect %v, got %v", expect, res)
			}
		}
	}
}

func TestTextualDataList_Swap(t *testing.T) {
	testList := pngtext.TextualDataList{
		{
			Keyword: "keyword-1",
			Text:    "text-1",
		},
		{
			Keyword: "keyword-2",
			Text:    "text-2",
		},
		{
			Keyword: "keyword-3",
			Text:    "text-3",
		},
	}
	expect := pngtext.TextualDataList{testList[2], testList[1], testList[0]}

	testList.Swap(0, 2)
	if !reflect.DeepEqual(testList, expect) {
		t.Errorf("expect %v, got %v", expect, testList)
	}
}

func TestParseTextData(t *testing.T) {
	r, err := os.Open("test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = r.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	res, err := pngtext.ParseTextualData(r)
	if err != nil {
		t.Fatal(err)
	}

	if v := res.Find("Text"); v == nil {
		t.Error("Text tag not found")
	} else if v.Text != expectTextData {
		t.Errorf("expect %v, got %v", expectTextData, v.Text)
	}

	if v := res.Find("ZTxt"); v == nil {
		t.Error("ZTxt tag not found")
	} else if v.Text != expectZTxtData {
		t.Errorf("expect %v, got %v", expectZTxtData, v.Text)
	}

	if v := res.Find("ITxt"); v == nil {
		t.Error("ITxt tag not found")
	} else if v.Text != expectITxtData {
		t.Errorf("expect %v, got %v", expectITxtData, v.Text)
	}

	if v := res.Find("ITxtCompressed"); v == nil {
		t.Error("ITxtCompressed tag not found")
	} else if v.Text != expectITxtCompressData {
		t.Errorf("expect %v, got %v", expectITxtCompressData, v.Text)
	}
}

func TestParseNotPNGData(t *testing.T) {
	r, err := os.Open("LICENSE")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = r.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = pngtext.ParseTextualData(r)
	if err != pngtext.ErrNotPngData {
		t.Errorf("expect %v, got %v", pngtext.ErrNotPngData, err)
	}
}
