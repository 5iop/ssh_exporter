package collector

import (
	"reflect"
	"testing"
)

// TestParseProcessOutput 测试进程输出解析
func TestParseProcessOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []string
	}{
		{
			name: "normal process cmdlines with null separators",
			input: "/usr/bin/nginx\x00-c\x00/etc/nginx/nginx.conf\n" +
				"/usr/bin/java\x00-jar\x00app.jar\n" +
				"/usr/bin/python\x00script.py\x00--verbose",
			want: []string{
				"/usr/bin/nginx -c /etc/nginx/nginx.conf",
				"/usr/bin/java -jar app.jar",
				"/usr/bin/python script.py --verbose",
			},
		},
		{
			name: "single process",
			input: "/bin/bash\x00-c\x00echo hello",
			want: []string{
				"/bin/bash -c echo hello",
			},
		},
		{
			name: "processes without arguments",
			input: "/sbin/init\n" +
				"/usr/sbin/sshd\n" +
				"/usr/bin/systemd",
			want: []string{
				"/sbin/init",
				"/usr/sbin/sshd",
				"/usr/bin/systemd",
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name: "input with only newlines",
			input: "\n\n\n",
			want:  nil,
		},
		{
			name: "input with whitespace",
			input: "   \n  \n   ",
			want:  nil,
		},
		{
			name: "mixed empty and non-empty lines",
			input: "/usr/bin/nginx\x00-c\x00/etc/nginx/nginx.conf\n" +
				"\n" +
				"/usr/bin/java\x00-jar\x00app.jar\n" +
				"   \n" +
				"/usr/bin/python\x00script.py",
			want: []string{
				"/usr/bin/nginx -c /etc/nginx/nginx.conf",
				"/usr/bin/java -jar app.jar",
				"/usr/bin/python script.py",
			},
		},
		{
			name: "process with multiple consecutive null characters",
			input: "/usr/bin/test\x00\x00arg1\x00\x00\x00arg2",
			want: []string{
				"/usr/bin/test  arg1   arg2",
			},
		},
		{
			name: "process with special characters in arguments",
			input: "/bin/bash\x00-c\x00echo $PATH\n" +
				"/usr/bin/grep\x00-E\x00^[0-9]+$",
			want: []string{
				"/bin/bash -c echo $PATH",
				"/usr/bin/grep -E ^[0-9]+$",
			},
		},
		{
			name: "long command line",
			input: "/usr/bin/java\x00-Xmx2048m\x00-Xms512m\x00-XX:+UseG1GC\x00-jar\x00application.jar\x00--server.port=8080\x00--spring.profiles.active=prod",
			want: []string{
				"/usr/bin/java -Xmx2048m -Xms512m -XX:+UseG1GC -jar application.jar --server.port=8080 --spring.profiles.active=prod",
			},
		},
		{
			name: "process name only (kernel threads)",
			input: "[kthreadd]\n" +
				"[kworker/0:0]\n" +
				"[ksoftirqd/0]",
			want: []string{
				"[kthreadd]",
				"[kworker/0:0]",
				"[ksoftirqd/0]",
			},
		},
		{
			name: "processes with paths containing spaces (represented as null)",
			input: "/opt/my\x00app/bin/server\x00--config\x00/etc/config.yml",
			want: []string{
				"/opt/my app/bin/server --config /etc/config.yml",
			},
		},
		{
			name: "trailing newline",
			input: "/usr/bin/nginx\n" +
				"/usr/bin/java\n",
			want: []string{
				"/usr/bin/nginx",
				"/usr/bin/java",
			},
		},
		{
			name: "trailing whitespace after null replacement",
			input: "/usr/bin/test\x00arg\x00\n",
			want: []string{
				"/usr/bin/test arg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseProcessOutput(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseProcessOutput() = %v, want %v", got, tt.want)
				t.Logf("got length: %d, want length: %d", len(got), len(tt.want))
				for i := 0; i < len(got) || i < len(tt.want); i++ {
					if i < len(got) && i < len(tt.want) {
						if got[i] != tt.want[i] {
							t.Logf("  [%d] got:  %q", i, got[i])
							t.Logf("  [%d] want: %q", i, tt.want[i])
						}
					} else if i < len(got) {
						t.Logf("  [%d] got:  %q (extra)", i, got[i])
					} else {
						t.Logf("  [%d] want: %q (missing)", i, tt.want[i])
					}
				}
			}
		})
	}
}

// BenchmarkParseProcessOutput 性能基准测试
func BenchmarkParseProcessOutput(b *testing.B) {
	// 模拟包含100个进程的输出
	input := ""
	for i := 0; i < 100; i++ {
		input += "/usr/bin/process" + string(rune(i)) + "\x00arg1\x00arg2\x00arg3\n"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseProcessOutput(input)
	}
}
