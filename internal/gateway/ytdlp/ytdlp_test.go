package ytdlp

import "testing"

func TestSubtitleTracksIncludesAutomaticAndPrefersManual(t *testing.T) {
	manual := map[string][]subtitleFormat{
		"en": {{Extension: "vtt", Name: "English"}},
	}
	automatic := map[string][]subtitleFormat{
		"en": {{Extension: "json3", Name: "English (auto-generated)"}},
		"vi": {
			{Extension: "vtt", Name: "Vietnamese (auto-generated)"},
			{Extension: "vtt", Name: "Vietnamese (auto-generated)"},
			{Extension: "json3", Name: "Vietnamese (auto-generated)"},
		},
	}

	tracks := subtitleTracks(manual, automatic)
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(tracks))
	}
	if tracks[0].LanguageCode != "en" || tracks[0].IsAutomatic {
		t.Fatalf("expected manual English track, got %+v", tracks[0])
	}
	if tracks[1].LanguageCode != "vi" || !tracks[1].IsAutomatic {
		t.Fatalf("expected automatic Vietnamese track, got %+v", tracks[1])
	}
	if len(tracks[1].Formats) != 2 || tracks[1].Formats[0] != "json3" || tracks[1].Formats[1] != "vtt" {
		t.Fatalf("expected sorted unique formats, got %v", tracks[1].Formats)
	}
}
