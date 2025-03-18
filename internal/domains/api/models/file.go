package models

type FileUpload struct {
	BucketName, ObjectName string
	FileSize               int64
	FileBytes              []byte
}
