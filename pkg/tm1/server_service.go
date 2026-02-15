package tm1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/andreyea/tm1go/pkg/models"
)

const (
	loggerLevelFatal   = 0
	loggerLevelError   = 1
	loggerLevelWarning = 2
	loggerLevelInfo    = 3
	loggerLevelDebug   = 4
	loggerLevelUnknown = 5
	loggerLevelOff     = 6
)

// LogLevel defines TM1 logger levels.
type LogLevel string

const (
	LogLevelFatal   LogLevel = "fatal"
	LogLevelError   LogLevel = "error"
	LogLevelWarning LogLevel = "warning"
	LogLevelInfo    LogLevel = "info"
	LogLevelDebug   LogLevel = "debug"
	LogLevelOff     LogLevel = "off"
)

// MessageLogQuery options for message log retrieval.
type MessageLogQuery struct {
	Reverse           bool
	Since             string
	Until             string
	Top               int
	Logger            string
	Level             string
	MessageContains   []string
	MessageContainsOp string
}

// TransactionLogQuery options for transaction log retrieval.
type TransactionLogQuery struct {
	Reverse            bool
	User               string
	Cube               string
	Since              string
	Until              string
	Top                int
	ElementTupleFilter map[string]string
}

// AuditLogQuery options for audit log retrieval.
type AuditLogQuery struct {
	User       string
	ObjectType string
	ObjectName string
	Since      string
	Until      string
	Top        int
}

// ServerService exposes server-level operations similar to TM1py ServerService.
type ServerService struct {
	rest *RestService

	process       *ProcessService
	users         *UserService
	configuration *ConfigurationService

	lastTransactionLogDeltaRequest string
	lastAuditLogDeltaRequest       string
	lastMessageLogDeltaRequest     string
}

// NewServerService creates a new ServerService instance.
func NewServerService(rest *RestService) *ServerService {
	return &ServerService{
		rest:          rest,
		process:       NewProcessService(rest),
		users:         NewUserService(rest),
		configuration: NewConfigurationService(rest),
	}
}

func (ss *ServerService) InitializeTransactionLogDeltaRequests(ctx context.Context, filter string) error {
	if err := ss.requirePreV12(); err != nil {
		return err
	}

	endpoint := "/TailTransactionLog()"
	if strings.TrimSpace(filter) != "" {
		endpoint += "?$filter=" + url.QueryEscape(filter)
	}

	payload, err := ss.getRawMap(ctx, endpoint)
	if err != nil {
		return err
	}
	ss.lastTransactionLogDeltaRequest = extractDeltaLink(payload)
	return nil
}

func (ss *ServerService) ExecuteTransactionLogDeltaRequest(ctx context.Context) ([]map[string]interface{}, error) {
	if strings.TrimSpace(ss.lastTransactionLogDeltaRequest) == "" {
		return nil, fmt.Errorf("transaction delta request not initialized")
	}

	payload, err := ss.getRawMap(ctx, "/"+strings.TrimPrefix(ss.lastTransactionLogDeltaRequest, "/"))
	if err != nil {
		return nil, err
	}
	ss.lastTransactionLogDeltaRequest = extractDeltaLink(payload)
	return extractValueMapSlice(payload), nil
}

func (ss *ServerService) InitializeAuditLogDeltaRequests(ctx context.Context, filter string) error {
	if err := ss.requirePreV12(); err != nil {
		return err
	}

	endpoint := "/TailAuditLog()"
	if strings.TrimSpace(filter) != "" {
		endpoint += "?$filter=" + url.QueryEscape(filter)
	}

	payload, err := ss.getRawMap(ctx, endpoint)
	if err != nil {
		return err
	}
	ss.lastAuditLogDeltaRequest = extractDeltaLink(payload)
	return nil
}

func (ss *ServerService) ExecuteAuditLogDeltaRequest(ctx context.Context) ([]map[string]interface{}, error) {
	if strings.TrimSpace(ss.lastAuditLogDeltaRequest) == "" {
		return nil, fmt.Errorf("audit delta request not initialized")
	}

	payload, err := ss.getRawMap(ctx, "/"+strings.TrimPrefix(ss.lastAuditLogDeltaRequest, "/"))
	if err != nil {
		return nil, err
	}
	ss.lastAuditLogDeltaRequest = extractDeltaLink(payload)
	return extractValueMapSlice(payload), nil
}

