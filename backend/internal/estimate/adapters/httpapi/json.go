package httpapi

// appendJSONString дописывает строку в JSON-буфер без HTML-экранирования.
func appendJSONString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for _, r := range s {
		switch r {
		case '"':
			dst = append(dst, '\\', '"')
		case '\\':
			dst = append(dst, '\\', '\\')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			if r < 0x20 {
				dst = append(dst, '\\', 'u', '0', '0',
					"0123456789abcdef"[byte(r)>>4], "0123456789abcdef"[byte(r)&0xF])
				continue
			}
			dst = append(dst, string(r)...)
		}
	}
	return append(dst, '"')
}
