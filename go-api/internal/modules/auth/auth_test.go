package auth

import "testing"

func TestGradeToRoles(t *testing.T) {
	tests := []struct {
		grade string
		want  []string
	}{
		{grade: "3", want: []string{"super", "admin", "user"}},
		{grade: "2", want: []string{"admin", "user"}},
		{grade: "1", want: []string{"user"}},
	}

	for _, tt := range tests {
		got := gradeToRoles(tt.grade)
		if len(got) != len(tt.want) {
			t.Fatalf("grade %s: got %v want %v", tt.grade, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("grade %s: got %v want %v", tt.grade, got, tt.want)
			}
		}
	}
}

func TestGetAccessCodes(t *testing.T) {
	svc := NewService()

	tests := []struct {
		grade string
		want  []string
	}{
		{grade: "0", want: []string{"AC_USER"}},
		{grade: "1", want: []string{"AC_USER", "AC_AGENT"}},
		{grade: "2", want: []string{"AC_USER", "AC_AGENT", "admin", "AC_ADMIN", "AC_ADMIN_CONFIG", "AC_ADMIN_CLASS", "AC_ADMIN_STATS"}},
		{grade: "3", want: []string{"AC_USER", "AC_AGENT", "admin", "AC_ADMIN", "AC_ADMIN_CONFIG", "AC_ADMIN_CLASS", "AC_ADMIN_STATS", "super"}},
	}

	for _, tt := range tests {
		got := svc.GetAccessCodes(tt.grade)
		if len(got) != len(tt.want) {
			t.Fatalf("grade %s: got %v want %v", tt.grade, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("grade %s: got %v want %v", tt.grade, got, tt.want)
			}
		}
	}
}
