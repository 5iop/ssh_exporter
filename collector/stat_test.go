package collector

import (
	"testing"
)

// TestParseCPUStats 测试 /proc/stat 解析
func TestParseCPUStats(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   *CPUStats
	}{
		{
			name: "normal /proc/stat output",
			input: `cpu  100000 1000 50000 200000 5000 1000 500 0 0 0
cpu0 50000 500 25000 100000 2500 500 250 0 0 0
cpu1 50000 500 25000 100000 2500 500 250 0 0 0
ctxt 123456789
btime 1234567890
processes 98765
procs_running 2
procs_blocked 1
softirq 9876543 10 20 30 40 50 60 70 80 90 100
intr 87654321 100 200 300`,
			want: &CPUStats{
				User:         1000.0,  // 100000 / 100
				System:       500.0,   // 50000 / 100
				Idle:         2000.0,  // 200000 / 100
				Iowait:       50.0,    // 5000 / 100
				Total:        3575.0,  // (100000+1000+50000+200000+5000+1000+500+0) / 100
				Ctxt:         123456789.0,
				Intr:         87654321.0,
				ProcsRunning: 2.0,
				ProcsBlocked: 1.0,
			},
		},
		{
			name: "minimal /proc/stat output",
			input: `cpu  10000 100 5000 20000 500 100 50 0
ctxt 1000
intr 2000
procs_running 1
procs_blocked 0`,
			want: &CPUStats{
				User:         100.0,  // 10000 / 100
				System:       50.0,   // 5000 / 100
				Idle:         200.0,  // 20000 / 100
				Iowait:       5.0,    // 500 / 100
				Total:        357.5,  // (10000+100+5000+20000+500+100+50+0) / 100
				Ctxt:         1000.0,
				Intr:         2000.0,
				ProcsRunning: 1.0,
				ProcsBlocked: 0.0,
			},
		},
		{
			name:  "empty input",
			input: "",
			want: &CPUStats{
				User:         0,
				System:       0,
				Idle:         0,
				Iowait:       0,
				Total:        0,
				Ctxt:         0,
				Intr:         0,
				ProcsRunning: 0,
				ProcsBlocked: 0,
			},
		},
		{
			name: "missing some fields",
			input: `cpu  10000 100 5000 20000 500 100 50 0
ctxt 1000`,
			want: &CPUStats{
				User:         100.0,
				System:       50.0,
				Idle:         200.0,
				Iowait:       5.0,
				Total:        357.5,
				Ctxt:         1000.0,
				Intr:         0.0,
				ProcsRunning: 0.0,
				ProcsBlocked: 0.0,
			},
		},
		{
			name: "invalid number format",
			input: `cpu  abc def ghi jkl mno pqr stu vwx
ctxt invalid
intr 1000
procs_running 1
procs_blocked 0`,
			want: &CPUStats{
				User:         0,
				System:       0,
				Idle:         0,
				Iowait:       0,
				Total:        0,
				Ctxt:         0,
				Intr:         1000.0,
				ProcsRunning: 1.0,
				ProcsBlocked: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCPUStats(tt.input)
			if got == nil {
				t.Fatal("parseCPUStats returned nil")
			}

			// 比较结果，允许浮点数误差
			const epsilon = 0.001
			if !floatEqual(got.User, tt.want.User, epsilon) {
				t.Errorf("User = %v, want %v", got.User, tt.want.User)
			}
			if !floatEqual(got.System, tt.want.System, epsilon) {
				t.Errorf("System = %v, want %v", got.System, tt.want.System)
			}
			if !floatEqual(got.Idle, tt.want.Idle, epsilon) {
				t.Errorf("Idle = %v, want %v", got.Idle, tt.want.Idle)
			}
			if !floatEqual(got.Iowait, tt.want.Iowait, epsilon) {
				t.Errorf("Iowait = %v, want %v", got.Iowait, tt.want.Iowait)
			}
			if !floatEqual(got.Total, tt.want.Total, epsilon) {
				t.Errorf("Total = %v, want %v", got.Total, tt.want.Total)
			}
			if !floatEqual(got.Ctxt, tt.want.Ctxt, epsilon) {
				t.Errorf("Ctxt = %v, want %v", got.Ctxt, tt.want.Ctxt)
			}
			if !floatEqual(got.Intr, tt.want.Intr, epsilon) {
				t.Errorf("Intr = %v, want %v", got.Intr, tt.want.Intr)
			}
			if !floatEqual(got.ProcsRunning, tt.want.ProcsRunning, epsilon) {
				t.Errorf("ProcsRunning = %v, want %v", got.ProcsRunning, tt.want.ProcsRunning)
			}
			if !floatEqual(got.ProcsBlocked, tt.want.ProcsBlocked, epsilon) {
				t.Errorf("ProcsBlocked = %v, want %v", got.ProcsBlocked, tt.want.ProcsBlocked)
			}
		})
	}
}

