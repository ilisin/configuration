package configuration

import (
	"errors"
	"os"
)

var (
	ErrUnkownProvider = errors.New("unkown configuration type")
)

type ConfigType int

const (
	// config with file
	CTFileConf ConfigType = iota
	// config with etcd
	//CTEtcd
	// path
	CTEnv
	// unsupport config type
	CTUnkown = -1
)

// parse config , must call before use config
func Parse(pro ...string) {
	var provider string
	if len(pro) == 0 {
		provider = os.Getenv("GLOBAL_CONF")
		if len(provider) == 0 {
			provider = DefaultProvider
		}
	} else {
		provider = pro[0]
	}
	driver = &Driver{}
	err := driver.ParseProvider(provider)
	if err != nil {
		panic(err)
	}
	_, err = driver.LoadProvider()
	if err != nil {
		panic(err)
	}
}

func loadDriver() {
	if driver == nil {
		Parse()
	}
}

// get config value of type string
func String(key string) (string, error) {
	loadDriver()
	v, e := driver.Buffer().GetString(key, "")
	if e != nil {
		return "", errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config value of type string
func Bool(key string) (bool, error) {
	loadDriver()
	v, e := driver.Buffer().GetBool(key, "")
	if e != nil {
		return false, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config value of type int
func Int(key string) (int, error) {
	loadDriver()
	i, e := driver.Buffer().GetInt(key, "")
	if e != nil {
		return 0, errors.New(e.Error())
	} else {
		return i, nil
	}
}

// get config value of type float32
func Float32(key string) (float32, error) {
	loadDriver()
	v, e := driver.Buffer().GetFloat32(key, "")
	if e != nil {
		return 0, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config value of type float64
func Float64(key string) (float64, error) {
	loadDriver()
	v, e := driver.Buffer().GetFloat64(key, "")
	if e != nil {
		return 0, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config values of type string slice
func Strings(key string) ([]string, error) {
	loadDriver()
	v, e := driver.Buffer().GetStrings(key, "")
	if e != nil {
		return []string{}, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config value of type bool slice
func Bools(key string) ([]bool, error) {
	loadDriver()
	v, e := driver.Buffer().GetBools(key, "")
	if e != nil {
		return []bool{}, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config values of type int slice
func Ints(key string) ([]int, error) {
	loadDriver()
	v, e := driver.Buffer().GetInts(key, "")
	if e != nil {
		return []int{}, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config values of type int slice
func Int64s(key string) ([]int64, error) {
	loadDriver()
	v, e := driver.Buffer().GetInt64s(key, "")
	if e != nil {
		return []int64{}, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config values of type float32 slice
func Float32s(key string) ([]float32, error) {
	loadDriver()
	v, e := driver.Buffer().GetFloat32s(key, "")
	if e != nil {
		return []float32{}, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config values of type float64 slice
func Float64s(key string) ([]float64, error) {
	loadDriver()
	v, e := driver.Buffer().GetFloat64s(key, "")
	if e != nil {
		return []float64{}, errors.New(e.Error())
	} else {
		return v, nil
	}
}

// get config value of custom struct type
// panic if in param isn't a pointer
func Var(o interface{}) error {
	loadDriver()
	return driver.Buffer().Var(o)
}
