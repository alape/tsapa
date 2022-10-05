package main

func stringInSlice(a string, list []string) int {
	for index, b := range list {
		if b == a {
			return index
		}
	}
	return -1
}

func getNextItem(arr []string, index int) string {
	if (len(arr) > (index + 1)) && (index != (len(arr) - 1)) {
		return arr[index+1]
	} else {
		panic("too few arguments")
	}
}

func getPreviousItem(arr []string, index int) string {
	if (len(arr) > 1) && (index > 0) {
		return arr[index-1]
	} else {
		panic("too few arguments")
	}
}

func keys(dict map[string]interface{}) []string {
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	return keys
}
