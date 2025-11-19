package collector

import (
	"reflect"
	"testing"
	"time"
)

// TestParseFileOutput 测试文件输出解析
func TestParseFileOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		basePath string
		want     []FileInfo
	}{
		{
			name: "normal ls output",
			input: "1024\t2024-01-15 10:30:45.123456000 +0000\tapp.log\n" +
				"2048\t2024-01-15 11:45:30.654321000 +0000\terror.log\n" +
				"4096\t2024-01-15 12:00:00.000000000 +0000\taccess.log",
			basePath: "/var/log/",
			want: []FileInfo{
				{
					Path:         "app.log",
					Size:         1024,
					LastModified: parseTime("2024-01-15 10:30:45.123456"),
				},
				{
					Path:         "error.log",
					Size:         2048,
					LastModified: parseTime("2024-01-15 11:45:30.654321"),
				},
				{
					Path:         "access.log",
					Size:         4096,
					LastModified: parseTime("2024-01-15 12:00:00.000000"),
				},
			},
		},
		{
			name: "single file",
			input: "512\t2024-01-15 09:00:00.000000000 +0000\ttest.txt",
			basePath: "/tmp/",
			want: []FileInfo{
				{
					Path:         "test.txt",
					Size:         512,
					LastModified: parseTime("2024-01-15 09:00:00.000000"),
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			basePath: "/var/log/",
			want:     nil,
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			basePath: "/var/log/",
			want:     nil,
		},
		{
			name: "zero-size file",
			input: "0\t2024-01-15 10:00:00.000000000 +0000\tempty.txt",
			basePath: "/tmp/",
			want: []FileInfo{
				{
					Path:         "empty.txt",
					Size:         0,
					LastModified: parseTime("2024-01-15 10:00:00.000000"),
				},
			},
		},
		{
			name: "large file size",
			input: "10737418240\t2024-01-15 10:00:00.000000000 +0000\tbigfile.dat",
			basePath: "/data/",
			want: []FileInfo{
				{
					Path:         "bigfile.dat",
					Size:         10737418240,
					LastModified: parseTime("2024-01-15 10:00:00.000000"),
				},
			},
		},
		{
			name: "file with special characters in name",
			input: "1024\t2024-01-15 10:00:00.000000000 +0000\tfile-with_special.chars[1].txt",
			basePath: "/tmp/",
			want: []FileInfo{
				{
					Path:         "file-with_special.chars[1].txt",
					Size:         1024,
					LastModified: parseTime("2024-01-15 10:00:00.000000"),
				},
			},
		},
		{
			name: "invalid line format - missing fields",
			input: "1024\t2024-01-15 10:00:00.000000000\n" + // Missing filename
				"2048\t2024-01-15 11:00:00.000000000 +0000\tvalid.log",
			basePath: "/var/log/",
			want: []FileInfo{
				{
					Path:         "valid.log",
					Size:         2048,
					LastModified: parseTime("2024-01-15 11:00:00.000000"),
				},
			},
		},
		{
			name: "invalid size format",
			input: "abc\t2024-01-15 10:00:00.000000000 +0000\tinvalid.log\n" +
				"1024\t2024-01-15 11:00:00.000000000 +0000\tvalid.log",
			basePath: "/var/log/",
			want: []FileInfo{
				{
					Path:         "valid.log",
					Size:         1024,
					LastModified: parseTime("2024-01-15 11:00:00.000000"),
				},
			},
		},
		{
			name: "invalid time format",
			input: "1024\tinvalid-time\tskipped.log\n" +
				"2048\t2024-01-15 10:00:00.000000000 +0000\tvalid.log",
			basePath: "/var/log/",
			want: []FileInfo{
				{
					Path:         "valid.log",
					Size:         2048,
					LastModified: parseTime("2024-01-15 10:00:00.000000"),
				},
			},
		},
		{
			name: "mixed valid and invalid lines",
			input: "1024\t2024-01-15 10:00:00.000000000 +0000\tfile1.log\n" +
				"invalid\n" +
				"2048\t2024-01-15 11:00:00.000000000 +0000\tfile2.log\n" +
				"\n" +
				"4096\tinvalid-time\tskipped.log\n" +
				"8192\t2024-01-15 12:00:00.000000000 +0000\tfile3.log",
			basePath: "/var/log/",
			want: []FileInfo{
				{
					Path:         "file1.log",
					Size:         1024,
					LastModified: parseTime("2024-01-15 10:00:00.000000"),
				},
				{
					Path:         "file2.log",
					Size:         2048,
					LastModified: parseTime("2024-01-15 11:00:00.000000"),
				},
				{
					Path:         "file3.log",
					Size:         8192,
					LastModified: parseTime("2024-01-15 12:00:00.000000"),
				},
			},
		},
		{
			name: "whitespace in fields",
			input: "  1024  \t  2024-01-15 10:00:00.000000000 +0000  \t  test.log  ",
			basePath: "/tmp/",
			want: []FileInfo{
				{
					Path:         "test.log",
					Size:         1024,
					LastModified: parseTime("2024-01-15 10:00:00.000000"),
				},
			},
		},
		{
			name: "nanosecond precision timestamp",
			input: "1024\t2024-01-15 10:30:45.123456789 +0000\tprecise.log",
			basePath: "/var/log/",
			want: []FileInfo{
				{
					Path:         "precise.log",
					Size:         1024,
					LastModified: parseTime("2024-01-15 10:30:45.123456"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFileOutput(tt.input, tt.basePath)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFileOutput() mismatch")
				t.Logf("got length: %d, want length: %d", len(got), len(tt.want))
				for i := 0; i < len(got) || i < len(tt.want); i++ {
					if i < len(got) && i < len(tt.want) {
						if !fileInfoEqual(got[i], tt.want[i]) {
							t.Logf("  [%d] got:  %+v", i, got[i])
							t.Logf("  [%d] want: %+v", i, tt.want[i])
						}
					} else if i < len(got) {
						t.Logf("  [%d] got:  %+v (extra)", i, got[i])
					} else {
						t.Logf("  [%d] want: %+v (missing)", i, tt.want[i])
					}
				}
			}
		})
	}
}

// TestMatchesPattern 测试正则表达式匹配
func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		pattern  string
		want     bool
	}{
		{
			name:     "log file pattern - match",
			filename: "app.log",
			pattern:  `.*\.log$`,
			want:     true,
		},
		{
			name:     "log file pattern - no match",
			filename: "app.txt",
			pattern:  `.*\.log$`,
			want:     false,
		},
		{
			name:     "any file pattern",
			filename: "anyfile.txt",
			pattern:  `.*`,
			want:     true,
		},
		{
			name:     "date pattern in filename - match",
			filename: "backup-2024-01-15.tar.gz",
			pattern:  `backup-\d{4}-\d{2}-\d{2}\.tar\.gz$`,
			want:     true,
		},
		{
			name:     "date pattern in filename - no match",
			filename: "backup-latest.tar.gz",
			pattern:  `backup-\d{4}-\d{2}-\d{2}\.tar\.gz$`,
			want:     false,
		},
		{
			name:     "prefix match",
			filename: "error.log",
			pattern:  `^error`,
			want:     true,
		},
		{
			name:     "prefix no match",
			filename: "app-error.log",
			pattern:  `^error`,
			want:     false,
		},
		{
			name:     "extension alternatives - log",
			filename: "file.log",
			pattern:  `.*\.(log|txt|err)$`,
			want:     true,
		},
		{
			name:     "extension alternatives - txt",
			filename: "file.txt",
			pattern:  `.*\.(log|txt|err)$`,
			want:     true,
		},
		{
			name:     "extension alternatives - err",
			filename: "file.err",
			pattern:  `.*\.(log|txt|err)$`,
			want:     true,
		},
		{
			name:     "extension alternatives - no match",
			filename: "file.dat",
			pattern:  `.*\.(log|txt|err)$`,
			want:     false,
		},
		{
			name:     "case sensitive match",
			filename: "Error.LOG",
			pattern:  `^Error\.LOG$`,
			want:     true,
		},
		{
			name:     "case sensitive no match",
			filename: "error.log",
			pattern:  `^Error\.LOG$`,
			want:     false,
		},
		{
			name:     "invalid regex pattern",
			filename: "test.log",
			pattern:  `[invalid(regex`,
			want:     false,
		},
		{
			name:     "empty pattern matches empty string",
			filename: "",
			pattern:  `^$`,
			want:     true,
		},
		{
			name:     "empty pattern does not match non-empty",
			filename: "file.log",
			pattern:  `^$`,
			want:     false,
		},
		{
			name:     "wildcard character class",
			filename: "file123.log",
			pattern:  `file\d+\.log`,
			want:     true,
		},
		{
			name:     "negated character class - match",
			filename: "file_abc.log",
			pattern:  `file[^0-9]+\.log`,
			want:     true,
		},
		{
			name:     "negated character class - no match",
			filename: "file123.log",
			pattern:  `file[^0-9]+\.log`,
			want:     false,
		},
		{
			name:     "quantifier match",
			filename: "error-2024-01-15-001.log",
			pattern:  `error-\d{4}-\d{2}-\d{2}-\d{3}\.log`,
			want:     true,
		},
		{
			name:     "quantifier no match",
			filename: "error-2024-01-15-1.log",
			pattern:  `error-\d{4}-\d{2}-\d{2}-\d{3}\.log`,
			want:     false,
		},
		{
			name:     "optional group match with group",
			filename: "backup.tar.gz",
			pattern:  `backup(\.tar)?\.gz`,
			want:     true,
		},
		{
			name:     "optional group match without group",
			filename: "backup.gz",
			pattern:  `backup(\.tar)?\.gz`,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesPattern(tt.filename, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v",
					tt.filename, tt.pattern, got, tt.want)
			}
		})
	}
}

// parseTime 辅助函数：将时间字符串解析为 Unix 时间戳
func parseTime(timeStr string) float64 {
	t, err := time.Parse("2006-01-02 15:04:05.000000", timeStr)
	if err != nil {
		panic(err)
	}
	return float64(t.Unix())
}

// fileInfoEqual 比较两个 FileInfo 是否相等
func fileInfoEqual(a, b FileInfo) bool {
	return a.Path == b.Path &&
		a.Size == b.Size &&
		a.LastModified == b.LastModified
}

// BenchmarkParseFileOutput 性能基准测试
func BenchmarkParseFileOutput(b *testing.B) {
	// 模拟包含100个文件的输出
	input := ""
	for i := 0; i < 100; i++ {
		input += "1024\t2024-01-15 10:00:00.000000000 +0000\tfile" + string(rune(i)) + ".log\n"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseFileOutput(input, "/var/log/")
	}
}

// BenchmarkMatchesPattern 性能基准测试
func BenchmarkMatchesPattern(b *testing.B) {
	filename := "app-2024-01-15.log"
	pattern := `app-\d{4}-\d{2}-\d{2}\.log`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matchesPattern(filename, pattern)
	}
}
