package str

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// RemoveElement 从切片中移除指定元素，支持任何类型的切片
func RemoveElement[T comparable](slice []T, element T) []T {
	// 使用反射来避免不必要的元素复制
	for i, item := range slice {
		if item == element {
			// 删除指定元素并返回新的切片
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// LeftUpper 首字母转大写
func LeftUpper(s string) string {
	if len(s) > 0 {
		return strings.ToUpper(string(s[0])) + s[1:]
	}
	return s
}

// LeftLower 首字母转小写
func LeftLower(s string) string {
	if len(s) > 0 {
		return strings.ToLower(string(s[0])) + s[1:]
	}
	return s
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GenerateTableName
func GenerateTableName(baseTable string, dataTimeStamp int) string {
	nowMonth := time.Unix(int64(dataTimeStamp/1000), 0).In(time.Local).Format("200601")
	return baseTable + "_" + nowMonth
}

// IsZeroValue 检查一个值是否为零值
func IsZeroValue(value interface{}) bool {
	if value == nil {
		return true
	}
	// 通过反射检测零值
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	default:
		return false
	}
}

func StringToInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err // 失败时返回 0
	}
	return n, nil
}

// Strval 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func Strval(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func StructToMapDemo(obj interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < obj1.NumField(); i++ {
		data[obj1.Field(i).Name] = obj2.Field(i).Interface()
	}
	return data
}

func MapMerge(x map[string]interface{}, y map[string]interface{}) map[string]interface{} {

	n := make(map[string]interface{})
	for i, v := range x {
		for j, w := range y {
			if i == j {
				n[i] = w

			} else {
				if _, ok := n[i]; !ok {
					n[i] = v
				}
				if _, ok := n[j]; !ok {
					n[j] = w
				}
			}
		}
	}

	return n

}

func EqualAttrName(sourceKey string, currentKey string) bool {
	sk := strings.Replace(sourceKey, "_", "", -1)
	ck := strings.Replace(currentKey, "_", "", -1)

	return strings.ToLower(sk) == strings.ToLower(ck)
}

// SetStructFormSame
func SetStructFormSame(in interface{}, out interface{}) {
	it := reflect.TypeOf(in).Elem()
	ot := reflect.TypeOf(out).Elem()
	iv := reflect.ValueOf(in).Elem()
	ov := reflect.ValueOf(out).Elem()

	for i := 0; i < it.NumField(); i++ {
		inFieldName := it.Field(i).Name
		inFieldType := it.Field(i).Type

		for o := 0; o < ot.NumField(); o++ {
			outFieldName := ot.Field(o).Name
			outFieldType := ot.Field(o).Type

			if EqualAttrName(inFieldName, outFieldName) {
				if iv.FieldByName(inFieldName).IsValid() && ov.FieldByName(outFieldName).CanSet() {
					if outFieldType == inFieldType {
						ov.FieldByName(outFieldName).Set(iv.FieldByName(inFieldName))
					} else {
						if inFieldType.ConvertibleTo(outFieldType) {
							ov.FieldByName(outFieldName).Set(reflect.ValueOf(iv.FieldByName(inFieldName)).Convert(outFieldType))
						} else {
							// core.App.Logger().Debugf("set attr failed: name [%s],type [%s] - [%s] ", outFieldName, inFieldType.String(), outFieldType.String())
						}
					}
				}
				break
			}
		}
	}
}

func SetStructFromOther(inStruct interface{}, outStruct interface{}) {
	t := reflect.TypeOf(outStruct).Elem()
	o := reflect.ValueOf(outStruct).Elem()
	i := reflect.ValueOf(inStruct).Elem()

	for k := 0; k < t.NumField(); k++ {
		fieldName := t.Field(k).Name
		fieldType := t.Field(k).Type

		if i.FieldByName(fieldName).IsValid() && o.FieldByName(fieldName).CanSet() {
			dataType := i.FieldByName(fieldName).Type()
			if dataType == fieldType {
				o.FieldByName(fieldName).Set(i.FieldByName(fieldName))
			} else {
				if dataType.ConvertibleTo(fieldType) {
					o.FieldByName(fieldName).Set(reflect.ValueOf(i.FieldByName(fieldName)).Convert(fieldType))
				} else {
					// core.App.Logger().Debugf("set attr failed: name [%s],type [%s] - [%s] ", fieldName, fieldType.String(), dataType.String())
				}
			}
		}
	}
}

