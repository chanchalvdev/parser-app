package search

import (
\t"bytes"
\t"context"
\t"encoding/json"
\t"fmt"

\t"github.com/example/file-platform/backend/internal/models"
\t"github.com/opensearch-project/opensearch-go/v2"
\t"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type Client struct {
\tClient *opensearch.Client
\tIndex  string
}

type Config struct {
\tURL      string
\tUser     string
\tPassword string
\tIndex    string
}

func New(cfg Config) (*Client, error) {
\tcli, err := opensearch.NewClient(opensearch.Config{
\t\tAddresses: []string{cfg.URL},
\t\tUsername:  cfg.User,
\t\tPassword:  cfg.Password,
\t})
\tif err != nil {
\t\treturn nil, err
\t}
\tsearcher := &Client{Client: cli, Index: cfg.Index}
\tif err := searcher.EnsureIndex(context.Background()); err != nil {
\t\treturn nil, err
\t}
\treturn searcher, nil
}

func (c *Client) EnsureIndex(ctx context.Context) error {
\treq := opensearchapi.IndicesExistsRequest{Index: []string{c.Index}}
\tresp, err := req.Do(ctx, c.Client)
\tif err != nil {
\t\treturn err
\t}
\tdefer resp.Body.Close()
\tif resp.StatusCode == 200 {
\t\treturn nil
\t}
\tbody := map[string]any{
\t\t"settings": map[string]any{
\t\t\t"number_of_shards":   1,
\t\t\t"number_of_replicas": 0,
\t\t},
\t\t"mappings": map[string]any{
\t\t\t"properties": map[string]any{
\t\t\t\t"upload_id":     map[string]any{"type": "keyword"},
\t\t\t\t"job_id":        map[string]any{"type": "keyword"},
\t\t\t\t"file_id":       map[string]any{"type": "keyword"},
\t\t\t\t"filename":      map[string]any{"type": "text"},
\t\t\t\t"path":          map[string]any{"type": "text"},
\t\t\t\t"content":       map[string]any{"type": "text"},
\t\t\t\t"source_format": map[string]any{"type": "keyword"},
\t\t\t},
\t\t},
\t}
\tb, _ := json.Marshal(body)
\tcreateReq := opensearchapi.IndicesCreateRequest{
\t\tIndex: c.Index,
\t\tBody:  bytes.NewReader(b),
\t}
\tcreateResp, err := createReq.Do(ctx, c.Client)
\tif err != nil {
\t\treturn err
\t}
\tdefer createResp.Body.Close()
\tif createResp.StatusCode > 299 {
\t\treturn fmt.Errorf("create index failed status=%d", createResp.StatusCode)
\t}
\treturn nil
}

func (c *Client) IndexDocument(ctx context.Context, docID string, doc map[string]any) error {
\tb, err := json.Marshal(doc)
\tif err != nil {
\t\treturn err
\t}
\treq := opensearchapi.IndexRequest{
\t\tIndex:      c.Index,
\t\tDocumentID: docID,
\t\tBody:       bytes.NewReader(b),
\t}
\tresp, err := req.Do(ctx, c.Client)
\tif err != nil {
\t\treturn err
\t}
\tdefer resp.Body.Close()
\tif resp.StatusCode > 299 {
\t\treturn fmt.Errorf("index document failed status=%d", resp.StatusCode)
\t}
\treturn nil
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]models.SearchResult, error) {
\tif limit <= 0 {
\t\tlimit = 20
\t}
\tbody := map[string]any{
\t\t"size": limit,
\t\t"query": map[string]any{
\t\t\t"multi_match": map[string]any{
\t\t\t\t"query":  query,
\t\t\t\t"fields": []string{"filename", "path", "content"},
\t\t\t},
\t\t},
\t}
\tb, _ := json.Marshal(body)
\tsearchReq := opensearchapi.SearchRequest{
\t\tIndex: []string{c.Index},
\t\tBody:  bytes.NewReader(b),
\t}
\tresp, err := searchReq.Do(ctx, c.Client)
\tif err != nil {
\t\treturn nil, err
\t}
\tdefer resp.Body.Close()
\tif resp.StatusCode > 299 {
\t\treturn nil, fmt.Errorf("search failed status=%d", resp.StatusCode)
\t}
\tvar respBody struct {
\t\tHits struct {
\t\t\tHits []struct {
\t\t\t\tID     string          `json:"_id"`
\t\t\t\tScore  float64         `json:"_score"`
\t\t\t\tSource json.RawMessage `json:"_source"`
\t\t\t} `json:"hits"`
\t\t} `json:"hits"`
\t}
\tif err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
\t\treturn nil, err
\t}
\tresults := make([]models.SearchResult, 0, len(respBody.Hits.Hits))
\tfor _, hit := range respBody.Hits.Hits {
\t\tvar source struct {
\t\t\tUploadID string `json:"upload_id"`
\t\t\tJobID    string `json:"job_id"`
\t\t\tFileID   string `json:"file_id"`
\t\t\tFilename string `json:"filename"`
\t\t\tPath     string `json:"path"`
\t\t\tContent  string `json:"content"`
\t\t}
\t\tif err := json.Unmarshal(hit.Source, &source); err != nil {
\t\t\tcontinue
\t\t}
\t\tresults = append(results, models.SearchResult{
\t\t\tIndexID:   hit.ID,
\t\t\tUploadID:  source.UploadID,
\t\t\tJobID:     source.JobID,
\t\t\tFileID:    source.FileID,
\t\t\tFilename:  source.Filename,
\t\t\tPath:      source.Path,
\t\t\tContent:   source.Content,
\t\t\tScore:     hit.Score,
\t\t})
\t}
\treturn results, nil
}
