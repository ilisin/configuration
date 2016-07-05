package configuration

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	errKeyNotFound = errors.New("cann't found the value with the key")
	errType        = errors.New("cann't fetch the type")
)

type BufferError struct {
	err error
	msg string
}

func IsKeyNotFound(b *BufferError) bool {
	if b == nil {
		return false
	}
	return b.err == errKeyNotFound
}

func (b BufferError) Error() string {
	return fmt.Sprintf("[buffer error] %v [%v]", b.err, b.msg)
}

func NewBufferError(err error, msg string) *BufferError {
	return &BufferError{err, msg}
}

type TreeBuffer struct {
	Data     map[string]string
	Children map[string]*TreeBuffer
	//lock for data map
	DataLock     sync.RWMutex
	ChildrenLock sync.RWMutex
}

func NewTreeBuffer() *TreeBuffer {
	return &TreeBuffer{
		Data:     make(map[string]string),
		Children: make(map[string]*TreeBuffer),
	}
}

func (t *TreeBuffer) Set(key, value string) {
	value = strings.Trim(value, `"`)
	var ks []string
	if strings.Count(key, `"`) == 2 && strings.HasSuffix(key, `"`) {
		i := strings.LastIndex(string(key[:len(key)-1]), `"`)
		kt := string(key[:i])
		kt = strings.TrimRight(kt, ".")
		ks = strings.Split(kt, `.`)
		ks = append(ks, string(key[i+1:len(key)-1]))
	} else {
		ks = strings.Split(key, ".")
	}
	t.SetIn(ks, value)
}

func (t *TreeBuffer) SetIn(ks []string, value string) {
	if len(ks) == 1 {
		t.DataLock.Lock()
		defer t.DataLock.Unlock()
		t.Data[ks[0]] = value
	} else if len(ks) > 1 {
		var (
			tb *TreeBuffer
			ok bool
		)
		t.ChildrenLock.Lock()
		defer t.ChildrenLock.Unlock()
		if tb, ok = t.Children[ks[0]]; !ok {
			tb = NewTreeBuffer()
			t.Children[ks[0]] = tb
		}
		tb.SetIn(ks[1:], value)
	}
}

func (t *TreeBuffer) Delete(key string) *BufferError {
	ks := strings.Split(key, ".")
	b, err := t.GetBuffer(ks)
	if err != nil {
		return NewBufferError(err, key)
	}
	delete(b.Data, ks[len(ks)-1])
	return nil
}

func (t *TreeBuffer) GetIn(ks []string) (string, error) {
	if len(ks) == 1 {
		t.DataLock.RLock()
		defer t.DataLock.RUnlock()
		if s, ok := t.Data[ks[0]]; ok {
			return s, nil
		} else {
			return "", errKeyNotFound
		}
	} else if len(ks) > 1 {
		t.ChildrenLock.RLock()
		defer t.ChildrenLock.RUnlock()
		if tb, ok := t.Children[ks[0]]; ok {
			return tb.GetIn(ks[1:])
		} else {
			return "", errKeyNotFound
		}
	} else {
		return "", errKeyNotFound
	}
}

func (t *TreeBuffer) GetBuffer(ks []string) (*TreeBuffer, error) {
	if len(ks) == 1 {
		return t, nil
	} else if len(ks) > 1 {
		t.ChildrenLock.RLock()
		defer t.ChildrenLock.RUnlock()
		if tb, ok := t.Children[ks[0]]; ok {
			return tb.GetBuffer(ks[1:])
		} else {
			return nil, errKeyNotFound
		}
	} else {
		return nil, errKeyNotFound
	}
}

func (t *TreeBuffer) GetString(key, def string) (string, *BufferError) {
	ks := strings.Split(key, ".")
	s, err := t.GetIn(ks)
	if err != nil {
		if err == errKeyNotFound && len(def) > 0 {
			return def, nil
		}
		return "", NewBufferError(err, key)
	} else {
		return s, nil
	}
}

