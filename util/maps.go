package util

type StringMap map[string]string

func (m StringMap) Keys() []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
