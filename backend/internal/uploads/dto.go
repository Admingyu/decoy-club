package uploads

import "time"

type UploadedFileView struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	PublicURL string    `json:"public_url"`
	CreatedAt time.Time `json:"created_at"`
}

type UploadImageResponse struct {
	File *UploadedFileView `json:"file"`
}

func NewUploadedFileView(file *UploadedFile) *UploadedFileView {
	if file == nil {
		return nil
	}
	return &UploadedFileView{
		ID:        file.ID.Hex(),
		FileName:  file.FileName,
		MimeType:  file.MimeType,
		Size:      file.Size,
		PublicURL: file.PublicURL,
		CreatedAt: file.CreatedAt,
	}
}