func (t *TreeBuffer) GetStrings(key, def string) ([]string, *BufferError) {
	ks := strings.Split(key, ".")
	tb, err := t.GetBuffer(ks)
	if err != nil {
		if err == errKeyNotFound && len(def) > 0 {
			return strings.Split(def, ";"), nil
		}
		return []string{}, NewBufferError(err, key)
	}
	tb.DataLock.RLock()
	defer tb.DataLock.RUnlock()
	if v, ok := tb.Data[ks[len(ks)-1]]; ok {
		return strings.Split(v, ";"), nil
	}
	tb.ChildrenLock.RLock()
	defer tb.ChildrenLock.RUnlock()
	if tbc, ok := tb.Children[ks[len(ks)-1]]; ok {
		rets := make([]string, len(tbc.Data))
		tbc.DataLock.RLock()
		defer tbc.DataLock.RUnlock()
		for i := 0; i < len(tbc.Data); i++ {
			if s, ok := tbc.Data[fmt.Sprintf("%v", i)]; ok {
				rets[i] = s
			} else {
				return []string{}, NewBufferError(errKeyNotFound, key)
			}
		}
		return rets, nil
	} else {
		return []string{}, NewBufferError(errKeyNotFound, key)
	}
}

func (t *TreeBuffer) GetMap(key, def string) (map[string]string, *BufferError) {
	ks := strings.Split(key, ".")
	tb, err := t.GetBuffer(ks)
	if err != nil {
		if err == errKeyNotFound && len(def) > 0 {
			ss := strings.Split(def, ";") //分号
			m := make(map[string]string)
			for _, s := range ss {
				kvs := strings.SplitN(s, ":", 2) //冒号
				if len(kvs) < 2 {
					m[kvs[0]] = ""
				} else {
					m[kvs[0]] = kvs[1]
				}
			}
			return m, nil
		}
		return map[string]string{}, NewBufferError(err, key)
	}
	tb.ChildrenLock.RLock()
	defer tb.ChildrenLock.RUnlock()
	if ttb, ok := tb.Children[ks[len(ks)-1]]; ok {
		ttb.DataLock.RLock()
		defer ttb.DataLock.RUnlock()
		return ttb.Data, nil
	} else {
		return nil, NewBufferError(errKeyNotFound, key)
	}
}

// get map child
func (t *TreeBuffer) GetMapChild(key string) (map[string]*TreeBuffer, *BufferError) {
	ks := strings.Split(key, ".")
	tb, err := t.GetBuffer(ks)
	if err != nil {
		return nil, NewBufferError(err, key)
	}
	tb.ChildrenLock.RLock()
	defer tb.ChildrenLock.RUnlock()
	if ttb, ok := tb.Children[ks[len(ks)-1]]; ok {
		ttb.ChildrenLock.RLock()
		defer ttb.ChildrenLock.RUnlock()
		return ttb.Children, nil
	}
	return nil, NewBufferError(errKeyNotFound, key)
}

func (t *TreeBuffer) GetInt(key, def string) (int, *BufferError) {
	s, err := t.GetString(key, def)
	if err != nil {
		return 0, err
	}
	if i64, err := strconv.ParseInt(s, 10, 32); err != nil {
		return 0, NewBufferError(errType, key)
	} else {
		return int(i64), nil
	}
}

func (t *TreeBuffer) GetInts(key, def string) ([]int, *BufferError) {
	ss, err := t.GetStrings(key, def)
	if err != nil {
		return []int{}, err
	}
	rets := make([]int, len(ss))
	for i, s := range ss {
		if i64, err := strconv.ParseInt(s, 10, 32); err != nil {
			return []int{}, NewBufferError(errType, key)
		} else {
			rets[i] = int(i64)
		}
	}
	return rets, nil
}

func (t *TreeBuffer) GetInt64(key, def string) (int64, *BufferError) {
	s, err := t.GetString(key, def)
	if err != nil {
		return 0, err
	}
	if i64, err := strconv.ParseInt(s, 10, 64); err != nil {
		return 0, NewBufferError(errType, key)
	} else {
		return i64, nil
	}
}

func (t *TreeBuffer) GetInt64s(key, def string) ([]int64, *BufferError) {
	ss, err := t.GetStrings(key, def)
	if err != nil {
		return []int64{}, err
	}
	rets := make([]int64, len(ss))
	for i, s := range ss {
		if i64, err := strconv.ParseInt(s, 10, 64); err != nil {
			return []int64{}, NewBufferError(errType, key)
		} else {
			rets[i] = i64
		}
	}
	return rets, nil
}

func (t *TreeBuffer) convertBool(s string) bool {
	if s == "1" || s == "T" || s == "t" || strings.ToLower(s) == "true" {
		return true
	} else {
		return false
	}
}

func (t *TreeBuffer) GetBool(key, def string) (bool, *BufferError) {
	s, err := t.GetString(key, def)
	if err != nil {
		return false, err
	}
	return t.convertBool(s), nil
}

