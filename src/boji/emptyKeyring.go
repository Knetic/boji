package boji
import "golang.org/x/crypto/openpgp"

type emptyKeyring uint8

func (this emptyKeyring) KeysById(id uint64) []openpgp.Key {
	return []openpgp.Key{}
}
func (this emptyKeyring) KeysByIdUsage(id uint64, requiredUsage byte) []openpgp.Key {
	return []openpgp.Key{}
}
func (this emptyKeyring) DecryptionKeys() []openpgp.Key {
	return []openpgp.Key{}
}