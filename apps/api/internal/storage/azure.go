package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

// sasClockSkew backdates the SAS start time so a browser or server whose clock runs
// slightly ahead of Azure's is not rejected with "signature not yet valid".
const sasClockSkew = 5 * time.Minute

type azureStore struct {
	container      *container.Client
	credential     *azblob.SharedKeyCredential
	containerName  string
	publicEndpoint string
	sasProtocol    sas.Protocol
}

func newAzureStore(cfg Config) (*azureStore, error) {
	account := strings.TrimSpace(cfg.AzureAccountName)
	key := strings.TrimSpace(cfg.AzureAccountKey)
	if account == "" || key == "" {
		return nil, errors.New("AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY are required when STORAGE_BACKEND=azure")
	}
	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return nil, fmt.Errorf("azure shared key credential: %w", err)
	}
	endpoint := strings.TrimSuffix(strings.TrimSpace(cfg.AzureEndpoint), "/")
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", account)
	}
	client, err := container.NewClientWithSharedKeyCredential(endpoint+"/"+cfg.Bucket, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure container client: %w", err)
	}
	// Real Azure endpoints are HTTPS, so the SAS refuses plain HTTP; only the
	// Azurite emulator (http://) needs HTTP allowed.
	protocol := sas.ProtocolHTTPS
	if strings.HasPrefix(endpoint, "http://") {
		protocol = sas.ProtocolHTTPSandHTTP
	}
	return &azureStore{
		container:      client,
		credential:     cred,
		containerName:  cfg.Bucket,
		publicEndpoint: cfg.AzurePublicEndpoint,
		sasProtocol:    protocol,
	}, nil
}

func (s *azureStore) Backend() string { return BackendAzure }
func (s *azureStore) Bucket() string  { return s.containerName }

func (s *azureStore) EnsureBucket(ctx context.Context) error {
	_, err := s.container.Create(ctx, nil)
	if err == nil || bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return nil
	}
	return fmt.Errorf("create blob container: %w", err)
}

// PresignPut returns a service SAS scoped to this one blob, allowing only create+write,
// valid for `expiry`. The account key never leaves the API.
//
// The container and blob names are signed explicitly rather than via blob.Client.GetSASURL,
// which re-parses the URL and only recognises path-style endpoints (Azurite's
// http://host:port/<account>/...) when the host is an IP address.
func (s *azureStore) PresignPut(_ context.Context, objectKey string, expiry time.Duration) (*PresignedUpload, error) {
	now := time.Now().UTC()
	query, err := sas.BlobSignatureValues{
		Protocol:      s.sasProtocol,
		StartTime:     now.Add(-sasClockSkew),
		ExpiryTime:    now.Add(expiry),
		Permissions:   (&sas.BlobPermissions{Create: true, Write: true}).String(),
		ContainerName: s.containerName,
		BlobName:      objectKey,
	}.SignWithSharedKey(s.credential)
	if err != nil {
		return nil, fmt.Errorf("sign blob SAS: %w", err)
	}
	sasURL := s.container.NewBlobClient(objectKey).URL() + "?" + query.Encode()
	publicURL, err := RewriteURLHost(sasURL, s.publicEndpoint)
	if err != nil {
		return nil, fmt.Errorf("rewrite blob SAS url: %w", err)
	}
	return &PresignedUpload{
		URL: publicURL,
		// Put Blob creates a block blob only when told to; without this header Azure returns 400.
		Headers: map[string]string{"x-ms-blob-type": "BlockBlob"},
	}, nil
}

func (s *azureStore) Stat(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	props, err := s.container.NewBlobClient(objectKey).GetProperties(ctx, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrObjectNotFound, objectKey)
		}
		return nil, err
	}
	var size int64
	if props.ContentLength != nil {
		size = *props.ContentLength
	}
	return &ObjectInfo{Size: size}, nil
}
