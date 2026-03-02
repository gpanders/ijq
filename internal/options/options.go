package options

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
)

//go:generate go run gen.go

type Option interface {
	flag.Value
	Flag() string
}

func (o *Options) Toggle(option Option) {
	o.Each(func(opt Option) {
		if reflect.TypeOf(option) == reflect.TypeOf(opt) {
			value := reflect.ValueOf(opt).Elem()
			value.SetBool(!value.Bool())
		}
	})
}

func (o Options) ToSlice() []string {
	var flags []string

	o.Each(func(option Option) {
		flagName := option.Flag()
		if len(flagName) != 1 {
			return
		}

		value := reflect.Indirect(reflect.ValueOf(option))
		if !value.IsValid() {
			return
		}

		switch value.Kind() {
		case reflect.Bool:
			if !value.Bool() {
				return
			}

			flags = append(flags, "-"+flagName)
		case reflect.Array, reflect.Slice:
			for i := 0; i < value.Len(); i++ {
				flags = append(flags, "-"+flagName, fmt.Sprint(value.Index(i).Interface()))
			}
		}
	})

	return flags
}

func (o *Options) Each(f func(option Option)) {
	if o == nil {
		return
	}

	value := reflect.ValueOf(o).Elem()
	for _, field := range value.Fields() {
		option, ok := field.Addr().Interface().(Option)
		if !ok {
			continue
		}

		f(option)
	}
}

func setBoolOption[T ~bool](option *T, value string) error {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}

	*option = T(b)
	return nil
}

func unmarshalTextValue[T interface{ Set(string) error }](option T, text []byte) error {
	return option.Set(string(text))
}
