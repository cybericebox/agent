package tools

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/cybericebox/agent/internal/model"
	"strings"
)

func RecordsToStr(records []model.DNSRecordConfig) string {
	var strRecords []string

	for _, r := range records {
		join, _ := strings.CutSuffix(strings.Join([]string{r.Type, r.Name, r.Data}, "___"), "___")
		strRecords = append(strRecords, join)
	}

	return strings.Join(strRecords, "---")
}

func RecordsFromStr(str string) []model.DNSRecordConfig {
	var records []model.DNSRecordConfig

	rData := strings.Split(str, "---")
	for _, r := range rData {
		rItem := strings.Split(r, "___")
		// if the recordData is empty, append an empty string to the end
		rItem = append(rItem, "")
		records = append(records, model.DNSRecordConfig{
			Type: rItem[0],
			Name: rItem[1],
			Data: rItem[2],
		})
	}

	return records
}

func GetLabel(values ...string) string {
	labelSHA := sha256.Sum256([]byte(strings.Join(values, "")))
	cut, _ := strings.CutSuffix(base64.URLEncoding.EncodeToString(labelSHA[:]), "=")
	return fmt.Sprintf("A%sA", cut)
}
