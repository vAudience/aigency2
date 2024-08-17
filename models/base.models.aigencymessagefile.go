package models

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/itsatony/go-filemanager"
	nuts "github.com/vaudience/go-nuts"
)

const (
	IDPREFIX_AGENCYMESSAGEFILE = "amf"
	IDLENGTH_AGENCYMESSAGEFILE = 16
)

// cut version of go-filemanager.ManagedFile
type AIgencyMessageFile struct {
	ID            string         `json:"id"`
	FileName      string         `json:"fileName"`
	MimeType      string         `json:"mimetype"`
	URL           string         `json:"url"`
	LocalFilePath string         `json:"localFilePath"` // TODO: filter out when returning to client
	FileSize      int64          `json:"fileSize"`
	MetaData      map[string]any `json:"metaData"`
} //@name AIgencyMessageFile

func NewAIgencyMessageFile(fileName string, localFilePath string, mimeType string, url string) *AIgencyMessageFile {
	ID := CreateAIgencyMessageFileID()
	entity := AIgencyMessageFile{
		ID:            ID,
		FileName:      fileName,
		LocalFilePath: localFilePath,
		MimeType:      mimeType,
		URL:           url,
	}
	return &entity
}

func CreateAIgencyMessageFileID() string {
	return nuts.NID(IDPREFIX_AGENCYMESSAGEFILE, IDLENGTH_AGENCYMESSAGEFILE)
}

func IsAIgencyMessageFileID(id string) (isModelID bool) {
	isModelID = len(id) == IDLENGTH_AGENCYMESSAGEFILE+len(IDPREFIX_AGENCYMESSAGEFILE)+1 && strings.HasPrefix(id, IDPREFIX_AGENCYMESSAGEFILE)
	return isModelID
}

func (file *AIgencyMessageFile) ReadFileContentFromLocal() (fileContent []byte, err error) {
	// Check if file exists
	if _, err := os.Stat(file.LocalFilePath); os.IsNotExist(err) {
		return nil, err
	}

	// Open file
	f, err := os.Open(file.LocalFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read file content
	fileContent, err = io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}

// ---------------------------------------------------------

type AIgencyMessageFileList struct {
	Files []*AIgencyMessageFile `json:"files"`
	mu    sync.Mutex
} //@name AIgencyMessageFileList

func NewAIgencyMessageFileList() *AIgencyMessageFileList {
	return &AIgencyMessageFileList{
		Files: []*AIgencyMessageFile{},
	}
}

func (fl *AIgencyMessageFileList) AddFileFromManagedFile(managedFile *filemanager.ManagedFile) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	file := NewAIgencyMessageFile(managedFile.FileName, managedFile.LocalFilePath, managedFile.MimeType, managedFile.URL)
	fl.Files = append(fl.Files, file)
}

func (fl *AIgencyMessageFileList) AddFile(file *AIgencyMessageFile) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.Files = append(fl.Files, file)
}

func (fl *AIgencyMessageFileList) GetFiles() []*AIgencyMessageFile {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	return fl.Files
}

func (fl *AIgencyMessageFileList) FindFileByMimeType(mimeType string) []*AIgencyMessageFile {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	var matchedFiles []*AIgencyMessageFile
	for _, file := range fl.Files {
		if file.MimeType == mimeType {
			matchedFiles = append(matchedFiles, file)
		}
	}
	return matchedFiles
}

func (fl *AIgencyMessageFileList) FindFileByMimeTypes(mimeTypes []string) []*AIgencyMessageFile {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	var matchedFiles []*AIgencyMessageFile
	for _, file := range fl.Files {
		if nuts.StringSliceContains(mimeTypes, file.MimeType) {
			matchedFiles = append(matchedFiles, file)
		}
	}
	return matchedFiles
}

func (fl *AIgencyMessageFileList) FindFileByMimeTypePatterns(mimeTypePatterns []string) []*AIgencyMessageFile {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	var matchedFiles []*AIgencyMessageFile
	for _, file := range fl.Files {
		for _, pattern := range mimeTypePatterns {
			if strings.HasPrefix(file.MimeType, pattern) || strings.HasSuffix(file.MimeType, pattern) || strings.Contains(file.MimeType, pattern) {
				matchedFiles = append(matchedFiles, file)
			}
		}
	}
	return matchedFiles
}

func (fl *AIgencyMessageFileList) FindFileByName(fileName string) *AIgencyMessageFile {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	for _, file := range fl.Files {
		if file.FileName == fileName {
			return file
		}
	}
	return nil
}

func (fl *AIgencyMessageFileList) FindFilesByPattern(pattern string) []*AIgencyMessageFile {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	var matchedFiles []*AIgencyMessageFile
	for _, file := range fl.Files {
		if strings.HasPrefix(file.FileName, pattern) || strings.HasSuffix(file.FileName, pattern) || strings.Contains(file.FileName, pattern) {
			matchedFiles = append(matchedFiles, file)
		}
	}
	return matchedFiles
}
