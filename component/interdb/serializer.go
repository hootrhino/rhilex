package interdb

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm/schema"
	"reflect"
)

type CustomTypeSerializer struct {
}

func (cs CustomTypeSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	bytes, err := json.Marshal(fieldValue)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

func (cs CustomTypeSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}
	bytes, ok := dbValue.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", dbValue)
	}
	val := dst.FieldByName(field.Name)
	if !val.CanSet() {
		return fmt.Errorf("cannot set value for field %s", field.Name)
	}
	return json.Unmarshal(bytes, val.Addr().Interface())
}

func (cs CustomTypeSerializer) GormDataType() string {
	return "text"
}
