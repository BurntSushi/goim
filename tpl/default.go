package tpl

import (
	"strings"

	"github.com/BurntSushi/goim/imdb"
)

var tplDB *imdb.DB

// SetDB should be called by clients of this package to set the database to
// be used to query information.
//
// This unfortunately relies on global state, but this makes it more convenient
// and simpler to write templates.
func SetDB(db *imdb.DB) {
	tplDB = db
}

type Attrs map[string]interface{}

type Formatted struct {
	E interface{}
	A Attrs
}

var Defaults = defaults

var defaults = strings.TrimSpace(`
{{ define "running-times" }}

	{{ printf "Running times for %s" .E | underlined "=" }}

	{{ $runtimes := running_times .E }}
	{{ if eq 0 (len $runtimes) }}
		None found.

	{{ else }}
		{{ range $runtime := $runtimes }}
			{{ $runtime }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "release-dates" }}

	{{ printf "Release dates for %s" .E | underlined "=" }}

	{{ $dates := release_dates .E }}
	{{ if eq 0 (len $dates) }}
		None found.

	{{ else }}
		{{ range $date := $dates }}
			{{ $date }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "aka-titles" }}

	{{ printf "AKA titles for %s" .E | underlined "=" }}

	{{ $akas := aka_titles .E }}
	{{ if eq 0 (len $akas) }}
		None found.

	{{ else }}
		{{ range $aka := $akas }}
			{{ $aka }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "alternate-versions" }}

	{{ printf "Alternate versions for %s" .E | underlined "=" }}

	{{ $alts := alternate_versions .E }}
	{{ if eq 0 (len $alts) }}
		None found.

	{{ else }}
		{{ range $alt := $alts }}
			{{ $alt | wrap 80 }}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "color-info" }}

	{{ printf "Color information for %s" .E | underlined "=" }}

	{{ $colors := color_info .E }}
	{{ if eq 0 (len $colors) }}
		None found.

	{{ else }}
		{{ range $color := $colors }}
			{{ $color }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "mpaa" }}

	{{ printf "Rating for %s" .E | underlined "=" }}

	{{ $mpaa := mpaa .E }}
	{{ if $mpaa.Unrated }}
		None found.

	{{ else }}
		{{ $mpaa }}


	{{ end }}
{{ end }}

{{ define "sound-mix" }}

	{{ printf "Sound mixes for %s" .E | underlined "=" }}

	{{ $mixs := sound_mixes .E }}
	{{ if eq 0 (len $mixs) }}
		None found.

	{{ else }}
		{{ range $mix := $mixs }}
			{{ $mix }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "taglines" }}

	{{ printf "Taglines for %s" .E | underlined "=" }}

	{{ $tags := taglines .E }}
	{{ if eq 0 (len $tags) }}
		None found.

	{{ else }}
		{{ range $tag := $tags }}
			{{ $tag | wrap 80}}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "trivia" }}

	{{ printf "Trivia for %s" .E | underlined "=" }}

	{{ $trivias := trivia .E }}
	{{ if eq 0 (len $trivias) }}
		None found.

	{{ else }}
		{{ range $trivia := $trivias }}
			{{ $trivia | wrap 80}}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "genres" }}

	{{ printf "Genre tags for %s" .E | underlined "=" }}

	{{ $genres := genres .E }}
	{{ if eq 0 (len $genres) }}
		None found.

	{{ else }}
		{{ range $genre := $genres }}
			{{ $genre }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "goofs" }}

	{{ printf "Goofs for %s" .E | underlined "=" }}

	{{ $goofs := goofs .E }}
	{{ if eq 0 (len $goofs) }}
		None found.

	{{ else }}
		{{ range $goof := $goofs }}
			{{ $goof | wrap 80}}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "languages" }}

	{{ printf "Languages for %s" .E | underlined "=" }}

	{{ $langs := languages .E }}
	{{ if eq 0 (len $langs) }}
		None found.

	{{ else }}
		{{ range $lang := $langs }}
			{{ $lang }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "literature" }}

	{{ printf "Literature references for %s" .E | underlined "=" }}

	{{ $lits := literature .E }}
	{{ if eq 0 (len $lits) }}
		None found.

	{{ else }}
		{{ range $lit := $lits }}
			{{ $lit | wrap 80 }}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "locations" }}

	{{ printf "Locations for %s" .E | underlined "=" }}

	{{ $locs := locations .E }}
	{{ if eq 0 (len $locs) }}
		None found.

	{{ else }}
		{{ range $loc := $locs }}
			{{ $loc | wrap 80 }}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "links" }}

	{{ printf "Links for %s" .E | underlined "=" }}

	{{ $links := links .E | sort }}
	{{ if eq 0 (len $links) }}
		None found.

	{{ else }}
		{{ $ent := .E }}
		{{ range $lk := $links }}
			{{ if eq $ent.Type $lk.Entity.Type }}
				{{ printf "%s %s" $lk.Type $lk.Entity }}
			{{ else }}
				{{ printf "%s %s (%s)" $lk.Type $lk.Entity $lk.Entity.Type }}
			{{ end }}

		{{ end }}

	{{ end }}
{{ end }}

{{ define "plots" }}

	{{ printf "Plot summaries for %s" .E | underlined "=" }}

	{{ $plots := plots .E }}
	{{ if eq 0 (len $plots) }}
		None found.

	{{ else }}
		{{ range $plot := $plots }}
			{{ $plot.Entry | wrap 80 }}

			{{ printf "-- %s" $plot.By | wrap 80 }}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "quotes" }}

	{{ printf "Quotes for %s" .E | underlined "=" }}

	{{ $quotes := quotes .E }}
	{{ if eq 0 (len $quotes) }}
		None found.

	{{ else }}
		{{ range $quote := $quotes }}
			{{ range $line := lines $quote.Entry }}
				{{ wrap 80 $line }}\
			{{ end }}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "rank" }}

	{{ printf "User rank for %s" .E | underlined "=" }}

	{{ $rank := rank .E }}
	{{ if $rank.Unrated }}
		None found.

	{{ else }}
		{{ $rank }}


	{{ end }}
{{ end }}

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
	{{ printf "%3d. %-8s" .A.Index .E.Entity }}
	{{ if gt .E.Similarity -1.0 }}
		{{ printf " (%0.2f) " .E.Similarity }}
	{{ end }}
	{{ printf " %s" .E.Name }}
	{{ if and (gt .E.Year 0) (ne .E.Entity.String "tvshow") }}
		{{ printf " (%d)" .E.Year }}
	{{ end }}
	{{ if .E.Attrs }}
		{{ printf " %s" .E.Attrs }}
	{{ end }}
	{{ if not .E.Rating.Unrated }}
		{{ printf " (rank: %d/100)" .E.Rating.Rank }}
	{{ end }}
	{{ if .E.Credit.Valid }}
		{{ if gt (len .E.Credit.Character) 0 }}
			{{ printf " [%s]" .E.Credit.Character }}
		{{ end }}
		{{ if gt .E.Credit.Position 0 }}
			{{ printf " <%d>" .E.Credit.Position }}
		{{ end }}
	{{ end }}

{{ end }}
`)
