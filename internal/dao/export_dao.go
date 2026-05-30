package dao

import (
	"context"
	"filler/internal/database"
	"fmt"
	"os"
	"strings"
)

type fileRotator struct {
	baseName    string
	author      string
	fileIndex   int
	lineCount   int
	currentFile *os.File
}

func newFileRotator(baseName, author string) *fileRotator {
	baseName = strings.TrimSuffix(baseName, ".sql")
	return &fileRotator{
		baseName:  baseName,
		author:    author,
		fileIndex: 0,
		lineCount: 0,
	}
}

func (fr *fileRotator) writeString(str string) error {
	linesInStr := strings.Count(str, "\n")
	if fr.currentFile != nil && fr.lineCount+linesInStr > 5000 {
		fr.currentFile.Close()
		fr.currentFile = nil
	}
	if fr.currentFile == nil {
		fr.fileIndex += 1
		fr.lineCount = 0
		fullName := fmt.Sprintf("%s-%d.sql", fr.baseName, fr.fileIndex)
		if _, err := os.Stat(fullName); err == nil {
			if err := os.Remove(fullName); err != nil {
				return fmt.Errorf("не удалось удалить существующий файл %s: %w", fullName, err)
			}
		}
		f, err := os.Create(fullName)
		if err != nil {
			return fmt.Errorf("не удалось создать файл %s: %w", fullName, err)
		}
		fr.currentFile = f

		header := "-- liquibase formatted sql\n"
		if _, err := fr.currentFile.WriteString(header); err != nil {
			return err
		}
		fr.lineCount += strings.Count(header, "\n")
	}
	if _, err := fr.currentFile.WriteString(str); err != nil {
		return err
	}
	fr.lineCount += linesInStr
	return nil
}

func (fr *fileRotator) close() {
	if fr.currentFile != nil {
		fr.currentFile.Close()
	}
}

func ExportDatabaseToLiquibase(ctx context.Context, filename string, author string, startValue int) error {
	conn := database.Get()
	rotator := newFileRotator(filename, author)
	defer rotator.close()
	currentID := startValue
	tables := []struct {
		name    string
		columns []string
	}{
		{"public.param", []string{"id", "title", "name_en", "name_ru", "type", "value", "description_en", "description_ru", "units_en", "units_ru", "access", "saved", "visible"}},
		{"public.component", []string{"id", "title", "name_en", "name_ru", "base_component", "description_en", "description_ru", "access"}},
		{"public.component_param", []string{"component_id", "param_id"}},
		{"public.device_indicator", []string{"id", "description", "object_id", "contact", "name", "location", "services"}},
		{"public.param_indicator", []string{"id", "oid_id", "dotter_notation"}},
		{"public.mapping", []string{"id", "indicator", "param", "frequency", "coefficient", "enum"}},
		{"public.device_component", []string{"id", "model", "internal_order", "parent"}},
		{"public.device_component_mapping", []string{"device_component_id", "mapping_id"}},
		{"public.configuration", []string{"id", "indicator", "device_component_id"}},
		{"public.default_configuration", []string{"id", "indicator", "device_component_id"}},
		{"public.threshold", []string{"id", "source_model", "source_internal_order", "source_param", "value", "type", "operator", "enabled", "target_param", "target_device", "level", "prev_operator", "previous"}},
		{"public.configuration_threshold", []string{"configuration_id", "threshold_id"}},
		{"public.default_configuration_threshold", []string{"default_configuration_id", "threshold_id"}},
	}
	for _, t := range tables {
		quotedCols := make([]string, len(t.columns))
		selectCols := make([]string, len(t.columns))
		for i, col := range t.columns {
			quotedCols[i] = fmt.Sprintf(`"%s"`, col)
			if col == "oid_id" || col == "id" && (t.name == "public.oid" || t.name == "public.param_indicator") {
				selectCols[i] = fmt.Sprintf(`"%s"::text`, col)
			} else if col == "enum" || col == "coefficient" {
				selectCols[i] = fmt.Sprintf(`"%s"::text`, col)
			} else {
				selectCols[i] = fmt.Sprintf(`"%s"`, col)
			}
		}
		colsStrSelect := strings.Join(selectCols, ", ")
		colsStrInsert := strings.Join(quotedCols, ", ")
		selectQuery := fmt.Sprintf("SELECT %s FROM %s ORDER BY 1", colsStrSelect, t.name)
		rows, err := conn.Query(ctx, selectQuery)
		if err != nil {
			return fmt.Errorf("ошибка вычитки таблицы %s: %w", t.name, err)
		}
		var allValues []string
		for rows.Next() {
			scannedValues := make([]interface{}, len(t.columns))
			scannedPointers := make([]interface{}, len(t.columns))
			for i := range scannedValues {
				scannedPointers[i] = &scannedValues[i]
			}
			if err := rows.Scan(scannedPointers...); err != nil {
				rows.Close()
				return fmt.Errorf("ошибка сканирования строки таблицы %s: %w", t.name, err)
			}
			var rowValues []string
			for _, val := range scannedValues {
				if val == nil {
					rowValues = append(rowValues, "NULL")
				} else {
					switch v := val.(type) {
					case string:
						escaped := strings.ReplaceAll(v, "'", "''")
						rowValues = append(rowValues, fmt.Sprintf("'%s'", escaped))
					case []byte:
						escaped := strings.ReplaceAll(string(v), "'", "''")
						rowValues = append(rowValues, fmt.Sprintf("'%s'", escaped))
					case bool:
						if v {
							rowValues = append(rowValues, "TRUE")
						} else {
							rowValues = append(rowValues, "FALSE")
						}
					default:
						rowValues = append(rowValues, fmt.Sprintf("%v", v))
					}
				}
			}
			allValues = append(allValues, fmt.Sprintf("(%s)", strings.Join(rowValues, ", ")))
		}
		rows.Close()
		if 0 == len(allValues) {
			continue
		}
		limit := 5000
		for i := 0; len(allValues) > i; i += limit {
			end := i + limit
			if end > len(allValues) {
				end = len(allValues)
			}
			chunk := allValues[i:end]
			var sb strings.Builder
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("-- changeset %s:%d\n", author, currentID))
			currentID += 1
			sb.WriteString(fmt.Sprintf("INSERT INTO %s (%s)\n", t.name, colsStrInsert))
			sb.WriteString("VALUES ")
			sb.WriteString(chunk[0]) // Исправлено: берем первый элемент слайса строк
			indent := "       "
			for j := 1; len(chunk) > j; j++ {
				sb.WriteString(",\n")
				sb.WriteString(indent)
				sb.WriteString(chunk[j])
			}
			sb.WriteString(";\n")
			if err := rotator.writeString(sb.String()); err != nil {
				return err
			}
		}
	}
	return nil
}