// StructToMap 将结构体转换为 map[string]interface{}，并删除零值字段
func StructToMap(data interface{}) (map[string]string, error) {
	// 将结构体序列化为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	// 反序列化为 map[string]interface{}
	var tempMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &tempMap); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	// 删除零值字段
	for key, value := range tempMap {
		if IsZeroValue(value) {
			continue
		}
		result[key] = fmt.Sprintf("%v", value)
	}
	return result, nil
}

func StringToMap(data string) (map[string]string, error) {
	// 反序列化为 map[string]interface{}
	var tempMap map[string]interface{}
	if err := json.Unmarshal([]byte(data), &tempMap); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	// 删除零值字段
	for key, value := range tempMap {
		if IsZeroValue(value) {
			continue
		}
		result[key] = fmt.Sprintf("%v", value)
	}
	return result, nil
}

func MergeMaps[K comparable, V any](map1, map2 map[K]V) map[K]V {
	mergedMap := make(map[K]V, len(map1)+len(map2))

	for key, value := range map1 {
		mergedMap[key] = value
	}
	for key, value := range map2 {
		mergedMap[key] = value
	}
	return mergedMap
}

/**
 * 字符串转map
 */
func JsonToMap(jsonStr []byte) map[string]interface{} {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonStr, &data); err == nil {
		return data
	}
	return data
}

/**
 * 生成毫秒时间戳
 */
func GetMillisecond() string {
	signTime := time.Now().UnixNano() / 1e6
	signTimeStr := strconv.Itoa(int(signTime))
	return signTimeStr
}

func JoinStringsInOrder(data map[string]string, sep string, onlyValues, includeEmpty bool, exceptKeys ...string) string {
	var list []string
	var keyList []string
	m := make(map[string]int)
	if len(exceptKeys) > 0 {
		for _, except := range exceptKeys {
			m[except] = 1
		}
	}
	for k := range data {
		if _, ok := m[k]; ok {
			continue
		}
		value := data[k]
		if !includeEmpty && value == "" {
			continue
		}
		if onlyValues {
			keyList = append(keyList, k)
		} else {
			list = append(list, fmt.Sprintf("%s=%s", k, value))
		}
	}
	if onlyValues {
		sort.Strings(keyList)
		for _, v := range keyList {
			list = append(list, data[v])
		}
	} else {
		sort.Strings(list)
	}
	return strings.Join(list, sep)

}

func GenerateRandString(n int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // 创建新的随机数生成器
	b := make([]rune, n)
	for i := range b {
		b[i] = rune(chars[r.Intn(len(chars))])
	}
	return string(b)
}

func GenerateStringKey(args ...string) string {
	return strings.Join(args, ":")
}

func GetRandomStringFromSlice(slice []string) string {
	// 如果切片为空，返回空字符串
	if len(slice) == 0 {
		return ""
	}
	// 使用当前时间戳作为随机种子
	rand.Seed(time.Now().UnixNano())
	// 获取随机索引
	randomIndex := rand.Intn(len(slice))
	return slice[randomIndex]
}

// IsAllEmojiOrTags 判断字符串是否全由 emoji 或 [xxx] 标签构成
func IsAllEmojiOrTags(str string) bool {
	if str == "" {
		return false
	}

	// emoji Unicode 范围（部分示例，可扩展）
	emojiRanges := [][2]rune{
		{0x1F600, 0x1F64F}, // 表情符号
		{0x1F300, 0x1F5FF}, // 杂项符号和象形文字
		{0x1F680, 0x1F6FF}, // 交通和地图符号
		{0x1F900, 0x1F9FF}, // 补充符号
		{0x200D, 0x200D},   // 零宽连接符
	}

	for i := 0; i < len(str); {
		// 检查 [xxx] 标签
		if i < len(str)-1 && str[i] == '[' {
			end := strings.Index(str[i:], "]")
			if end == -1 {
				return false // 未闭合的 [
			}
			end += i
			if end == i+1 {
				return false // 空标签 []
			}
			i = end + 1 // 跳过整个标签
			continue
		}

		// 检查 Unicode emoji
		r, size := utf8.DecodeRuneInString(str[i:])
		isEmoji := false
		for _, rRange := range emojiRanges {
			if r >= rRange[0] && r <= rRange[1] {
				isEmoji = true
				break
			}
		}
		if !isEmoji {
			return false
		}
		i += size
	}
	return true
}

// 版本比较
func CompareVersions(v1, v2 string) int {
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		v1Num := atoi(v1Parts[i])
		v2Num := atoi(v2Parts[i])
		if v1Num > v2Num {
			return 1
		} else if v1Num < v2Num {
			return -1
		}
	}
	return 0
}

// 字符串转整数（简易版）
func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
