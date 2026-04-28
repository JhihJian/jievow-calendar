package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// FlowerRecord represents a single province-solar_term-flower combination
type FlowerRecord struct {
	Province  string   `json:"province"`
	SolarTerm string   `json:"solar_term"`
	Flowers   []string `json:"flowers"`
}

// Base flower data from traditional 花信风 (Flower Signal Wind)
var traditionalFlowers = map[string][]string{
	"小寒": {"梅花", "山茶花", "水仙"},
	"大寒": {"瑞香花", "兰花", "山矾花"},
	"立春": {"迎春花", "樱桃", "望春"},
	"雨水": {"菜花", "杏花", "李花"},
	"惊蛰": {"桃花", "棠棣", "蔷薇"},
	"春分": {"海棠", "梨花", "木兰"},
	"清明": {"桐花", "麦花", "柳花"},
	"谷雨": {"牡丹", "荼蘼", "楝花"},
}

// Summer/Autumn/Winter flowers (after 谷雨)
var seasonalFlowers = map[string][]string{
	"立夏": {"月季", "芍药", "石榴花"},
	"小满": {"栀子花", "金银花", "合欢花"},
	"芒种": {"玉簪花", "百合", "萱草"},
	"夏至": {"荷花", "紫薇", "夹竹桃"},
	"小暑": {"茉莉花", "凌霄花", "睡莲"},
	"大暑": {"木槿", "凤仙花", "鸡冠花"},
	"立秋": {"桂花", "秋海棠", "紫薇"},
	"处暑": {"桂花", "木槿", "美人蕉"},
	"白露": {"桂花", "菊花", "木芙蓉"},
	"秋分": {"菊花", "桂花", "木芙蓉"},
	"寒露": {"菊花", "秋海棠", "木芙蓉"},
	"霜降": {"菊花", "山茶花", "木芙蓉"},
	"立冬": {"山茶花", "梅花", "八角金盘"},
	"小雪": {"梅花", "山茶花", "蜡梅"},
	"大雪": {"梅花", "蜡梅", "山茶花"},
	"冬至": {"梅花", "蜡梅", "水仙"},
}

// 34 Provinces in exact order
var provinces = []string{
	"北京", "天津", "河北", "山西", "内蒙古",
	"辽宁", "吉林", "黑龙江",
	"上海", "江苏", "浙江", "安徽", "福建", "江西", "山东",
	"河南", "湖北", "湖南",
	"广东", "广西", "海南",
	"重庆", "四川", "贵州", "云南",
	"西藏",
	"陕西", "甘肃", "青海", "宁夏", "新疆",
	"香港", "澳门", "台湾",
}

// 24 Solar Terms in exact order
var solarTerms = []string{
	"小寒", "大寒", "立春", "雨水", "惊蛰", "春分",
	"清明", "谷雨", "立夏", "小满", "芒种", "夏至",
	"小暑", "大暑", "立秋", "处暑", "白露", "秋分",
	"寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
}

// Regional specialty flowers
var regionalSpecialties = map[string][]string{
	"北京":   {"腊梅", "玉兰", "紫藤"},
	"天津":   {"月季", "海棠"},
	"河北":   {"枣花", "苹果花"},
	"山西":   {"杏花", "枣花"},
	"内蒙古": {"沙棘花", "马兰花"},
	"辽宁":   {"梨花", "苹果花"},
	"吉林":   {"梨花", "向日葵"},
	"黑龙江": {"兴安杜鹃", "杏花"},
	"上海":   {"白玉兰", "桂花"},
	"江苏":   {"茉莉", "琼花"},
	"浙江":   {"桂花", "茶花"},
	"安徽":   {"黄山杜鹃", "金银花"},
	"福建":   {"水仙", "茉莉", "三角梅"},
	"江西":   {"杜鹃", "荷花"},
	"山东":   {"牡丹", "蔷薇"},
	"河南":   {"牡丹", "菊花"},
	"湖北":   {"梅花", "樱花"},
	"湖南":   {"杜鹃", "荷花"},
	"广东":   {"木棉花", "紫荆花", "簕杜鹃"},
	"广西":   {"桂花", "三角梅"},
	"海南":   {"三角梅", "鸡蛋花", "木棉花", "凤凰花"},
	"重庆":   {"山茶花", "杜鹃"},
	"四川":   {"杜鹃", "芙蓉花"},
	"贵州":   {"杜鹃", "油菜花"},
	"云南":   {"山茶花", "杜鹃", "报春花", "龙胆花", "绿绒蒿"},
	"西藏":   {"格桑花", "雪莲花"},
	"陕西":   {"山丹花", "芍药"},
	"甘肃":   {"马兰花", "杏花"},
	"青海":   {"格桑花", "油菜花"},
	"宁夏":   {"沙枣花", "枸杞花"},
	"新疆":   {"棉花", "杏花", "薰衣草"},
	"香港":   {"紫荆花", "杜鹃", "木棉花"},
	"澳门":   {"荷花", "菊花"},
	"台湾":   {"樱花", "蝴蝶兰", "姜花"},
}

