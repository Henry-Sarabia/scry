package scry

type Buffer struct {
	serv MusicService
	user *User
	artists []Artist
	tracks []Track
}

func NewBuffer(ms MusicService) (*Buffer, error) {
	return &Buffer{serv: ms}, nil
}

func (b Buffer) CurrentUser() (*User, error) {
	var u *User
	if b.user != nil {
		u = b.user
	}

	var err error
	u, err = b.serv.CurrentUser()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (b Buffer) TopArtists() ([]Artist, error) {
	var top []Artist
	if b.artists != nil {
		top = b.artists
	}

	var err error
	top, err = b.serv.TopArtists()
	if err != nil {
		return nil, err
	}

	return top, nil
}

func (b Buffer) RecentTracks() ([]Track, error) {
	var rec []Track
	if b.tracks != nil {
		rec = b.tracks
	}

	var err error
	rec, err = b.serv.RecentTracks()
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (b Buffer) Playlist(name string, list []Track) (*Playlist, error) {
	return b.serv.Playlist(name, list)
}
