package hexonet

import (
	"context"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
  "io"
	"fmt"
  "regexp"

	CL "github.com/hexonet/go-sdk/apiclient"
	"github.com/libdns/libdns"
)

type client struct {
	client *CL.APIClient
	mutex  sync.Mutex
}

func newClient(username, password string, debug io.Writer) (*client, error) {
	if debug == nil {
		debug = ioutil.Discard
	}
	c := CL.NewAPIClient()
	c.SetCredentials(username, password) //username, password
	resp := c.Login()
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("login failed, response code:%d description:%s\n", resp.GetCode(), resp.GetDescription())
	}

	return &client{client: c}, nil
}

func (c *client) getDNSEntries(ctx context.Context, zone string) ([]libdns.Record, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cmd := map[string]interface{}{
		"COMMAND": "QueryDNSZoneRRList",
		"dnszone": zone,
	}
	resp := c.client.Request(cmd)
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("getDNSEntries failed, response code:%d description:%s\n", resp.GetCode(), resp.GetDescription())
	}

	records := resp.GetRecords()
	recs := make([]libdns.Record, 0, len(records))
	for i := range records {
		rr := records[i].GetData()["RR"]
		name, ttlStr, typ, value := ParseRR(rr)
		ttl, _ := strconv.Atoi(ttlStr)
		recs = append(recs, libdns.Record{
			Type:  typ,
			Name:  name,
			Value: value,
			TTL:   time.Second * time.Duration(ttl),
		})
	}
	return recs, nil
}

func (c *client) addDNSEntry(ctx context.Context, zone string, records []libdns.Record) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cmd := map[string]interface{}{
		"COMMAND": "UpdateDNSZone",
		"dnszone": zone,
	}
	for i := range records {
		key := fmt.Sprintf("addrr%d", i)
		val := fmt.Sprintf("%s %d IN %s %s", records[i].Name, int(records[i].TTL.Seconds()), records[i].Type, records[i].Value)
		cmd[key] = val
	}
	resp := c.client.Request(cmd)
	if !resp.IsSuccess() {
		return fmt.Errorf("addDNSEntry failed, response code:%d description:%s\n", resp.GetCode(), resp.GetDescription())
	}
	return nil
}

func (c *client) removeDNSEntry(ctx context.Context, zone string, records []libdns.Record) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cmd := map[string]interface{}{
		"COMMAND": "UpdateDNSZone",
		"dnszone": zone,
	}
	for i := range records {
		key := fmt.Sprintf("delrr%d", i)
		val := fmt.Sprintf("%s %d IN %s %s", records[i].Name, int(records[i].TTL.Seconds()), records[i].Type, records[i].Value)
		cmd[key] = val
	}
	resp := c.client.Request(cmd)
	if !resp.IsSuccess() {
		return fmt.Errorf("removeDNSEntry failed, response code:%d description:%s\n", resp.GetCode(), resp.GetDescription())
	}

	return nil
}

func (c *client) updateDNSEntry(ctx context.Context, zone string, records []libdns.Record) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cmd := map[string]interface{}{
		"COMMAND": "UpdateDNSZone",
		"dnszone": zone,
	}
	for i := range records {
		key := fmt.Sprintf("rr%d", i)
		val := fmt.Sprintf("%s %d IN %s %s", records[i].Name, int(records[i].TTL.Seconds()), records[i].Type, records[i].Value)
		cmd[key] = val
	}
	resp := c.client.Request(cmd)
	if !resp.IsSuccess() {
		return fmt.Errorf("updateDNSEntry failed, response code:%d description:%s\n", resp.GetCode(), resp.GetDescription())
	}

	return nil
}

// rr format  "gomeing.com. 3600 IN NS ns1191.hexonet.net."
func ParseRR(rr string) (name, ttl, typ, value string) {
	pattern := `^(.+\.)\s+(\d+)\s+IN\s+(\w+)\s+(.+)$`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(rr)
	// extra
	name = matches[1]
	ttl = matches[2]
	typ = matches[3]
	value = matches[4]

	return
}
