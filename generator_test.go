package refind

import (
	"github.com/pkg/errors"
	"reflect"
	"testing"
)

var (
	testErrFetchArtists = errors.New("cannot fetch artists")
	testErrFetchTracks = errors.New("cannot fetch tracks")
	testErrFetchRecommendations = errors.New("cannot fetch recommendation tracks")
)

const testTotal int = 30

type fakeMusicService struct {
	artists []Artist
	artistErr error
	tracks []Track
	trackErr error
}

func (f fakeMusicService) TopArtists() ([]Artist, error) {
	return f.artists, f.artistErr
}

func (f fakeMusicService) RecentTracks() ([]Track, error) {
	return f.tracks, f.trackErr
}

type fakeRecommender struct{
	tracks []Track
	err error
}

func (f fakeRecommender) Recommendations(int, []Seed) ([]Track, error) {
	return f.tracks, f.err
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		serv MusicService
		rec Recommender
		wantGen *generator
		wantErr error
	}{
		{
			"Valid interfaces",
			fakeMusicService{},
			fakeRecommender{},
			&generator{serv: fakeMusicService{}, rec: fakeRecommender{}},
			nil,
		},
		{
			"Nil MusicService",
			nil,
			fakeRecommender{},
			nil,
			errNilGen,
		},
		{
			"Nil Recommender",
			fakeMusicService{},
			nil,
			nil,
			errNilGen,
		},
		{
			"Nil interfaces",
			nil,
			nil,
			nil,
			errNilGen,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := New(test.serv, test.rec)
			if err != test.wantErr {
				t.Errorf("got: <%v>, want: <%v>", err, test.wantErr)
			}

			if !reflect.DeepEqual(got, test.wantGen) {
				t.Errorf("got: <%v>, want: <%v>", got, test.wantGen)
			}
		})
	}
}

func TestGenerator_Tracklist(t *testing.T) {
	tests := []struct {
		name string
		gen generator
		total int
		wantList []Track
		wantErr error
	}{
		{
			"Valid responses",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
					tracks: []Track {
						{ID: "10", Name: "baz", Artist: Artist{ID: "0", Name: "foo"}},
					},
					trackErr: nil,
				},
				rec: fakeRecommender{
					tracks: []Track {
						{ID: "21", Name: "qux"},
					},
					err: nil,
				},
			},
			testTotal,
			[]Track{
				{ID: "21", Name: "qux"},
			},
			nil,
		},
		{
			"Empty top artists response",
			generator{
				serv: fakeMusicService{
					artists: nil,
					artistErr: testErrFetchArtists,
					tracks: []Track {
						{ID: "10", Name: "baz", Artist: Artist{ID: "0", Name: "foo"}},
					},
					trackErr: nil,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: nil,
				},
			},
			testTotal,
			nil,
			testErrFetchArtists,
		},
		{
			"Empty recent tracks response",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
					tracks: nil,
					trackErr: testErrFetchTracks,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: nil,
				},
			},
			testTotal,
			nil,
			testErrFetchTracks,
		},
		{
			"Empty recommendation response",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
					tracks: []Track {
						{ID: "10", Name: "baz", Artist: Artist{ID: "0", Name: "foo"}},
					},
					trackErr: nil,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: testErrFetchRecommendations,
				},
			},
			testTotal,
			nil,
			testErrFetchRecommendations,
		},
		{
			"Invalid track seed",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
					tracks: []Track {
						{ID: "", Name: "baz", Artist: Artist{ID: "0", Name: "foo"}},
					},
					trackErr: nil,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: nil,
				},
			},
			testTotal,
			nil,
			errTrackSeed,
		},
		{
			"n out of range",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
					tracks: []Track {
						{ID: "10", Name: "baz", Artist: Artist{ID: "0", Name: "foo"}},
					},
					trackErr: nil,
				},
				rec: fakeRecommender{
					tracks: []Track {
						{ID: "21", Name: "qux"},
					},
					err: nil,
				},
			},
			0,
			nil,
			errRangeInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list, err := test.gen.Tracklist(test.total)
			if !reflect.DeepEqual(errors.Cause(err), test.wantErr) {
				t.Errorf("got: <%v>, want: <%v>", errors.Cause(err), test.wantErr)
			}

			if !reflect.DeepEqual(list, test.wantList) {
				t.Errorf("got: <%v>, want: <%v>", list, test.wantList)
			}
		})
	}
}

