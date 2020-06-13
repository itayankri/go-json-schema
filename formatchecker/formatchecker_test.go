package formatchecker_test

import (
	"testing"

	"github.com/omeryahud/caf/internal/pkg/validators/formatchecker"
)

type test struct {
	data        string
	valid       bool
	description string
}

type format func(string) error

const succeed = "V"
const failed = "X"

const (
	FORMAT_DATE_TIME             = "date-time"
	FORMAT_TIME                  = "time"
	FORMAT_DATE                  = "date"
	FORMAT_EMAIL                 = "email"
	FORMAT_IDN_EMAIL             = "idn-email"
	FORMAT_HOSTNAME              = "hostname"
	FORMAT_IDN_HOSTNAME          = "idn-hostname"
	FORMAT_IPV4                  = "ipv4"
	FORMAT_IPV6                  = "ipv6"
	FORMAT_URI                   = "uri"
	FORMAT_URI_REFERENCE         = "uri-reference"
	FORMAT_IRI                   = "iri"
	FORMAT_IRI_REFERENCE         = "iri-reference"
	FORMAT_URI_TEMPLATE          = "uri-template"
	FORMAT_JSON_POINTER          = "json-pointer"
	FORMAT_RELATIVE_JSON_POINTER = "relative-json-pointer"
	FORMAT_REGEX                 = "regex"
)

func TestIsValidDateTime(t *testing.T) {
	testCases := []test{
		{
			description: "a valid date-time string",
			data:        "1985-04-12T23:20:50.52Z",
			valid:       true,
		},
		{
			description: "a valid date-time string",
			data:        "1996-12-19T16:39:57-08:00",
			valid:       true,
		},
		{
			description: "an invalid date-time string",
			data:        "06/19/1963 08:30:06 PST",
			valid:       false,
		},
	}

	isValidFormat(t, testCases, FORMAT_DATE_TIME, formatchecker.IsValidDateTime)
}