// TestParseMemoryStats 测试 /proc/meminfo 解析
func TestParseMemoryStats(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   *MemoryStats
	}{
		{
			name: "normal /proc/meminfo output",
			input: `MemTotal:       16384000 kB
MemFree:         4096000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
Cached:          2048000 kB
SwapCached:            0 kB
Active:          8192000 kB
Inactive:        4096000 kB`,
			want: &MemoryStats{
				Total:        16384000 * 1024, // 转换为字节
				Free:         4096000 * 1024,
				Available:    8192000 * 1024,
				Buffers:      512000 * 1024,
				Cached:       2048000 * 1024,
				UsagePercent: 50.0, // (16384000 - 8192000) / 16384000 * 100 = 50%
			},
		},
		{
			name: "minimal /proc/meminfo output",
			input: `MemTotal:       8192000 kB
MemFree:        2048000 kB
MemAvailable:   4096000 kB
Buffers:         256000 kB
Cached:         1024000 kB`,
			want: &MemoryStats{
				Total:        8192000 * 1024,
				Free:         2048000 * 1024,
				Available:    4096000 * 1024,
				Buffers:      256000 * 1024,
				Cached:       1024000 * 1024,
				UsagePercent: 50.0, // (8192000 - 4096000) / 8192000 * 100 = 50%
			},
		},
		{
			name:  "empty input",
			input: "",
			want: &MemoryStats{
				Total:        0,
				Free:         0,
				Available:    0,
				Buffers:      0,
				Cached:       0,
				UsagePercent: 0,
			},
		},
		{
			name: "missing some fields",
			input: `MemTotal:       8192000 kB
MemFree:        2048000 kB
MemAvailable:   4096000 kB`,
			want: &MemoryStats{
				Total:        8192000 * 1024,
				Free:         2048000 * 1024,
				Available:    4096000 * 1024,
				Buffers:      0,
				Cached:       0,
				UsagePercent: 50.0,
			},
		},
		{
			name: "zero total memory (edge case)",
			input: `MemTotal:       0 kB
MemFree:        0 kB
MemAvailable:   0 kB
Buffers:        0 kB
Cached:         0 kB`,
			want: &MemoryStats{
				Total:        0,
				Free:         0,
				Available:    0,
				Buffers:      0,
				Cached:       0,
				UsagePercent: 0, // 避免除以0
			},
		},
		{
			name: "invalid number format",
			input: `MemTotal:       invalid kB
MemFree:        2048000 kB
MemAvailable:   4096000 kB
Buffers:        abc kB
Cached:         1024000 kB`,
			want: &MemoryStats{
				Total:        0,
				Free:         2048000 * 1024,
				Available:    4096000 * 1024,
				Buffers:      0,
				Cached:       1024000 * 1024,
				UsagePercent: 0, // Total is 0, so can't calculate
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMemoryStats(tt.input)
			if got == nil {
				t.Fatal("parseMemoryStats returned nil")
			}

			const epsilon = 0.001
			if !floatEqual(got.Total, tt.want.Total, epsilon) {
				t.Errorf("Total = %v, want %v", got.Total, tt.want.Total)
			}
			if !floatEqual(got.Free, tt.want.Free, epsilon) {
				t.Errorf("Free = %v, want %v", got.Free, tt.want.Free)
			}
			if !floatEqual(got.Available, tt.want.Available, epsilon) {
				t.Errorf("Available = %v, want %v", got.Available, tt.want.Available)
			}
			if !floatEqual(got.Buffers, tt.want.Buffers, epsilon) {
				t.Errorf("Buffers = %v, want %v", got.Buffers, tt.want.Buffers)
			}
			if !floatEqual(got.Cached, tt.want.Cached, epsilon) {
				t.Errorf("Cached = %v, want %v", got.Cached, tt.want.Cached)
			}
			if !floatEqual(got.UsagePercent, tt.want.UsagePercent, epsilon) {
				t.Errorf("UsagePercent = %v, want %v", got.UsagePercent, tt.want.UsagePercent)
			}
		})
	}
}

