package download

import (
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/koralski/run-command-extension-linux/pkg/blobutil"
	"github.com/pkg/errors"
)

const (
	// blobSASDuration describes the duration for which the generated
	// Shared Access Signature for the blob is valid.
	blobSASDuration = time.Minute * 30
)

type blobDownload struct {
	accountName, accountKey string
	blob                    blobutil.AzureBlobRef
}

func (b blobDownload) GetRequest() (*http.Request, error) {
	url, err := b.getURL()
	if err != nil {
		return nil, err
	}
	return http.NewRequest("GET", url, nil)
}

// getURL returns publicly downloadable URL of the Azure Blob
// by generating a URL with a temporary Shared Access Signature.
func (b blobDownload) getURL() (string, error) {
	client, err := storage.NewClient(b.accountName, b.accountKey,
		b.blob.StorageBase, storage.DefaultAPIVersion, true)
	if err != nil {
		return "", errors.Wrap(err, "failed to initialize azure storage client")
	}

	// get read-only
	blobStorageClient := client.GetBlobService()
	container := blobStorageClient.GetContainerReference(b.blob.Container)
	blob := container.GetBlobReference(b.blob.Blob)
	options := storage.BlobSASOptions{
		BlobServiceSASPermissions: storage.BlobServiceSASPermissions{Read: true},
		SASOptions: storage.SASOptions{
			Expiry: time.Now().UTC().Add(blobSASDuration),
		},
	}
	sasURL, err := blob.GetSASURI(options)

	if err != nil {
		return "", errors.Wrap(err, "failed to generate SAS key for blob")
	}
	return sasURL, nil
}

// NewBlobDownload creates a new Downloader for a blob hosted in Azure Blob Storage.
func NewBlobDownload(accountName, accountKey string, blob blobutil.AzureBlobRef) Downloader {
	return blobDownload{accountName, accountKey, blob}
}
