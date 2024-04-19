package tm1go

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"strings"
)

type FileService struct {
	rest   *RestService
	object *ObjectService
}

func NewFileService(rest *RestService, object *ObjectService) *FileService {
	return &FileService{rest: rest, object: object}
}

// Get Version Content Path
func (fs *FileService) getVersionContentPath() string {
	if isV1GreaterOrEqualToV2(fs.rest.version, "12.0.0") {
		return "Files"
	}
	return "Blobs"
}

// GetAllNames retrieves all names
func (fs *FileService) GetAllNames(numberOfLevels int) ([]string, error) {
	contentPath := fs.getVersionContentPath()
	if contentPath == "Blobs" {
		numberOfLevels = 0
	}

	url := "/Contents('" + contentPath + "')?$select=ID,Name&$expand=tm1.Folder/Contents"

	// Add additional contents query parameters for each level
	for i := 0; i < numberOfLevels; i++ {
		url += "($select=ID,Name;$expand=tm1.Folder/Contents"
	}

	// Close parenthesis of contents query
	for i := 0; i < numberOfLevels; i++ {
		url += ")"
	}

	response, err := fs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := Content{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	names := ExtractNamesFromContent(result)
	return names, nil
}

// Get retrieves a file
// If contentPath is empty, the file will be retrieved from the root folder
func (fs *FileService) Get(fileName string, contentPath []string) ([]byte, error) {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents('" + fileName + "')/Content"

	response, err := fs.rest.GET(url, nil, 0, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	result := []byte{}
	_, err = response.Body.Read(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create a new file
// If contentPath is empty, the file will be created in the root folder
func (fs *FileService) Create(fileName string, contentPath []string, file []byte) error {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents"

	body := map[string]interface{}{
		"@odata.type": "#ibm.tm1.api.v1.Document",
		"ID":          fileName,
		"Name":        fileName,
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = fs.rest.POST(url, string(bodyJson), nil, 0, nil)
	if err != nil {
		return err
	}

	return fs.Update(fileName, contentPath, file)
}

// Update a file
// If contentPath is empty, the file will be updated in the root folder
func (fs *FileService) Update(fileName string, contentPath []string, file []byte) error {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents('" + fileName + "')/Content"

	// check if exists
	exists, err := fs.object.Exists(url)
	if err != nil {
		return err
	}
	if !exists {
		filePath := strings.Join(contentPath, "/")
		filePath += "/" + fileName
		return fmt.Errorf("file %s does not exist", filePath)
	}

	headers := map[string]string{
		"Content-Type": "application/octet-stream",
	}

	_, err = fs.rest.PATCH(url, string(file), headers, 0, nil)

	return err
}

// UpdateCompressed compresses file content and updates a file
// If contentPath is empty, the file will be updated in the root folder
func (fs *FileService) UpdateCompressed(fileName string, contentPath []string, file []byte) error {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents('" + fileName + "')/Content"

	// check if exists
	exists, err := fs.object.Exists(url)
	if err != nil {
		return err
	}
	if !exists {
		filePath := strings.Join(contentPath, "/")
		filePath += "/" + fileName
		return fmt.Errorf("file %s does not exist", filePath)
	}

	// Compress the file data using gzip
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)
	if _, err := gz.Write(file); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	headers := map[string]string{
		"Content-Type":     "application/octet-stream",
		"Content-Encoding": "gzip",
	}

	_, err = fs.rest.PATCH(url, compressedData.String(), headers, 0, nil)

	return err
}

// UpdateOrCreate updates a file if it exists, otherwise creates it
// If contentPath is empty, the file will be updated in the root folder
func (fs *FileService) UpdateOrCreate(fileName string, contentPath []string, file []byte) error {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents('" + fileName + "')/Content"

	// check if exists
	exists, err := fs.object.Exists(url)
	if err != nil {
		return err
	}

	if exists {
		return fs.Update(fileName, contentPath, file)
	}

	return fs.Create(fileName, contentPath, file)
}

// Exists checks if a file exists
// If contentPath is empty, the file will be checked in the root folder
func (fs *FileService) Exists(fileName string, contentPath []string) (bool, error) {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents('" + fileName + "')"

	exists, err := fs.object.Exists(url)

	return exists, err
}

// Delete a file
func (fs *FileService) Delete(fileName string, contentPath []string) error {

	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents('" + fileName + "')"

	_, err := fs.rest.DELETE(url, nil, 0, nil)

	return err
}

// Create compresses file content and creates a new file entity
func (fs *FileService) CreateCompressed(fileName string, contentPath []string, file []byte) error {
	url := "/Contents('" + fs.getVersionContentPath() + "')"

	for _, path := range contentPath {
		url += "/Contents('" + path + "')"
	}

	url += "/Contents"

	body := map[string]interface{}{
		"@odata.type": "#ibm.tm1.api.v1.Document",
		"ID":          fileName,
		"Name":        fileName,
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = fs.rest.POST(url, string(bodyJson), nil, 0, nil)
	if err != nil {
		return err
	}

	url += "('" + fileName + "')/Content"

	return fs.UpdateCompressed(fileName, contentPath, file)

}
