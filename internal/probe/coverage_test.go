package probe

import "testing"

func TestMeasureCoverage(t *testing.T) {
	flags := ParseHelp(`
  -b, --bytes=LIST        select only these bytes
  -d, --delimiter=DELIM   use DELIM instead of TAB
  -f, --fields=LIST       select only these fields
  -s, --only-delimited    do not print lines not containing delimiters
  -z, --zero-terminated   line delimiter is NUL
      --complement        complement the set
      --help              display this help and exit
`)
	cov := MeasureCoverage(flags, [][]string{
		{"-d", ",", "-f", "2"},
		{"-sf1", "file.txt"},
		{"--compl", "-b1-3"},
		{"--fields=1", "-d:", "--", "-z"},
		{"-q"},
		{"-", "-5"},
	})
	got := map[string]int{}
	for _, u := range cov.Flags {
		got[u.Flag.Name()] = u.Cases
	}
	want := map[string]int{"--bytes": 1, "--delimiter": 2, "--fields": 3, "--only-delimited": 1,
		"--zero-terminated": 0, "--complement": 1, "--help": 0}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %d cases, want %d", k, got[k], v)
		}
	}
	if cov.Covered() != 5 {
		t.Errorf("covered = %d, want 5", cov.Covered())
	}
	if len(cov.Unknown) != 1 || cov.Unknown[0] != "-q" {
		t.Errorf("unknown = %v, want [-q]", cov.Unknown)
	}
}
