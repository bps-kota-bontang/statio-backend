package utils

import (
	"sort"
	"statio/internal/models"
	"strings"
)

// helpers
func MapDimensionValues(table *models.Table) map[string]string {
	dimMap := make(map[string]string)
	for _, td := range table.Dimensions {
		if td.Dimension == nil {
			continue
		}
		for _, val := range td.Dimension.Values {
			dimMap[val.ID] = td.ID
		}
	}
	return dimMap
}

func DimensionValueKeyFromIDs(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	sorted := append([]string{}, ids...) // copy agar original tidak berubah
	sort.Strings(sorted)
	key := strings.Join(sorted, "|") // gunakan separator aman untuk UUID

	return key
}