func TestGenerator_LimitedTracklist(t *testing.T) {
	tests := []struct {
		name string
		gen generator
		total int
		wantList []Track
		wantErr error
	}{
		{
			"Valid responses",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
				},
				rec: fakeRecommender{
					tracks: []Track {
						{ID: "10", Name: "qux"},
					},
					err: nil,
				},
			},
			testTotal,
			[]Track{
				{ID: "10", Name: "qux"},
			},
			nil,
		},
		{
			"Empty top artists response",
			generator{
				serv: fakeMusicService{
					artists: nil,
					artistErr: testErrFetchArtists,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: nil,
				},
			},
			testTotal,
			nil,
			testErrFetchArtists,
		},
		{
			"Empty recommendation response",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: testErrFetchRecommendations,
				},
			},
			testTotal,
			nil,
			testErrFetchRecommendations,
		},
		{
			"Invalid artist seed",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "", Name: "bar"},
					},
					artistErr: nil,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: nil,
				},
			},
			testTotal,
			nil,
			errArtistSeed,
		},
		{
			"n out of range",
			generator{
				serv: fakeMusicService{
					artists: []Artist{
						{ID: "0", Name: "foo"},
						{ID: "1", Name: "bar"},
					},
					artistErr: nil,
				},
				rec: fakeRecommender{
					tracks: nil,
					err: nil,
				},
			},
			0,
			nil,
			errRangeInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			list, err := test.gen.LimitedTracklist(test.total)
			if !reflect.DeepEqual(errors.Cause(err), test.wantErr) {
				t.Errorf("got: <%v>, want: <%v>", errors.Cause(err), test.wantErr)
			}

			if !reflect.DeepEqual(list, test.wantList) {
				t.Errorf("got: <%v>, want: <%v>", list, test.wantList)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	tests := []struct {
		name string
		prev []Artist
		want map[string]Artist
	}{
		{
			"Nil artist slice",
			nil,
			nil,
		},
		{
			"Empty artist slice",
			[]Artist{},
			nil,
		},
		{
			"Single artist slice",
			[]Artist{
				{ID: "1", Name: "one"},
			},
			map[string]Artist{
				"one": {ID: "1", Name: "one"},
			},
		},
		{
			"Multiple artists slice",
			[]Artist{
				{ID: "1", Name: "one"},
				{ID: "2", Name: "two"},
				{ID: "3", Name: "three"},
				{ID: "4", Name: "four"},
				{ID: "5", Name: "five"},
				{ID: "6", Name: "six"},
			},
			map[string]Artist{
				"one": {ID: "1", Name: "one"},
				"two": {ID: "2", Name: "two"},
				"three": {ID: "3", Name: "three"},
				"four": {ID: "4", Name: "four"},
				"five": {ID: "5", Name: "five"},
				"six": {ID: "6", Name: "six"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := toMap(test.prev)
			if !reflect.DeepEqual(got, test.want){
				t.Errorf("got: <%v>, want: <%v>", got, test.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name string
		prev []Track
		rmv map[string]Artist
		want []Track
	}{
		{
			"Nil tracks and nil artist map",
			nil,
			nil,
			nil,
		},
		{
			"Multiple tracks and nil artist map",
			[]Track{
				{ID: "1", Name: "foo", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "2", Name: "bar", Artist: Artist{ID: "12", Name: "garply"}},
				{ID: "3", Name: "baz", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "4", Name: "qux", Artist: Artist{ID: "14", Name: "fred"}},
				{ID: "5", Name: "quux", Artist: Artist{ID: "14", Name: "fred"}},
			},
			nil,
			[]Track{
				{ID: "1", Name: "foo", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "2", Name: "bar", Artist: Artist{ID: "12", Name: "garply"}},
				{ID: "3", Name: "baz", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "4", Name: "qux", Artist: Artist{ID: "14", Name: "fred"}},
				{ID: "5", Name: "quux", Artist: Artist{ID: "14", Name: "fred"}},
			},
		},
		{
			"Nil tracks and multiple artist map",
			nil,
			map[string]Artist{
				"grault": {ID: "11", Name: "grault"},
				"fred": {ID: "14", Name: "fred"},
			},
			nil,
		},
		{
			"Multiple tracks and single artist map",
			[]Track{
				{ID: "1", Name: "foo", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "2", Name: "bar", Artist: Artist{ID: "12", Name: "garply"}},
				{ID: "3", Name: "baz", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "4", Name: "qux", Artist: Artist{ID: "14", Name: "fred"}},
				{ID: "5", Name: "quux", Artist: Artist{ID: "14", Name: "fred"}},
			},
			map[string]Artist{
				"garply": {ID: "12", Name: "garply"},
			},
			[]Track{
				{ID: "1", Name: "foo", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "4", Name: "qux", Artist: Artist{ID: "14", Name: "fred"}},
			},
		},
		{
			"Multiple tracks with same artist and artist map with that same artist",
			[]Track{
				{ID: "1", Name: "foo", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "2", Name: "bar", Artist: Artist{ID: "12", Name: "garply"}},
				{ID: "3", Name: "baz", Artist: Artist{ID: "11", Name: "grault"}},
				{ID: "4", Name: "qux", Artist: Artist{ID: "14", Name: "fred"}},
				{ID: "5", Name: "quux", Artist: Artist{ID: "14", Name: "fred"}},
			},
			map[string]Artist{
				"grault": {ID: "11", Name: "grault"},
				"fred": {ID: "14", Name: "fred"},
			},
			[]Track{
				{ID: "2", Name: "bar", Artist: Artist{ID: "12", Name: "garply"}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filter(test.prev, test.rmv)

			if !reflect.DeepEqual(got, test.want){
				t.Errorf("got: <%v>, want: <%v>", got, test.want)
			}
		})
	}
}
