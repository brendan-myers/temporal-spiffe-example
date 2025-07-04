package spiffeauth

import (
	"context"
	"fmt"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"sync"
	"time"
)

type SpiffeConfig struct {
	SocketPath string
	SpiffeID   spiffeid.ID
	Audience   string
}

func (s SpiffeConfig) GenerateSvid(ctx context.Context) (*jwtsvid.SVID, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	clientOptions := workloadapi.WithClientOptions(workloadapi.WithAddr(s.SocketPath))

	jwtSource, err := workloadapi.NewJWTSource(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to create JWTSource: %w", err)
	}
	defer jwtSource.Close()

	svid, err := jwtSource.FetchJWTSVID(ctx, jwtsvid.Params{
		Subject:  s.SpiffeID,
		Audience: s.Audience,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch SVID: %w", err)
	}

	return svid, nil
}

type SpiffeHeadersProvider struct {
	Config   SpiffeConfig
	Svid     *jwtsvid.SVID
	SvidLock sync.RWMutex
}

func (s *SpiffeHeadersProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	var err error

	s.SvidLock.RLock()
	svid := s.Svid
	s.SvidLock.RUnlock()

	if svid == nil || time.Now().After(svid.Expiry) {
		svid, err = s.Config.GenerateSvid(ctx)
		if err != nil {
			return nil, err
		}
		s.SvidLock.Lock()
		s.Svid = svid
		s.SvidLock.Unlock()
	}

	return map[string]string{"Authorization": "Bearer " + svid.Marshal()}, nil
}
