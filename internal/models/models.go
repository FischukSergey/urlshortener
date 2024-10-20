package models

import (
	"net"
)

// TrustedSubnet структура для хранения доверенной подсети
type TrustedSubnet struct {
	IP *net.IPNet
}

// AllURLUserID структура для хранения короткого и оригинального URL
type AllURLUserID struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// NewTrustedSubnet создает структуру TrustedSubnet на основе переданной подсети
func NewTrustedSubnet(flagTrustedSubnet string) (TrustedSubnet, error) {
	_, subnet, err := net.ParseCIDR(flagTrustedSubnet)
	if err != nil {
		return TrustedSubnet{}, err
	}
	return TrustedSubnet{IP: subnet}, nil
}

// IsTrusted проверяет, принадлежит ли IP-адрес к доверенной подсети
func (t *TrustedSubnet) IsTrusted(ip net.IP) bool {
	return t.IP.Contains(ip)
}
