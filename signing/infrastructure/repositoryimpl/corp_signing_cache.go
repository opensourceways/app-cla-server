package repositoryimpl

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

const (
	corpSigningPageCacheTTL    = 5 * time.Minute
	corpSigningPageCachePrefix = "corp_signing:page:"
)

// cacheDAO defines the Redis operations needed by the cache decorator.
type cacheDAO interface {
	SetWithExpiry(key string, val interface{}, expiry time.Duration) error
	Get(key string, data interface{}) error
	Del(keys ...string) error
	Keys(pattern string) ([]string, error)
	IsDocNotExists(err error) bool
}

// cachedCorpSigningPage is a serialisable snapshot of CorpSigningSummaryPage.
type cachedCorpSigningPage struct {
	Total int64                      `json:"total"`
	Data  []cachedCorpSigningSummary `json:"data"`
}

type cachedCorpSigningSummary struct {
	Id        string `json:"id"`
	Date      string `json:"date"`
	HasPDF    bool   `json:"has_pdf"`
	CLANotify string `json:"cla_notify"`
	// Link
	LinkId   string `json:"link_id"`
	CLAId    string `json:"cla_id"`
	Language string `json:"language"`
	// Rep
	RepName  string `json:"rep_name"`
	RepEmail string `json:"rep_email"`
	// Corp
	CorpName           string   `json:"corp_name"`
	PrimaryEmailDomain string   `json:"primary_email_domain"`
	AllEmailDomains    []string `json:"all_email_domains"`
	// Admin
	AdminId    string `json:"admin_id"`
	AdminName  string `json:"admin_name"`
	AdminEmail string `json:"admin_email"`
}

func toCache(p repository.CorpSigningSummaryPage) cachedCorpSigningPage {
	items := make([]cachedCorpSigningSummary, len(p.Data))
	for i := range p.Data {
		s := &p.Data[i]
		adminName := ""
		adminEmail := ""
		if !s.Admin.IsEmpty() {
			adminName = s.Admin.Representative.Name.Name()
			adminEmail = s.Admin.Representative.EmailAddr.EmailAddr()
		}
		items[i] = cachedCorpSigningSummary{
			Id:                 s.Id,
			Date:               s.Date,
			HasPDF:             s.HasPDF,
			CLANotify:          s.CLANotify,
			LinkId:             s.Link.Id,
			CLAId:              s.Link.CLAId,
			Language:           s.Link.Language.Language(),
			RepName:            s.Rep.Name.Name(),
			RepEmail:           s.Rep.EmailAddr.EmailAddr(),
			CorpName:           s.Corp.Name.CorpName(),
			PrimaryEmailDomain: s.Corp.PrimaryEmailDomain,
			AllEmailDomains:    s.Corp.AllEmailDomains,
			AdminId:            s.Admin.Id,
			AdminName:          adminName,
			AdminEmail:         adminEmail,
		}
	}
	return cachedCorpSigningPage{Total: p.Total, Data: items}
}

func fromCache(c cachedCorpSigningPage) repository.CorpSigningSummaryPage {
	items := make([]repository.CorpSigningSummary, len(c.Data))
	for i := range c.Data {
		s := &c.Data[i]
		items[i] = repository.CorpSigningSummary{
			Id:        s.Id,
			Date:      s.Date,
			HasPDF:    s.HasPDF,
			CLANotify: s.CLANotify,
			Link: domain.LinkInfo{
				Id: s.LinkId,
				CLAInfo: domain.CLAInfo{
					CLAId:    s.CLAId,
					Language: dp.CreateLanguage(s.Language),
				},
			},
			Rep: domain.Representative{
				Name:      dp.CreateName(s.RepName),
				EmailAddr: dp.CreateEmailAddr(s.RepEmail),
			},
			Corp: domain.Corporation{
				Name:               dp.CreateCorpName(s.CorpName),
				PrimaryEmailDomain: s.PrimaryEmailDomain,
				AllEmailDomains:    s.AllEmailDomains,
			},
			Admin: domain.Manager{
				Id: s.AdminId,
				Representative: domain.Representative{
					Name:      dp.CreateName(s.AdminName),
					EmailAddr: dp.CreateEmailAddr(s.AdminEmail),
				},
			},
		}
	}
	return repository.CorpSigningSummaryPage{Total: c.Total, Data: items}
}

// pageKey builds the cache key for a specific page query.
func pageKey(linkId string, page, pageSize int, adminAdded bool, searchQuery string) string {
	return fmt.Sprintf("%s%s:%d:%d:%v:%s", corpSigningPageCachePrefix, linkId, page, pageSize, adminAdded, searchQuery)
}

// linkPattern returns the glob pattern to match all page cache keys for a linkId.
func linkPattern(linkId string) string {
	return fmt.Sprintf("%s%s:*", corpSigningPageCachePrefix, linkId)
}

