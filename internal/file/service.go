package file

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"nordik-drive-api/internal/auth"
	"path/filepath"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type FileService struct {
	DB *gorm.DB
}

func (fs *FileService) SaveFilesMultipart(uploadedFiles []*multipart.FileHeader, filenames FileUploadInput, userID uint) ([]File, error) {
	var savedFiles []File
	files := filenames.FileNames
	privateList := filenames.Private

	for i, fileHeader := range uploadedFiles {
		filename := files[i]
		private := privateList[i]

		// check duplicate
		var existing File
		if err := fs.DB.Where("filename = ?", filename).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("file with name %s already exists", filename)
		}

		f, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()

		ext := filepath.Ext(fileHeader.Filename)
		var headers []string
		var dataRows [][]string

		if ext == ".xlsx" || ext == ".xls" {
			headers, dataRows, err = parseExcelReader(f)
		} else if ext == ".csv" {
			headers, dataRows, err = parseCSVReader(f)
		} else {
			return nil, fmt.Errorf("unsupported file type: %s", ext)
		}
		if err != nil {
			return nil, err
		}

		newFile := File{
			Filename:   filename,
			InsertedBy: userID,
			CreatedAt:  time.Now(),
			Private:    private,
			Version:    1,
			IsDelete:   false,
			Rows:       len(dataRows),
			Size:       float64(fileHeader.Size) / 1024.0,
		}
		if err := fs.DB.Create(&newFile).Error; err != nil {
			return nil, err
		}

		fileVersion := FileVersion{
			FileID:     newFile.ID,
			Filename:   filename,
			InsertedBy: userID,
			CreatedAt:  time.Now(),
			Private:    private,
			Version:    1,
			IsDelete:   false,
			Rows:       len(dataRows),
			Size:       float64(fileHeader.Size) / 1024.0,
		}
		if err := fs.DB.Create(&fileVersion).Error; err != nil {
			return nil, err
		}

		// save each row as ordered JSON
		for _, row := range dataRows {
			rowMap := orderedmap.New()
			for j, header := range headers {
				val := ""
				if j < len(row) {
					val = row[j]
				}
				rowMap.Set(header, val) // preserves column order
			}

			jsonBytes, err := rowMap.MarshalJSON()
			if err != nil {
				return nil, err
			}

			record := FileData{
				FileID:     newFile.ID,
				RowData:    jsonBytes,
				InsertedBy: userID,
				CreatedAt:  time.Now(),
				Version:    1,
			}
			if err := fs.DB.Create(&record).Error; err != nil {
				return nil, err
			}
		}

		savedFiles = append(savedFiles, newFile)
	}

	return savedFiles, nil
}

// parseExcelReader ensures each row has all columns, aligned with headers
func parseExcelReader(file multipart.File) ([]string, [][]string, error) {
	defer file.Seek(0, 0)

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read excel file: %w", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse excel file: %w", err)
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 1 {
		return nil, nil, fmt.Errorf("excel file is empty")
	}

	headers := rows[0]
	var dataRows [][]string

	for rowIdx, _ := range rows[1:] {
		newRow := make([]string, len(headers))
		for colIdx := range headers {
			cellRef, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			val, _ := f.GetCellValue(sheetName, cellRef)
			newRow[colIdx] = val
		}
		dataRows = append(dataRows, newRow)
	}

	return headers, dataRows, nil
}

func (fs *FileService) ReplaceFiles(uploadedFile *multipart.FileHeader, fileID uint, userID uint) error {
	var existing File
	if err := fs.DB.First(&existing, fileID).Error; err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	f, err := uploadedFile.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	// 3. Parse file
	ext := filepath.Ext(uploadedFile.Filename)
	var headers []string
	var dataRows [][]string

	if ext == ".xlsx" || ext == ".xls" {
		headers, dataRows, err = parseExcelReader(f)
	} else if ext == ".csv" {
		headers, dataRows, err = parseCSVReader(f)
	} else {
		return fmt.Errorf("unsupported file type: %s", ext)
	}
	if err != nil {
		return err
	}

	sizeInBytes := uploadedFile.Size
	sizeInKB := float64(sizeInBytes) / 1024.0
	newVersion := existing.Version + 1

	// 4. Update file metadata (only certain fields, keep same ID)
	existing.Version = newVersion
	existing.Rows = len(dataRows)
	existing.Size = sizeInKB

	if err := fs.DB.Save(&existing).Error; err != nil {
		return err
	}

	// 5. Insert into FileVersion table
	fileVersion := FileVersion{
		FileID:     existing.ID,
		Filename:   existing.Filename,
		InsertedBy: userID,
		CreatedAt:  time.Now(),
		Private:    existing.Private,
		Version:    newVersion,
		IsDelete:   false,
		Rows:       len(dataRows),
		Size:       sizeInKB,
	}
	if err := fs.DB.Create(&fileVersion).Error; err != nil {
		return err
	}

	for _, row := range dataRows {
		recordMap := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				recordMap[header] = row[j]
			} else {
				recordMap[header] = ""
			}
		}

		jsonBytes, err := json.Marshal(recordMap)
		if err != nil {
			return err
		}

		record := FileData{
			FileID:     existing.ID,
			RowData:    jsonBytes,
			InsertedBy: userID,
			CreatedAt:  time.Now(),
			Version:    newVersion,
		}
		if err := fs.DB.Create(&record).Error; err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileService) GetUserRole(userID uint) (string, error) {
	var user auth.Auth
	if err := fs.DB.First(&user, userID).Error; err != nil {
		return "", err
	}
	return user.Role, nil
}

