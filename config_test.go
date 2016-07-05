package configuration

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	f := filepath.Join(os.Getenv("GOPATH"), "src", "gogs.xlh", "tools", "configuration", "testdata", "config.conf")
	Parse("file::" + f)
}

//func init() {
//	Parse("etcd::http://localhost:2379")
//}

type InlineCustomValue struct {
	InStringValue string `conf:"struct.string"`
	InBoolsValue  []bool `conf:"struct.bools"`
}

type CustomConfig struct {
	defValue             int
	IntValue             int                  `conf:"comp.struct.int"`
	StringValue          string               `conf:"comp.struct.string"`
	OmitValue            bool                 `conf:"comp.struct.omit,omit"`
	BoolValue            bool                 `conf:"comp.struct.bool"`
	Float32Value         float32              `conf:"comp.struct.float32"`
	Float64Value         float64              `conf:"comp.struct.float64"`
	StringDefValue       string               `conf:"comp.string.def,default(defaultvalue)"`
	InlineValue          *InlineCustomValue   `conf:"comp.struct"`
	InLineStructArray    []InlineCustomValue  `conf:"comp.array"`
	InLineStructPtrArray []*InlineCustomValue `conf:"comp.array"`
	MapValue             map[string]string    `conf:"comp.map"`
	OmitInline           *InlineCustomValue   `conf:"comp.omit,omit"`

	AnotherMap map[string]string `conf:"wmds.wxoauth2.redirects"`

	MapSt    map[string]MapStruct  `conf:"map.struct"`
	MapStPtr map[string]*MapStruct `conf:"map.struct"`
}

type MapStruct struct {
	Field1 string `conf:"field1"`
	Field2 string `conf:"field2"`
}

func TestString(t *testing.T) {
	if v, err := Int("test.int"); err != nil || v != 1020 {
		if err != nil {
			t.Fatalf("err:", err.Error())
		}
		if v != 1020 {
			t.Fatalf("value err")
		}
		t.Fatal("config int error", err, v)
	}
	if v, err := String("test.string"); err != nil || v != "my name string" {
		t.Fatal("config string error")
	}
	if v, err := Bool("test.bool"); err != nil || v != true {
		t.Fatal("config bool error")
	}
	if v, err := Float32("test.float.32"); err != nil || v != 102.00 {
		t.Fatal("config float32 error")
	}
	if v, err := Float64("test.float.64"); err != nil || v != 100.04324323 {
		t.Fatal("config float64 error", err)
	}
	if v, err := Ints("comp.array.ints"); err != nil || len(v) != 5 {
		t.Fatal("config int array error")
	} else {
		for i, vv := range v {
			if i+1 != vv {
				t.Fatal("config int array error")
			}
		}
	}
	estrings := []string{"my", "name", "string"}
	if v, err := Strings("comp.array.strings"); err != nil || len(v) != len(estrings) {
		t.Fatal("config string array error")
	} else {
		for i, vv := range v {
			if estrings[i] != vv {
				t.Fatal("config string array error")
			}
		}
	}
	ebools := []bool{true, false}
	if v, err := Bools("comp.array.bools"); err != nil || len(v) != len(ebools) {
		t.Fatal("config bool array error")
	} else {
		for i, vv := range v {
			if ebools[i] != vv {
				t.Fatal("config bool array error")
			}
		}
	}
	efloat32s := []float32{1, 2, 10.1}
	if v, err := Float32s("comp.array.float32"); err != nil || len(v) != len(efloat32s) {
		t.Fatal("config float32 array error")
	} else {
		for i, vv := range v {
			if efloat32s[i] != vv {
				t.Fatal("config float32 array error")
			}
		}
	}
	efloat64s := []float32{11.11, 22.22}
	if v, err := Float32s("comp.array.float64"); err != nil || len(v) != len(efloat64s) {
		t.Fatal("config float64 array error")
	} else {
		for i, vv := range v {
			if efloat64s[i] != vv {
				t.Fatal("config float64 array error")
			}
		}
	}
	cfg := CustomConfig{}
	if err := Var(&cfg); err != nil {
		t.Fatal("config var error", err)
	} else {
		t.Log(cfg, cfg.InlineValue)
		t.Log("xxx", cfg.InLineStructArray)
		t.Log("yyy", cfg.InLineStructArray)
		t.Log("zzz", cfg.MapValue)
		t.Log("EEE", cfg.AnotherMap)
		t.Log("MMM", cfg.MapSt)
		t.Log("MMM", cfg.MapStPtr)
		t.Log("MMM_VALUE", cfg.MapStPtr["key1"])
		if cfg.IntValue != 55 || cfg.StringValue != "stringv" || cfg.BoolValue != true || cfg.Float32Value != 101.101 ||
			cfg.Float64Value != 1011.1011 {
			t.Fatal("config var error")
		}
		if cfg.InlineValue.InStringValue != "ReString" {
			t.Fatal("config var inline error")
		}
		if len(cfg.InlineValue.InBoolsValue) != 2 || cfg.InlineValue.InBoolsValue[0] != true || cfg.InlineValue.InBoolsValue[1] != false {
			t.Fatal("config var inline error")
		}
	}
	if cfg.StringDefValue != "defaultvalue" {
		t.Fatalf("default string value error", cfg.StringDefValue)
	}
	return
}
