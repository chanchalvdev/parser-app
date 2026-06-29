package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultParsedRecordsIndex = "parsed-records"
	defaultFilesIndex        = "files"
)

type FacetBucket struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type ParsedHit struct {
	ID        string
	Source    map[string]any
	Score     float64
	Highlight map[string][]string
}

type ParsedSearchResponse struct {
	Total        int64
	Hits         []ParsedHit
	Aggregations map[string][]FacetBucket
}

type FileMetadata struct {
	FileID           string
	Extension        string
	DetectedFileType string
}

type SearchRequest struct {
	Query      string
	TenantID   string
	FileID     string
	JobID      string
	RecordType string
	StartDate  string
	EndDate    string
	IP         string
	Email      string
	Domain     string
	FileIDs    []string
	Page       int
	PageSize   int
	Sort       string
}

type FileMetadataFilter struct {
	TenantID         string
	FileID           string
	Extension        string
	DetectedFileType string
}

type ClientConfig struct {
	URL                string
	ParsedRecordsIndex string
	FilesIndex         string
	Username           string
	Password           string
}

type Client struct {
	baseURL           string
	parsedRecordsIndex string
	filesIndex        string
	username          string
	password          string
	httpClient        *http.Client
}

func NewClient(cfg ClientConfig) *Client {
	parsedIndex := strings.TrimSpace(cfg.ParsedRecordsIndex)
	if parsedIndex == "" {
		parsedIndex = defaultParsedRecordsIndex
	}
	filesIndex := strings.TrimSpace(cfg.FilesIndex)
	if filesIndex == "" {
		filesIndex = defaultFilesIndex
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	if baseURL == "" {
		baseURL = "http://opensearch:9200"
	}

	return &Client{
		baseURL:           baseURL,
		parsedRecordsIndex: parsedIndex,
		filesIndex:        filesIndex,
		username:          strings.TrimSpace(cfg.Username),
		password:          cfg.Password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SearchParsedRecords(ctx context.Context, req SearchRequest) (*ParsedSearchResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize < 0 {
		req.PageSize = 0
	}

	query := c.buildParsedQuery(req)

	requestBody := map[string]any{
		"query": query,
		"track_total_hits": true,
		"size":           req.PageSize,
		"from":           (req.Page - 1) * req.PageSize,
	}

	if req.Sort == "created_at" {
		requestBody["sort"] = []any{
			map[string]any{"created_at": "desc"},
		}
	} else {
		requestBody["sort"] = []any{
			map[string]any{"_score": "desc"},
		}
	}

	requestBody["highlight"] = map[string]any{
		"fields": map[string]any{
			"content": map[string]any{
				"number_of_fragments": 1,
				"fragment_size":      220,
			},
		},
	}

	requestBody["aggs"] = map[string]any{
		"record_type": map[string]any{
			"terms": map[string]any{
				"field": "record_type",
				"size":  25,
			},
		},
		"file_id": map[string]any{
			"terms": map[string]any{
				"field": "file_id",
				"size":  2000,
			},
		},
		"entities_ip_addresses": map[string]any{
			"terms": map[string]any{
				"field": "entities.ip_addresses",
				"size":  25,
			},
		},
		"entities_emails": map[string]any{
			"terms": map[string]any{
				"field": "entities.emails",
				"size":  25,
			},
		},
		"entities_domains": map[string]any{
			"terms": map[string]any{
				"field": "entities.domains",
				"size":  25,
			},
		},
		"entities_urls": map[string]any{
			"terms": map[string]any{
				"field": "entities.urls",
				"size":  25,
			},
		},
		"entities_hashes": map[string]any{
			"terms": map[string]any{
				"field": "entities.hashes",
				"size":  25,
			},
		},
	}

	payload, err := c.doRequest(ctx, c.endpoint("/"+c.parsedRecordsIndex+"/_search"), requestBody)
	if err != nil {
		return nil, err
	}

	var response struct {
		Hits struct {
			Total        json.RawMessage              `json:"total"`
			Hits         []map[string]json.RawMessage `json:"hits"`
		} `json:"hits"`
		Aggregations map[string]json.RawMessage `json:"aggregations"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, err
	}

	parsed := &ParsedSearchResponse{
		Aggregations: map[string][]FacetBucket{},
	}

	parsed.Total, err = parseTotalHits(response.Hits.Total)
	if err != nil {
		return nil, err
	}

	for _, rawHit := range response.Hits.Hits {
		var hit struct {
			ID        string                 `json:"_id"`
			Score     float64                `json:"_score"`
			Source    map[string]any         `json:"_source"`
			Highlight map[string][]string     `json:"highlight"`
		}
		if err := unmarshalRawMap(rawHit, &hit); err != nil {
			continue
		}
		parsed.Hits = append(parsed.Hits, ParsedHit{
			ID:        hit.ID,
			Source:    hit.Source,
			Score:     hit.Score,
			Highlight: hit.Highlight,
		})
	}

	for field, rawAgg := range response.Aggregations {
		var parsedAgg struct {
			Buckets []struct {
				Key   any   `json:"key"`
				Count int64 `json:"doc_count"`
			} `json:"buckets"`
		}
		if err := json.Unmarshal(rawAgg, &parsedAgg); err != nil {
			continue
		}
		buckets := make([]FacetBucket, 0, len(parsedAgg.Buckets))
		for _, bucket := range parsedAgg.Buckets {
			buckets = append(buckets, FacetBucket{
				Value: strings.TrimSpace(fmt.Sprint(bucket.Key)),
				Count: bucket.Count,
			})
		}
		parsed.Aggregations[field] = buckets
	}

	return parsed, nil
}

func (c *Client) ResolveFileMetadata(ctx context.Context, filter FileMetadataFilter) ([]FileMetadata, error) {
	query := c.buildFileMetadataQuery(filter)
	if query == nil {
		return []FileMetadata{}, nil
	}

	requestBody := map[string]any{
		"query": query,
		"size":  10000,
		"_source": []string{"file_id", "extension", "detected_file_type"},
	}

	payload, err := c.doRequest(ctx, c.endpoint("/"+c.filesIndex+"/_search"), requestBody)
	if err != nil {
		return nil, err
	}

	var response struct {
		Hits struct {
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, err
	}

	files := make([]FileMetadata, 0, len(response.Hits.Hits))
	seen := map[string]struct{}{}
	for _, hit := range response.Hits.Hits {
		fileID := toString(hit.Source["file_id"])
		if fileID == "" {
			continue
		}
		if _, ok := seen[fileID]; ok {
			continue
		}
		seen[fileID] = struct{}{}

		files = append(files, FileMetadata{
			FileID:           fileID,
			Extension:        toString(hit.Source["extension"]),
			DetectedFileType: toString(hit.Source["detected_file_type"]),
		})
	}

	return files, nil
}

func (c *Client) AggregateDetectedFileTypeByFileIDs(ctx context.Context, tenantID string, fileIDs []string) ([]FacetBucket, error) {
	fileIDs = dedupeStrings(fileIDs)
	if len(fileIDs) == 0 {
		return []FacetBucket{}, nil
	}
	if len(fileIDs) > 2000 {
		fileIDs = fileIDs[:2000]
	}

	filters := []any{}
	if tenantID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"tenant_id": tenantID}})
	}
	filters = append(filters, map[string]any{"terms": map[string]any{"file_id": fileIDs}})

	requestBody := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{"filter": filters},
		},
		"aggs": map[string]any{
			"detected_file_type": map[string]any{
				"terms": map[string]any{
					"field": "detected_file_type",
					"size":  25,
				},
			},
		},
	}

	payload, err := c.doRequest(ctx, c.endpoint("/"+c.filesIndex+"/_search"), requestBody)
	if err != nil {
		return nil, err
	}

	var response struct {
		Aggregations map[string]json.RawMessage `json:"aggregations"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, err
	}

	rawAgg, ok := response.Aggregations["detected_file_type"]
	if !ok {
		return []FacetBucket{}, nil
	}

	var parsedAgg struct {
		Buckets []struct {
			Key   any   `json:"key"`
			Count int64 `json:"doc_count"`
		} `json:"buckets"`
	}
	if err := json.Unmarshal(rawAgg, &parsedAgg); err != nil {
		return nil, err
	}

	buckets := make([]FacetBucket, 0, len(parsedAgg.Buckets))
	for _, bucket := range parsedAgg.Buckets {
		value := strings.TrimSpace(fmt.Sprint(bucket.Key))
		if value == "" {
			continue
		}
		buckets = append(buckets, FacetBucket{Value: value, Count: bucket.Count})
	}

	return buckets, nil
}

func (c *Client) doRequest(ctx context.Context, fullPath string, body any) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullPath, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(responseBody) == 0 {
			return nil, fmt.Errorf("opensearch request failed: status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("opensearch request failed: status %d, body %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return responseBody, nil
}

func (c *Client) endpoint(path string) string {
	return c.baseURL + path
}

func (c *Client) buildParsedQuery(req SearchRequest) map[string]any {
	filters := []any{}
	if strings.TrimSpace(req.TenantID) != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"tenant_id": strings.TrimSpace(req.TenantID)}})
	}
	if req.FileID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"file_id": strings.TrimSpace(req.FileID)}})
	}
	if len(req.FileIDs) > 0 {
		filters = append(filters, map[string]any{"terms": map[string]any{"file_id": dedupeStrings(req.FileIDs)}})
	}
	if req.JobID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"job_id": strings.TrimSpace(req.JobID)}})
	}
	if req.RecordType != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"record_type": strings.TrimSpace(req.RecordType)}})
	}
	if req.IP != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"entities.ip_addresses": strings.TrimSpace(req.IP)}})
	}
	if req.Email != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"entities.emails": strings.TrimSpace(req.Email)}})
	}
	if req.Domain != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"entities.domains": strings.TrimSpace(req.Domain)}})
	}
	if req.StartDate != "" || req.EndDate != "" {
		rangeFilter := map[string]any{}
		if req.StartDate != "" {
			rangeFilter["gte"] = strings.TrimSpace(req.StartDate)
		}
		if req.EndDate != "" {
			rangeFilter["lte"] = strings.TrimSpace(req.EndDate)
		}
		filters = append(filters, map[string]any{"range": map[string]any{"created_at": rangeFilter}})
	}

	boolQuery := map[string]any{}
	if req.Query != "" {
		boolQuery["must"] = []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":  strings.TrimSpace(req.Query),
					"fields": []string{"content"},
				},
			},
		}
	}
	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}
	if len(boolQuery) == 0 {
		return map[string]any{"match_all": map[string]any{}}
	}

	return map[string]any{"bool": boolQuery}
}

