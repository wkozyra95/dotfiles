package fn

func Filter[T any](list []T, cond func(T) bool) []T {
	results := []T{}
	for _, element := range list {
		if cond(element) {
			results = append(results, element)
		}
	}
	return results
}

func Map[T any, P any](list []T, mapFn func(T) P) []P {
	results := make([]P, len(list))
	for index, element := range list {
		results[index] = mapFn(element)
	}
	return results
}

func Identity[T any](t T) T {
	return t
}
