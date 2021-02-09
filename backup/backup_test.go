package backup

import "testing"

func TestMgd_prefixName(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		backupName string
		want       string
	}{
		{name: "path include",
			prefix:     "XXXXX",
			backupName: "/file/path/backup.bolt",
			want:       "/file/path/XXXXX_backup.bolt"},
		{name: "no path",
			prefix:     "XXXXX",
			backupName: "backup.bolt",
			want:       "XXXXX_backup.bolt"},
		{name: "no path no prefix",
			prefix:     "XXXXX",
			backupName: "backup.bolt",
			want:       "XXXXX_backup.bolt"},
		{name: "path include no prefix",
			prefix:     "",
			backupName: "/file/path/backup.bolt",
			want:       "/file/path/_backup.bolt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrefixName(tt.prefix, tt.backupName); got != tt.want {
				t.Errorf("backupName() = %v, want %v", got, tt.want)
			}
		})
	}
}
