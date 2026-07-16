package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/enterprise-file-platform/api/internal/models"
	"github.com/enterprise-file-platform/api/internal/repositories"
	"github.com/enterprise-file-platform/api/internal/search"
	"github.com/enterprise-file-platform/api/internal/http/middleware"
)

const (
	defaultSearchPageSize  = 25
	maxSearchPageSize      = 100
	searchDefaultSort      = "relevance"
	searchSuggestionLimit  = 10
	searchPreviewMaxLen    = 420
	searchDefaultTenantID  = "00000000-0000-0000-0000-000000000001"
)

var (
	ErrInvalidSearchRequest   = errors.New("invalid search request")
	ErrSearchServiceUnavailable = errors.New("search service unavailable")
)

type SearchService struct {
	searchClient *search.Client
	auditLogRepo repositories.AuditLogRepository
}

type SearchRecordsRequest struct {
	Q                string `json:"q"`
	FileID           string `json:"file_id"`
	Extension        string `json:"extension"`
	DetectedFileType string `json:"detected_file_type"`
	RecordType       string `json:"record_type"`
	DateFrom         string `json:"date_from"`
	DateTo           string `json:"date_to"`
	IP               string `json:"ip"`
	Email            string `json:"email"`
	Domain           string `json:"domain"`
	JobID            string `json:"job_id"`
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
	Sort             string `json:"sort"`
	TenantID         string `json:"tenant_id"`
}

type SearchResponse struct {
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Results  []SearchResult `json:"results"`
	Facets   SearchFacets  `json:"facets"`
}

type SearchResult struct {
	RecordID       string         `json:"record_id"`
	FileID         string         `json:"file_id"`
	JobID          string         `json:"job_id"`
	SourceFileName string         `json:"source_file_name"`
	RecordType     string         `json:"record_type"`
	ContentPreview string         `json:"content_preview"`
	Highlight      string         `json:"highlight"`
	Entities       map[string]any `json:"entities"`
	CreatedAt      time.Time      `json:"created_at"`
}

type SearchFacets struct {
	RecordType       []SearchFacetItem                `json:"record_type"`
	DetectedFileType []SearchFacetItem                `json:"detected_file_type"`
	Entities         map[string][]SearchFacetItem `json:"entities"`
}

type SearchFacetItem struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type SearchSuggestionResponse struct {
	Suggestions SearchSuggestions `json:"suggestions"`
}

type SearchSuggestions struct {
	RecordType       []SearchFacetItem                `json:"record_type"`
	DetectedFileType []SearchFacetItem                `json:"detected_file_type"`
	Entities         map[string][]SearchFacetItem `json:"entities"`
}

func NewSearchService(client *search.Client, auditLogRepo repositories.AuditLogRepository) *SearchService {
	return &SearchService{searchClient: client, auditLogRepo: auditLogRepo}
}

