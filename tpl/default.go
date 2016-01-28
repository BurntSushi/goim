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

// Defaults is a string containing the default templates. It corresponds
// precisely to the content in command.tpl.
var Defaults string = defaults

var defaults = strings.TrimSpace(`
{{/*
The templates below are Go templates from the standard library text/template
package. Documentation is here: http://golang.org/pkg/text/template/

Documentation specific to the data available in each template for Goim is
at http://godoc.burntsushi.net/pkg/github.com/BurntSushi/goim/tpl/

Note that in an effort to control whitespace, lines ending with '}}' are
completely ignored. Lines ending with '}}\' are not ignored. Three or more
successive new lines (LF) are replaced with two new lines (LF).
*/}}

{{ define "search_result" }}
	{{ printf "%3d. %-8s" .A.Index .E.Entity }}
	{{ if gtf .E.Similarity -1.0 }}
		{{ printf " (%0.2f) " .E.Similarity }}
	{{ end }}
	{{ printf " %s" .E.Name }}
	{{ if and (gt .E.Year 0) (ne .E.Entity.String "tvshow") }}
		{{ printf " (%d)" .E.Year }}
	{{ end }}
	{{ if .E.Attrs }}
		{{ printf " %s" .E.Attrs }}
	{{ end }}
	{{ if not .E.Rank.Unranked }}
		{{ printf " (rank: %d/100, votes: %d)" .E.Rank.Rank .E.Rank.Votes }}
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

{{ define "rename_movie" }}
	{{ if gt .E.Year 0 }}
		{{ printf "%s (%d)%s" .E.Title .E.Year .A.Ext }}
	{{ else }}
		{{ printf "%s%s" .E.Title .A.Ext }}
	{{ end }}
{{ end }}

{{ define "rename_tvshow" }}
	{{ if gt .E.Year 0 }}
		{{ printf "%s (%d)%s" .E.Title .E.Year .A.Ext }}
	{{ else }}
		{{ printf "%s%s" .E.Title .A.Ext }}
	{{ end }}
{{ end }}

{{ define "rename_episode" }}
	{{ $nums := printf "S%02dE%02d" .E.Season .E.EpisodeNum }}
	{{ if .E.Title }}
		{{ if .A.ShowTv }}
			{{ $tv := tvshow .E }}
			{{ printf "%s - %s - %s" $tv.Title .E.Title .A.Ext }}
		{{ else }}
			{{ printf "%s - %s%s" $nums .E.Title .A.Ext }}
		{{ end }}
	{{ else }}
		{{ if .A.ShowTv }}
			{{ $tv := tvshow .E }}
			{{ printf "%s - %s%s" $tv.Title $nums .A.Ext }}
		{{ else }}
			{{ printf "%s%s" $nums .A.Ext }}
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


	{{ template "short_media_details" .E }}
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

	{{ template "short_media_details" .E }}
{{ end }}

{{ define "short_episode" }}

	{{ $tv := tvshow .E }}
	{{ $tvname := printf "(TV show: %s (%d))" $tv.Title $tv.Year }}
	{{ printf "%s (%d) %s" .E.Title .E.Year $tvname | underlined "=" }}

	{{ if and (gt .E.Season 0) (gt .E.EpisodeNum 0) }}
		{{ printf "Season %d, Episode %d" .E.Season .E.EpisodeNum }}

	{{ end }}

	{{ template "short_media_details" .E }}
{{ end }}

{{ define "short_actor" }}

	{{ .E.Name | underlined "=" }}


{{ end }}

{{ define "short_media_details" }}
	{{ $plots := plots . }}
	{{ $runtimes := running_times . }}
	{{ $dates := release_dates . }}
	{{ $mpaa := mpaa . }}
	{{ $rank := rank . }}
	{{ if gt (len $runtimes) 0 }}
		{{ printf "Running time: %s" (index $runtimes 0) }}


	{{ end }}
	{{ if gt (len $dates) 0 }}
		{{ printf "Release date: %s" (index $dates 0) }}


	{{ end }}
	{{ if not $rank.Unranked }}
		{{ printf "IMDb rank: %s" $rank }}


	{{ end }}
	{{ if not $mpaa.Unrated }}
		{{ printf "MPAA rating: %s" $mpaa }}


	{{ end }}
	{{ if gt (len $plots) 0 }}
		{{ $p := index $plots 0 }}
		{{ "Plot" | underlined "-" }}

		{{ $p.Entry | wrap 80 }}

		{{ printf "-- %s" $p.By | wrap 80 }}


	{{ end }}
{{ end }}

{{ define "running-times" }}

	{{ printf "Running times for %s" .E | underlined "=" }}

	{{ $runtimes := running_times .E }}
	{{ if not (len $runtimes) }}
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
	{{ if not (len $dates) }}
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
	{{ if not (len $akas) }}
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
	{{ if not (len $alts) }}
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
	{{ if not (len $colors) }}
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
	{{ if not (len $mixs) }}
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
	{{ if not (len $tags) }}
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
	{{ if not (len $trivias) }}
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
	{{ if not (len $genres) }}
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
	{{ if not (len $goofs) }}
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
	{{ if not (len $langs) }}
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
	{{ if not (len $lits) }}
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
	{{ if not (len $locs) }}
		None found.

	{{ else }}
		{{ range $loc := $locs }}
			{{ $loc | wrap 80 }}


		{{ end }}
	{{ end }}
{{ end }}

{{ define "links" }}

	{{ printf "Links for %s" .E | underlined "=" }}

	{{ $links := links .E }}
	{{ if not (len $links) }}
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
	{{ if not (len $plots) }}
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
	{{ if not (len $quotes) }}
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
	{{ if $rank.Unranked }}
		None found.

	{{ else }}
		{{ $rank }}


	{{ end }}
{{ end }}

{{ define "credits" }}

	{{ printf "Credits for %s" .E | underlined "=" }}

	{{ $credits := credits .E }}
	{{ if not (len $credits) }}
		None found.

	{{ else }}
		{{ range $c := $credits }}
			{{ if eq "actor" $.E.Type.String }}
				{{ if eq "episode" $c.Media.Type.String }}
					{{ $tv := printf "(TV show: %s)" (tvshow $c.Media) }}
					{{ printf "%s %s %s" $c.Media $tv $c }}
				{{ else }}
					{{ printf "%s %s" $c.Media $c }}
				{{ end }}
			{{ else }}
				{{ printf "%s %s" $c.Actor $c }}
			{{ end }}

		{{ end }}

	{{ end }}
{{ end }}
`)
