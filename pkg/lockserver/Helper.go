package lockserver

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func reverse(slice []string) {
	for i, j := 0, len(slice) - 1; i < j; i, j = i + 1, j - 1 {
		temp := slice[j]
		slice[j] = slice[i]
		slice[i] = temp
	}
}

func remove(slice []string, item string) []string {
	newSlice := make([]string, 0)
	for _, element := range slice {
		if element == item {
			continue
		}
		newSlice = append(newSlice, element)
	}
	return newSlice
}