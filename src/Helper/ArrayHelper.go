package Helper

import (
	"math/rand"
	"reflect"
	"time"
)

func ArrayContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GetMinMax(arr interface{}) (float32, float32) {
	v := reflect.ValueOf(arr)
	if v.Len() == 0 {
		return 0.0, 0.0
	}

	min, max := v.Index(0).Interface().(float32), v.Index(0).Interface().(float32)

	for i := 0; i < v.Len(); i++ {
		value := v.Index(i).Interface()
		switch value := value.(type) {
		case *float32:
			if *value < min {
				min = *value
			}
			if *value > max {
				max = *value
			}
		case float32:
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
		}
	}
	return min, max
}

func IntArrayContains(array []int, value int) bool {
	for _, a := range array {
		if a == value {
			return true
		}
	}
	return false
}

func GenerateRandomNumber(min float32, max float32) float32 {
	rand.Seed(time.Now().UnixNano())

	return min + rand.Float32()*(max-min)
}
