package server

import (
	"cmp"
)

func iterators[T any](list []T, fn func(T)) {
	for i := 0; i < len(list); i++ {
		fn(list[i])
	}
}

// 对列表按照从小到大进行排序
func sortList[T cmp.Ordered](list []T) {
	for end := len(list); end > 0; end-- {
		for i := 0; i < end-1; i++ {
			if list[i] < list[i+1] {
				list[i], list[i+1] = list[i+1], list[i]
			}
		}
	}
}

// 检查list中是否至少含有item中的一个
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

// 检查list是否含有所有的item
func isListContainAllTheItem[T comparable](list []T, items ...T) bool {
	for _, item := range items {
		if isItemInList(list, item) {
			continue
		}
		return false
	}
	return true
}

// 检查列表是否含有某种元素,传入ok函数进行检测
func isListContain[T any](list []T, ok func(T) bool) bool {
	for i := 0; i < len(list); i++ {
		if ok(list[i]) {
			return true
		}
	}
	return false
}

// 从列表中删除第一个指定元素
func delFromList[T any](list []T, ok func(T) bool) []T {
	for i := 0; i < len(list); i++ {
		if ok(list[i]) {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