func TestIsValidDate(t *testing.T) {
	testCases := []test{
		{
			description: "a valid date string",
			data:        "1963-06-19",
			valid:       true,
		},
		{
			description: "an invalid date string (/ is invalid)",
			data:        "06/19/1963",
			valid:       false,
		},
		{
			description: "an invalid RFC3339 date",
			data:        "02-2002",
			valid:       false,
		},
		{
			description: "an invalid month 350",
			data:        "2010-350",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_DATE, formatchecker.IsValidDate)
}

func TestIsValidTime(t *testing.T) {
	testCases := []test{
		{
			description: "a valid time",
			data:        "08:30:06.283185Z",
			valid:       true,
		},
		{
			description: "a valid time",
			data:        "10:05:08+01:00",
			valid:       true,
		},
		{
			description: "an invalid time",
			data:        "09:45:10 PST",
			valid:       false,
		},
		{
			description: "an invalid RFC3339 time",
			data:        "01:02:03,121212",
			valid:       false,
		},
		{
			description: "an invalid seconds",
			data:        "45:59:62",
			valid:       false,
		},
		{
			description: "an invalid time",
			data:        "1234",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_TIME, formatchecker.IsValidTime)
}

func TestIsValidEmail(t *testing.T) {
	testCases := []test{
		{
			description: "a valid email",
			data:        "john@example.com",
			valid:       true,
		},
		{
			description: "an invalid email address",
			data:        "@",
			valid:       false,
		},
		{
			description: "@ is missing",
			data:        "john(at)example.com",
			valid:       false,
		},
		{
			description: "an invalid email address",
			data:        "1234",
			valid:       false,
		},
		{
			description: "an invalid email address",
			data:        "",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_EMAIL, formatchecker.IsValidEmail)
}

func TestIsValidIdnEmail(t *testing.T) {
	testCases := []test{
		{
			description: "a valid idn email (example@example.test in Hangul)",
			data:        "실례@실례.테스트",
			valid:       true,
		},
		{
			description: "a valid idn email",
			data:        "john@example.com",
			valid:       true,
		},
		{
			description: "an invalid idn email",
			data:        "1234",
			valid:       false,
		},
		{
			description: "an invalid idn email",
			data:        "",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_IDN_EMAIL, formatchecker.IsValidIdnEmail)
}

func TestIsValidHostname(t *testing.T) {
	testCases := []test{
		{
			description: "a valid host name",
			data:        "www.example.com",
			valid:       true,
		},
		{
			description: "a valid host name",
			data:        "xn--4gbwdl.xn--wgbh1c",
			valid:       true,
		},
		{
			description: "a host name containing illegal characters (_)",
			data:        "not_a_valid_host_name",
			valid:       false,
		},
		{
			description: "a host name starting with an illegal character",
			data:        "-a-host-name-that-starts-with--",
			valid:       false,
		},
		{
			description: "a host name with a component too long",
			data: "a-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-long-host-name-component",
			valid: false,
		},
	}
	isValidFormat(t, testCases, FORMAT_HOSTNAME, formatchecker.IsValidHostname)
}

func TestIsValidIdnHostname(t *testing.T) {
	testCases := []test{
		{
			description: "a valid host name (example.test in Hangul)",
			data:        "실례.테스트",
			valid:       true,
		},
		{
			description: "illegal first char",
			data:        "〮실례.테스트",
			valid:       false,
		},
		{
			description: "contains illegal",
			data:        "실〮례.테스트",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_IDN_HOSTNAME, formatchecker.IsValidIdnHostname)
}

func TestIsValidIPv4(t *testing.T) {
	testCases := []test{
		{
			description: "a valid IPv4 address",
			data:        "192.168.0.1",
			valid:       true,
		},
		{
			description: "too many components",
			data:        "127.0.0.0.1",
			valid:       false,
		},
		{
			description: "IPv4 out of range",
			data:        "256.256.256.256",
			valid:       false,
		},
		{
			description: "not enough components (4 needed)",
			data:        "127",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_IPV4, formatchecker.IsValidIPv4)
}

func TestIsValidIPv6(t *testing.T) {
	testCases := []test{
		{
			description: "a valid IPv6 address",
			data:        "::1",
			valid:       true,
		},
		{
			description: "IPv6 out of range",
			data:        "12345::",
			valid:       false,
		},
		{
			description: "too many components",
			data:        "1:1:1:1:1",
			valid:       false,
		},
		{
			description: "IPv6 containing illegal characters",
			data:        "::string",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_IPV6, formatchecker.IsValidIPv6)
}

func TestIsValidURI(t *testing.T) {
	testCases := []test{
		{
			description: "a valid URL",
			data:        "http://foo.bar/?baz=qux#quux",
			valid:       true,
		},
		{
			description: "a valid URL with URL-encoded",
			data:        "http://foo.bar/?q=Test%20URL-encoded%20stuff",
			valid:       true,
		},
		{
			description: "a valid URL with  special characters",
			data:        "http://-.~_!$&'()*+,;=:%40:80%2f::::::@example.com",
			valid:       true,
		},
		{
			description: "a valid URL for a simple text file",
			data:        "http: //www.fff.com/rfc/rfc2396.txt",
			valid:       true,
		},
		{
			description: "an invalid URI with spaces",
			data:        "http:// shouldfail.com",
			valid:       false,
		},
		{
			description: "an invalid URI missing scheme",
			data:        ":// houldfail",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_URI, formatchecker.IsValidURI)

}

func TestIsValidUriRef(t *testing.T) {
	testCases := []test{
		{
			description: "a valid uri reference",
			data:        "aaa/bbb.html",
			valid:       true,
		},
		{
			description: "a valid uri reference",
			data:        "?a=b",
			valid:       true,
		},
		{
			description: "a valid uri reference",
			data:        "#fragment",
			valid:       true,
		},
		{
			description: "a valid uri reference",
			data:        "http://example.com",
			valid:       true,
		},
		{
			description: "an invalid URI fragment",
			data:        "#frag\\ment",
			valid:       false,
		},
		{
			description: "an invalid URI Reference",
			data:        "\\\\WINDOWS\\fileshare",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_URI_REFERENCE, formatchecker.IsValidUriRef)

}

func TestIsValidIri(t *testing.T) {
	testCases := []test{
		{
			description: "a valid IRI with anchor tag",
			data:        "http://ƒøø.ßår/?∂éœ=πîx#πîüx",
			valid:       true,
		},
		{
			description: "a valid IRI with anchor tag and parantheses",
			data:        "http://ƒøø.com/blah_(wîkïpédiå)_blah#ßité-1",
			valid:       true,
		},
		{
			description: "an invalid IRI",
			data:        "http:// ƒøø.com",
			valid:       false,
		},
		{
			description: "an invalid relative IRI Reference",
			data:        "/abc",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_IRI, formatchecker.IsValidIri)
}

func TestIsValidIriRef(t *testing.T) {
	testCases := []test{
		{
			description: "a valid IRI",
			data:        "http://ƒøø.ßår/?∂éœ=πîx#πîüx",
			valid:       true,
		},
		{
			description: "a valid IRI fragment",
			data:        "#ƒrägmênt",
			valid:       true,
		},
		{
			description: "a valid IRI",
			data:        "http://ƒøø.com/blah_(wîkïpédiå)_blah#ßité-1",
			valid:       true,
		},
		{
			description: "an invalid IRI Reference",
			data:        "\\\\WINDOWS\\filëßåré",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_IRI_REFERENCE, formatchecker.IsValidIriRef)
}

func TestIsValidURITemplate(t *testing.T) {
	testCases := []test{
		{
			description: "a valid URI template",
			data:        "http://example.com/dictionary/{term:1}/{term}",
			valid:       true,
		},
		{
			description: "a valid relative URI template",
			data:        "dictionary/{term:1}/{term}",
			valid:       true,
		},
		{
			description: "an invalid URI template",
			data:        "http://example.com/dictionary/{term:1}/{term",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_URI_TEMPLATE, formatchecker.IsValidURITemplate)
}

func TestIsValidJSONPointer(t *testing.T) {
	testCases := []test{
		{
			description: "a valid JSON-pointer",
			data:        "/foo/bar~0/baz~1/%a",
			valid:       true,
		},
		{
			description: "valid JSON-pointer",
			data:        "",
			valid:       true,
		},
		{
			description: "valid JSON-pointer",
			data:        "/foo/0",
			valid:       true,
		},
		{
			description: "valid JSON-pointer",
			data:        "/",
			valid:       true,
		},
		{
			description: "valid JSON-pointer",
			data:        "/a~1b",
			valid:       true,
		},
		{
			description: "valid JSON-pointer",
			data:        "/ ",
			valid:       true,
		},
		{
			description: "invalid JSON-pointer (~ not escaped)",
			data:        "/foo/bar~",
			valid:       false,
		},
		{
			description: "invalid JSON-pointer (URI Fragment Identifier)",
			data:        "#/",
			valid:       false,
		},
		{
			description: "invalid JSON-pointer (URI Fragment Identifier)",
			data:        "#a",
			valid:       false,
		},
		{
			description: "not a valid JSON-pointer (isn't empty nor starts with /)",
			data:        "0",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_JSON_POINTER, formatchecker.IsValidJSONPointer)
}

func TestIsValidRelJSONPointer(t *testing.T) {
	testCases := []test{
		{
			description: "a valid relative json pointer",
			data:        "0/a/b",
			valid:       true,
		},
		{
			description: "a valid relative json pointer",
			data:        "5/a/b#",
			valid:       true,
		},
		{
			description: "a valid relative json pointer",
			data:        "2#",
			valid:       true,
		},
		{
			description: "a valid relative json pointer",
			data:        "2",
			valid:       true,
		},
		{
			description: "an invalid relative json pointer",
			data:        "/a/b",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_RELATIVE_JSON_POINTER, formatchecker.IsValidRelJSONPointer)
}

func TestIsValidRegex(t *testing.T) {
	testCases := []test{
		{
			description: "a valid regex",
			data:        "^[a-z]+$",
			valid:       true,
		},
		{
			description: "incomplete group",
			data:        "(a",
			valid:       false,
		},
	}
	isValidFormat(t, testCases, FORMAT_REGEX, formatchecker.IsValidRegex)
}

func isValidFormat(t *testing.T, tests []test, formatType string, fn format) {
	t.Logf("Given the need to test %s format", formatType)
	{
		for index, testCase := range tests {
			t.Logf("\tTest %d: When trying to format %s => %s", index, testCase.data, testCase.description)
			{
				var valid bool
				if err := fn(testCase.data); err != nil {
					valid = false
				} else {
					valid = true
				}

				if valid != testCase.valid {
					t.Errorf("\t%s\tShould get valid = %t but got valid = %t", failed, testCase.valid, valid)
				} else {
					t.Logf("\t%s\tvalid = %t", succeed, testCase.valid)
				}
			}
		}
	}
}
