package spotify

import (
	"github.com/Henry-Sarabia/refind"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

const (
	popTarget      int  = 40
	popMax         int  = 50
	publicPlaylist bool = true
)

var (
	errNilClient = errors.New("client pointer is nil")
	errInvalidData = errors.New("invalid or empty data returned")
	errMissingSeeds = errors.New("missing seed input")
)
type clienter interface {
	artister
	tracker
	recenter
	recommender
	playlister
}

type artister interface {
	CurrentUsersTopArtists() (*spotify.FullArtistPage, error)
}

type tracker interface {
	CurrentUsersTopTracks() (*spotify.FullTrackPage, error)
}

type recenter interface {
	PlayerRecentlyPlayed() ([]spotify.RecentlyPlayedItem, error)
}

type recommender interface {
	GetRecommendations(spotify.Seeds, *spotify.TrackAttributes, *spotify.Options) (*spotify.Recommendations, error)
}

type playlister interface {
	AddTracksToPlaylist(spotify.ID, ...spotify.ID) (string, error)
	CreatePlaylistForUser(string, string, string, bool) (*spotify.FullPlaylist, error)
	CurrentUser() (*spotify.PrivateUser, error)
}

type service struct {
	art artister
	track tracker
	rec recenter
	recom recommender
	play playlister
}

func New(c clienter) (*service, error) {
	if c == nil {
		return nil, errNilClient
	}
	s := &service{
		art: c,
		track: c,
		rec: c,
		recom: c,
		play: c,
	}

	return s, nil
}

func (s *service) TopArtists() ([]refind.Artist, error) {
	top, err := s.art.CurrentUsersTopArtists()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch top artists")
	}

	if top == nil {
		return nil, errInvalidData
	}

	return parseArtists(top.Artists...), nil
}

func (s *service) topTracks() ([]refind.Track, error) {
	top, err := s.track.CurrentUsersTopTracks()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch top tracks")
	}

	if top == nil {
		return nil, errInvalidData
	}

	return parseFullTracks(top.Tracks...), nil
}

func (s *service) RecentTracks() ([]refind.Track, error) {
	rec, err := s.rec.PlayerRecentlyPlayed()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch recently played tracks")
	}

	if len(rec) <= 0 {
		return nil, errInvalidData
	}

	var t []refind.Track
	for _, r := range rec {
		t = append(t, parseTrack(r.Track))
	}

	return t, nil
}

func (s *service) Recommendations(seeds []refind.Seed) ([]refind.Track, error) {
	if len(seeds) <= 0 {
		return nil, errMissingSeeds
	}

	sds, err := parseSeeds(seeds)
	if err != nil {
		return nil, err
	}

	var tracks []refind.Track
	attr := spotify.NewTrackAttributes().TargetPopularity(popTarget).MaxPopularity(popMax)

	for _, sd := range sds {
		recs, err := s.recom.GetRecommendations(sd, attr, nil)
		if err != nil {
			return nil, errors.Wrap(err, "cannot fetch recommendations")
		}

		if recs == nil {
			return nil, errInvalidData
		}

		t := parseSimpleTracks(recs.Tracks...)
		tracks = append(tracks, t...)
	}

	return tracks, nil
}

func (s *service) Playlist(name string, list []refind.Track) (*spotify.FullPlaylist, error) {
	u, err := s.play.CurrentUser()
	if err != nil {
		return nil, errors.Wrap(err, "cannot fetch user")
	}

	pl, err := s.play.CreatePlaylistForUser(u.ID, name, "description", publicPlaylist)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create playlist")
	}

	var IDs []spotify.ID
	for _, t := range list {
		IDs = append(IDs, spotify.ID(t.ID))
	}

	_, err = s.play.AddTracksToPlaylist(pl.ID, IDs...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot add tracks to playlist")
	}

	return pl, nil
}
