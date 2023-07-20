package model

import (
	"context"
	"gorm.io/gorm/schema"
	"reflect"
)

func init() {
	schema.RegisterSerializer("nilGob", &NilGobSerializer{schema.GobSerializer{}})
}

type NilGobSerializer struct {
	schema.GobSerializer
}

func (s NilGobSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		if value.IsNil() {
			return nil, nil
		}
	case reflect.Invalid:
		return nil, nil
	}
	return s.GobSerializer.Value(ctx, field, dst, fieldValue)
}
