// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat for the canonical source repository
// @license     https://github.com/chanxuehong/wechat/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chanxuehong/wechat/media"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// 上传多媒体文件, MediaUpload 的一个简单封装
func (c *Client) MediaUploadFromFile(mediaType, filePath string) (*media.UploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return c.MediaUpload(mediaType, filepath.Base(filePath), file)
}

// 上传多媒体文件.
//  NOTE:
//  1. 媒体文件在后台保存时间为3天，即3天后 media_id 失效。
//  2. 返回的 media_id 是可复用的;
//  3. 图片（image）: 1M，支持JPG格式
//  4. 语音（voice）：2M，播放长度不超过60s，支持AMR\MP3格式
//  5. 视频（video）：10MB，支持MP4格式
//  6. 缩略图（thumb）：64KB，支持JPG格式
func (c *Client) MediaUpload(mediaType, filename string, mediaReader io.Reader) (*media.UploadResponse, error) {
	switch mediaType {
	case media.MEDIA_TYPE_IMAGE,
		media.MEDIA_TYPE_VOICE,
		media.MEDIA_TYPE_VIDEO,
		media.MEDIA_TYPE_THUMB:
	default:
		return nil, errors.New("错误的 mediaType")
	}
	if filename == "" {
		return nil, errors.New(`filename == ""`)
	}
	if mediaReader == nil {
		return nil, errors.New("mediaReader == nil")
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}

	bodyBuf := c.getBufferFromPool() // io.ReadWriter
	defer c.putBufferToPool(bodyBuf) // important!

	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fileWriter, mediaReader); err != nil {
		return nil, err
	}

	bodyContentType := bodyWriter.FormDataContentType()

	if err = bodyWriter.Close(); err != nil {
		return nil, err
	}

	_url := mediaUploadURL(token, mediaType)
	resp, err := c.httpClient.Post(_url, bodyContentType, bodyBuf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http.Status: %s", resp.Status)
	}

	switch mediaType {
	case media.MEDIA_TYPE_THUMB: // 返回的是 thumb_media_id 而不是 media_id
		var result struct {
			MediaType string `json:"type"`
			MediaId   string `json:"thumb_media_id"`
			CreatedAt int64  `json:"created_at"`
			Error
		}
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		if result.ErrCode != 0 {
			return nil, &result.Error
		}

		var resp media.UploadResponse
		resp.MediaType = result.MediaType
		resp.MediaId = result.MediaId
		resp.CreatedAt = result.CreatedAt
		return &resp, nil

	default:
		var result struct {
			media.UploadResponse
			Error
		}
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		if result.ErrCode != 0 {
			return nil, &result.Error
		}
		return &result.UploadResponse, nil
	}
}

// 下载多媒体文件.
//  NOTE: 视频文件不支持下载.
func (c *Client) MediaDownloadToFile(mediaId, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return c.MediaDownload(mediaId, file)
}

// 下载多媒体文件.
//  NOTE: 视频文件不支持下载.
func (c *Client) MediaDownload(mediaId string, writer io.Writer) error {
	if mediaId == "" {
		return errors.New(`mediaId == ""`)
	}
	if writer == nil {
		return errors.New("writer == nil")
	}

	token, err := c.Token()
	if err != nil {
		return err
	}

	_url := mediaDownloadURL(token, mediaId)
	resp, err := c.httpClient.Get(_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", resp.Status)
	}

	contentType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if contentType != "text/plain" && contentType != "application/json" {
		_, err = io.Copy(writer, resp.Body)
		return err
	}

	// 返回的是错误信息

	var result Error
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	return &result
}

// 上传图文消息素材
func (c *Client) MediaUploadNews(news *media.News) (*media.UploadResponse, error) {
	if news == nil {
		return nil, errors.New("news == nil")
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := mediaUploadNewsURL(token)

	var result struct {
		media.UploadResponse
		Error
	}
	if err = c.postJSON(_url, news, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}
	return &result.UploadResponse, nil
}

// 上传视频消息
func (c *Client) MediaUploadVideo(video *media.Video) (*media.UploadResponse, error) {
	if video == nil {
		return nil, errors.New("video == nil")
	}

	token, err := c.Token()
	if err != nil {
		return nil, err
	}
	_url := mediaUploadVideoURL(token)

	var result struct {
		media.UploadResponse
		Error
	}
	if err = c.postJSON(_url, video, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, &result.Error
	}
	return &result.UploadResponse, nil
}