func (fs *FileService) GetAllFiles(userID uint, role string) ([]FileWithUser, error) {
	var files []FileWithUser

	if role == "Admin" {
		// Admin → all files with uploader info
		if err := fs.DB.
			Table("file f").
			Select("f.*, u.firstname, u.lastname").
			Joins("LEFT JOIN users u ON u.id = f.inserted_by").
			Scan(&files).Error; err != nil {
			return nil, err
		}
		return files, nil
	}

	// User → public files OR private files they have access to
	err := fs.DB.
		Raw(`
			SELECT f.*, u.firstname, u.lastname
			FROM file f
			LEFT JOIN users u ON u.id = f.inserted_by
			LEFT JOIN file_access fa ON f.id = fa.file_id AND fa.user_id = ?
			WHERE f.private = false OR (fa.user_id = ? AND f.is_delete = ?)
		`, userID, userID, false).
		Scan(&files).Error

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (fs *FileService) GetFileData(filename string, version int) ([]FileData, error) {
	var file File

	// Fetch file by filename
	if err := fs.DB.Where("filename = ? AND is_delete = ?", filename, false).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var fileData []FileData

	if err := fs.DB.Where("file_id = ? AND version = ?", file.ID, version).Find(&fileData).Error; err != nil {
		return nil, err
	}

	return fileData, nil
}

func (fs *FileService) DeleteFile(fileID string) (File, error) {
	// Check if file exists
	var file File
	if err := fs.DB.Where("id = ?", fileID).First(&file).Error; err != nil {
		return file, err
	}

	// Soft delete: just mark is_delete = true
	if err := fs.DB.Model(&file).Update("is_delete", true).Error; err != nil {
		return file, err
	}

	return file, nil
}

func (fs *FileService) ResetFile(fileID string) (File, error) {
	var file File
	if err := fs.DB.Where("id = ?", fileID).First(&file).Error; err != nil {
		return file, err
	}

	// Reset soft delete: mark is_delete = false
	if err := fs.DB.Model(&file).Update("is_delete", false).Error; err != nil {
		return file, err
	}

	return file, nil
}

// parseCSVReader reads CSV file from multipart.File and returns headers + data rows
func parseCSVReader(file multipart.File) ([]string, [][]string, error) {
	defer file.Seek(0, 0) // reset file pointer if needed

	reader := csv.NewReader(file)
	allRows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read csv file: %w", err)
	}

	if len(allRows) < 1 {
		return nil, nil, fmt.Errorf("csv file is empty")
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	return headers, dataRows, nil
}

func (fs *FileService) CreateAccess(input []FileAccess) error {
	if err := fs.DB.Create(&input).Error; err != nil {
		return err
	}
	return nil
}

func (fs *FileService) DeleteAccess(accessId string) error {
	// Check if access record exists
	var access FileAccess
	if err := fs.DB.Where("id = ?", accessId).First(&access).Error; err != nil {
		return err
	}

	// Delete access record
	if err := fs.DB.Delete(&access).Error; err != nil {
		return err
	}

	return nil
}

func (fs *FileService) GetFileAccess(fileId string) ([]FileAccessWithUser, error) {
	var results []FileAccessWithUser

	err := fs.DB.Table("file_access").
		Select("file_access.id, file_access.user_id, file_access.file_id, users.firstname, users.lastname").
		Joins("JOIN users ON users.id = file_access.user_id").
		Where("file_access.file_id = ?", fileId).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (fs *FileService) GetFileHistory(fileId string) ([]FileVersionWithUser, error) {
	var results []FileVersionWithUser

	err := fs.DB.Table("file_version").
		Select(`file_version.id, file_version.file_id, file_version.filename, 
		        users.firstname AS firstname, users.lastname AS lastname,
		        file_version.created_at, file_version.private, file_version.is_delete,
		        file_version.size, file_version.version, file_version.rows`).
		Joins("JOIN users ON users.id = file_version.inserted_by").
		Where("file_version.file_id = ?", fileId).
		Order("file_version.version DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (fs *FileService) RevertFile(filename string, version int, userID uint) error {
	var file File
	if err := fs.DB.Where("filename = ?", filename).First(&file).Error; err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// get target version from file_version
	var targetVersion FileVersion
	if err := fs.DB.Where("file_id = ? AND version = ?", file.ID, version).First(&targetVersion).Error; err != nil {
		return fmt.Errorf("target version not found: %w", err)
	}

	// new version number
	newVersion := file.Version + 1

	// update file table to new version
	if err := fs.DB.Model(&file).Updates(File{
		Version: newVersion,
		Rows:    targetVersion.Rows,
		Size:    targetVersion.Size,
		Private: targetVersion.Private,
	}).Error; err != nil {
		return err
	}

	// insert new row in file_version
	newFileVersion := FileVersion{
		FileID:     file.ID,
		Filename:   filename,
		InsertedBy: userID,
		CreatedAt:  time.Now(),
		Private:    targetVersion.Private,
		Version:    newVersion,
		IsDelete:   false,
		Rows:       targetVersion.Rows,
		Size:       targetVersion.Size,
	}
	if err := fs.DB.Create(&newFileVersion).Error; err != nil {
		return err
	}

	// copy file_data rows of target version into new version
	var dataRows []FileData
	if err := fs.DB.Where("file_id = ? AND version = ?", file.ID, version).Find(&dataRows).Error; err != nil {
		return err
	}

	for _, row := range dataRows {
		newRow := FileData{
			FileID:     file.ID,
			RowData:    row.RowData,
			InsertedBy: userID,
			CreatedAt:  time.Now(),
			Version:    newVersion,
		}
		if err := fs.DB.Create(&newRow).Error; err != nil {
			return err
		}
	}

	return nil
}
