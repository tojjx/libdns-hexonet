package hexonet

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Hexonet.
type Provider struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`

	// Debug - can set this to stdout or stderr to dump
	// debugging information about the API interaction with
	// hexonet.  This will dump your auth token in plain text
	// so be careful.
	Debug string `json:"debug,omitempty"`

	mu sync.Mutex
	c  *client
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	recs, err := c.getDNSEntries(ctx, zone)

	return recs, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	err = c.addDNSEntry(ctx, zone, records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	err = c.updateDNSEntry(ctx, zone, records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	c, err := p.client()
	if err != nil {
		return nil, err
	}

	err = c.removeDNSEntry(ctx, zone, records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (p *Provider) client() (*client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.c == nil {
		var err error
		var debug io.Writer
		switch strings.ToLower(p.Debug) {
		case "stdout", "yes", "true", "1":
			debug = os.Stdout
		case "stderr":
			debug = os.Stderr
		}
		p.c, err = newClient(p.Username, p.Password, debug)
		if err != nil {
			return nil, err
		}
	}
	return p.c, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
