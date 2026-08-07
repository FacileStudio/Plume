package schemas

import "testing"

func TestAvatarPrecedence(t *testing.T) {
	const porte = "https://porte.facile.studio/media/user-avatars/x.png"

	cases := []struct {
		name       string
		user       User
		wantURL    string
		wantOrigin string
	}{
		{"Porte photo wins over an upload", User{OIDCPictureURL: porte, AvatarUploadPath: "avatars/user-3-1.png"}, porte, "oidc"},
		{"upload is the fallback", User{AvatarUploadPath: "avatars/user-3-1.png"}, "/api/files/avatars/user-3-1.png", "upload"},
		{"only Porte", User{OIDCPictureURL: porte}, porte, "oidc"},
		{"neither, so the client draws initials", User{}, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.user.Avatar(); got != tc.wantURL {
				t.Errorf("Avatar() = %q, want %q", got, tc.wantURL)
			}
			if got := tc.user.AvatarOrigin(); got != tc.wantOrigin {
				t.Errorf("AvatarOrigin() = %q, want %q", got, tc.wantOrigin)
			}
		})
	}
}

// The client uses avatar_url as an src attribute exactly as it arrives, because the other
// branch is an absolute Porte URL and concatenating "/api" onto that yields
// "/apihttps://…". So the upload branch has to carry the prefix the file route is mounted
// on — this is the test that fails if main.go's route and this constant drift apart.
func TestUploadedAvatarIsUsableAsAnSrcUnchanged(t *testing.T) {
	if AvatarFilePrefix != "/api/files/" {
		t.Fatalf("AvatarFilePrefix is %q; main.go serves the upload volume at /api/files/", AvatarFilePrefix)
	}
}

// The avatar is also readable in SQL, for a join that wants a picture without loading the
// row. The two spellings of one rule have to agree; this is the test that fails when
// someone edits one and forgets the other.
func TestAvatarSelectExprMatchesAvatar(t *testing.T) {
	orm := requireDB(t)

	users := []User{
		{Email: "both@example.com", OIDCPictureURL: "https://porte.facile.studio/media/user-avatars/a.png", AvatarUploadPath: "avatars/user-1-1.png"},
		{Email: "upload@example.com", AvatarUploadPath: "avatars/user-2-1.png"},
		{Email: "oidc@example.com", OIDCPictureURL: "https://porte.facile.studio/media/user-avatars/b.png"},
		{Email: "neither@example.com"},
	}
	for i := range users {
		if err := orm.Create(&users[i]).Error; err != nil {
			t.Fatalf("create %s: %v", users[i].Email, err)
		}
	}

	for _, want := range users {
		var got string
		if err := orm.Model(&User{}).
			Select(AvatarSelectExpr).
			Where("users.id = ?", want.ID).
			Scan(&got).Error; err != nil {
			t.Fatalf("select for %s: %v", want.Email, err)
		}
		if got != want.Avatar() {
			t.Errorf("%s: SQL gave %q, Avatar() gave %q", want.Email, got, want.Avatar())
		}
	}
}

// Row 2 is why this test exists: an uploaded file with avatar_source EMPTY, because that
// column was added after the upload feature. A backfill keyed on avatar_source = 'upload'
// drops its picture without a word. Row 5 is the other half — a data: URI parked in
// oidc_picture_url by the old code, which under the new rule would claim an SSO photo the
// user does not have and suppress the fallback forever.
func TestBackfillAvatarColumns(t *testing.T) {
	orm := requireDB(t)

	const initials = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjwvc3ZnPg=="

	rows := []struct {
		email      string
		url        string
		source     string
		oidc       string
		wantUpload string
		wantOIDC   string
	}{
		{"oidc-copy@example.com", "/files/avatars/oidc-1-178006.png", "oidc", "https://porte.facile.studio/media/user-avatars/a.png", "", "https://porte.facile.studio/media/user-avatars/a.png"},
		{"legacy-upload@example.com", "/files/avatars/user-2-177802.jpg", "", "", "avatars/user-2-177802.jpg", ""},
		{"upload-and-sso@example.com", "/files/avatars/user-3-178096.jpg", "upload", "https://porte.facile.studio/media/user-avatars/b.jpeg", "avatars/user-3-178096.jpg", "https://porte.facile.studio/media/user-avatars/b.jpeg"},
		{"no-avatar@example.com", "", "", "", "", ""},
		{"placeholder-claim@example.com", "", "", initials, "", ""},
	}
	for _, row := range rows {
		if err := orm.Exec(
			`INSERT INTO users (email, password_hash, avatar_url, avatar_source, oidc_picture_url) VALUES (?, 'hash', ?, ?, ?)`,
			row.email, row.url, row.source, row.oidc).Error; err != nil {
			t.Fatalf("insert %s: %v", row.email, err)
		}
	}

	if err := backfillAvatarColumns(orm); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	for _, row := range rows {
		var got User
		if err := orm.Where("email = ?", row.email).First(&got).Error; err != nil {
			t.Fatalf("read %s: %v", row.email, err)
		}
		if got.AvatarUploadPath != row.wantUpload {
			t.Errorf("%s: avatar_upload_path = %q, want %q", row.email, got.AvatarUploadPath, row.wantUpload)
		}
		if got.OIDCPictureURL != row.wantOIDC {
			t.Errorf("%s: oidc_picture_url = %q, want %q", row.email, got.OIDCPictureURL, row.wantOIDC)
		}
	}

	// The row that carries both keeps its file, and still renders the Porte photo.
	var both User
	if err := orm.Where("email = ?", "upload-and-sso@example.com").First(&both).Error; err != nil {
		t.Fatalf("read both: %v", err)
	}
	if both.Avatar() != "https://porte.facile.studio/media/user-avatars/b.jpeg" {
		t.Errorf("SSO photo should win, got %q", both.Avatar())
	}

	// And the user Authentik only ever drew initials for falls all the way back.
	var placeholder User
	if err := orm.Where("email = ?", "placeholder-claim@example.com").First(&placeholder).Error; err != nil {
		t.Fatalf("read placeholder: %v", err)
	}
	if placeholder.Avatar() != "" || placeholder.AvatarOrigin() != "" {
		t.Errorf("a data: placeholder still reads as a photo: %q / %q", placeholder.Avatar(), placeholder.AvatarOrigin())
	}
}
