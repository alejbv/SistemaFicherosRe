package storage

type FileEncoding struct {
	FileName      string   `json:"file_name"`
	FileExtension string   `json:"file_extension"`
	UploadDate    string   `json:"upload_date"`
	Size          int      `json:"size"`
	Tags          []string `json:"tags"`
}