func (t *TreeBuffer) GetBools(key, def string) ([]bool, *BufferError) {
	ss, err := t.GetStrings(key, def)
	if err != nil {
		return []bool{}, err
	}
	rets := make([]bool, len(ss))
	for i, s := range ss {
		rets[i] = t.convertBool(s)
	}
	return rets, nil
}

func (t *TreeBuffer) GetFloat32(key, def string) (float32, *BufferError) {
	s, err := t.GetString(key, def)
	if err != nil {
		return 0, err
	}
	if f64, err := strconv.ParseFloat(s, 32); err != nil {
		return 0, NewBufferError(errType, key)
	} else {
		return float32(f64), nil
	}
}

func (t *TreeBuffer) GetFloat32s(key, def string) ([]float32, *BufferError) {
	ss, err := t.GetStrings(key, def)
	if err != nil {
		return []float32{}, err
	}
	rets := make([]float32, len(ss))
	for i, s := range ss {
		if f64, err := strconv.ParseFloat(s, 32); err != nil {
			return []float32{}, NewBufferError(errType, key)
		} else {
			rets[i] = float32(f64)
		}
	}
	return rets, nil
}

func (t *TreeBuffer) GetFloat64(key, def string) (float64, *BufferError) {
	s, err := t.GetString(key, def)
	if err != nil {
		return 0, err
	}
	if f64, err := strconv.ParseFloat(s, 64); err != nil {
		return 0, NewBufferError(errType, key)
	} else {
		return f64, nil
	}
}

func (t *TreeBuffer) GetFloat64s(key, def string) ([]float64, *BufferError) {
	ss, err := t.GetStrings(key, def)
	if err != nil {
		return []float64{}, err
	}
	rets := make([]float64, len(ss))
	for i, s := range ss {
		if f64, err := strconv.ParseFloat(s, 64); err != nil {
			return []float64{}, NewBufferError(errType, key)
		} else {
			rets[i] = f64
		}
	}
	return rets, nil
}

