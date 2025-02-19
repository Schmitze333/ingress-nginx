/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package redirect

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	networking "k8s.io/api/networking/v1"

	"k8s.io/ingress-nginx/internal/ingress/annotations/parser"
	"k8s.io/ingress-nginx/internal/ingress/errors"
	"k8s.io/ingress-nginx/internal/ingress/resolver"
)

const (
	defRedirectURL = "http://some-site.com"
)

func TestPermanentRedirectWithDefaultCode(t *testing.T) {
	rp := NewParser(resolver.Mock{})
	if rp == nil {
		t.Fatalf("Expected a parser.IngressAnnotation but returned nil")
	}

	ing := new(networking.Ingress)

	data := make(map[string]string, 1)
	data[parser.GetAnnotationWithPrefix(permanentRedirectAnnotation)] = defRedirectURL
	ing.SetAnnotations(data)

	i, err := rp.Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("Expected a Redirect type")
	}
	if redirect.URL != defRedirectURL {
		t.Errorf("Expected %v as redirect but returned %s", defRedirectURL, redirect.URL)
	}
	if redirect.Code != defaultPermanentRedirectCode {
		t.Errorf("Expected %v as redirect to have a code %d but had %d", defRedirectURL, defaultPermanentRedirectCode, redirect.Code)
	}
}

func TestPermanentRedirectWithCustomCode(t *testing.T) {
	rp := NewParser(resolver.Mock{})
	if rp == nil {
		t.Fatalf("Expected a parser.IngressAnnotation but returned nil")
	}

	testCases := map[string]struct {
		input        int
		expectOutput int
	}{
		"valid code":   {http.StatusPermanentRedirect, http.StatusPermanentRedirect},
		"invalid code": {http.StatusTeapot, defaultPermanentRedirectCode},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ing := new(networking.Ingress)

			data := make(map[string]string, 2)
			data[parser.GetAnnotationWithPrefix(permanentRedirectAnnotation)] = defRedirectURL
			data[parser.GetAnnotationWithPrefix(permanentRedirectAnnotationCode)] = strconv.Itoa(tc.input)
			ing.SetAnnotations(data)

			i, err := rp.Parse(ing)
			if err != nil {
				t.Errorf("Unexpected error with ingress: %v", err)
			}
			redirect, ok := i.(*Config)
			if !ok {
				t.Errorf("Expected a redirect Config type")
			}
			if redirect.URL != defRedirectURL {
				t.Errorf("Expected %v as redirect but returned %s", defRedirectURL, redirect.URL)
			}
			if redirect.Code != tc.expectOutput {
				t.Errorf("Expected %v as redirect to have a code %d but had %d", defRedirectURL, tc.expectOutput, redirect.Code)
			}
		})
	}
}

func TestTemporalRedirectWithDefaultCode(t *testing.T) {
	rp := NewParser(resolver.Mock{})
	if rp == nil {
		t.Fatalf("Expected a parser.IngressAnnotation but returned nil")
	}

	ing := new(networking.Ingress)

	data := make(map[string]string, 1)
	data[parser.GetAnnotationWithPrefix(fromToWWWRedirAnnotation)] = "true"
	data[parser.GetAnnotationWithPrefix(temporalRedirectAnnotation)] = defRedirectURL
	ing.SetAnnotations(data)

	i, err := rp.Parse(ing)
	if err != nil {
		t.Errorf("Unexpected error with ingress: %v", err)
	}
	redirect, ok := i.(*Config)
	if !ok {
		t.Errorf("Expected a Redirect type")
	}
	if redirect.URL != defRedirectURL {
		t.Errorf("Expected %v as redirect but returned %s", defRedirectURL, redirect.URL)
	}
	if redirect.Code != http.StatusFound {
		t.Errorf("Expected %v as redirect to have a code %d but had %d", defRedirectURL, http.StatusFound, redirect.Code)
	}
}

func TestTemporalRedirectWithCustomCode(t *testing.T) {
	rp := NewParser(resolver.Mock{})
	if rp == nil {
		t.Fatalf("Expected a parser.IngressAnnotation but returned nil")
	}

	testCases := map[string]struct {
		input        int
		expectOutput int
	}{
		"valid code":   {http.StatusTemporaryRedirect, http.StatusTemporaryRedirect},
		"invalid code": {http.StatusTeapot, http.StatusFound},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ing := new(networking.Ingress)

			data := make(map[string]string, 2)
			data[parser.GetAnnotationWithPrefix(fromToWWWRedirAnnotation)] = "true"
			data[parser.GetAnnotationWithPrefix(temporalRedirectAnnotation)] = defRedirectURL
			data[parser.GetAnnotationWithPrefix(temporalRedirectAnnotationCode)] = strconv.Itoa(tc.input)
			ing.SetAnnotations(data)

			i, err := rp.Parse(ing)
			if err != nil {
				t.Errorf("Unexpected error with ingress: %v", err)
			}
			redirect, ok := i.(*Config)
			if !ok {
				t.Errorf("Expected a Redirect type")
			}
			if redirect.URL != defRedirectURL {
				t.Errorf("Expected %v as redirect but returned %s", defRedirectURL, redirect.URL)
			}
			if redirect.Code != tc.expectOutput {
				t.Errorf("Expected %v as redirect to have a code %d but had %d", defRedirectURL, tc.expectOutput, redirect.Code)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	invalid := "ok.com"
	urlParse, err := url.Parse(invalid)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	expected := errors.Errorf("only http and https are valid protocols (%v)", urlParse.Scheme)
	err = isValidURL(invalid)
	if reflect.DeepEqual(expected.Error, err.Error) {
		t.Errorf("expected '%v' but got '%v'", expected, err)
	}

	valid := "http://ok.com"
	err = isValidURL(valid)
	if err != nil {
		t.Errorf("expected nil but got %v", err)
	}
}

func TestParseAnnotations(t *testing.T) {
	ing := new(networking.Ingress)

	data := map[string]string{}
	data[parser.GetAnnotationWithPrefix(relativeRedirectsAnnotation)] = "true"
	ing.SetAnnotations(data)

	_, err := NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test ingress using the annotation without a TLS section
	_, err = NewParser(&resolver.Mock{}).Parse(ing)
	if err != nil {
		t.Errorf("unexpected error parsing ingress with relative-redirects")
	}
}