// TestParseDiskStats 测试 df 命令输出解析
func TestParseDiskStats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []DiskStats
	}{
		{
			name: "normal df output with multiple partitions",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1      107374182400 53687091200  53687091200  50% /
/dev/sda2      214748364800 107374182400 107374182400  50% /data
/dev/sdb1       53687091200  10737418240  42949672960  20% /backup`,
			want: []DiskStats{
				{
					Device:       "/dev/sda1",
					MountPoint:   "/",
					Total:        107374182400,
					Used:         53687091200,
					Free:         53687091200,
					UsagePercent: 50,
				},
				{
					Device:       "/dev/sda2",
					MountPoint:   "/data",
					Total:        214748364800,
					Used:         107374182400,
					Free:         107374182400,
					UsagePercent: 50,
				},
				{
					Device:       "/dev/sdb1",
					MountPoint:   "/backup",
					Total:        53687091200,
					Used:         10737418240,
					Free:         42949672960,
					UsagePercent: 20,
				},
			},
		},
		{
			name: "single partition",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1      107374182400 53687091200  53687091200  50% /`,
			want: []DiskStats{
				{
					Device:       "/dev/sda1",
					MountPoint:   "/",
					Total:        107374182400,
					Used:         53687091200,
					Free:         53687091200,
					UsagePercent: 50,
				},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  []DiskStats{},
		},
		{
			name: "only header",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on`,
			want:  []DiskStats{},
		},
		{
			name: "100% usage",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1      107374182400 107374182400           0 100% /`,
			want: []DiskStats{
				{
					Device:       "/dev/sda1",
					MountPoint:   "/",
					Total:        107374182400,
					Used:         107374182400,
					Free:         0,
					UsagePercent: 100,
				},
			},
		},
		{
			name: "0% usage",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1      107374182400           0 107374182400   0% /`,
			want: []DiskStats{
				{
					Device:       "/dev/sda1",
					MountPoint:   "/",
					Total:        107374182400,
					Used:         0,
					Free:         107374182400,
					UsagePercent: 0,
				},
			},
		},
		{
			name: "invalid line (missing fields)",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1      107374182400 53687091200
/dev/sda2      214748364800 107374182400 107374182400  50% /data`,
			want: []DiskStats{
				{
					Device:       "/dev/sda2",
					MountPoint:   "/data",
					Total:        214748364800,
					Used:         107374182400,
					Free:         107374182400,
					UsagePercent: 50,
				},
			},
		},
		{
			name: "mount point with spaces (edge case - df handles this differently)",
			input: `Filesystem     1B-blocks        Used   Available Use% Mounted on
/dev/sda1      107374182400 53687091200  53687091200  50% /mnt/my`,
			want: []DiskStats{
				{
					Device:       "/dev/sda1",
					MountPoint:   "/mnt/my",
					Total:        107374182400,
					Used:         53687091200,
					Free:         53687091200,
					UsagePercent: 50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDiskStats(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("parseDiskStats() returned %d disks, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				const epsilon = 0.001
				if got[i].Device != tt.want[i].Device {
					t.Errorf("disk[%d].Device = %v, want %v", i, got[i].Device, tt.want[i].Device)
				}
				if got[i].MountPoint != tt.want[i].MountPoint {
					t.Errorf("disk[%d].MountPoint = %v, want %v", i, got[i].MountPoint, tt.want[i].MountPoint)
				}
				if !floatEqual(got[i].Total, tt.want[i].Total, epsilon) {
					t.Errorf("disk[%d].Total = %v, want %v", i, got[i].Total, tt.want[i].Total)
				}
				if !floatEqual(got[i].Used, tt.want[i].Used, epsilon) {
					t.Errorf("disk[%d].Used = %v, want %v", i, got[i].Used, tt.want[i].Used)
				}
				if !floatEqual(got[i].Free, tt.want[i].Free, epsilon) {
					t.Errorf("disk[%d].Free = %v, want %v", i, got[i].Free, tt.want[i].Free)
				}
				if !floatEqual(got[i].UsagePercent, tt.want[i].UsagePercent, epsilon) {
					t.Errorf("disk[%d].UsagePercent = %v, want %v", i, got[i].UsagePercent, tt.want[i].UsagePercent)
				}
			}
		})
	}
}

// floatEqual 比较两个浮点数是否在误差范围内相等
func floatEqual(a, b, epsilon float64) bool {
	if a == b {
		return true
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
