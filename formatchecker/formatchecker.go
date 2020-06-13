package formatchecker

import (
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// from RFC 3339, section 5.6 [RFC3339]
// https://tools.ietf.org/html/rfc3339#section-5.6
func IsValidDateTime(dateTime string) error {
	if _, err := time.Parse(time.RFC3339, dateTime); err != nil {
		return err
	}
	return nil
}

// RFC 3339, section 5.6 [RFC3339]
// https://tools.ietf.org/html/rfc3339#section-5.6
func IsValidDate(date string) error {
	timeToAppend := "T00:00:00.0Z"
	dateTime := fmt.Sprintf("%s%s", date, timeToAppend)
	return IsValidDateTime(dateTime)
}

// RFC 3339, section 5.6 [RFC3339]
// https://tools.ietf.org/html/rfc3339#section-5.6
func IsValidTime(time string) error {
	dateToAppend := "1991-02-21"
	dateTime := fmt.Sprintf("%sT%s", dateToAppend, time)
	return IsValidDateTime(dateTime)
}

// RFC 5322, section 3.4.1 [RFC5322].
// https://tools.ietf.org/html/rfc5322#section-3.4.1
func IsValidEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return err
	}
	return nil
}

// RFC 6531 [RFC6531]
// https://tools.ietf.org/html/rfc6531
func IsValidIdnEmail(idnEmail string) error {
	if _, err := mail.ParseAddress(idnEmail); err != nil {
		return err
	}
	return nil
}

// RFC 1034, section 3.1 [RFC1034]
// https://tools.ietf.org/html/rfc1034#section-3.1
func IsValidHostname(hostname string) error {
	hostnamePattern := `^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`
	hostnamePatternCompiled := regexp.MustCompile(hostnamePattern)
	if len(hostname) > 255 {
		return errors.New("hostname is too long (more then 255 characters)")
	}
	if valid := hostnamePatternCompiled.MatchString(hostname); !valid {
		return errors.New(hostname + "is not valid hostname")
	}
	return nil
}

// RFC 1034 as for hostname, or
// an internationalized hostname as defined by RFC 5890, section
// 2.3.2.3 [RFC5890].
// https://tools.ietf.org/html/rfc1034
// https://tools.ietf.org/html/rfc5890#section-2.3.2.3
func IsValidIdnHostname(idnHostname string) error {
	disallowedIdnChars := map[string]bool{"\u0020": true, "\u002D": true, "\u00A2": true, "\u00A3": true,
		"\u00A4": true, "\u00A5": true, "\u034F": true, "\u0640": true, "\u07FA": true, "\u180B": true,
		"\u180C": true, "\u180D": true, "\u200B": true, "\u2060": true, "\u2104": true, "\u2108": true,
		"\u2114": true, "\u2117": true, "\u2118": true, "\u211E": true, "\u211F": true, "\u2123": true,
		"\u2125": true, "\u2282": true, "\u2283": true, "\u2284": true, "\u2285": true, "\u2286": true,
		"\u2287": true, "\u2288": true, "\u2616": true, "\u2617": true, "\u2619": true, "\u262F": true,
		"\u2638": true, "\u266C": true, "\u266D": true, "\u266F": true, "\u2752": true, "\u2756": true,
		"\u2758": true, "\u275E": true, "\u2761": true, "\u2775": true, "\u2794": true, "\u2798": true,
		"\u27AF": true, "\u27B1": true, "\u27BE": true, "\u3004": true, "\u3012": true, "\u3013": true,
		"\u3020": true, "\u302E": true, "\u302F": true, "\u3031": true, "\u3032": true, "\u3035": true,
		"\u303B": true, "\u3164": true, "\uFFA0": true}
	if len(idnHostname) > 255 {
		return errors.New("hostname is too long (more then 255 characters)")
	}
	for _, r := range idnHostname {
		s := string(r)
		if disallowedIdnChars[s] {
			return errors.New(fmt.Sprintf("invalid hostname: contains illegal character %#U", r))
		}
	}

	return nil
}

// RFC 2673, section 3.2 [RFC2673].
// https://tools.ietf.org/html/rfc2673#section-3.2
func IsValidIPv4(ipv4 string) error {
	parsed := net.ParseIP(ipv4)
	hasDots := strings.Contains(ipv4, ".")
	if parsed == nil || !hasDots {
		return errors.New("invalid ipv4 address " + ipv4)
	}

	return nil
}

