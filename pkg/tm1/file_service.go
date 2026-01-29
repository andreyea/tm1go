package tm1

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// FileService handles operations for server-side files.
type FileService struct {
	rest *RestService
}

// NewFileService creates a new FileService instance.
func NewFileService(rest *RestService) *FileService {
	return &FileService{rest: rest}
}

// GetAllNames retrieves all file and folder names up to the specified depth.
func (fs *FileService) GetAllNames(ctx context.Context, numberOfLevels int) ([]string, error) {
	contentPath := fs.getVersionContentPath()
	if contentPath == "Blobs" {
		numberOfLevels = 0
	}

	endpoint := fmt.Sprintf("/Contents('%s')?$select=ID,Name&$expand=tm1.Folder/Contents", contentPath)

	for i := 0; i < numberOfLevels; i++ {
		endpoint += "($select=ID,Name;$expand=tm1.Folder/Contents"
	}

	for i := 0; i < numberOfLevels; i++ {
		endpoint += ")"
	}

	resp, err := fs.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := Content{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	names := ExtractNamesFromContent(result)
	return names, nil
}

// Get retrieves a file.
// If contentPath is empty, the file will be retrieved from the root folder.
func (fs *FileService) Get(ctx context.Context, fileName string, contentPath []string) ([]byte, error) {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += fmt.Sprintf("/Contents('%s')/Content", url.PathEscape(fileName))

	resp, err := fs.rest.Get(ctx, endpoint, WithHeader("Accept-Encoding", "gzip, deflate"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := io.Reader(resp.Body)
	if encoding := strings.ToLower(resp.Header.Get("Content-Encoding")); strings.Contains(encoding, "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Create creates a new file.
// If contentPath is empty, the file will be created in the root folder.
func (fs *FileService) Create(ctx context.Context, fileName string, contentPath []string, file []byte) error {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += "/Contents"

	body := map[string]interface{}{
		"@odata.type": "#ibm.tm1.api.v1.Document",
		"ID":          fileName,
		"Name":        fileName,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := fs.rest.Post(ctx, endpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return fs.Update(ctx, fileName, contentPath, file)
}

// Update updates a file.
// If contentPath is empty, the file will be updated in the root folder.
func (fs *FileService) Update(ctx context.Context, fileName string, contentPath []string, file []byte) error {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += fmt.Sprintf("/Contents('%s')/Content", url.PathEscape(fileName))

	exists, err := fs.existsByEndpoint(ctx, endpoint)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %s does not exist", fs.buildFilePath(fileName, contentPath))
	}

	resp, err := fs.rest.Patch(ctx, endpoint, bytes.NewReader(file), WithHeader("Content-Type", "application/octet-stream"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// UpdateCompressed compresses file content and updates a file.
// If contentPath is empty, the file will be updated in the root folder.
func (fs *FileService) UpdateCompressed(ctx context.Context, fileName string, contentPath []string, file []byte) error {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += fmt.Sprintf("/Contents('%s')/Content", url.PathEscape(fileName))

	exists, err := fs.existsByEndpoint(ctx, endpoint)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %s does not exist", fs.buildFilePath(fileName, contentPath))
	}

	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)
	if _, err := gz.Write(file); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	resp, err := fs.rest.Patch(
		ctx,
		endpoint,
		bytes.NewReader(compressedData.Bytes()),
		WithHeader("Content-Type", "application/octet-stream"),
		WithHeader("Content-Encoding", "gzip"),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// UpdateOrCreate updates a file if it exists, otherwise creates it.
// If contentPath is empty, the file will be updated in the root folder.
func (fs *FileService) UpdateOrCreate(ctx context.Context, fileName string, contentPath []string, file []byte) error {
	exists, err := fs.Exists(ctx, fileName, contentPath)
	if err != nil {
		return err
	}

	if exists {
		return fs.Update(ctx, fileName, contentPath, file)
	}

	return fs.Create(ctx, fileName, contentPath, file)
}

// Exists checks if a file exists.
// If contentPath is empty, the file will be checked in the root folder.
func (fs *FileService) Exists(ctx context.Context, fileName string, contentPath []string) (bool, error) {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += fmt.Sprintf("/Contents('%s')", url.PathEscape(fileName))

	resp, err := fs.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// Delete deletes a file.
func (fs *FileService) Delete(ctx context.Context, fileName string, contentPath []string) error {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += fmt.Sprintf("/Contents('%s')", url.PathEscape(fileName))

	resp, err := fs.rest.Delete(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// CreateCompressed compresses file content and creates a new file entity.
func (fs *FileService) CreateCompressed(ctx context.Context, fileName string, contentPath []string, file []byte) error {
	endpoint := fs.buildContentsEndpoint(contentPath)
	endpoint += "/Contents"

	body := map[string]interface{}{
		"@odata.type": "#ibm.tm1.api.v1.Document",
		"ID":          fileName,
		"Name":        fileName,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := fs.rest.Post(ctx, endpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return fs.UpdateCompressed(ctx, fileName, contentPath, file)
}

func (fs *FileService) buildContentsEndpoint(contentPath []string) string {
	base := fs.getVersionContentPath()
	endpoint := fmt.Sprintf("/Contents('%s')", base)

	for _, path := range contentPath {
		endpoint += fmt.Sprintf("/Contents('%s')", url.PathEscape(path))
	}

	return endpoint
}

func (fs *FileService) buildFilePath(fileName string, contentPath []string) string {
	if len(contentPath) == 0 {
		return fileName
	}

	return strings.Join(append(contentPath, fileName), "/")
}

func (fs *FileService) existsByEndpoint(ctx context.Context, endpoint string) (bool, error) {
	resp, err := fs.rest.Get(ctx, endpoint)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

// getVersionContentPath returns the REST content root name for the current TM1 version.
func (fs *FileService) getVersionContentPath() string {
	version := strings.TrimSpace(fs.rest.version)
	if version == "" {
		return "Blobs"
	}

	if IsV1GreaterOrEqualToV2(version, "12.0.0") {
		return "Files"
	}

	return "Blobs"
}