// cachedCorpSigning wraps corpSigning with a Redis page cache (Decorator pattern).
type cachedCorpSigning struct {
	repo  *corpSigning
	cache cacheDAO
	sf    singleflightGroup
	sem   semaphoreWeighted
}

const cacheWriteConcurrency = 100

var cacheWriteSem semaphoreWeighted

type singleflightGroup interface {
	Do(key string, fn func() (interface{}, error)) (interface{}, error, bool)
}

type semaphoreWeighted interface {
	TryAcquire(n int64) bool
	Release(n int64)
}

func init() {
	cacheWriteSem = newInternalSemaphore(cacheWriteConcurrency)
}

func newInternalSemaphore(n int64) semaphoreWeighted {
	return &internalSemaphore{ch: make(chan struct{}, n)}
}

type internalSemaphore struct {
	ch chan struct{}
}

func (s *internalSemaphore) TryAcquire(n int64) bool {
	for i := int64(0); i < n; i++ {
		select {
		case s.ch <- struct{}{}:
		default:
			for j := int64(0); j < i; j++ {
				<-s.ch
			}
			return false
		}
	}
	return true
}

func (s *internalSemaphore) Release(n int64) {
	for i := int64(0); i < n; i++ {
		<-s.ch
	}
}

type internalSingleflight struct {
	mu sync.Mutex
	m  map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
	dup bool
}

func newInternalSingleflight() singleflightGroup {
	return &internalSingleflight{m: make(map[string]*call)}
}

func (g *internalSingleflight) Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	g.mu.Lock()
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err, true
	}
	c := &call{}
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.m, key)
		g.mu.Unlock()
		c.wg.Done()
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				c.err = fmt.Errorf("singleflight panic: %v", r)
			}
		}()
		c.val, c.err = fn()
	}()
	return c.val, c.err, false
}

// NewCachedCorpSigning wraps the base repo with a Redis cache layer.
func NewCachedCorpSigning(dao dao, cache cacheDAO) *cachedCorpSigning {
	return &cachedCorpSigning{
		repo:  NewCorpSigning(dao),
		cache: cache,
		sf:    newInternalSingleflight(),
		sem:   cacheWriteSem,
	}
}

// FindPage tries the cache first; on miss it queries MongoDB and populates the cache.
func (c *cachedCorpSigning) FindPage(linkId string, page, pageSize int, adminAdded bool, searchQuery string) (repository.CorpSigningSummaryPage, error) {
	key := pageKey(linkId, page, pageSize, adminAdded, searchQuery)

	// --- cache hit ---
	var cached cachedCorpSigningPage
	if err := c.cache.Get(key, &cached); err == nil {
		return fromCache(cached), nil
	}

	// --- cache miss: singleflight to deduplicate concurrent same-key queries ---
	v, err, _ := c.sf.Do(key, func() (interface{}, error) {
		result, queryErr := c.repo.FindPage(linkId, page, pageSize, adminAdded, searchQuery)
		if queryErr != nil {
			return result, queryErr
		}

		// populate cache synchronously (inside singleflight, no race)
		if c.sem.TryAcquire(1) {
			defer c.sem.Release(1)

			payload, jsonErr := json.Marshal(toCache(result))
			if jsonErr != nil {
				log.Printf("corp_signing cache marshal error: %v", jsonErr)
				return result, nil
			}
			if setErr := c.cache.SetWithExpiry(key, string(payload), corpSigningPageCacheTTL); setErr != nil {
				log.Printf("corp_signing cache set error: %v", setErr)
			}
		}
		return result, nil
	})

	if err != nil {
		return repository.CorpSigningSummaryPage{}, err
	}

	return v.(repository.CorpSigningSummaryPage), nil
}