func (s *SearchService) Search(ctx context.Context, req SearchRecordsRequest) (*SearchResponse, error) {
	if s.searchClient == nil {
		return nil, fmt.Errorf("%w: search client is not initialized", ErrSearchServiceUnavailable)
	}

	parsedReq, err := normalizeSearchRequest(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSearchRequest, err)
	}

	tenantID := strings.TrimSpace(parsedReq.TenantID)

	fileIDsForFilter := []string{}
	if parsedReq.Extension != "" || parsedReq.DetectedFileType != "" {
		files, err := s.searchClient.ResolveFileMetadata(ctx, search.FileMetadataFilter{
			TenantID:         tenantID,
			FileID:           parsedReq.FileID,
			Extension:        parsedReq.Extension,
			DetectedFileType: parsedReq.DetectedFileType,
		})
		if err != nil {
			return nil, fmt.Errorf("resolve file metadata: %w", err)
		}
		for _, file := range files {
			if file.FileID != "" {
				fileIDsForFilter = append(fileIDsForFilter, file.FileID)
			}
		}
		if len(fileIDsForFilter) == 0 {
			return &SearchResponse{
				Total:    0,
				Page:     parsedReq.Page,
				PageSize: parsedReq.PageSize,
				Results:  []SearchResult{},
				Facets: SearchFacets{
					RecordType:       []SearchFacetItem{},
					DetectedFileType: []SearchFacetItem{},
					Entities: map[string][]SearchFacetItem{
						"ip_addresses": {},
						"emails":       {},
						"domains":      {},
						"urls":         {},
						"hashes":       {},
					},
				},
			}, nil
		}
	}

	searchQuery := search.SearchRequest{
		Query:      parsedReq.Q,
		TenantID:   tenantID,
		FileID:     parsedReq.FileID,
		JobID:      parsedReq.JobID,
		RecordType: parsedReq.RecordType,
		StartDate:  parsedReq.DateFrom,
		EndDate:    parsedReq.DateTo,
		IP:         parsedReq.IP,
		Email:      parsedReq.Email,
		Domain:     parsedReq.Domain,
		FileIDs:    fileIDsForFilter,
		Page:       parsedReq.Page,
		PageSize:   parsedReq.PageSize,
		Sort:       parsedReq.Sort,
	}
	searchResp, err := s.searchClient.SearchParsedRecords(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search parsed records: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Hits))
	for _, hit := range searchResp.Hits {
		source := hit.Source
		if source == nil {
			source = map[string]any{}
		}
		content := toString(source["content"])
		record := SearchResult{
			RecordID:       hit.ID,
			FileID:         toString(source["file_id"]),
			JobID:          toString(source["job_id"]),
			SourceFileName: toString(source["source_file_name"]),
			RecordType:     toString(source["record_type"]),
			ContentPreview: preview(content, searchPreviewMaxLen),
			Entities:       toEntitiesMap(source["entities"]),
			CreatedAt:      parseTime(source["created_at"]),
		}
		if hit.Highlight != nil {
			if values, ok := hit.Highlight["content"]; ok {
				if len(values) > 0 {
					record.Highlight = values[0]
				}
			}
		}
		results = append(results, record)
	}

	recordTypeBuckets := bucketsToFacetItems(searchResp.Aggregations["record_type"])
	fileIDBuckets := searchResp.Aggregations["file_id"]
	detectedFileTypeBuckets, err := s.searchClient.AggregateDetectedFileTypeByFileIDs(
		ctx,
		tenantID,
		extraFileIDsFromBuckets(bucketsToFacetItems(fileIDBuckets)),
	)
	if err != nil {
		return nil, fmt.Errorf("aggregate detected file type facets: %w", err)
	}

	facets := SearchFacets{
		RecordType: recordTypeBuckets,
		DetectedFileType: dedupeSearchFacetItems(bucketsToFacetItems(detectedFileTypeBuckets)),
		Entities: map[string][]SearchFacetItem{
			"ip_addresses": bucketsToFacetItems(searchResp.Aggregations["entities_ip_addresses"]),
			"emails":       bucketsToFacetItems(searchResp.Aggregations["entities_emails"]),
			"domains":      bucketsToFacetItems(searchResp.Aggregations["entities_domains"]),
			"urls":         bucketsToFacetItems(searchResp.Aggregations["entities_urls"]),
			"hashes":       bucketsToFacetItems(searchResp.Aggregations["entities_hashes"]),
		},
	}
	if err := s.createSearchAuditLog(
		ctx,
		parsedReq,
		"search.executed",
		map[string]any{
			"result_count": searchResp.Total,
			"source":       "search",
		},
	); err != nil {
		return nil, err
	}

	return &SearchResponse{
		Total:    searchResp.Total,
		Page:     parsedReq.Page,
		PageSize: parsedReq.PageSize,
		Results:  results,
		Facets:   facets,
	}, nil
}

