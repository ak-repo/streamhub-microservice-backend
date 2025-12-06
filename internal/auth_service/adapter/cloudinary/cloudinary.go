package authcloudinary

import (
	"bytes"
	"context"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploader struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryUploader(cloudName, apiKey, apiSecret string) (*CloudinaryUploader, error) {

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &CloudinaryUploader{client: cld}, nil
}

// UploadAvatar uploads the file bytes to the "avatars" folder on Cloudinary.
func (c *CloudinaryUploader) UploadAvatar(ctx context.Context, data []byte, filename string) (string, error) {
	reader := bytes.NewReader(data) // <-- wrap []byte in io.Reader

	resp, err := c.client.Upload.Upload(ctx, reader, uploader.UploadParams{
		PublicID: "avatars/" + filename,
		Folder:   "avatars",
	})
	if err != nil {
		return "", err
	}

	return resp.SecureURL, nil
}