// invalidateLink deletes all cached pages for a given linkId.
func (c *cachedCorpSigning) invalidateLink(linkId string) {
	keys, err := c.cache.Keys(linkPattern(linkId))
	if err != nil {
		log.Printf("corp_signing cache keys error: %v", err)
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := c.cache.Del(keys...); err != nil {
		log.Printf("corp_signing cache invalidate error: %v", err)
	}
}

// --- write-through: invalidate cache on every mutation ---

func (c *cachedCorpSigning) Add(v *domain.CorpSigning) error {
	err := c.repo.Add(v)
	if err == nil {
		go c.invalidateLink(v.Link.Id)
	}
	return err
}

func (c *cachedCorpSigning) AddForMigrate(v *domain.CorpSigning) error {
	err := c.repo.AddForMigrate(v)
	if err == nil {
		go c.invalidateLink(v.Link.Id)
	}
	return err
}

func (c *cachedCorpSigning) Remove(cs *domain.CorpSigning) error {
	err := c.repo.Remove(cs)
	if err == nil {
		go c.invalidateLink(cs.Link.Id)
	}
	return err
}

func (c *cachedCorpSigning) AddAdmin(cs *domain.CorpSigning) error {
	err := c.repo.AddAdmin(cs)
	if err == nil {
		go c.invalidateLink(cs.Link.Id)
	}
	return err
}

func (c *cachedCorpSigning) SaveCorpPDF(cs *domain.CorpSigning, pdf []byte) error {
	err := c.repo.SaveCorpPDF(cs, pdf)
	if err == nil {
		go c.invalidateLink(cs.Link.Id)
	}
	return err
}

func (c *cachedCorpSigning) Update(cs *domain.CorpSigning) error {
	err := c.repo.Update(cs)
	if err == nil {
		go c.invalidateLink(cs.Link.Id)
	}
	return err
}

// --- pass-through for all other methods ---

func (c *cachedCorpSigning) Find(index string) (domain.CorpSigning, error) {
	return c.repo.Find(index)
}

func (c *cachedCorpSigning) FindAll(linkId string) ([]repository.CorpSigningSummary, error) {
	return c.repo.FindAll(linkId)
}

func (c *cachedCorpSigning) FindAllWithPagination(linkId string, offset, limit int) ([]repository.CorpSigningSummary, error) {
	return c.repo.FindAllWithPagination(linkId, offset, limit)
}

func (c *cachedCorpSigning) CountByLinkId(linkId string) (int64, error) {
	return c.repo.CountByLinkId(linkId)
}

func (c *cachedCorpSigning) FindCorpSummary(linkId, emailDomain string) ([]repository.CorpSummary, error) {
	return c.repo.FindCorpSummary(linkId, emailDomain)
}

func (c *cachedCorpSigning) FindCorpManagers(linkId, emailDomain string) ([]domain.Manager, error) {
	return c.repo.FindCorpManagers(linkId, emailDomain)
}

func (c *cachedCorpSigning) AddEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	return c.repo.AddEmployee(cs, es)
}

func (c *cachedCorpSigning) SaveEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	return c.repo.SaveEmployee(cs, es)
}

func (c *cachedCorpSigning) FindEmployees(csId string) ([]domain.EmployeeSigning, error) {
	return c.repo.FindEmployees(csId)
}

func (c *cachedCorpSigning) RemoveEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	return c.repo.RemoveEmployee(cs, es)
}

func (c *cachedCorpSigning) FindEmployeesByEmail(linkId string, email dp.EmailAddr) (repository.EmployeeSigningSummary, error) {
	return c.repo.FindEmployeesByEmail(linkId, email)
}

func (c *cachedCorpSigning) AddEmployeeManagers(cs *domain.CorpSigning, ms []domain.Manager) error {
	return c.repo.AddEmployeeManagers(cs, ms)
}

func (c *cachedCorpSigning) RemoveEmployeeManagers(cs *domain.CorpSigning, ms []string) error {
	return c.repo.RemoveEmployeeManagers(cs, ms)
}

func (c *cachedCorpSigning) FindEmployeeManagers(csId string) ([]domain.Manager, error) {
	return c.repo.FindEmployeeManagers(csId)
}

func (c *cachedCorpSigning) AddEmailDomain(cs *domain.CorpSigning, emailDomain string) error {
	return c.repo.AddEmailDomain(cs, emailDomain)
}

func (c *cachedCorpSigning) FindEmailDomains(csId string) ([]string, error) {
	return c.repo.FindEmailDomains(csId)
}

func (c *cachedCorpSigning) FindCorpPDF(csId string) ([]byte, error) {
	return c.repo.FindCorpPDF(csId)
}

func (c *cachedCorpSigning) HasSignedLink(linkId string) (bool, error) {
	return c.repo.HasSignedLink(linkId)
}

func (c *cachedCorpSigning) HasSignedCLA(index *domain.CLAIndex, t dp.CLAType) (bool, error) {
	return c.repo.HasSignedCLA(index, t)
}

func (c *cachedCorpSigning) UpdateClaId(cs *domain.CorpSigning) error {
	return c.repo.UpdateClaId(cs)
}

func (c *cachedCorpSigning) UpdateCLANotify(summary *repository.CorpSigningSummary) error {
	return c.repo.UpdateCLANotify(summary)
}

func (c *cachedCorpSigning) ListTriggered() ([]TriggeredCorp, error) {
	return c.repo.ListTriggered()
}

func (c *cachedCorpSigning) ResetTriggered(csId string, version int) error {
	return c.repo.ResetTriggered(csId, version)
}

func (c *cachedCorpSigning) SetPendingCLAForLink(linkId, newClaId string) error {
	return c.repo.SetPendingCLAForLink(linkId, newClaId)
}

func (c *cachedCorpSigning) FindPendingAgreements(linkId string) ([]repository.CorpSigningSummary, error) {
	return c.repo.FindPendingAgreements(linkId)
}