func (s *SearchService) Suggestions(ctx context.Context, req SearchRecordsRequest) (*SearchSuggestionResponse, error) {
	if s.searchClient == nil {
		return nil, fmt.Errorf("%w: search client is not initialized", ErrSearchServiceUnavailable)
	}

	parsedReq, err := normalizeSearchRequest(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSearchRequest, err)
	}

	tenantID := strings.TrimSpace(parsedReq.TenantID)
	fileIDsForFilter := []string{}
	if parsedReq.Extension != "" || parsedReq.DetectedFileType != "" {
		files, err := s.searchClient.ResolveFileMetadata(ctx, search.FileMetadataFilter{
			TenantID:         tenantID,
			FileID:           parsedReq.FileID,
			Extension:        parsedReq.Extension,
			DetectedFileType: parsedReq.DetectedFileType,
		})
		if err != nil {
			return nil, fmt.Errorf("resolve file metadata: %w", err)
		}
		for _, file := range files {
			if file.FileID != "" {
				fileIDsForFilter = append(fileIDsForFilter, file.FileID)
			}
		}
		if len(fileIDsForFilter) == 0 {
			return &SearchSuggestionResponse{
				Suggestions: SearchSuggestions{
					RecordType:       []SearchFacetItem{},
					DetectedFileType: []SearchFacetItem{},
					Entities: map[string][]SearchFacetItem{
						"ip_addresses": {},
						"emails":       {},
						"domains":      {},
						"urls":         {},
						"hashes":       {},
					},
				},
			}, nil
		}
	}

	searchQuery := search.SearchRequest{
		Query:      parsedReq.Q,
		TenantID:   tenantID,
		FileID:     parsedReq.FileID,
		JobID:      parsedReq.JobID,
		RecordType: parsedReq.RecordType,
		StartDate:  parsedReq.DateFrom,
		EndDate:    parsedReq.DateTo,
		IP:         parsedReq.IP,
		Email:      parsedReq.Email,
		Domain:     parsedReq.Domain,
		FileIDs:    fileIDsForFilter,
		Page:       1,
		PageSize:   0,
		Sort:       "relevance",
	}

	searchResp, err := s.searchClient.SearchParsedRecords(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search parsed records: %w", err)
	}

	recordTypeBuckets := bucketsToFacetItems(searchResp.Aggregations["record_type"])
	fileIDBuckets := searchResp.Aggregations["file_id"]
	detectedFileTypeBuckets, err := s.searchClient.AggregateDetectedFileTypeByFileIDs(
		ctx,
		tenantID,
		extraFileIDsFromBuckets(bucketsToFacetItems(fileIDBuckets)),
	)
	if err != nil {
		return nil, fmt.Errorf("aggregate detected file type facets: %w", err)
	}

	response := SearchSuggestionResponse{
		Suggestions: SearchSuggestions{
			RecordType:       filterFacetsByQuery(recordTypeBuckets, parsedReq.Q, searchSuggestionLimit),
			DetectedFileType: filterFacetsByQuery(bucketsToFacetItems(detectedFileTypeBuckets), parsedReq.Q, searchSuggestionLimit),
			Entities: map[string][]SearchFacetItem{
				"ip_addresses": filterFacetsByQuery(bucketsToFacetItems(searchResp.Aggregations["entities_ip_addresses"]), parsedReq.Q, searchSuggestionLimit),
				"emails":       filterFacetsByQuery(bucketsToFacetItems(searchResp.Aggregations["entities_emails"]), parsedReq.Q, searchSuggestionLimit),
				"domains":      filterFacetsByQuery(bucketsToFacetItems(searchResp.Aggregations["entities_domains"]), parsedReq.Q, searchSuggestionLimit),
				"urls":         filterFacetsByQuery(bucketsToFacetItems(searchResp.Aggregations["entities_urls"]), parsedReq.Q, searchSuggestionLimit),
				"hashes":       filterFacetsByQuery(bucketsToFacetItems(searchResp.Aggregations["entities_hashes"]), parsedReq.Q, searchSuggestionLimit),
			},
		},
	}
	if err := s.createSearchAuditLog(
		ctx,
		parsedReq,
		"search.executed",
		map[string]any{
			"result_count": len(response.Suggestions.RecordType) + len(response.Suggestions.DetectedFileType),
			"source":       "suggestions",
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *SearchService) createSearchAuditLog(
	ctx context.Context,
	req SearchRecordsRequest,
	action string,
	stats map[string]any,
) error {
	if s.auditLogRepo == nil {
		return nil
	}
	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = searchDefaultTenantID
	}

	details := map[string]any{
		"request": map[string]any{
			"tenant_id":          tenantID,
			"query":              req.Q,
			"file_id":            req.FileID,
			"extension":          req.Extension,
			"detected_file_type": req.DetectedFileType,
			"record_type":        req.RecordType,
			"date_from":          req.DateFrom,
			"date_to":            req.DateTo,
			"ip":                 req.IP,
			"email":              req.Email,
			"domain":             req.Domain,
			"job_id":             req.JobID,
			"page":               req.Page,
			"page_size":          req.PageSize,
			"sort":               req.Sort,
			"source":             stats["source"],
		},
		"stats":       stats,
		"request_id":  middleware.RequestIDFromContext(ctx),
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
	}

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	_, err = s.auditLogRepo.CreateAuditLog(ctx, &models.AuditLog{
		TenantID:   tenantID,
		ActorUserID: middleware.ActorUserIDFromContext(ctx),
		Action:     action,
		EntityType: ptrString("search"),
		Details:    detailsJSON,
		IPAddress:  middleware.RequestIPAddressFromContext(ctx),
		UserAgent:  middleware.RequestUserAgentFromContext(ctx),
	})
	return err
}