// RFC 4291, section 2.2 [RFC4291].
// https://tools.ietf.org/html/rfc4291#section-2.2
func IsValidIPv6(ipv6 string) error {
	parsed := net.ParseIP(ipv6)
	hasColons := strings.Contains(ipv6, ":")
	if parsed == nil || !hasColons {
		return errors.New("invalid ipv6 address " + ipv6)
	}

	return nil
}

// RFC3986
// https://tools.ietf.org/html/rfc3986
func IsValidURI(uri string) error {
	schemePrefix := `^[^\:]+\:`
	schemePrefixPattern := regexp.MustCompile(schemePrefix)
	if _, err := url.Parse(uri); err != nil {
		return err
	}
	if !schemePrefixPattern.MatchString(uri) {
		return fmt.Errorf("uri missing scheme prefix")
	}
	return nil
}

// RFC3986
// https://tools.ietf.org/html/rfc3986
func IsValidUriRef(uriRef string) error {
	if _, err := url.Parse(uriRef); err != nil {
		return err
	}
	if strings.Contains(uriRef, "\\") {
		return errors.New("invalid uri-ref " + uriRef)
	}
	return nil
}

// A string instance is a valid against "iri" if it is a valid IRI,
// according to [RFC3987].
// https://tools.ietf.org/html/rfc3987
func IsValidIri(iri string) error {
	return IsValidURI(iri)
}

// A string instance is a valid against "iri-reference" if it is a
// valid IRI Reference (either an IRI or a relative-reference),
// according to [RFC3987].
// https://tools.ietf.org/html/rfc3987
func IsValidIriRef(iriRef string) error {
	return IsValidUriRef(iriRef)
}

// A string instance is a valid against "uri-template" if it is a
// valid URI Template (of any level), according to [RFC6570]. Note
// that URI Templates may be used for IRIs; there is no separate IRI
// Template specification.
// https://tools.ietf.org/html/rfc6570
func IsValidURITemplate(uriTemplate string) error {
	//uriTemplatePattern := regexp.MustCompile(`\{[^\{\}\\]*\}`)
	uriTemplatePattern := regexp.MustCompile(`{[^{}\\]*}`)
	arbitraryValue := "tmp"
	uriRef := uriTemplatePattern.ReplaceAllString(uriTemplate, arbitraryValue)
	if strings.Contains(uriRef, "{") || strings.Contains(uriRef, "}") {
		return errors.New("invalid uri template " + uriTemplate)
	}
	return IsValidUriRef(uriRef)
}

// RFC 6901, section 5 [RFC6901].
// https://tools.ietf.org/html/rfc6901#section-5
func IsValidJSONPointer(jsonPointer string) error {
	unescapedTilda := `\~[^01]`
	endingTilda := `\~$`
	unescaptedTildaPattern := regexp.MustCompile(unescapedTilda)
	endingTildaPattern := regexp.MustCompile(endingTilda)

	if len(jsonPointer) == 0 {
		return nil
	}
	if jsonPointer[0] != '/' {
		return errors.New("non-empty references must begin with a '/' character: " + jsonPointer)
	}
	str := jsonPointer[1:]
	if unescaptedTildaPattern.MatchString(str) {
		return errors.New("unescaped tilda error")
	}
	if endingTildaPattern.MatchString(str) {
		return errors.New("ending tilda error")
	}
	return nil
}

// https://tools.ietf.org/html/draft-handrews-relative-json-pointer-00
func IsValidRelJSONPointer(relJSONPointer string) error {
	parts := strings.Split(relJSONPointer, "/")
	if len(parts) == 1 {
		parts = strings.Split(relJSONPointer, "#")
	}
	if i, err := strconv.Atoi(parts[0]); err != nil || i < 0 {
		return err
	}
	//skip over first part
	str := relJSONPointer[len(parts[0]):]
	if len(str) > 0 && str[0] == '#' {
		return nil
	}
	return IsValidJSONPointer(str)
}

// http://www.ecma-international.org/publications/files/ECMA-ST/Ecma-262.pdf
// https://tools.ietf.org/html/rfc7159
func IsValidRegex(regex string) error {
	if _, err := regexp.Compile(regex); err != nil {
		return err
	}
	return nil
}
