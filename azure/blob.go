package azure

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"cognitube.com/transcoding-service/env"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type IBlobClient interface {
	UploadBlob(container string, blobName string, data []byte) (url string, err error)
}

type BlobClient struct {
	_client *azblob.Client
}

func (b *BlobClient) UploadBlob(container string, blobName string, data []byte) (string, error) {
	_, err := b._client.UploadBuffer(context.TODO(), container, blobName, data, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s/%s", b._client.URL(), container, blobName), err
}

func (b *BlobClient) ExtractContainerAndBlob(path string) (string, error) {
	// 解析 URL
	u, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	// 分割路径部分
	parts := strings.Split(u.Path, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid path format: %s", u.Path)
	}

	// 提取容器名和 Blob 名
	containerAndBlob := strings.Join(parts[1:], "/")
	return containerAndBlob, nil
}

func (b *BlobClient) DownloadFromBlob(url string) ([]byte, error) {
	containerName, blobName, err := parseBlobURL(url)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading blob: %s/%s", containerName, blobName)
	log.Println(env.GetInstance().BlobConnectString)
	containerClient := b._client.ServiceClient().NewContainerClient(containerName)
	blobClient := containerClient.NewBlobClient(blobName)

	resp, err := blobClient.DownloadStream(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob data: %w", err)
	}

	return data, nil
}

func parseBlobURL(blobURL string) (containerName, blobName string, err error) {
	// 解析 URL
	uri, err := url.Parse(blobURL)
	if err != nil {
		return "", "", fmt.Errorf("error parsing blob URL: %w", err)
	}

	// 获取路径部分
	path := uri.Path

	// 根据 '/' 分割路径
	parts := strings.Split(path, "/")

	// 检查分割后的部分是否足够多以包含容器名和 Blob 名称
	if len(parts) < 3 {
		return "", "", fmt.Errorf("path '%s' is too short, must include container and blob name", path)
	}

	// 容器名通常是第一个有效部分（跳过空字符串）
	containerName = parts[1]

	// Blob 名称是容器名之后的所有部分，合并为完整的 Blob 名称
	blobName = strings.Join(parts[2:], "/")

	return containerName, blobName, nil
}

func NewBlobClient() *BlobClient {
	client, _ := azblob.NewClientFromConnectionString(env.GetInstance().BlobConnectString, nil)

	return &BlobClient{
		_client: client,
	}
}
