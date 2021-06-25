/*
Copyright © 2020 David Muckle <dvdmuckle@dvdmuckle.xyz>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dvdmuckle/spc/cmd/helper"
	"github.com/golang/glog"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Seach for and play a track, album, or playlist",
	Long: `Seach takes two arguments: the search type, and the query.
Search type can be an album, a track, or a playlist, with the rest of the arguments
making up the search query. For example:

	spc search album moving pictures
	spc search track tom sawyer
	spc search playlist prog monsters

If a track is queried for, additional similar songs will be queued up.

More advanced options are availble for the search query. For this,
please see https://pkg.go.dev/github.com/zmb3/spotify?tab=doc#Client.Search`,
	Run: func(cmd *cobra.Command, args []string) {
		helper.SetClient(&conf)
		var searchType string
		if len(args) < 2 {
			fmt.Println("Please search for a track, album, playlist, or artist")
			os.Exit(1)
		}
		if args[0] == "track" || args[0] == "album" || args[0] == "playlist" || args[0] == "artist" {
			searchType = args[0]
		} else {
			fmt.Println("Please search for a track, album, playlist, or artist")
			os.Exit(1)
		}
		searchTerm := strings.Join(args[1:], " ")
		var searchResults *spotify.SearchResult
		var err error
		switch searchType {
		case "track":
			searchResults, err = conf.Client.Search(searchTerm, spotify.SearchTypeTrack)
		case "album":
			searchResults, err = conf.Client.Search(searchTerm, spotify.SearchTypeAlbum)
		case "playlist":
			searchResults, err = conf.Client.Search(searchTerm, spotify.SearchTypePlaylist)
		case "artist":
			searchResults, err = conf.Client.Search(searchTerm, spotify.SearchTypeArtist)
		}
		if err != nil {
			glog.Fatal(err)
		}
		toPlay := fuzzySearchResults(*searchResults, searchType)
		var opts spotify.PlayOptions
		opts.DeviceID = &conf.DeviceID
		switch searchType {
		case "track":
			//This lines up some songs to play after the search result plays
			regexID := regexp.MustCompile(`(spotify:track:)(.*)`)
			trackID := spotify.ID(regexID.FindStringSubmatch(string(toPlay))[2])
			seeds := spotify.Seeds{Tracks: []spotify.ID{trackID}}
			recommends, err := conf.Client.GetRecommendations(seeds, nil, nil)
			if err != nil {
				glog.Fatal(err)
			}
			var recommendURIs []spotify.URI
			for _, track := range recommends.Tracks {
				recommendURIs = append(recommendURIs, track.URI)
			}
			opts.URIs = append(opts.URIs, toPlay)
			opts.URIs = append(opts.URIs, recommendURIs...)
		case "album", "playlist":
			opts.PlaybackContext = &toPlay
		}
		//If a user tries to play a track with shuffle on,
		//they'll instead get the related tracks first
		if err := conf.Client.Shuffle(false); err != nil {
			glog.Fatal(err)
		}
		if err := conf.Client.PlayOpt(&opts); err != nil {
			glog.Fatal(err)
		}
	},
	ValidArgs: []string{"track", "album", "playlist", "artist"},
}

func fuzzySearchResults(results spotify.SearchResult, searchType string) spotify.URI {
	var idx int
	var err error
	switch searchType {
	case "track":
		idx, err = fuzzyfinder.Find(
			results.Tracks.Tracks,
			func(i int) string {
				return fmt.Sprintf("%s - %s - %s", results.Tracks.Tracks[i].Artists[0].Name,
					results.Tracks.Tracks[i].Name,
					results.Tracks.Tracks[i].Album.Name)
			})
	case "album":
		idx, err = fuzzyfinder.Find(
			results.Albums.Albums,
			func(i int) string {
				return fmt.Sprintf("%s - %s", results.Albums.Albums[i].Name,
					results.Albums.Albums[i].Artists[0].Name)
			})
	case "playlist":
		idx, err = fuzzyfinder.Find(
			results.Playlists.Playlists,
			func(i int) string {
				return fmt.Sprintf("%s - %s", results.Playlists.Playlists[i].Name,
					results.Playlists.Playlists[i].Owner.DisplayName)
			})
		case "artist":
			idx, err = fuzzyfinder.Find(
				results.Artists.Artists,
				func(i int) string {
					return fmt.Sprintf("%s", results.Artists.Artists[i].Name)

				})
	}
	if err != nil {
		if err.Error() == "abort" {
			fmt.Println("Aborted search")
			os.Exit(0)
		}
		glog.Fatal(err)
	}
	switch searchType {
	case "track":
		return results.Tracks.Tracks[idx].URI
	case "album":
		return results.Albums.Albums[idx].URI
	case "playlist":
		return results.Playlists.Playlists[idx].URI
	case "artist":
		return results.Artists.Artists[idx].URI
		//func (c *Client) GetArtistsTopTracks(artistID ID, country string) ([]FullTrack, error) {

		}
	}
	//The code should never get here because of our check of
	//search types earlier, this is just to make the compiler
	//happy
	return ""
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
