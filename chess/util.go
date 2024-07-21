package chess

// Find and remove the first matching element from the slice, returning the updated slice
// The second parameter is true if the element was found
func removeElement[T comparable](slice []T, el T) ([]T, bool) {
	foundIndex := -1
	for index, element := range slice {
		if element == el {
			foundIndex = index
			break
		}
	}

	if foundIndex < 0 {
		return slice, false
	}

	return append(slice[:foundIndex], slice[foundIndex+1:]...), true
}

func signum(x int) int {
	if x < 0 {
		return -1
	} else if x == 0 {
		return 0
	} else {
		return 1
	}
}
