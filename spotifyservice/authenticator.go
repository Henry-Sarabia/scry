package spotifyservice

import (
	"github.com/Henry-Sarabia/blank"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

func Authenticator(URI string) (*spotify.Authenticator, error) {
	if blank.Is(URI) {
		return nil, errors.New("URI is blank")
	}

	auth := spotify.NewAuthenticator(
		URI,
		spotify.ScopePlaylistModifyPublic,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserTopRead,
		spotify.ScopeUserReadRecentlyPlayed,
	)

	return &auth, nil
}
