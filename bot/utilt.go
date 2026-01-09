package bot

// 检测list中是否含有item
func isItemInList[T comparable](list []T, item ...T) bool {
	for i := 0; i < len(list); i++ {
		for j := 0; j < len(item); j++ {
			if list[i] == item[j] {
				return true
			}
		}
	}
	return false
}

func iterators[T any](list []T, fn func(T)) {
	for i := 0; i < len(list); i++ {
		fn(list[i])
	}
}

//比较两个数组是否相等
func compareArray[T comparable](l1 []T, l2 []T) bool {
	if len(l1) != len(l2) {
		return false
	}
	for i := 0; i < len(l1); i++ {
		if l1[i] != l2[i] {
			return false
		}
	}
	return true
}
