package config

import (
	"encoding/json"
	"errors"
	"github.com/451008604/nets/logs"
	"os"
	"strconv"
	"strings"
)

// 获取配置数据到字节
func getConfigDataToBytes(configPath string, configName string, configStruct any) any {
	bytes, err := os.ReadFile(configPath + configName)
	logs.PrintLogPanic(err)
	logs.PrintLogErr(json.Unmarshal(bytes, configStruct))
	return configStruct
}

/*
将字符串字段转为数组。** Example：sep1="-" **

"0-1-2-3-4-5" -> [0,1,2,3,4,5]
*/
func StringFieldToOneDimensionalArray(StringField, sep1 string) []int {
	tempField := make([]int, 0)
	strArr := strings.Split(StringField, sep1)
	for i := 0; i < len(strArr); i++ {
		num, err := strconv.Atoi(strArr[i])
		tempField = append(tempField, num)
		if err != nil && err.Error() != "" {
			logs.PrintLogErr(err, "StringFieldToOneDimensionalArray")
			return []int{}
		}
	}
	return tempField
}

/*
将字符串字段转为数组。** Example：sep1="," sep2=":" **

"0:1,2:3,4:5" -> [[0,1],[2,3],[4,5]]
*/
func StringFieldToTwoDimensionalArray(StringField, sep1, sep2 string) [][]int {
	tempField := make([][]int, 0)
	strArr := strings.Split(StringField, sep1)
	for i := 0; i < len(strArr); i++ {
		temp, num, err := make([]string, 0), 0, errors.New("")
		temp = strings.Split(strArr[i], sep2)
		arr := make([]int, 0)
		for j := 0; j < len(temp); j++ {
			num, err = strconv.Atoi(temp[j])
			if err != nil && err.Error() != "" {
				logs.PrintLogErr(err, "StringFieldToTwoDimensionalArray")
				return [][]int{}
			}
			arr = append(arr, num)
		}
		tempField = append(tempField, arr)
	}
	return tempField
}
