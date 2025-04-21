package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/nslaughter/codecourt/api-gateway/config"
)

// ServiceProxy represents a proxy for a microservice
type ServiceProxy struct {
	cfg *config.Config
}

// NewServiceProxy creates a new service proxy
func NewServiceProxy(cfg *config.Config) *ServiceProxy {
	return &ServiceProxy{
		cfg: cfg,
	}
}

// ProxyRequest proxies a request to the appropriate microservice
func (p *ServiceProxy) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Determine the target service based on the request path
	targetURL, err := p.getTargetURL(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Create a reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Modify the request to match the target URL
	r.URL.Host = targetURL.Host
	r.URL.Scheme = targetURL.Scheme
	r.Host = targetURL.Host

	// Remove the service prefix from the path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/v1")

	// Log the proxy request
	log.Printf("Proxying request to %s%s", targetURL.String(), r.URL.Path)

	// Serve the request
	proxy.ServeHTTP(w, r)
}

// getTargetURL determines the target URL based on the request path
func (p *ServiceProxy) getTargetURL(path string) (*url.URL, error) {
	var targetURLStr string

	// Determine the target service based on the path
	switch {
	case strings.HasPrefix(path, "/api/v1/problems"):
		targetURLStr = p.cfg.ProblemServiceURL
	case strings.HasPrefix(path, "/api/v1/submissions"):
		targetURLStr = p.cfg.SubmissionServiceURL
	case strings.HasPrefix(path, "/api/v1/judging"):
		targetURLStr = p.cfg.JudgingServiceURL
	case strings.HasPrefix(path, "/api/v1/auth"):
		targetURLStr = p.cfg.AuthServiceURL
	default:
		// Default to the problem service for now
		targetURLStr = p.cfg.ProblemServiceURL
	}

	// Parse the target URL
	return url.Parse(targetURLStr)
}

// ForwardRequest forwards a request to another service and returns the response
func (p *ServiceProxy) ForwardRequest(method, path string, body []byte, headers http.Header) (*http.Response, error) {
	// Determine the target service based on the path
	targetURL, err := p.getTargetURL(path)
	if err != nil {
		return nil, err
	}

	// Create a new URL with the target and path
	targetURL.Path = strings.TrimPrefix(path, "/api/v1")

	// Create a new request
	req, err := http.NewRequest(method, targetURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send the request
	client := &http.Client{}
	return client.Do(req)
}

// HandleResponse handles a response from a service
func (p *ServiceProxy) HandleResponse(w http.ResponseWriter, resp *http.Response) error {
	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy body
	_, err := io.Copy(w, resp.Body)
	return err
}
