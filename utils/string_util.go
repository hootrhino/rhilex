package utils

import (
	"fmt"
	"strings"
)

/*
*
* 检查列表元素是不是不重复了
*
 */
func IsListDuplicated(list []string) bool {
	tmpMap := make(map[string]int)
	for _, value := range list {
		tmpMap[value] = 1
	}
	var keys []interface{}
	for k := range tmpMap {
		keys = append(keys, k)
	}
	// 对比列表的Key元素和列表长度是否等长
	return len(keys) != len(list)
}

/*
*
* 列表包含
*
 */
func SContains(slice []string, e string) bool {
	for _, s := range slice {
		if s == e {
			return true
		}
	}
	return false
}
func IsValidNameLength(username string) (bool, string) {
	const minLen = 4
	const maxLen = 64

	if len(username) < minLen {
		return false, fmt.Sprintf("Name is too short (minimum %d characters)", minLen)
	} else if len(username) > maxLen {
		return false, fmt.Sprintf("Name is too long (maximum %d characters)", maxLen)
	}

	return true, ""
}

/*
*
* Tag name 不可出现非法字符
*
 */
func IsValidColumnName(columnName string) bool {
	// 列名不能以数字开头
	if len(columnName) == 0 || (columnName[0] >= '0' && columnName[0] <= '9') {
		return false
	}
	invalidChars := []string{" ", "-", ";"}
	for _, char := range invalidChars {
		if strings.Contains(columnName, char) {
			return false
		}
	}
	return true
}
