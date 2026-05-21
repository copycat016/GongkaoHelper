package services

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"gkweb/backend/internal/config"
	"gkweb/backend/internal/models"
)

const baiduOCRProvider = "baidu"
const ocrLocalConfigPath = "data/ocr_config.json"

type OCRServerConfig struct {
	Enabled        bool           `json:"enabled"`
	APIKey         string         `json:"api_key,omitempty"`
	SecretKey      string         `json:"secret_key,omitempty"`
	APIKeyMasked   string         `json:"api_key_masked,omitempty"`
	SecretMasked   string         `json:"secret_masked,omitempty"`
	MonthlyLimit   int            `json:"monthly_limit"`
	EngineLimits   map[string]int `json:"engine_limits,omitempty"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	Source         string         `json:"source"`
}

type OCREngine struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Async       bool   `json:"async"`
	Endpoint    string `json:"-"`
}

type OCRScene struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Description   string `json:"description"`
	DefaultEngine string `json:"default_engine"`
}

type OCRResult struct {
	TaskID      uint           `json:"task_id"`
	Source      string         `json:"source"`
	Provider    string         `json:"provider"`
	Engine      string         `json:"engine"`
	Status      string         `json:"status"`
	Text        string         `json:"text"`
	Quality     PDFTextQuality `json:"quality"`
	LineCount   int            `json:"line_count"`
	FromCache   bool           `json:"from_cache"`
	RawResponse string         `json:"raw_response,omitempty"`
}

type OCREngineUsage struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Used        int64  `json:"used"`
	Limit       int    `json:"limit"`
	Description string `json:"description"`
}

type BaiduOCRService struct {
	db          *gorm.DB
	cfg         config.Config
	client      *http.Client
	token       string
	tokenExpiry time.Time
	tokenMu     sync.Mutex
}

func NewBaiduOCRService(db *gorm.DB, cfg config.Config) *BaiduOCRService {
	timeout := time.Duration(cfg.BaiduOCRTimeoutSecond) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &BaiduOCRService{
		db:     db,
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

func (s *BaiduOCRService) Engines() []OCREngine {
	return []OCREngine{
		{Key: "general_basic", Label: "通用文字识别", Description: "适合普通截图、PDF 可解析图片、题干文字。", Endpoint: "general_basic"},
		{Key: "general", Label: "通用文字识别含位置", Description: "返回文字和位置信息，后续可用于版面分析。", Endpoint: "general"},
		{Key: "accurate_basic", Label: "高精度文字识别", Description: "更适合试题截图、低清晰度文字。", Endpoint: "accurate_basic"},
		{Key: "accurate", Label: "高精度含位置", Description: "高精度并保留位置，适合复杂题面。", Endpoint: "accurate"},
		{Key: "handwriting", Label: "手写文字识别", Description: "适合短手写答案、手写笔记。", Endpoint: "handwriting"},
		{Key: "essay_handwriting", Label: "手写作文识别", Description: "异步接口预留，适合申论长篇手写答案。", Async: true},
	}
}

func (s *BaiduOCRService) PublicConfig() OCRServerConfig {
	cfg := s.effectiveConfig()
	cfg.EngineLimits = s.normalizedEngineLimits(cfg.EngineLimits, cfg.MonthlyLimit)
	cfg.APIKeyMasked = maskSecret(cfg.APIKey)
	cfg.SecretMasked = maskSecret(cfg.SecretKey)
	cfg.APIKey = ""
	cfg.SecretKey = ""
	return cfg
}

func (s *BaiduOCRService) UpdateConfig(update OCRServerConfig) (OCRServerConfig, error) {
	current := s.effectiveConfig()
	if update.APIKey == "" {
		update.APIKey = current.APIKey
	}
	if update.SecretKey == "" {
		update.SecretKey = current.SecretKey
	}
	if update.MonthlyLimit <= 0 {
		update.MonthlyLimit = current.MonthlyLimit
	}
	update.EngineLimits = s.normalizedEngineLimits(update.EngineLimits, update.MonthlyLimit)
	if update.TimeoutSeconds <= 0 {
		update.TimeoutSeconds = current.TimeoutSeconds
	}
	update.Source = "local_file"

	if err := os.MkdirAll(filepath.Dir(ocrLocalConfigPath), 0755); err != nil {
		return OCRServerConfig{}, err
	}
	content, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return OCRServerConfig{}, err
	}
	if err := os.WriteFile(ocrLocalConfigPath, content, 0600); err != nil {
		return OCRServerConfig{}, err
	}

	s.tokenMu.Lock()
	s.token = ""
	s.tokenExpiry = time.Time{}
	s.tokenMu.Unlock()

	return s.PublicConfig(), nil
}

func (s *BaiduOCRService) Scenes() []OCRScene {
	return []OCRScene{
		{Key: "printed", Label: "印刷体题目", Description: "适合行测题、教材截图、普通 PDF 截图。", DefaultEngine: "general_basic"},
		{Key: "printed_accurate", Label: "高精度印刷体", Description: "适合文字较小、截图压缩或版面复杂的题面。", DefaultEngine: "accurate_basic"},
		{Key: "printed_layout", Label: "印刷体含位置", Description: "保留文字位置，后续可用于题面版面分析。", DefaultEngine: "accurate"},
		{Key: "handwriting", Label: "手写内容", Description: "适合短手写答案、手写笔记、草稿。", DefaultEngine: "handwriting"},
		{Key: "essay_handwriting", Label: "申论手写作文", Description: "异步接口预留，适合整篇申论手写答案。", DefaultEngine: "essay_handwriting"},
	}
}

func (s *BaiduOCRService) Recognize(userID uint, sceneKey string, engineKey string, fileHeader *multipart.FileHeader) (*OCRResult, error) {
	cfg := s.effectiveConfig()
	scene := s.sceneByKey(sceneKey)
	if engineKey == "" {
		engineKey = scene.DefaultEngine
	}
	engine, ok := s.engineByKey(engineKey)
	if !ok {
		return nil, fmt.Errorf("unsupported ocr engine: %s", engineKey)
	}
	if engine.Async {
		return nil, errors.New("async ocr engine is reserved but not implemented yet")
	}
	if !cfg.Enabled {
		return nil, errors.New("baidu ocr is disabled")
	}
	if cfg.APIKey == "" || cfg.SecretKey == "" {
		return nil, errors.New("baidu ocr api key or secret key is empty")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(content)
	fileHash := hex.EncodeToString(hash[:])

	if cached, err := s.cachedTask(userID, engine.Key, fileHash); err == nil && cached != nil {
		return &OCRResult{
			TaskID:    cached.ID,
			Source:    "ocr",
			Provider:  cached.Provider,
			Engine:    cached.Engine,
			Status:    cached.Status,
			Text:      cached.RawText,
			Quality:   ocrTextQuality(cached.RawText),
			LineCount: countTextLines(cached.RawText),
			FromCache: true,
		}, nil
	}

	if err := s.checkMonthlyLimit(userID, engine.Key); err != nil {
		return nil, err
	}

	task := models.OCRTask{
		BaseModel:   models.BaseModel{UserID: userID},
		Provider:    baiduOCRProvider,
		Engine:      engine.Key,
		Scene:       scene.Key,
		Status:      "processing",
		FileName:    fileHeader.Filename,
		FileSHA256:  fileHash,
		MimeType:    fileHeader.Header.Get("Content-Type"),
		SizeBytes:   fileHeader.Size,
		RawResponse: "",
	}
	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}

	rawText, rawResponse, err := s.callBaiduSync(engine.Endpoint, content)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		_ = s.db.Save(&task).Error
		return nil, err
	}

	task.Status = "success"
	task.RawText = rawText
	task.RawResponse = rawResponse
	if err := s.db.Save(&task).Error; err != nil {
		return nil, err
	}

	return &OCRResult{
		TaskID:      task.ID,
		Source:      "ocr",
		Provider:    task.Provider,
		Engine:      task.Engine,
		Status:      task.Status,
		Text:        task.RawText,
		Quality:     ocrTextQuality(task.RawText),
		LineCount:   countTextLines(task.RawText),
		RawResponse: task.RawResponse,
	}, nil
}

func ocrTextQuality(text string) PDFTextQuality {
	if strings.TrimSpace(text) == "" {
		return PDFTextQuality{OK: false, Reason: "ocr returned no text"}
	}
	return PDFTextQuality{OK: true, Reason: "ocr text captured"}
}

func (s *BaiduOCRService) MonthUsage(userID uint) (map[string]any, error) {
	cfg := s.effectiveConfig()
	start, end := currentMonthRange()
	var used int64
	if err := s.db.Model(&models.OCRTask{}).
		Where("user_id = ? AND provider = ? AND created_at >= ? AND created_at < ?", userID, baiduOCRProvider, start, end).
		Count(&used).Error; err != nil {
		return nil, err
	}
	return map[string]any{
		"used":            used,
		"local_limit":     cfg.MonthlyLimit,
		"engines":         s.engineUsages(userID, start, end, cfg),
		"month":           start.Format("2006-01"),
		"provider_quota":  "百度各 OCR 产品额度独立，当前仅统计本系统本月调用次数。",
		"quota_queryable": false,
	}, nil
}

func (s *BaiduOCRService) cachedTask(userID uint, engine string, fileHash string) (*models.OCRTask, error) {
	var task models.OCRTask
	err := s.db.
		Where("user_id = ? AND provider = ? AND engine = ? AND file_sha256 = ? AND status = ?", userID, baiduOCRProvider, engine, fileHash, "success").
		Order("created_at desc").
		First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &task, err
}

func (s *BaiduOCRService) checkMonthlyLimit(userID uint, engine string) error {
	cfg := s.effectiveConfig()
	limit := s.limitForEngine(cfg, engine)
	if limit <= 0 {
		return nil
	}
	start, end := currentMonthRange()
	var used int64
	if err := s.db.Model(&models.OCRTask{}).
		Where("user_id = ? AND provider = ? AND engine = ? AND created_at >= ? AND created_at < ?", userID, baiduOCRProvider, engine, start, end).
		Count(&used).Error; err != nil {
		return err
	}
	if used >= int64(limit) {
		return fmt.Errorf("baidu ocr monthly limit exceeded for %s", engine)
	}
	return nil
}

func (s *BaiduOCRService) callBaiduSync(endpoint string, content []byte) (string, string, error) {
	token, err := s.accessToken()
	if err != nil {
		return "", "", err
	}

	values := url.Values{}
	values.Set("image", base64.StdEncoding.EncodeToString(content))
	values.Set("detect_direction", "true")
	values.Set("paragraph", "true")
	values.Set("probability", "false")

	apiURL := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/ocr/v1/%s?access_token=%s", endpoint, url.QueryEscape(token))
	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode >= 400 {
		return "", string(body), fmt.Errorf("baidu ocr http status %d", resp.StatusCode)
	}

	text, err := parseBaiduWords(body)
	if err != nil {
		return "", string(body), err
	}
	return text, string(body), nil
}

func (s *BaiduOCRService) accessToken() (string, error) {
	cfg := s.effectiveConfig()
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	if s.token != "" && time.Now().Before(s.tokenExpiry.Add(-10*time.Minute)) {
		return s.token, nil
	}

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", cfg.APIKey)
	values.Set("client_secret", cfg.SecretKey)

	req, err := http.NewRequest(http.MethodPost, "https://aip.baidubce.com/oauth/2.0/token", strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("baidu token failed: %s %s", payload.Error, payload.ErrorDesc)
	}

	s.token = payload.AccessToken
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = 2592000
	}
	s.tokenExpiry = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	return s.token, nil
}

func (s *BaiduOCRService) effectiveConfig() OCRServerConfig {
	cfg := OCRServerConfig{
		Enabled:        s.cfg.BaiduOCREnabled,
		APIKey:         s.cfg.BaiduOCRAPIKey,
		SecretKey:      s.cfg.BaiduOCRSecretKey,
		MonthlyLimit:   s.cfg.BaiduOCRMonthlyLimit,
		EngineLimits:   defaultEngineLimits(s.cfg.BaiduOCRMonthlyLimit),
		TimeoutSeconds: s.cfg.BaiduOCRTimeoutSecond,
		Source:         "env",
	}
	content, err := os.ReadFile(ocrLocalConfigPath)
	if err != nil {
		return cfg
	}
	var local OCRServerConfig
	if err := json.Unmarshal(content, &local); err != nil {
		return cfg
	}
	if local.APIKey == "" {
		local.APIKey = cfg.APIKey
	}
	if local.SecretKey == "" {
		local.SecretKey = cfg.SecretKey
	}
	if local.MonthlyLimit <= 0 {
		local.MonthlyLimit = cfg.MonthlyLimit
	}
	local.EngineLimits = s.normalizedEngineLimits(local.EngineLimits, local.MonthlyLimit)
	if local.TimeoutSeconds <= 0 {
		local.TimeoutSeconds = cfg.TimeoutSeconds
	}
	local.Source = "local_file"
	return local
}

func (s *BaiduOCRService) engineUsages(userID uint, start time.Time, end time.Time, cfg OCRServerConfig) []OCREngineUsage {
	limits := s.normalizedEngineLimits(cfg.EngineLimits, cfg.MonthlyLimit)
	usages := make([]OCREngineUsage, 0, len(s.Engines()))
	for _, engine := range s.Engines() {
		var used int64
		_ = s.db.Model(&models.OCRTask{}).
			Where("user_id = ? AND provider = ? AND engine = ? AND created_at >= ? AND created_at < ?", userID, baiduOCRProvider, engine.Key, start, end).
			Count(&used).Error
		usages = append(usages, OCREngineUsage{
			Key:         engine.Key,
			Label:       engine.Label,
			Used:        used,
			Limit:       limits[engine.Key],
			Description: engine.Description,
		})
	}
	return usages
}

func (s *BaiduOCRService) limitForEngine(cfg OCRServerConfig, engine string) int {
	limits := s.normalizedEngineLimits(cfg.EngineLimits, cfg.MonthlyLimit)
	return limits[engine]
}

func (s *BaiduOCRService) normalizedEngineLimits(limits map[string]int, fallback int) map[string]int {
	if fallback < 0 {
		fallback = 0
	}
	next := defaultEngineLimits(fallback)
	for key, value := range limits {
		if value >= 0 {
			next[key] = value
		}
	}
	return next
}

func defaultEngineLimits(fallback int) map[string]int {
	return map[string]int{
		"general_basic":     fallback,
		"general":           fallback,
		"accurate_basic":    fallback,
		"accurate":          fallback,
		"handwriting":       fallback,
		"essay_handwriting": fallback,
	}
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= 8 {
		return "****"
	}
	return string(runes[:4]) + "****" + string(runes[len(runes)-4:])
}

func (s *BaiduOCRService) engineByKey(key string) (OCREngine, bool) {
	if key == "" {
		key = "general_basic"
	}
	for _, engine := range s.Engines() {
		if engine.Key == key {
			return engine, true
		}
	}
	return OCREngine{}, false
}

func (s *BaiduOCRService) sceneByKey(key string) OCRScene {
	if key == "" {
		key = "printed"
	}
	for _, scene := range s.Scenes() {
		if scene.Key == key {
			return scene
		}
	}
	return s.Scenes()[0]
}

func parseBaiduWords(body []byte) (string, error) {
	var payload struct {
		ErrorCode   int    `json:"error_code"`
		ErrorMsg    string `json:"error_msg"`
		WordsResult []struct {
			Words string `json:"words"`
		} `json:"words_result"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.ErrorCode != 0 {
		return "", fmt.Errorf("baidu ocr failed: %d %s", payload.ErrorCode, payload.ErrorMsg)
	}
	lines := make([]string, 0, len(payload.WordsResult))
	for _, item := range payload.WordsResult {
		if item.Words != "" {
			lines = append(lines, item.Words)
		}
	}
	return normalizeOCRLines(lines), nil
}

func normalizeOCRLines(lines []string) string {
	cleaned := make([]string, 0, len(lines))
	singleRuneLines := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len([]rune(line)) == 1 {
			singleRuneLines++
		}
		cleaned = append(cleaned, line)
	}
	if len(cleaned) == 0 {
		return ""
	}
	if len(cleaned) >= 12 && float64(singleRuneLines)/float64(len(cleaned)) > 0.72 {
		return mergeSingleRuneOCRLines(cleaned)
	}
	return strings.Join(cleaned, "\n")
}

func mergeSingleRuneOCRLines(lines []string) string {
	var builder strings.Builder
	for _, line := range lines {
		runes := []rune(line)
		if len(runes) == 1 {
			builder.WriteString(line)
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func currentMonthRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return start, start.AddDate(0, 1, 0)
}
