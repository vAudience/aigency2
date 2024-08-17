package models

type AdapterFileInfo struct {
	Description string `json:"description"`
	FileName    string `json:"fileName"`
	MimeType    string `json:"mimeType"`
	LocalPath   string `json:"localPath"`
	PublicUrl   string `json:"publicUrl"`
}

func NewAdapterFileInfo(description string, fileName string, mimeType string, localPath string, publicUrl string) AdapterFileInfo {
	return AdapterFileInfo{
		Description: description,
		FileName:    fileName,
		MimeType:    mimeType,
		LocalPath:   localPath,
		PublicUrl:   publicUrl,
	}
}