func (ss *ServerService) InitializeMessageLogDeltaRequests(ctx context.Context, filter string) error {
	if err := ss.requirePreV12(); err != nil {
		return err
	}

	endpoint := "/TailMessageLog()"
	if strings.TrimSpace(filter) != "" {
		endpoint += "?$filter=" + url.QueryEscape(filter)
	}

	payload, err := ss.getRawMap(ctx, endpoint)
	if err != nil {
		return err
	}
	ss.lastMessageLogDeltaRequest = extractDeltaLink(payload)
	return nil
}

func (ss *ServerService) ExecuteMessageLogDeltaRequest(ctx context.Context) ([]map[string]interface{}, error) {
	if strings.TrimSpace(ss.lastMessageLogDeltaRequest) == "" {
		return nil, fmt.Errorf("message delta request not initialized")
	}

	payload, err := ss.getRawMap(ctx, "/"+strings.TrimPrefix(ss.lastMessageLogDeltaRequest, "/"))
	if err != nil {
		return nil, err
	}
	ss.lastMessageLogDeltaRequest = extractDeltaLink(payload)
	return extractValueMapSlice(payload), nil
}

func (ss *ServerService) GetMessageLogEntries(ctx context.Context, q MessageLogQuery) ([]map[string]interface{}, error) {
	if err := ss.requirePreV12(); err != nil {
		return nil, err
	}
	if err := ss.requireOpsAdmin(ctx); err != nil {
		return nil, err
	}

	reverse := "desc"
	if !q.Reverse {
		reverse = "asc"
	}
	endpoint := "/MessageLogEntries?$orderby=TimeStamp " + reverse

	filters := make([]string, 0)
	if q.Since != "" {
		filters = append(filters, fmt.Sprintf("TimeStamp ge %s", q.Since))
	}
	if q.Until != "" {
		filters = append(filters, fmt.Sprintf("TimeStamp le %s", q.Until))
	}
	if q.Logger != "" {
		filters = append(filters, fmt.Sprintf("Logger eq '%s'", strings.ReplaceAll(q.Logger, "'", "''")))
	}
	if q.Level != "" {
		if idx, ok := messageLogLevelToIndex(q.Level); ok {
			filters = append(filters, fmt.Sprintf("Level eq %d", idx))
		}
	}
	if len(q.MessageContains) > 0 {
		op := strings.ToLower(strings.TrimSpace(q.MessageContainsOp))
		if op == "" {
			op = "and"
		}
		if op != "and" && op != "or" {
			return nil, fmt.Errorf("message contains operator must be 'and' or 'or'")
		}
		contains := make([]string, 0, len(q.MessageContains))
		for _, wildcard := range q.MessageContains {
			contains = append(contains, fmt.Sprintf("contains(toupper(Message),toupper('%s'))", strings.ReplaceAll(wildcard, "'", "''")))
		}
		filters = append(filters, "("+strings.Join(contains, " "+op+" ")+")")
	}
	if len(filters) > 0 {
		endpoint += "&$filter=" + strings.Join(filters, " and ")
	}
	if q.Top > 0 {
		endpoint += fmt.Sprintf("&$top=%d", q.Top)
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (ss *ServerService) WriteToMessageLog(ctx context.Context, level string, message string) error {
	if err := ss.requireDataAdmin(ctx); err != nil {
		return err
	}
	lvl := strings.ToUpper(strings.TrimSpace(level))
	if lvl != "FATAL" && lvl != "ERROR" && lvl != "WARN" && lvl != "INFO" && lvl != "DEBUG" {
		return fmt.Errorf("invalid message level: %s", level)
	}
	msg := strings.ReplaceAll(message, "'", "''")

	process := models.NewProcess("")
	process.PrologProcedure = fmt.Sprintf("LogOutput('%s', '%s');", lvl, msg)
	success, status, _, err := ss.process.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("failed to write to message log, status: %s", status)
	}
	return nil
}

func (ss *ServerService) GetTransactionLogEntries(ctx context.Context, q TransactionLogQuery) ([]map[string]interface{}, error) {
	if err := ss.requirePreV12(); err != nil {
		return nil, err
	}
	if err := ss.requireDataAdmin(ctx); err != nil {
		return nil, err
	}

	reverse := "desc"
	if !q.Reverse {
		reverse = "asc"
	}
	endpoint := "/TransactionLogEntries?$orderby=TimeStamp " + reverse

	filters := make([]string, 0)
	if q.User != "" {
		filters = append(filters, fmt.Sprintf("User eq '%s'", strings.ReplaceAll(q.User, "'", "''")))
	}
	if q.Cube != "" {
		filters = append(filters, fmt.Sprintf("Cube eq '%s'", strings.ReplaceAll(q.Cube, "'", "''")))
	}
	if len(q.ElementTupleFilter) > 0 {
		tupleFilters := make([]string, 0, len(q.ElementTupleFilter))
		for elem, op := range q.ElementTupleFilter {
			tupleFilters = append(tupleFilters, fmt.Sprintf("e %s '%s'", op, strings.ReplaceAll(elem, "'", "''")))
		}
		filters = append(filters, fmt.Sprintf("Tuple/any(e: %s)", strings.Join(tupleFilters, " or ")))
	}
	if q.Since != "" {
		filters = append(filters, fmt.Sprintf("TimeStamp ge %s", q.Since))
	}
	if q.Until != "" {
		filters = append(filters, fmt.Sprintf("TimeStamp le %s", q.Until))
	}
	if len(filters) > 0 {
		endpoint += "&$filter=" + strings.Join(filters, " and ")
	}
	if q.Top > 0 {
		endpoint += fmt.Sprintf("&$top=%d", q.Top)
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (ss *ServerService) GetAuditLogEntries(ctx context.Context, q AuditLogQuery) ([]map[string]interface{}, error) {
	if err := ss.requirePreV12(); err != nil {
		return nil, err
	}
	if err := ss.requireDataAdmin(ctx); err != nil {
		return nil, err
	}
	if !IsV1GreaterOrEqualToV2(ss.rest.version, "11.6") {
		return nil, fmt.Errorf("audit logs require TM1 version >= 11.6")
	}

	endpoint := "/AuditLogEntries?$expand=AuditDetails"
	filters := make([]string, 0)
	if q.User != "" {
		filters = append(filters, fmt.Sprintf("UserName eq '%s'", strings.ReplaceAll(q.User, "'", "''")))
	}
	if q.ObjectType != "" {
		filters = append(filters, fmt.Sprintf("ObjectType eq '%s'", strings.ReplaceAll(q.ObjectType, "'", "''")))
	}
	if q.ObjectName != "" {
		filters = append(filters, fmt.Sprintf("ObjectName eq '%s'", strings.ReplaceAll(q.ObjectName, "'", "''")))
	}
	if q.Since != "" {
		filters = append(filters, fmt.Sprintf("TimeStamp ge %s", q.Since))
	}
	if q.Until != "" {
		filters = append(filters, fmt.Sprintf("TimeStamp le %s", q.Until))
	}
	if len(filters) > 0 {
		endpoint += "&$filter=" + strings.Join(filters, " and ")
	}
	if q.Top > 0 {
		endpoint += fmt.Sprintf("&$top=%d", q.Top)
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (ss *ServerService) GetLastProcessMessageFromMessageLog(ctx context.Context, processName string) (string, error) {
	if err := ss.requirePreV12(); err != nil {
		return "", err
	}
	if err := ss.requireOpsAdmin(ctx); err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("/MessageLog()?$orderby='TimeStamp'&$filter=Logger eq 'TM1.Process' and contains(Message, '%s')", strings.ReplaceAll(processName, "'", "''"))
	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", endpoint, nil, &response); err != nil {
		return "", err
	}
	if len(response.Value) == 0 {
		return "", nil
	}
	msg, _ := response.Value[0]["Message"].(string)
	return msg, nil
}

func (ss *ServerService) GetServerName(ctx context.Context) (string, error) {
	return ss.configuration.GetServerName(ctx)
}

func (ss *ServerService) GetProductVersion(ctx context.Context) (string, error) {
	return ss.configuration.GetProductVersion(ctx)
}

func (ss *ServerService) GetAdminHost(ctx context.Context) (string, error) {
	return ss.configuration.GetAdminHost(ctx)
}

func (ss *ServerService) GetDataDirectory(ctx context.Context) (string, error) {
	return ss.configuration.GetDataDirectory(ctx)
}

func (ss *ServerService) GetConfiguration(ctx context.Context) (map[string]interface{}, error) {
	return ss.configuration.GetAll(ctx)
}

func (ss *ServerService) GetStaticConfiguration(ctx context.Context) (map[string]interface{}, error) {
	return ss.configuration.GetStatic(ctx)
}

func (ss *ServerService) GetActiveConfiguration(ctx context.Context) (map[string]interface{}, error) {
	return ss.configuration.GetActive(ctx)
}

func (ss *ServerService) GetAPIMetadata(ctx context.Context) ([]byte, error) {
	resp, err := ss.rest.Get(ctx, "/$metadata")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (ss *ServerService) UpdateStaticConfiguration(ctx context.Context, configuration map[string]interface{}) error {
	return ss.configuration.UpdateStatic(ctx, configuration)
}

func (ss *ServerService) SaveData(ctx context.Context) error {
	if err := ss.requirePreV12(); err != nil {
		return err
	}
	if err := ss.requireDataAdmin(ctx); err != nil {
		return err
	}

	process := models.NewProcess("")
	process.PrologProcedure = "SaveDataAll;"
	success, status, _, err := ss.process.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("SaveDataAll did not complete successfully: %s", status)
	}
	return nil
}

func (ss *ServerService) DeletePersistentFeeders(ctx context.Context) error {
	if err := ss.requireDataAdmin(ctx); err != nil {
		return err
	}

	process := models.NewProcess("")
	process.PrologProcedure = "DeleteAllPersistentFeeders;"
	success, status, _, err := ss.process.ExecuteProcessWithReturn(ctx, process, nil)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("DeleteAllPersistentFeeders did not complete successfully: %s", status)
	}
	return nil
}

func (ss *ServerService) StartPerformanceMonitor(ctx context.Context) error {
	config := map[string]interface{}{"Administration": map[string]interface{}{"PerformanceMonitorOn": true}}
	return ss.UpdateStaticConfiguration(ctx, config)
}

func (ss *ServerService) StopPerformanceMonitor(ctx context.Context) error {
	config := map[string]interface{}{"Administration": map[string]interface{}{"PerformanceMonitorOn": false}}
	return ss.UpdateStaticConfiguration(ctx, config)
}

func (ss *ServerService) ActivateAuditLog(ctx context.Context) error {
	config := map[string]interface{}{"Administration": map[string]interface{}{"AuditLog": map[string]interface{}{"Enable": true}}}
	return ss.UpdateStaticConfiguration(ctx, config)
}

func (ss *ServerService) DeactivateAuditLog(ctx context.Context) error {
	if err := ss.requireOpsAdmin(ctx); err != nil {
		return err
	}
	config := map[string]interface{}{"Administration": map[string]interface{}{"AuditLog": map[string]interface{}{"Enable": false}}}
	return ss.UpdateStaticConfiguration(ctx, config)
}

func (ss *ServerService) UpdateMessageLoggerLevel(ctx context.Context, logger string, level string) error {
	if err := ss.requireAdmin(ctx); err != nil {
		return err
	}

	idx, ok := loggerLevelToIndex(level)
	if !ok {
		return fmt.Errorf("%s is not a valid logger level", level)
	}

	endpoint := fmt.Sprintf("/Loggers('%s')", url.PathEscape(logger))
	return ss.rest.JSON(ctx, "PATCH", endpoint, map[string]int{"Level": idx}, nil)
}

func (ss *ServerService) GetAllMessageLoggerLevel(ctx context.Context) ([]map[string]interface{}, error) {
	if err := ss.requireAdmin(ctx); err != nil {
		return nil, err
	}

	var response struct {
		Value []map[string]interface{} `json:"value"`
	}
	if err := ss.rest.JSON(ctx, "GET", "/Loggers", nil, &response); err != nil {
		return nil, err
	}
	return response.Value, nil
}

func (ss *ServerService) getRawMap(ctx context.Context, endpoint string) (map[string]interface{}, error) {
	resp, err := ss.rest.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractDeltaLink(payload map[string]interface{}) string {
	deltaLinkRaw, ok := payload["@odata.deltaLink"]
	if !ok {
		return ""
	}
	deltaLink, ok := deltaLinkRaw.(string)
	if !ok {
		return ""
	}
	deltaLink = strings.TrimSpace(deltaLink)
	if deltaLink == "" {
		return ""
	}

	if idx := strings.Index(deltaLink, "/api/v1/"); idx >= 0 {
		return strings.TrimPrefix(deltaLink[idx+len("/api/v1/"):], "/")
	}
	return strings.TrimPrefix(deltaLink, "/")
}

func extractValueMapSlice(payload map[string]interface{}) []map[string]interface{} {
	valueRaw, ok := payload["value"]
	if !ok {
		return []map[string]interface{}{}
	}
	entries, ok := valueRaw.([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		if m, ok := entry.(map[string]interface{}); ok {
			result = append(result, m)
		}
	}
	return result
}

func messageLogLevelToIndex(level string) (int, bool) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "ERROR":
		return 1, true
	case "WARNING":
		return 2, true
	case "INFO":
		return 3, true
	case "DEBUG":
		return 4, true
	case "UNKNOWN":
		return 5, true
	default:
		return 0, false
	}
}

func loggerLevelToIndex(level string) (int, bool) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "FATAL":
		return loggerLevelFatal, true
	case "ERROR":
		return loggerLevelError, true
	case "WARNING":
		return loggerLevelWarning, true
	case "INFO":
		return loggerLevelInfo, true
	case "DEBUG":
		return loggerLevelDebug, true
	case "UNKNOWN":
		return loggerLevelUnknown, true
	case "OFF":
		return loggerLevelOff, true
	default:
		return 0, false
	}
}

func (ss *ServerService) requirePreV12() error {
	version := strings.TrimSpace(ss.rest.version)
	if version == "" {
		return nil
	}
	if IsV1GreaterOrEqualToV2(version, "12.0.0") {
		return fmt.Errorf("operation is deprecated and unavailable in TM1 version 12.0.0+")
	}
	return nil
}

func (ss *ServerService) requireAdmin(ctx context.Context) error {
	if ok, err := ss.checkAdminFlags(ctx); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("admin privileges required")
	}
	return nil
}

