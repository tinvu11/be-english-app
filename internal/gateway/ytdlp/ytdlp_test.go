package ytdlp

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCookieFileUsedForPreviewAndAudio(t *testing.T) {
	const cookiePath = "/run/secrets/youtube-cookies.txt"
	provider := &Provider{cookiesFile: cookiePath}
	for name, args := range map[string][]string{
		"preview": provider.previewArgs("dQw4w9WgXcQ"),
		"audio":   provider.audioArgs("dQw4w9WgXcQ", t.TempDir()),
	} {
		i := slices.Index(args, "--cookies")
		if i < 0 || i+1 >= len(args) || args[i+1] != cookiePath {
			t.Errorf("%s args do not include the configured cookies file: %v", name, args)
		}
	}
}

func TestCookieFileOptional(t *testing.T) {
	provider := &Provider{}
	for _, args := range [][]string{provider.previewArgs("dQw4w9WgXcQ"), provider.audioArgs("dQw4w9WgXcQ", t.TempDir())} {
		if slices.Contains(args, "--cookies") {
			t.Errorf("unexpected cookies option: %v", args)
		}
	}
}

func TestPrepareCookiesUsesWritableTemporaryCopy(t *testing.T) {
	source := filepath.Join(t.TempDir(), "cookie_yt.txt")
	if err := os.WriteFile(source, []byte("test cookie jar"), 0o400); err != nil {
		t.Fatal(err)
	}
	args, cleanup, err := prepareCookies([]string{"--cookies", source, "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	if args[1] == source {
		t.Fatal("yt-dlp would write to the original cookies file")
	}
	data, err := os.ReadFile(args[1])
	if err != nil || string(data) != "test cookie jar" {
		t.Fatalf("temporary cookies mismatch: %q, %v", data, err)
	}
	cleanup()
	if _, err := os.Stat(args[1]); !os.IsNotExist(err) {
		t.Fatalf("temporary cookies file was not removed: %v", err)
	}
}

func TestNetscapeCookiesFromJSONExport(t *testing.T) {
	data := []byte(`[{"domain":".youtube.com","expirationDate":1900000000,"hostOnly":false,"httpOnly":true,"name":"session","path":"/","secure":true,"session":false,"value":"example"}]`)
	got, err := netscapeCookies(data)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Netscape HTTP Cookie File\n#HttpOnly_.youtube.com\tTRUE\t/\tTRUE\t1900000000\tsession\texample\n"
	if string(got) != want {
		t.Fatalf("converted cookies mismatch: %q", got)
	}
}
