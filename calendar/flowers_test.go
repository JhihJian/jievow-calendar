// calendar/flowers_test.go
package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidSolarTerms(t *testing.T) {
	if len(ValidSolarTerms) != 24 {
		t.Errorf("ValidSolarTerms len want 24 got %d", len(ValidSolarTerms))
	}
	lookup := map[string]bool{}
	for _, term := range ValidSolarTerms {
		lookup[term] = true
	}
	for _, term := range []string{"立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
		"立夏", "小满", "芒种", "夏至", "小暑", "大暑",
		"立秋", "处暑", "白露", "秋分", "寒露", "霜降",
		"立冬", "小雪", "大雪", "冬至", "小寒", "大寒"} {
		if !lookup[term] {
			t.Errorf("missing solar term %q", term)
		}
	}
}

func TestFlowerStoreLoadAndQuery(t *testing.T) {
	dir := t.TempDir()
	records := []FlowerRecord{
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花", "山茶花"}},
		{Province: "浙江", SolarTerm: "立春", Flowers: []string{"梅花", "水仙"}},
		{Province: "北京", SolarTerm: "雨水", Flowers: []string{"迎春花", "玉兰花"}},
	}
	data, _ := json.MarshalIndent(records, "", "  ")
	dataFile := filepath.Join(dir, "flowers.json")
	dataWithNewline := append(data, '\n')
	os.WriteFile(dataFile, dataWithNewline, 0o644)

	hash := sha256.Sum256(dataWithNewline)
	checksumLine := fmt.Sprintf("%x  flowers.json\n", hash)
	os.WriteFile(filepath.Join(dir, "flowers.json.sha256"), []byte(checksumLine), 0o644)

	store, err := LoadFlowerStore(dir)
	if err != nil {
		t.Fatalf("LoadFlowerStore: %v", err)
	}

	flowers, ok := store.GetFlowers("北京", "立春")
	if !ok {
		t.Fatal("expected to find 北京 立春")
	}
	if len(flowers) != 2 || flowers[0] != "梅花" {
		t.Errorf("flowers want [梅花 山茶花] got %v", flowers)
	}

	flowers, ok = store.GetFlowers("浙江", "立春")
	if !ok {
		t.Fatal("expected to find 浙江 立春")
	}
	if flowers[0] != "梅花" {
		t.Errorf("first flower want 梅花 got %s", flowers[0])
	}

	_, ok = store.GetFlowers("上海", "立春")
	if ok {
		t.Error("expected not to find 上海 立春")
	}
}

func TestFlowerStoreListProvinces(t *testing.T) {
	dir := t.TempDir()
	records := []FlowerRecord{
		{Province: "浙江", SolarTerm: "立春", Flowers: []string{"梅花"}},
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花"}},
		{Province: "广东", SolarTerm: "立春", Flowers: []string{"桃花"}},
	}
	data, _ := json.MarshalIndent(records, "", "  ")
	dataFile := filepath.Join(dir, "flowers.json")
	dataWithNewline := append(data, '\n')
	os.WriteFile(dataFile, dataWithNewline, 0o644)

	hash := sha256.Sum256(dataWithNewline)
	checksumLine := fmt.Sprintf("%x  flowers.json\n", hash)
	os.WriteFile(filepath.Join(dir, "flowers.json.sha256"), []byte(checksumLine), 0o644)

	store, err := LoadFlowerStore(dir)
	if err != nil {
		t.Fatalf("LoadFlowerStore: %v", err)
	}

	provinces := store.ListProvinces()
	if len(provinces) != 3 {
		t.Fatalf("provinces want 3 got %d", len(provinces))
	}
	if provinces[0] != "北京" {
		t.Errorf("first province want 北京 got %s", provinces[0])
	}
}

func TestFlowerStoreBadChecksum(t *testing.T) {
	dir := t.TempDir()
	records := []FlowerRecord{{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花"}}}
	data, _ := json.Marshal(records)
	os.WriteFile(filepath.Join(dir, "flowers.json"), data, 0o644)
	os.WriteFile(filepath.Join(dir, "flowers.json.sha256"), []byte("0000  flowers.json\n"), 0o644)

	_, err := LoadFlowerStore(dir)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestNewFlowerStore(t *testing.T) {
	records := []FlowerRecord{
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花"}},
	}
	store := NewFlowerStore(records)
	flowers, ok := store.GetFlowers("北京", "立春")
	if !ok {
		t.Fatal("expected to find 北京 立春")
	}
	if len(flowers) != 1 || flowers[0] != "梅花" {
		t.Errorf("flowers want [梅花] got %v", flowers)
	}
}
