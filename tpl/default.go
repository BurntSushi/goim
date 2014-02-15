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

// Attrs represents a template-specific map of attributes.
type Attrs map[string]interface{}

// Args is the value that is passed to all templates.
type Args struct {
	// E is a value that satisfies the imdb.Entity interface, except when
	// showing search results where it corresponds to a search.Result value.
	E interface{}

	// Attrs is a map of attributes that is template specific. For example,
	// the "search_result" template is given an index that represents the
	// position of the result in the search results.
	A Attrs
}

var Defaults = defaults

var defaults = strings.TrimSpace(`
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

{{ define "short_movie" }}

	{{ printf "%s (%d)" .E.Title .E.Year | underlined "=" }}

	{{ if .E.Tv }}
		{{ "(made for tv)" }}
	{{ end }}
	{{ if .E.Video }}
		{{ "(made for video)" }}
	{{ end }}


{{ end }}

{{ define "short_tvshow" }}

	{{ printf "%s (%d)" .E.Title .E.Year | underlined "=" }}

	{{ if gt .E.YearStart 0 }}
		{{ printf "Years active: %d-" .E.YearStart }}
		{{ if gt .E.YearEnd 0 }}
			{{ printf "%d" .E.YearEnd }}
		{{ else }}
			{{ "????" }}
		{{ end }}

	{{ end }}
	{{ $seasons := count_seasons .E }}
	{{ $episodes := count_episodes .E }}
	{{ if gt $seasons 0 }}
		{{ printf "%d season(s) with %d episodes" $seasons $episodes }}

	{{ end }}

{{ end }}

{{ define "short_episode" }}

	{{ $tv := tvshow .E }}
	{{ $tvname := printf "(TV show: %s (%d))" $tv.Title $tv.Year }}
	{{ printf "%s (%d) %s" .E.Title .E.Year $tvname | underlined "=" }}

	{{ if and (gt .E.Season 0) (gt .E.EpisodeNum 0) }}
		{{ printf "Season %d, Episode %d" .E.Season .E.EpisodeNum }}

	{{ end }}

{{ end }}

{{ define "short_actor" }}

	{{ .E.Name | underlined "=" }}


{{ end }}

{{ define "short_media_details" }}
{{ end }}

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
`)