func (c *Client) buildFileMetadataQuery(filter FileMetadataFilter) map[string]any {
	tenantID := strings.TrimSpace(filter.TenantID)
	extension := strings.TrimSpace(filter.Extension)
	detectedFileType := strings.TrimSpace(filter.DetectedFileType)
	fileID := strings.TrimSpace(filter.FileID)

	filters := []any{}
	if tenantID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"tenant_id": tenantID}})
	}
	if fileID != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"file_id": fileID}})
	}
	if extension != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"extension": extension}})
	}
	if detectedFileType != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"detected_file_type": detectedFileType}})
	}
	if len(filters) == 0 {
		return nil
	}
	return map[string]any{"bool": map[string]any{"filter": filters}}
}

func parseTotalHits(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 {
		return 0, nil
	}

	var objectValue struct {
		Value int64 `json:"value"`
	}
	if err := json.Unmarshal(raw, &objectValue); err == nil {
		return objectValue.Value, nil
	}

	var simpleInt int64
	if err := json.Unmarshal(raw, &simpleInt); err == nil {
		return simpleInt, nil
	}

	var simpleFloat float64
	if err := json.Unmarshal(raw, &simpleFloat); err == nil {
		return int64(simpleFloat), nil
	}

	return 0, fmt.Errorf("unsupported total format")
}

func unmarshalRawMap(raw map[string]json.RawMessage, target any) error {
	payload, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}
