package proxy

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func (s *PGStore) SaveProxy(ctx context.Context, p *Proxy) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO proxies (url, protocol, host, port, region, latency_ms, status, source, success_rate, fail_count, last_checked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (url) DO UPDATE SET status=$7, latency_ms=$6, last_checked=$11`,
		p.URL, p.Protocol, p.Host, p.Port, p.Region, p.LatencyMs, p.Status, p.Source, p.SuccessRate, p.FailCount, time.Now(), time.Now())
	return err
}

func (s *PGStore) GetProxies(ctx context.Context, filter map[string]interface{}) ([]*Proxy, error) {
	rows, err := s.pool.Query(ctx, "SELECT url, protocol, COALESCE(host,''), COALESCE(port,''), COALESCE(region,''), latency_ms, status, COALESCE(source,''), COALESCE(success_rate,0), COALESCE(fail_count,0), last_checked, created_at FROM proxies ORDER BY latency_ms ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*Proxy
	for rows.Next() {
		p := &Proxy{}
		rows.Scan(&p.URL, &p.Protocol, &p.Host, &p.Port, &p.Region, &p.LatencyMs, &p.Status, &p.Source, &p.SuccessRate, &p.FailCount, &p.LastChecked, &p.CreatedAt)
		proxies = append(proxies, p)
	}
	return proxies, nil
}

func (s *PGStore) DeleteProxy(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM proxies WHERE id = $1", id)
	return err
}

func (s *PGStore) SavePool(ctx context.Context, p *ProxyPool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO proxy_pools (name, pool_type, proxy_url, no_proxy, strict_proxy, is_active, test_status, last_error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET name=$1, proxy_url=$3, is_active=$6, test_status=$7, updated_at=NOW()`,
		p.Name, p.Type, p.ProxyURL, p.NoProxy, p.StrictProxy, p.IsActive, p.TestStatus, p.LastError, time.Now(), time.Now())
	return err
}

func (s *PGStore) GetPools(ctx context.Context) ([]*ProxyPool, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, name, COALESCE(pool_type,'http'), COALESCE(proxy_url,''), COALESCE(no_proxy,''), COALESCE(strict_proxy,false), is_active, COALESCE(test_status,'unknown'), COALESCE(last_error,''), created_at, updated_at FROM proxy_pools")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []*ProxyPool
	for rows.Next() {
		p := &ProxyPool{}
		rows.Scan(&p.ID, &p.Name, &p.Type, &p.ProxyURL, &p.NoProxy, &p.StrictProxy, &p.IsActive, &p.TestStatus, &p.LastError, &p.CreatedAt, &p.UpdatedAt)
		pools = append(pools, p)
	}
	return pools, nil
}

func (s *PGStore) DeletePool(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM proxy_pools WHERE id = $1", id)
	return err
}

func (s *PGStore) SaveSettings(ctx context.Context, ps *ProxySettings) error {
	data, _ := json.Marshal(ps)
	_, err := s.pool.Exec(ctx, "UPDATE proxy_settings SET enabled=$1, proxy_for_kiro=$2, proxy_for_codex=$3, proxy_for_openai=$4, proxy_for_github=$5, proxy_for_claude=$6, max_latency_ms=$7, scraper_interval_min=$8, webshare_api_key=$9, updated_at=NOW() WHERE id=1",
		ps.Enabled, ps.ProxyForKiro, ps.ProxyForCodex, ps.ProxyForOpenAI, ps.ProxyForGitHub, ps.ProxyForClaude, ps.MaxLatencyMs, ps.ScraperIntervalMin, ps.WebshareAPIKey)
	_ = data
	return err
}

func (s *PGStore) GetSettings(ctx context.Context) (*ProxySettings, error) {
	ps := &ProxySettings{}
	err := s.pool.QueryRow(ctx, "SELECT enabled, proxy_for_kiro, proxy_for_codex, proxy_for_openai, proxy_for_github, proxy_for_claude, max_latency_ms, scraper_interval_min, COALESCE(webshare_api_key,'') FROM proxy_settings WHERE id=1").
		Scan(&ps.Enabled, &ps.ProxyForKiro, &ps.ProxyForCodex, &ps.ProxyForOpenAI, &ps.ProxyForGitHub, &ps.ProxyForClaude, &ps.MaxLatencyMs, &ps.ScraperIntervalMin, &ps.WebshareAPIKey)
	if err != nil {
		return nil, err
	}
	return ps, nil
}