func (ss *ServerService) requireDataAdmin(ctx context.Context) error {
	user, err := ss.getActiveUserMap(ctx)
	if err != nil {
		return err
	}
	if userType, _ := user["Type"].(string); strings.EqualFold(userType, "Admin") || strings.EqualFold(userType, "DataAdmin") {
		return nil
	}
	if isDataAdmin, _ := user["IsDataAdmin"].(bool); isDataAdmin {
		return nil
	}
	return fmt.Errorf("data admin privileges required")
}

func (ss *ServerService) requireOpsAdmin(ctx context.Context) error {
	user, err := ss.getActiveUserMap(ctx)
	if err != nil {
		return err
	}
	if userType, _ := user["Type"].(string); strings.EqualFold(userType, "Admin") || strings.EqualFold(userType, "OperationsAdmin") {
		return nil
	}
	if isOpsAdmin, _ := user["IsOpsAdmin"].(bool); isOpsAdmin {
		return nil
	}
	return fmt.Errorf("operations admin privileges required")
}

func (ss *ServerService) checkAdminFlags(ctx context.Context) (bool, error) {
	user, err := ss.getActiveUserMap(ctx)
	if err != nil {
		return false, err
	}
	userType, _ := user["Type"].(string)
	return strings.EqualFold(userType, "Admin"), nil
}

func (ss *ServerService) getActiveUserMap(ctx context.Context) (map[string]interface{}, error) {
	var user map[string]interface{}
	if err := ss.rest.JSON(ctx, "GET", "/ActiveUser", nil, &user); err != nil {
		return nil, err
	}
	return user, nil
}