// support type
// struct pointer
// struct slice
func (t *TreeBuffer) Var(o interface{}) error {
	ot := reflect.TypeOf(o)
	ov := reflect.ValueOf(o)
	if (ot.Kind() == reflect.Ptr && ot.Elem().Kind() == reflect.Struct) ||
		(ot.Kind() == reflect.Array && ot.Elem().Kind() == reflect.Struct) ||
		(ot.Kind() == reflect.Array && ot.Elem().Kind() == reflect.Ptr && ot.Elem().Elem().Kind() == reflect.Struct) {
		e := t.varSet(ot.Elem(), ov.Elem(), "")
		if e != nil {
			return errors.New(e.Error())
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("Configuration struct type in error!")
	}
}

func (this *TreeBuffer) hasChildBuffer(key string) bool {
	ks := strings.Split(key, ".")
	b, err := this.GetBuffer(ks)
	if err != nil {
		return false
	}
	b.ChildrenLock.RLock()
	defer b.ChildrenLock.RUnlock()
	_, ok := b.Children[ks[len(ks)-1]]
	return ok
}

func (this *TreeBuffer) varSet(ot reflect.Type, ov reflect.Value, ptag string) *BufferError {
	for imax := 0; imax < ot.NumField(); imax++ {
		oti := ot.Field(imax)
		ovi := ov.Field(imax)
		if !ovi.CanSet() {
			continue
		}
		confTag := oti.Tag.Get("conf")
		if len(confTag) == 0 && !(oti.Type.Kind() == reflect.Struct ||
			(oti.Type.Kind() == reflect.Ptr && oti.Type.Elem().Kind() == reflect.Struct)) {
			continue
		}
		confTag = strings.ToLower(confTag)
		confTags := strings.Split(confTag, ",")
		omit := false
		def := ""
		for _, t := range confTags {
			if t == "omit" {
				omit = true
			} else if strings.HasPrefix(t, "default(") {
				t = string(t[len("default("):])
				t = string(t[:len(t)-1])
				def = t
			} else {
				confTag = t
			}
		}
		ptag = strings.TrimRight(ptag, ".")
		if len(ptag) > 0 {
			confTag = ptag + "." + confTag
		}
		switch oti.Type.Kind() {
		case reflect.String:
			v, err := this.GetString(confTag, def)
			if IsKeyNotFound(err) && omit {
				break
			}
			if err != nil {
				return err
			}
			ovi.SetString(v)
		case reflect.Bool:
			v, err := this.GetBool(confTag, def)
			if IsKeyNotFound(err) && omit {
				break
			}
			if err != nil {
				return err
			}
			ovi.SetBool(v)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, err := this.GetInt64(confTag, def)
			if IsKeyNotFound(err) && omit {
				break
			}
			if err != nil {
				return err
			}
			ovi.SetInt(int64(v))
		case reflect.Float32, reflect.Float64:
			v, err := this.GetFloat64(confTag, def)
			if IsKeyNotFound(err) && omit {
				break
			}
			if err != nil {
				return err
			}
			ovi.SetFloat(v)
		case reflect.Slice:
			switch oti.Type.Elem().Kind() {
			case reflect.String:
				v, err := this.GetStrings(confTag, def)
				if IsKeyNotFound(err) && omit {
					break
				}
				if err != nil {
					return err
				}
				ovi.Set(reflect.ValueOf(v))
			case reflect.Bool:
				v, err := this.GetBools(confTag, def)
				if IsKeyNotFound(err) && omit {
					break
				}
				if err != nil {
					return err
				}
				ovi.Set(reflect.ValueOf(v))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v, err := this.GetInt64s(confTag, def)
				if IsKeyNotFound(err) && omit {
					break
				}
				if err != nil {
					return err
				}
				ovi.Set(reflect.ValueOf(v))
			case reflect.Float32:
				v, err := this.GetFloat32s(confTag, def)
				if IsKeyNotFound(err) && omit {
					break
				}
				if err != nil {
					return err
				}
				ovi.Set(reflect.ValueOf(v))
			case reflect.Float64:
				v, err := this.GetFloat64s(confTag, def)
				if IsKeyNotFound(err) && omit {
					break
				}
				if err != nil {
					return err
				}
				ovi.Set(reflect.ValueOf(v))
			case reflect.Struct:
				if !this.hasChildBuffer(confTag) && omit {
					break
				}
				ks := strings.Split(confTag, ".")
				ttb, err := this.GetBuffer(ks)
				if err != nil {
					return NewBufferError(err, confTag)
				}
				ttb.ChildrenLock.RLock()
				defer ttb.ChildrenLock.RUnlock()
				ovitemp := reflect.MakeSlice(oti.Type, 0, 0)
				if tb, ok := ttb.Children[ks[len(ks)-1]]; ok {

					tb.ChildrenLock.RLock()
					defer tb.ChildrenLock.RUnlock()
					var berr *BufferError
					for i := 0; i < len(tb.Children); i++ {
						if _, ok := tb.Children[fmt.Sprintf("%v", i)]; ok {
							tempv := reflect.New(oti.Type.Elem())
							berr = this.varSet(tempv.Type().Elem(), tempv.Elem(), fmt.Sprintf("%v.%v", confTag, i))
							if err != nil {
								break
							}
							ovitemp = reflect.Append(ovitemp, tempv.Elem())
						} else {
							berr = NewBufferError(errType, fmt.Sprintf("%v.%v", confTag, i))
							break
						}
					}
					if berr != nil {
						return berr
					}
				} else {
					return NewBufferError(err, confTag)
				}
				ovi.Set(ovitemp)
			case reflect.Ptr:
				if oti.Type.Elem().Elem().Kind() != reflect.Struct {
					break
				}
				if !this.hasChildBuffer(confTag) && omit {
					break
				}
				ks := strings.Split(confTag, ".")
				ttb, err := this.GetBuffer(ks)
				if err != nil {
					return NewBufferError(err, confTag)
				}
				ttb.ChildrenLock.RLock()
				defer ttb.ChildrenLock.RUnlock()
				ovitemp := reflect.MakeSlice(oti.Type, 0, 0)
				if tb, ok := ttb.Children[ks[len(ks)-1]]; ok {

					tb.ChildrenLock.RLock()
					defer tb.ChildrenLock.RUnlock()
					var berr *BufferError
					for i := 0; i < len(tb.Children); i++ {
						if _, ok := tb.Children[fmt.Sprintf("%v", i)]; ok {
							tempv := reflect.New(oti.Type.Elem().Elem())
							berr = this.varSet(tempv.Type().Elem(), tempv.Elem(), fmt.Sprintf("%v.%v", confTag, i))
							if err != nil {
								break
							}
							ovitemp = reflect.Append(ovitemp, tempv)
						} else {
							berr = NewBufferError(errType, fmt.Sprintf("%v.%v", confTag, i))
							break
						}
					}
					if berr != nil {
						return berr
					}
				} else {
					return NewBufferError(err, confTag)
				}
				ovi.Set(ovitemp)
			}
		case reflect.Struct:
			if !this.hasChildBuffer(confTag) && omit {
				break
			}
			err := this.varSet(oti.Type, ovi, confTag)
			if IsKeyNotFound(err) && omit {
				break
			}
			if err != nil {
				return err
			}
		case reflect.Ptr:
			if oti.Type.Elem().Kind() == reflect.Struct {
				if !this.hasChildBuffer(confTag) && omit {
					break
				}
				if ovi.IsNil() {
					ovi.Set(reflect.New(oti.Type.Elem()))
				}
				err := this.varSet(oti.Type.Elem(), ovi.Elem(), confTag)
				if IsKeyNotFound(err) && omit {
					break
				}
				if err != nil {
					return err
				}
			}
		case reflect.Map:
			if oti.Type.Key().Kind() != reflect.String {
				return NewBufferError(errType, confTag)
			}
			if !this.hasChildBuffer(confTag) && omit {
				break
			}
			// map value's type
			switch oti.Type.Elem().Kind() {
			case reflect.String:
				ovitemp := reflect.MakeMap(oti.Type)
				v, err := this.GetMap(confTag, def)
				if err != nil {
					return err
				}
				for kt, vt := range v {
					ovitemp.SetMapIndex(reflect.ValueOf(kt), reflect.ValueOf(vt))
				}
				ovi.Set(ovitemp)
			case reflect.Struct:
				ovitemp := reflect.MakeMap(oti.Type)
				mv, err := this.GetMapChild(confTag)
				if err != nil {
					return err
				}
				for kt, vt := range mv {
					vtemp := reflect.New(oti.Type.Elem())
					er := vt.Var(vtemp.Interface())
					if err != nil {
						return NewBufferError(er, confTag+"."+kt)
					}
					ovitemp.SetMapIndex(reflect.ValueOf(kt), vtemp.Elem())
				}
				ovi.Set(ovitemp)
			case reflect.Ptr:
				if oti.Type.Elem().Elem().Kind() != reflect.Struct {
					break
				}
				ovitemp := reflect.MakeMap(oti.Type)
				mv, err := this.GetMapChild(confTag)
				if err != nil {
					return err
				}
				for kt, vt := range mv {
					vtemp := reflect.New(oti.Type.Elem().Elem())
					er := vt.Var(vtemp.Interface())
					if err != nil {
						return NewBufferError(er, confTag+"."+kt)
					}
					ovitemp.SetMapIndex(reflect.ValueOf(kt), vtemp)
				}
				ovi.Set(ovitemp)
			} // end switch oti.Type.Elem().Kind()
		}
	}
	return nil
}

func (b *TreeBuffer) MergeFrom(b2 *TreeBuffer, cover bool) {
	b.DataLock.Lock()
	b2.DataLock.RLock()
	for k, v := range b2.Data {
		if _, ok := b.Data[k]; !ok || cover {
			b.Data[k] = v
		}
	}
	b2.DataLock.RUnlock()
	b.DataLock.Unlock()

	b.ChildrenLock.Lock()
	b2.ChildrenLock.RLock()
	for k, b2i := range b2.Children {
		if bi, ok := b.Children[k]; !ok {
			b.Children[k] = b2i
		} else {
			bi.MergeFrom(b2i, cover)
		}
	}
	b2.ChildrenLock.RUnlock()
	b.ChildrenLock.Unlock()
}

func (b *TreeBuffer) StringRecursive(pre string) string {
	str := ""
	b.DataLock.RLock()
	defer b.DataLock.RUnlock()
	b.ChildrenLock.RLock()
	defer b.ChildrenLock.RUnlock()
	for k, v := range b.Data {
		if len(pre) > 0 {
			str += fmt.Sprintf("%-10s = %v\n", pre+"."+k, v)
		} else {
			str += fmt.Sprintf("%-10s = %v\n", k, v)
		}
	}
	for k, v := range b.Children {
		if len(pre) > 0 {
			str += fmt.Sprintf("%v", v.StringRecursive(pre+"."+k))
		} else {
			str += fmt.Sprintf("%v", v.StringRecursive(k))
		}
	}
	return str
}

func (b *TreeBuffer) String() string {
	return b.StringRecursive("")
}
