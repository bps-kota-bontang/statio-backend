package utils

import (
	"errors"
	"math"
	"sort"
	"statio/internal/models"
	"strings"
)

// Helper fungsi median
func median(data []float64) float64 {
	n := len(data)
	mid := n / 2

	if n%2 == 0 {
		return (data[mid-1] + data[mid]) / 2
	}
	return data[mid]
}

const OutlierThreshold = 3.5

// DetectOutliersModifiedZ menerima slice float64 dan mengembalikan:
// - daftar index outlier
// - daftar skor Modified Z untuk setiap nilai
func DetectOutliersModifiedZ(values []float64) ([]int, []float64, error) {
	n := len(values)
	if n < 3 {
		return nil, nil, errors.New("minimal 3 data untuk mendeteksi outlier")
	}

	// 1. Hitung median
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	medianValue := median(sorted)

	// 2. Hitung |xi - median|
	absDiff := make([]float64, n)
	for i, v := range values {
		absDiff[i] = math.Abs(v - medianValue)
	}

	// 3. Hitung MAD (Median Absolute Deviation)
	sortedDiff := make([]float64, n)
	copy(sortedDiff, absDiff)
	sort.Float64s(sortedDiff)
	mad := median(sortedDiff)

	if mad == 0 {
		// semua data hampir sama → tidak bisa hitung Mz
		mz := make([]float64, n)
		return []int{}, mz, nil
	}

	// 4. Hitung Modified Z-score
	mz := make([]float64, n)
	for i, v := range values {
		mz[i] = 0.6745 * (v - medianValue) / mad
	}

	// 5. Outlier jika |Mz| > OutlierThreshold (3.5)
	var outliers []int
	for i, score := range mz {
		if math.Abs(score) > OutlierThreshold {
			outliers = append(outliers, i)
		}
	}

	return outliers, mz, nil
}

func BuildDimensionKey(f *models.Fact) string {
	ids := make([]string, len(f.FactDimensionValues))
	for i, dv := range f.FactDimensionValues {
		ids[i] = dv.DimensionValueID
	}
	sort.Strings(ids)
	return strings.Join(ids, "|")
}