// getBaseFlowers returns the base flower list for a solar term
func getBaseFlowers(solarTerm string) []string {
	if flowers, ok := traditionalFlowers[solarTerm]; ok {
		return makeCopy(flowers)
	}
	if flowers, ok := seasonalFlowers[solarTerm]; ok {
		return makeCopy(flowers)
	}
	return []string{}
}

// makeCopy creates a copy of a string slice
func makeCopy(s []string) []string {
	result := make([]string, len(s))
	copy(result, s)
	return result
}

// adjustForRegion modifies the flower list based on province characteristics
func adjustForRegion(province, solarTerm string, flowers []string) []string {
	// Add regional specialties when appropriate
	if specialties, ok := regionalSpecialties[province]; ok {
		// Add up to 2 specialty flowers if not already present
		added := 0
		for _, spec := range specialties {
			if added >= 2 {
				break
			}
			if !contains(flowers, spec) {
				flowers = append(flowers, spec)
				added++
			}
		}
	}

	// Regional timing adjustments
	switch {
	case isNorthChina(province):
		// North China: flowers bloom later
		if solarTerm == "小寒" || solarTerm == "大寒" {
			// Very limited flowers in deep winter
			if isNorthEast(province) {
				return []string{"蜡梅"}
			}
			return []string{"蜡梅", "梅花"}
		}
		if solarTerm == "立春" {
			// Spring flowers delayed
			flowers = []string{"梅花", "蜡梅"}
		}

	case isSouthChina(province):
		// South China: flowers bloom earlier
		if solarTerm == "大寒" || solarTerm == "立春" {
			// Add more early spring flowers
			if !contains(flowers, "桃花") {
				flowers = append(flowers, "桃花")
			}
			if !contains(flowers, "木棉花") {
				flowers = append(flowers, "木棉花")
			}
		}
		if province == "海南" {
			// Tropical flowers year-round
			tropicalFlowers := []string{"三角梅", "鸡蛋花", "凤凰花"}
			for _, tf := range tropicalFlowers {
				if !contains(flowers, tf) {
					flowers = append(flowers, tf)
				}
			}
		}

	case isHighAltitude(province):
		// High altitude: very limited flowers
		if isWinter(solarTerm) {
			return []string{}
		}
		if isEarlySpring(solarTerm) {
			return []string{"格桑花"}
		}
		// Summer has more flowers
		if isSummer(solarTerm) {
			if !contains(flowers, "格桑花") {
				flowers = append(flowers, "格桑花")
			}
		}

	case isNorthWest(province):
		// Northwest: dry and cold
		if isWinter(solarTerm) || isEarlySpring(solarTerm) {
			// Limited flowers
			if len(flowers) > 2 {
				return flowers[:2]
			}
		}

	case isSouthWest(province):
		// Southwest: mild climate, flowers year-round
		if province == "云南" {
			// Yunnan is "春城" - more flowers in winter
			if isWinter(solarTerm) {
				if !contains(flowers, "山茶花") {
					flowers = append(flowers, "山茶花")
				}
				if !contains(flowers, "报春花") {
					flowers = append(flowers, "报春花")
				}
			}
		}

	case isCoastal(province):
		// East/South coastal: moderate climate
		if isWinter(solarTerm) && !isNorthEast(province) {
			// Milder winter, more flowers
			if !contains(flowers, "水仙") {
				flowers = append(flowers, "水仙")
			}
		}

	case province == "香港" || province == "澳门":
		// Subtropical: flowers year-round
		if !contains(flowers, "紫荆花") && (province == "香港" || solarTerm == "立春") {
			flowers = append(flowers, "紫荆花")
		}
		if !contains(flowers, "木棉花") {
			flowers = append(flowers, "木棉花")
		}

	case province == "台湾":
		// Subtropical/tropical
		if !contains(flowers, "樱花") {
			flowers = append(flowers, "樱花")
		}
		if isWinter(solarTerm) {
			if !contains(flowers, "蝴蝶兰") {
				flowers = append(flowers, "蝴蝶兰")
			}
		}
	}

	// Limit to 2-5 flowers per record
	if len(flowers) > 5 {
		flowers = flowers[:5]
	}

	return flowers
}

