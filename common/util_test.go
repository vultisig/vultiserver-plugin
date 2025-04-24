package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDataCompression(t *testing.T) {
	data := "message"
	compressedData, err := CompressData([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	decompressedData, err := DecompressData(compressedData)
	if err != nil {
		t.Fatal(err)
	}

	if string(decompressedData) != data {
		t.Fatalf("decompressed: %s, expected: %s", decompressedData, data)
	}
}

func TestVaultEncryption(t *testing.T) {
	password := "password"
	src := "vault_bytes"
	encrypted, err := EncryptVault(password, []byte(src))
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := DecryptVault(password, encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted) != src {
		t.Fatalf("decrypted: %s, expected: %s", decrypted, src)
	}
}

func TestGetSortingCondition(t *testing.T) {
	tests := []struct {
		sort                   string
		expectedOrderBy        string
		expectedOrderDirection string
	}{
		{"created_at", "created_at", "ASC"},
		{"-created_at", "created_at", "DESC"},
		{"non_exist", "created_at", "ASC"},
		{"-non_exist", "created_at", "DESC"},
		{"title", "title", "ASC"},
		{"-title", "title", "DESC"},
		{"updated_at", "updated_at", "ASC"},
		{"-updated_at", "updated_at", "DESC"},
	}

	for _, tt := range tests {
		orderBy, orderDirection := GetSortingCondition(tt.sort, map[string]bool{"updated_at": true, "created_at": true, "title": true})

		if orderBy != tt.expectedOrderBy {
			t.Errorf("sort: %s -> orderBy: %s, expected: %s", tt.sort, orderBy, tt.expectedOrderBy)
		}

		if orderDirection != tt.expectedOrderDirection {
			t.Errorf("sort: %s -> orderDirection: %s, expected: %s", tt.sort, orderDirection, tt.expectedOrderDirection)
		}
	}
}

func TestVaultBackupCompatible(t *testing.T) {
	dir := "../test/test_2of2_vault_backups"
	vaultPwd := "888717"

	filePathName := filepath.Join(dir, "Fast Vault #2-0985-part1of2-Vultiserver.vult")
	_, err := os.Stat(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	firstVault, err := DecryptVaultFromBackup(vaultPwd, content)
	if err != nil {
		t.Fatal(err)
	}

	filePathName = filepath.Join(dir, "Fast Vault #2-0985-part2of2.vult")
	_, err = os.Stat(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	content, err = os.ReadFile(filePathName)
	if err != nil {
		t.Fatal(err)
	}

	secondVault, err := DecryptVaultFromBackup(vaultPwd, content)
	if err != nil {
		t.Fatal(err)
	}

	if firstVault.PublicKeyEcdsa != secondVault.PublicKeyEcdsa {
		t.Fatalf("ios backup is not compatible with android backup")
	}
}
