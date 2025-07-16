package util

type Set map[string]struct{}

func (s Set) Add(items ...string) {
	for _, x := range items {
		s[x] = struct{}{}
	}
}

func (s Set) Remove(x string) {
	delete(s, x)
}

func (s Set) Contains(x string) bool {
	_, found := s[x]
	return found
}

func (s Set) ToSlice() []string {
	var list []string
	for k := range s {
		list = append(list, k)
	}
	return list
}
