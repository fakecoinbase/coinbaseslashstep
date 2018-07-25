package mocks

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/coinbase/step/utils/to"
)

// S3Client
type GetObjectResponse struct {
	Resp  *s3.GetObjectOutput
	Body  string
	Error error
}

type PutObjectResponse struct {
	Resp  *s3.PutObjectOutput
	Error error
}

type DeleteObjectResponse struct {
	Resp  *s3.DeleteObjectOutput
	Error error
}

type MockS3Client struct {
	s3iface.S3API

	GetObjectResp map[string]*GetObjectResponse

	PutObjectResp map[string]*PutObjectResponse

	DeleteObjectResp map[string]*DeleteObjectResponse
}

func (m *MockS3Client) init() {
	if m.GetObjectResp == nil {
		m.GetObjectResp = map[string]*GetObjectResponse{}
	}

	if m.PutObjectResp == nil {
		m.PutObjectResp = map[string]*PutObjectResponse{}
	}

	if m.DeleteObjectResp == nil {
		m.DeleteObjectResp = map[string]*DeleteObjectResponse{}
	}
}

func MakeS3Body(ret string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(ret))
}

func MakeS3Resp(ret string) *s3.GetObjectOutput {
	return &s3.GetObjectOutput{
		Body:         MakeS3Body(ret),
		LastModified: to.Timep(time.Now()),
	}
}

func AWSS3NotFoundError() error {
	return awserr.New(s3.ErrCodeNoSuchKey, "not found", nil)
}

func (m *MockS3Client) AddGetObject(key string, body string, err error) {
	m.init()
	m.GetObjectResp[key] = &GetObjectResponse{
		Resp:  MakeS3Resp(body),
		Body:  body,
		Error: err,
	}
}

func (m *MockS3Client) AddPutObject(key string, err error) {
	m.init()
	m.PutObjectResp[key] = &PutObjectResponse{
		Resp:  &s3.PutObjectOutput{},
		Error: err,
	}
}

func (m *MockS3Client) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	m.init()
	resp := m.GetObjectResp[*in.Key]

	if resp == nil {
		return nil, AWSS3NotFoundError()
	}

	resp.Resp.Body = MakeS3Body(resp.Body)
	return resp.Resp, resp.Error
}

func (m *MockS3Client) ListObjects(in *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	return nil, nil
}

func (m *MockS3Client) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	m.init()

	resp := m.PutObjectResp[*in.Key]
	// Simulates adding the object
	buf := new(bytes.Buffer)
	buf.ReadFrom(in.Body)
	m.AddGetObject(*in.Key, buf.String(), nil)

	if resp == nil {
		return &s3.PutObjectOutput{}, nil
	}
	return resp.Resp, resp.Error
}

func (m *MockS3Client) DeleteObject(in *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	m.init()

	resp := m.DeleteObjectResp[*in.Key]

	delete(m.GetObjectResp, *in.Key)

	if resp == nil {
		return &s3.DeleteObjectOutput{}, nil
	}
	return resp.Resp, resp.Error
}
