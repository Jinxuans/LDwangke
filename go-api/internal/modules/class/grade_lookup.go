package class

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go-api/internal/database"
)

type GradeRecord struct {
	ID     int
	Sort   int
	Name   string
	Rate   float64
	Money  float64
	AddKF  string
	GJKF   string
	Status string
	Time   string
}

func formatGradeRate(rate float64) string {
	return strconv.FormatFloat(rate, 'f', 2, 64)
}

func gradeRatesEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.00001
}

func scanGradeRecord(scanner interface {
	Scan(dest ...interface{}) error
}) (*GradeRecord, error) {
	var record GradeRecord
	var rateText string
	if err := scanner.Scan(
		&record.ID,
		&record.Sort,
		&record.Name,
		&rateText,
		&record.Money,
		&record.AddKF,
		&record.GJKF,
		&record.Status,
		&record.Time,
	); err != nil {
		return nil, err
	}
	rate, err := strconv.ParseFloat(strings.TrimSpace(rateText), 64)
	if err != nil {
		return nil, fmt.Errorf("等级费率格式无效: %w", err)
	}
	record.Rate = rate
	return &record, nil
}

func (s *classService) GetGradeByID(id int, activeOnly bool) (*GradeRecord, error) {
	if id <= 0 {
		return nil, sql.ErrNoRows
	}
	query := `
SELECT id,
       COALESCE(sort, '0'),
       COALESCE(name, ''),
       COALESCE(rate, '1'),
       COALESCE(money, 0),
       COALESCE(addkf, '0'),
       COALESCE(gjkf, '0'),
       COALESCE(status, '1'),
       CASE
         WHEN time IS NOT NULL AND time != '' AND time != '0' THEN FROM_UNIXTIME(CAST(time AS UNSIGNED), '%Y-%m-%d %H:%i')
         ELSE ''
       END
FROM qingka_wangke_dengji
WHERE id = ?`
	args := []interface{}{id}
	if activeOnly {
		query += " AND status = '1'"
	}
	return scanGradeRecord(database.DB.QueryRow(query, args...))
}

func (s *classService) ResolveSelectedGrade(gradeID int, activeOnly bool) (*GradeRecord, error) {
	if gradeID > 0 {
		record, err := s.GetGradeByID(gradeID, activeOnly)
		if err == nil {
			return record, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return nil, sql.ErrNoRows
}

func (s *classService) CountGradeAssignments(id int) (int, error) {
	if _, err := s.GetGradeByID(id, false); err != nil {
		return 0, err
	}
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_wangke_user WHERE grade_id = ? OR invite_grade_id = ?",
		id, id,
	).Scan(&count)
	return count, err
}

func (s *classService) GradeRateExists(rate float64, excludeID int) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_wangke_dengji WHERE rate = ? AND id <> ?",
		formatGradeRate(rate),
		excludeID,
	).Scan(&count)
	return count > 0, err
}

func (s *classService) GradeNameExists(name string, excludeID int) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM qingka_wangke_dengji WHERE name = ? AND id <> ?",
		strings.TrimSpace(name),
		excludeID,
	).Scan(&count)
	return count > 0, err
}
