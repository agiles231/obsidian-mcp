package urn

import (
	"fmt"
	"errors"
	"strings"
	"net/url"
)

// NoteRef is a parsed reference to a note; section anchor optional
type NoteRef struct {
	Vault string // logical vault name
	Path string // vault-relative path
	Anchor Anchor // optional section anchor
}

// URN returns the canonical URN string
func (r *NoteRef) URN() string {
	var sb strings.Builder
	sb.WriteString("urn:obsidian::")
	sb.WriteString(encodeField(r.Vault))
	sb.WriteString(":note:")
	sb.WriteString(encodePath(r.Path))
	if !r.Anchor.IsZero() {
		sb.WriteByte('#')
		sb.WriteString(r.Anchor.encode())
	}
	return sb.String()
}

// Anchor identifies a location within a note. Headings and BlockID are mutually exclusive; one or the other must be used
type Anchor struct {
	Headings []string // e.g. ["Design", "Identity"] for #Design#Identity
	BlockID string // e.g. "a1b2c3" for #^a1b2c3
}

func (a Anchor) encode() string {
	if a.BlockID != "" {
		return "^" + a.BlockID
	}
	encoded := make([]string, len(a.Headings))
	for i, h := range a.Headings {
		encoded[i] = encodeField(h)
	}
	return strings.Join(encoded, "#")
}

func (a Anchor) IsZero() bool {
	return len(a.Headings) == 0 && a.BlockID == ""
}

func encodeField(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if shouldEncode(r) {
			sb.WriteString(url.PathEscape(string(r)))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func encodePath(p string) string {
	segments := strings.Split(p, "/")
	for i, seg := range segments {
		segments[i] = encodeField(seg)
	}
	return strings.Join(segments, "/")
}

func shouldEncode(r rune) bool {
	if r > 0x7E || r < 0x21 { // non-ASCII or control/space
		return true
	}
	switch r {
	case ':', '#', '%', '?', '/':
		return true
	}
	return false
}

var (
	ErrInvalidURN = errors.New("invalid URN")
	ErrEmptyVault = errors.New("empty vault name")
	ErrEmptyPath = errors.New("empty path")
	ErrUnknownType = errors.New("unknown resource type")
	ErrUserReserved = errors.New("uesr field must be empty in v1")
)

const urnPrefix = "urn:obsidian:"

// ParseRef parses a full URN or a bare vault-relative path.
// For bare paths, defaultVault is used
func ParseRef(input, defaultVault string) (*NoteRef, error) {
	if strings.HasPrefix(strings.ToLower(input), urnPrefix) {
		return parseURN(input)
	}
	return parseBare(input, defaultVault)
}

func parseURN(input string) (*NoteRef, error) {
	// Strip prefix - case insensitive
	nss := input[len(urnPrefix):]

	// Splot off anchor at first #
	var rawAnchor string
	if idx := strings.Index(nss, "#"); idx != -1 {
		rawAnchor = nss[idx+1:]
		nss = nss[:idx]
	}

	// Split NSS on ":" into exactly 4 fields: user, vault, type, identifier
	parts := strings.SplitN(nss, ":", 4)
	if len(parts) != 4 {
		return nil, fmt.Errorf("%w: expected 4 colon-separated fields", ErrInvalidURN)
	}
	user, vault, rtype, identifier := parts[0], parts[1], parts[2], parts[3]

	// Validate fields
	if user != "" {
		return nil, ErrUserReserved
	}
	vault, err := url.PathUnescape(vault)
	if err != nil || vault == "" {
		return nil, ErrEmptyVault
	}
	if rtype != "note" {
		return nil, fmt.Errorf("%w, %s", ErrUnknownType, rtype)
	}

	path, err := decodePath(identifier)
	if err != nil {
		return nil, err
	}

	// parse anchor
	anchor, err := parseAnchor(rawAnchor)
	if err != nil {
		return nil, err
	}

	return &NoteRef{Vault: vault, Path: path, Anchor: anchor}, nil
}

func parseBare(input, defaultVault string) (*NoteRef, error) {
	if defaultVault == "" {
		return nil, ErrEmptyVault
	}
	if input == "" {
		return nil, ErrEmptyPath
	}
	// Bare path: no percent-encoding expected, just clean it
	path := strings.TrimPrefix(input, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" || strings.Contains(path, "//") {
		return nil, ErrEmptyPath
	}
	return &NoteRef{Vault: defaultVault, Path: path}, nil
}


func decodePath(encoded string) (string, error) {
	if encoded == "" {
		return "", ErrEmptyPath
	}
	// Decode each segment
	segments := strings.Split(encoded, "/")
	for i, seg := range segments {
		if seg == "" {
		}
		decoded, err := url.PathUnescape(seg)
		if err != nil {
			return "", fmt.Errorf("%w: invalid encoding in path", ErrInvalidURN)
		}
		segments[i] = decoded
	}
	return strings.Join(segments, "/"), nil
}
func parseAnchor(raw string) (Anchor, error) {
	if raw == "" {
		return Anchor{}, nil
	}
	if strings.HasPrefix(raw, "^") {
		// Block reference
		blockID := raw[1:]
		if blockID == "" {
			return Anchor{}, fmt.Errorf("%w: empty block ID", ErrInvalidURN)
		}
		return Anchor{BlockID: blockID}, nil
	}
	// Heading path: split on # and decode each
	parts := strings.Split(raw, "#")
	headings := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue // tolerate trailing #
		}
		decoded, err := url.PathUnescape(p)
		if err != nil {
			return Anchor{}, fmt.Errorf("%w: invalid encoding in anchor", ErrInvalidURN)
		}
		headings = append(headings, decoded)
	}
	if len(headings) == 0 {
		return Anchor{}, fmt.Errorf("%w: empty anchor", ErrInvalidURN)
	}
	return Anchor{Headings: headings}, nil
}