func normalizeSearchRequest(req SearchRecordsRequest) (SearchRecordsRequest, error) {
	request := req
	request.Q = strings.TrimSpace(request.Q)
	request.TenantID = strings.TrimSpace(request.TenantID)
	request.FileID = strings.TrimSpace(request.FileID)
	request.Extension = strings.TrimSpace(request.Extension)
	request.DetectedFileType = strings.TrimSpace(request.DetectedFileType)
	request.RecordType = strings.TrimSpace(request.RecordType)
	request.DateFrom = strings.TrimSpace(request.DateFrom)
	request.DateTo = strings.TrimSpace(request.DateTo)
	request.IP = strings.TrimSpace(request.IP)
	request.Email = strings.TrimSpace(request.Email)
	request.Domain = strings.TrimSpace(request.Domain)
	request.JobID = strings.TrimSpace(request.JobID)

	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 {
		request.PageSize = defaultSearchPageSize
	}
	if request.PageSize > maxSearchPageSize {
		request.PageSize = maxSearchPageSize
	}

	request.Sort = strings.TrimSpace(strings.ToLower(request.Sort))
	if request.Sort == "" {
		request.Sort = searchDefaultSort
	}
	if request.Sort != "relevance" && request.Sort != "created_at" {
		return request, fmt.Errorf("invalid sort: %s", request.Sort)
	}

	if request.DateFrom != "" {
		if _, err := parseSearchDate(request.DateFrom); err != nil {
			return request, fmt.Errorf("invalid date_from: %w", err)
		}
	}
	if request.DateTo != "" {
		if _, err := parseSearchDate(request.DateTo); err != nil {
			return request, fmt.Errorf("invalid date_to: %w", err)
		}
	}

	return request, nil
}

func parseSearchDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err == nil {
		return parsed, nil
	}
	parsedDateOnly, dateOnlyErr := time.Parse("2006-01-02", trimmed)
	if dateOnlyErr == nil {
		return parsedDateOnly, nil
	}
	parsed, err = time.Parse("2006-01-02T15:04:05.999Z07:00", trimmed)
	if err == nil {
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("%w (tried RFC3339, date-only, and timestamp-with-millis)", err)
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

func parseTime(value any) time.Time {
	if value == nil {
		return time.Time{}
	}
	if text, ok := value.(string); ok {
		parsed, err := time.Parse(time.RFC3339, text)
		if err != nil {
			parsed, err = time.Parse("2006-01-02T15:04:05.999Z07:00", text)
			if err != nil {
				return time.Time{}
			}
		}
		return parsed
	}
	if timestamp, ok := value.(float64); ok {
		return time.Unix(int64(timestamp), 0).UTC()
	}
	if timestamp, ok := value.(int64); ok {
		return time.Unix(timestamp, 0).UTC()
	}
	return time.Time{}
}

func toEntitiesMap(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	entities, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return entities
}

func preview(content string, limit int) string {
	if len(content) <= limit {
		return content
	}
	return content[:limit] + "..."
}

func bucketsToFacetItems(buckets []search.FacetBucket) []SearchFacetItem {
	if len(buckets) == 0 {
		return []SearchFacetItem{}
	}
	items := make([]SearchFacetItem, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.Value == "" {
			continue
		}
		items = append(items, SearchFacetItem{Value: bucket.Value, Count: bucket.Count})
	}
	return items
}

func extraFileIDsFromBuckets(fileBuckets []SearchFacetItem) []string {
	ids := make([]string, 0, len(fileBuckets))
	for _, bucket := range fileBuckets {
		if bucket.Value != "" {
			ids = append(ids, bucket.Value)
		}
	}
	return dedupeStrings(ids)
}

func dedupeSearchFacetItems(items []SearchFacetItem) []SearchFacetItem {
	seen := map[string]struct{}{}
	deduped := make([]SearchFacetItem, 0, len(items))
	for _, item := range items {
		if item.Value == "" {
			continue
		}
		if _, exists := seen[item.Value]; exists {
			continue
		}
		seen[item.Value] = struct{}{}
		deduped = append(deduped, item)
	}
	return deduped
}

func filterFacetsByQuery(items []SearchFacetItem, query string, limit int) []SearchFacetItem {
	if len(items) == 0 {
		return []SearchFacetItem{}
	}
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		if len(items) <= limit {
			return items
		}
		return items[:limit]
	}
	filtered := make([]SearchFacetItem, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Value), query) {
			filtered = append(filtered, item)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Count == filtered[j].Count {
			return filtered[i].Value < filtered[j].Value
		}
		return filtered[i].Count > filtered[j].Count
	})
	if len(filtered) <= limit {
		return filtered
	}
	return filtered[:limit]
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
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
