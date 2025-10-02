package genesis

import (
	"fmt"
	"reflect"
)

type Config struct {
	BlockInterval              uint64 `json:"blockInterval"`              // time interval between two consecutive blocks.
	EpochLength                uint32 `json:"epochLength"`                // number of blocks per epoch, also the number of blocks between two checkpoints.
	SeederInterval             uint32 `json:"seederInterval"`             // blocks between two scheduler seeder epochs.
	ValidatorEvictionThreshold uint32 `json:"validatorEvictionThreshold"` // the number of blocks after which offline validator will be evicted from the leader group (7 days)
	EvictionCheckInterval      uint32 `json:"evictionCheckInterval"`      // blocks between two eviction function executions

	// staker parameters
	LowStakingPeriod    uint32  `json:"lowStakingPeriod"`
	MediumStakingPeriod uint32  `json:"mediumStakingPeriod"`
	HighStakingPeriod   uint32  `json:"highStakingPeriod"`
	CooldownPeriod      uint32  `json:"cooldownPeriod"`
	HayabusaTP          *uint32 `json:"hayabusaTP"`
}

// ConfigFromThor populates the Config struct from any source struct with matching field names
func (c *Config) ConfigFromThor(source interface{}) error {
	if source == nil {
		return fmt.Errorf("source cannot be nil")
	}

	sourceVal := reflect.ValueOf(source)
	sourceType := reflect.TypeOf(source)

	// Handle pointer to struct
	if sourceType.Kind() == reflect.Ptr {
		if sourceVal.IsNil() {
			return fmt.Errorf("source pointer cannot be nil")
		}
		sourceVal = sourceVal.Elem()
		sourceType = sourceType.Elem()
	}

	// Ensure source is a struct
	if sourceType.Kind() != reflect.Struct {
		return fmt.Errorf("source must be a struct, got %s", sourceType.Kind())
	}

	destVal := reflect.ValueOf(c).Elem()
	destType := reflect.TypeOf(*c)

	// Iterate through destination fields
	for i := 0; i < destType.NumField(); i++ {
		destField := destType.Field(i)
		destFieldVal := destVal.Field(i)

		// Skip unexported fields
		if !destFieldVal.CanSet() {
			continue
		}

		// Find matching field in source by name
		sourceFieldVal := sourceVal.FieldByName(destField.Name)
		if !sourceFieldVal.IsValid() {
			// Field doesn't exist in source, skip
			continue
		}

		// Convert and set the value
		if err := setFieldValue(destFieldVal, sourceFieldVal, destField.Name); err != nil {
			return fmt.Errorf("failed to set field %s: %v", destField.Name, err)
		}
	}

	return nil
}

// setFieldValue handles type conversion and pointer creation when setting field values
func setFieldValue(dest, src reflect.Value, fieldName string) error {
	destType := dest.Type()
	srcType := src.Type()

	// Handle pointer destination
	if destType.Kind() == reflect.Ptr {
		// If source is also pointer
		if srcType.Kind() == reflect.Ptr {
			if src.IsNil() {
				dest.Set(reflect.Zero(destType))
				return nil
			}
			// Create new pointer and set its value
			newPtr := reflect.New(destType.Elem())
			if err := setFieldValue(newPtr.Elem(), src.Elem(), fieldName); err != nil {
				return err
			}
			dest.Set(newPtr)
			return nil
		}
		// Source is not pointer, create pointer to source value
		newPtr := reflect.New(destType.Elem())
		if err := setFieldValue(newPtr.Elem(), src, fieldName); err != nil {
			return err
		}
		dest.Set(newPtr)
		return nil
	}

	// Handle source pointer but destination is not
	if srcType.Kind() == reflect.Ptr {
		if src.IsNil() {
			return fmt.Errorf("cannot set non-pointer field %s from nil pointer", fieldName)
		}
		return setFieldValue(dest, src.Elem(), fieldName)
	}

	// Both are non-pointers, handle type conversion
	if !srcType.ConvertibleTo(destType) {
		return fmt.Errorf("cannot convert %s to %s for field %s", srcType, destType, fieldName)
	}

	convertedVal := src.Convert(destType)
	dest.Set(convertedVal)
	return nil
}
