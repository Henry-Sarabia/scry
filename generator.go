package refind

import (
	"github.com/pkg/errors"
)

var (
	errNilGen       = errors.New("cannot initialize new generator using nil interface")
	errRangeInvalid = errors.New("integer parameter is out of range")
)

type generator struct {
	serv MusicService
	rec  Recommender
}

func New(serv MusicService, rec Recommender) (*generator, error) {
	if serv == nil || rec == nil {
		return nil, errNilGen
	}

	return &generator{serv: serv, rec: rec}, nil
}

type MusicService interface {
	TopArtists() ([]Artist, error)
	RecentTracks() ([]Track, error)
}

type Recommender interface {
	Recommendations(int, []Seed) ([]Track, error)
}

func (g generator) Tracklist(n int) ([]Track, error) {
	if n <= 0 {
		return nil, errRangeInvalid
	}

	tracks, err := g.serv.RecentTracks()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch recent tracks")
	}

	var sds []Seed
	for _, t := range tracks {
		sd, err := t.Seed()
		if err != nil {
			return nil, errors.Wrap(err, "one or more tracks are invalid seeds")
		}
		sds = append(sds, sd)
	}

	recs, err := g.rec.Recommendations(n, sds)
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch recommendations")
	}

	top, err := g.serv.TopArtists()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch top artists")
	}

	f := filter(recs, toMap(top))

	return f, nil
}

func (g generator) LimitedTracklist(n int) ([]Track, error) {
	if n <= 0 {
		return nil, errRangeInvalid
	}

	top, err := g.serv.TopArtists()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch top artists")
	}

	var sds []Seed
	for _, t := range top {
		sd, err := t.Seed()
		if err != nil {
			return nil, errors.Wrap(err, "one or more artists are invalid seeds")
		}
		sds = append(sds, sd)
	}

	recs, err := g.rec.Recommendations(n, sds)
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch recommendations")
	}

	f := filter(recs, toMap(top))

	return f, nil
}

func toMap(prev []Artist) map[string]Artist {
	if len(prev) == 0 {
		return nil
	}

	curr := make(map[string]Artist)
	for _, p := range prev {
		curr[p.Name] = p
	}

	return curr
}

func filter(prev []Track, rmv map[string]Artist) []Track {
	if len(prev) == 0 {
		return nil
	}

	if len(rmv) == 0 {
		return prev
	}

	var curr []Track
	for _, p := range prev {
		if _, ok := rmv[p.Artist.Name]; !ok {
			curr = append(curr, p)
			rmv[p.Artist.Name] = p.Artist
		}
	}

	return curr
}
