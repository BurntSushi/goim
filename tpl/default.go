package tpl

import "strings"

type Object interface{}

type Attrs map[string]interface{}

type Formatted struct {
	O Object
	A Attrs
}

var Defaults = defaults

var defaults = strings.TrimSpace(`
{{ define "info_movie" }}
	{{ $full := .A.Full }}
	{{ printf "%s (%d)" .O.Title .O.Year }}
	{{ if .O.Tv }}{{ printf " (made for tv)" }}{{ end }}
	{{ if .O.Video }}{{ printf " (made for video)" }}{{ end }}


	{{ $dates := (Combine "full" $full "dates" .O.ReleaseDates) }}
	{{ template "bit_release_dates" $dates }}
{{ end }}

{{ define "info_tvshow" }}
	{{ $full := .A.Full }}
	{{ printf "%s (%d)" .O.Title .O.Year }}
	{{ if gt .O.YearStart 0 }}
		{{ printf " [%d-" .O.YearStart }}
		{{ if gt .O.YearEnd 0 }}
			{{ printf "%d]" .O.YearEnd }}
		{{ else }}
			{{ "]" }}
		{{ end }}
	{{ end }}
	{{ $seasons := .O.CountSeasons }}
	{{ $episodes := .O.CountEpisodes }}
	{{ if gt $seasons 0 }}
		{{ printf " (%d season(s) with %d episodes)" $seasons $episodes }}
	{{ end }}


	{{ $dates := (Combine "full" $full "dates" .O.ReleaseDates) }}
	{{ template "bit_release_dates" $dates }}
{{ end }}

{{ define "info_episode" }}
	{{ $tv := .O.Tvshow }}
	{{ $full := .A.Full }}
	{{ printf "%s (%d) (TV show: %s)" .O.Title .O.Year $tv.Title }}\
	{{ printf "Season %d, Episode %d" .O.Season .O.EpisodeNum }}


	{{ $dates := (Combine "full" $full "dates" .O.ReleaseDates) }}
	{{ template "bit_release_dates" $dates }}
{{ end }}

{{ define "bit_release_dates" }}
	{{ $full := .Full }}
	{{ if gt (len .dates) 0 }}
		Release dates
		=============
		{{ range $i, $date := .dates }}
			{{ if or $full (lt $i 5) }}
				{{ $date }}\
			{{ end }}
		{{ end }}
	{{ end }}
{{ end }}

{{ define "search_result" }}
	{{ printf "%3d. %-9s %s (%d)" .A.Index .O.Entity .O.Name .O.Year }}
	{{ if .O.Attrs }}
		{{ printf " %s" .O.Attrs }}
	{{ end }}
	{{ if gt .O.Similarity -1.0 }}
		{{ printf " (score: %0.2f)" .O.Similarity }}
	{{ end }}

{{ end }}
`)