// Helper functions for region classification
func isNorthChina(province string) bool {
	north := []string{"北京", "天津", "河北", "山西", "内蒙古"}
	for _, p := range north {
		if province == p {
			return true
		}
	}
	return false
}

func isNorthEast(province string) bool {
	northeast := []string{"辽宁", "吉林", "黑龙江"}
	for _, p := range northeast {
		if province == p {
			return true
		}
	}
	return false
}

func isSouthChina(province string) bool {
	south := []string{"广东", "广西", "海南", "福建"}
	for _, p := range south {
		if province == p {
			return true
		}
	}
	return false
}

func isHighAltitude(province string) bool {
	return province == "西藏" || province == "青海"
}

func isNorthWest(province string) bool {
	northwest := []string{"新疆", "甘肃", "宁夏"}
	for _, p := range northwest {
		if province == p {
			return true
		}
	}
	return false
}

func isSouthWest(province string) bool {
	southwest := []string{"云南", "四川", "贵州"}
	for _, p := range southwest {
		if province == p {
			return true
		}
	}
	return false
}

func isCoastal(province string) bool {
	coastal := []string{"上海", "江苏", "浙江", "安徽"}
	for _, p := range coastal {
		if province == p {
			return true
		}
	}
	return false
}

func isWinter(solarTerm string) bool {
	winter := []string{"小寒", "大寒", "立冬", "小雪", "大雪", "冬至"}
	for _, st := range winter {
		if solarTerm == st {
			return true
		}
	}
	return false
}

func isEarlySpring(solarTerm string) bool {
	spring := []string{"立春", "雨水", "惊蛰"}
	for _, st := range spring {
		if solarTerm == st {
			return true
		}
	}
	return false
}

func isSummer(solarTerm string) bool {
	summer := []string{"立夏", "小满", "芒种", "夏至", "小暑", "大暑"}
	for _, st := range summer {
		if solarTerm == st {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	var records []FlowerRecord

	// Generate records for all province-solar_term combinations
	for _, province := range provinces {
		for _, solarTerm := range solarTerms {
			baseFlowers := getBaseFlowers(solarTerm)
			adjustedFlowers := adjustForRegion(province, solarTerm, baseFlowers)

			record := FlowerRecord{
				Province:  province,
				SolarTerm: solarTerm,
				Flowers:   adjustedFlowers,
			}
			records = append(records, record)
		}
	}

	// Validate count
	expectedCount := 34 * 24 // 816 records
	if len(records) != expectedCount {
		fmt.Printf("ERROR: Expected %d records, got %d\n", expectedCount, len(records))
		os.Exit(1)
	}

	// Write to JSON file with 2-space indent
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	// Add trailing newline
	data = append(data, '\n')

	outputPath := "/data/dev/jievow/jievow-calendar/data/flowers.json"
	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		fmt.Printf("ERROR: Failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated %d flower records\n", len(records))
	fmt.Printf("✓ Wrote to %s\n", outputPath)
}
