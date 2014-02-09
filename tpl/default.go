package tpl

import "strings"

type Object interface{}

type Attrs map[string]interface{}

type Formatted struct {
	X Object
	A Attrs
}

var Defaults = defaults

var defaults = strings.TrimSpace(`
{{ define "info_movie" }}
	{{ printf "%s (%d)" .X.Title .X.Year }}
	{{ if .X.Tv }}{{ printf " (made for tv)" }}{{ end }}
	{{ if .X.Video }}{{ printf " (made for video)" }}{{ end }}

	{{ template "info_media_details" . }}
{{ end }}

{{ define "info_tvshow" }}
	{{ $full := .A.Full }}
	{{ printf "%s (%d)" .X.Title .X.Year }}
	{{ if gt .X.YearStart 0 }}
		{{ printf " [%d-" .X.YearStart }}
		{{ if gt .X.YearEnd 0 }}
			{{ printf "%d]" .X.YearEnd }}
		{{ else }}
			{{ "]" }}
		{{ end }}
	{{ end }}
	{{ $seasons := .X.CountSeasons }}
	{{ $episodes := .X.CountEpisodes }}
	{{ if gt $seasons 0 }}
		{{ printf " (%d season(s) with %d episodes)" $seasons $episodes }}
	{{ end }}

	{{ template "info_media_details" . }}
{{ end }}

{{ define "info_episode" }}
	{{ $tv := .X.Tvshow }}
	{{ $full := .A.Full }}
	{{ printf "%s (%d) (TV show: %s)" .X.Title .X.Year $tv.Title }}\
	{{ printf "Season %d, Episode %d" .X.Season .X.EpisodeNum }}

	{{ template "info_media_details" . }}
{{ end }}

{{ define "info_media_details" }}
	{{ $full := .A.Full }}
	{{ template "bit_mpaa" .X.MPAARating }}
	{{ template "bit_runtime" .X.RunningTimes }}
	{{ template "bit_release_date" .X.ReleaseDates }}
	{{ template "bit_plot" .X.Plots }}
	{{ if $full }}

		{{ template "bit_aka_titles" .X.AkaTitles }}
		{{ template "bit_alternate_versions" .X.AlternateVersions }}
		{{ template "bit_runtimes" .X.RunningTimes }}
		{{ template "bit_release_dates" .X.ReleaseDates }}
		{{ template "bit_plots" .X.Plots }}
		{{ template "bit_quotes" .X.Quotes }}
		{{ template "bit_color_info" .X.ColorInfos }}
		{{ template "bit_sound_mixes" .X.SoundMixes }}
	{{ else }}

	{{ end }}
{{ end }}

{{ define "bit_mpaa" }}
	{{ if not .Unrated }}

		{{ printf "Rating: %s" . | wrap 80 }}
	{{ end }}
{{ end }}

{{ define "bit_runtime" }}
	{{ if gt (len .) 0 }}

		{{ printf "Running time: %s" (index . 0) }}
	{{ end }}
{{ end }}

{{ define "bit_release_date" }}
	{{ if gt (len .) 0 }}

		{{ printf "Release date: %s" (index . 0) }}
	{{ end }}
{{ end }}

{{ define "bit_plot" }}
	{{ if gt (len .) 0 }}


		Plot
		====
		{{ index . 0 | wrap 80 }}
	{{ end }}
{{ end }}

{{ define "bit_aka_titles" }}
	{{ if gt (len .) 0 }}

		Alternate titles
		================
		{{ range $aka := . }}
			{{ $aka }}\
		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_alternate_versions" }}
	{{ if gt (len .) 0 }}

		Alternate versions
		=================={{ range $alt := . }}

			{{ wrap 80 $alt }}

		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_runtimes" }}
	{{ if gt (len .) 0 }}

		Running times
		=============
		{{ range $time := . }}
			{{ $time }}\
		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_release_dates" }}
	{{ if gt (len .) 0 }}

		Release dates
		=============
		{{ range $date := . }}
			{{ $date }}\
		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_plots" }}
	{{ if gt (len .) 0 }}

		Plot
		===={{ range $plot := . }}

			{{ wrap 80 $plot }}

		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_quotes" }}
	{{ if gt (len .) 0 }}

		Quotes
		======{{ range $quote := . }}

			{{ range $line := lines $quote }}
				{{ wrap 80 $line }}\
			{{ end }}
		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_color_info" }}
	{{ if gt (len .) 0 }}

		Color info
		==========
		{{ range $info := . }}
			{{ $info }}\
		{{ end }}
	{{ end }}
{{ end }}

{{ define "bit_sound_mixes" }}
	{{ if gt (len .) 0 }}

		Sound Mixes
		===========
		{{ range $mix := . }}
			{{ $mix }}\
		{{ end }}
	{{ end }}
{{ end }}

{{ define "search_result" }}
	{{ printf "%3d. %-8s" .A.Index .X.Entity }}
	{{ if gt .X.Similarity -1.0 }}
		{{ printf " (%0.2f) " .X.Similarity }}
	{{ end }}
	{{ printf " %s" .X.Name }}
	{{ if and (gt .X.Year 0) (ne .X.Entity.String "tvshow") }}
		{{ printf " (%d)" .X.Year }}
	{{ end }}
	{{ if .X.Attrs }}
		{{ printf " %s" .X.Attrs }}
	{{ end }}
	{{ if not .X.Rating.Unrated }}
		{{ printf " (rank: %d/100)" .X.Rating.Rank }}
	{{ end }}
	{{ if .X.Credit.Valid }}
		{{ if gt (len .X.Credit.Character) 0 }}
			{{ printf " [%s]" .X.Credit.Character }}
		{{ end }}
		{{ if gt .X.Credit.Position 0 }}
			{{ printf " <%d>" .X.Credit.Position }}
		{{ end }}
	{{ end }}

{{ end }}
`)
