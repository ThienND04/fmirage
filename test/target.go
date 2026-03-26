package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const demoToken = "dev-token"

type serverState struct {
	startedAt    time.Time
	requestCount uint64

	mu       sync.Mutex
	hitByIP  map[string]int
	users    []apiUser
	products []apiProduct
}

type apiUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type apiProduct struct {
	ID       int     `json:"id"`
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	PriceUSD float64 `json:"price_usd"`
	Stock    int     `json:"stock"`
}

func main() {
	port := flag.Int("port", 8080, "Port to run test target server")
	flag.Parse()

	state := &serverState{
		startedAt: time.Now(),
		hitByIP:   make(map[string]int),
		users: []apiUser{
			{ID: 1, Name: "Alice Nguyen", Email: "alice@example.com", Role: "admin"},
			{ID: 2, Name: "Bob Tran", Email: "bob@example.com", Role: "editor"},
			{ID: 3, Name: "Carol Pham", Email: "carol@example.com", Role: "viewer"},
			{ID: 4, Name: "David Le", Email: "david@example.com", Role: "viewer"},
			{ID: 5, Name: "Emma Vu", Email: "emma@example.com", Role: "editor"},
		},
		products: []apiProduct{
			{ID: 101, SKU: "LAP-14-STD", Name: "Laptop 14 Standard", PriceUSD: 799.0, Stock: 17},
			{ID: 102, SKU: "MON-27-PRO", Name: "Monitor 27 Pro", PriceUSD: 289.0, Stock: 8},
			{ID: 103, SKU: "KEY-MECH-RGB", Name: "Mechanical Keyboard RGB", PriceUSD: 119.0, Stock: 31},
			{ID: 104, SKU: "MOU-WL-ERG", Name: "Wireless Ergonomic Mouse", PriceUSD: 59.0, Stock: 44},
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddUint64(&state.requestCount, 1)
		path := r.URL.Path
		q := r.URL.Query()

		if delayMs := parseIntWithDefault(q.Get("delay"), 0); delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}

		if location := q.Get("redirect"); location != "" {
			status := parseIntWithDefault(q.Get("status"), http.StatusFound)
			if status < 300 || status > 399 {
				status = http.StatusFound
			}
			http.Redirect(w, r, location, status)
			return
		}

		if q.Get("status") != "" {
			writeCustomResponse(w, q)
			return
		}

		if strings.HasPrefix(path, "/api/") && state.rateLimited(r) {
			w.Header().Set("Retry-After", "5")
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests from your IP",
			})
			return
		}

		switch {
		case r.Method == http.MethodGet && path == "/":
			serveHomePage(w)
			return
		case r.Method == http.MethodGet && path == "/robots.txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin\nAllow: /\n"))
			return
		case r.Method == http.MethodGet && path == "/assets/app.js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			_, _ = w.Write([]byte("console.log('Mock frontend loaded');"))
			return
		case r.Method == http.MethodGet && path == "/healthz":
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "fmirage-mock"})
			return
		case r.Method == http.MethodGet && path == "/readyz":
			if q.Get("fail") == "1" {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not_ready"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
			return
		case r.Method == http.MethodGet && path == "/metrics":
			writePrometheusLikeMetrics(w, current, time.Since(state.startedAt))
			return
		case r.Method == http.MethodPost && path == "/login":
			handleLogin(w, r)
			return
		case r.Method == http.MethodGet && path == "/api/v1/me":
			if !checkAuth(w, r) {
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"id": 999, "name": "Demo User", "role": "tester"})
			return
		case r.Method == http.MethodGet && path == "/api/v1/users":
			if !checkAuth(w, r) {
				return
			}
			handleListUsers(w, r, state.users)
			return
		case r.Method == http.MethodGet && path == "/api/v1/products":
			if !checkAuth(w, r) {
				return
			}
			handleListProducts(w, r, state.products)
			return
		case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/orders/"):
			if !checkAuth(w, r) {
				return
			}
			handleGetOrder(w, r)
			return
		case r.Method == http.MethodGet && path == "/api/v1/flaky":
			handleFlakyEndpoint(w, r)
			return
		case r.Method == http.MethodGet && path == "/api/v1/slow":
			handleSlowEndpoint(w, r)
			return
		}

		// Simulate known sensitive routes.
		if path == "/admin" || path == "/config.php" || path == "/api" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("BINGO! Ban da tim thay thu muc an!"))
			return
		}

		// Simulate a WAF-like page with fixed size.
		if strings.HasPrefix(path, "/fake_") {
			w.WriteHeader(http.StatusOK)
			fakeHTML := fmt.Sprintf("%-100s", "<html><body>Please verify you are human (Fake CAPTCHA)</body></html>")
			_, _ = w.Write([]byte(fakeHTML))
			return
		}

		// Simulate common status variants by route.
		switch path {
		case "/forbidden":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("Forbidden"))
			return
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
			return
		case "/redirect":
			http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
			return
		}

		fmt.Printf("[%d] Received request for: %s\n", current, path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	})

	fmt.Printf("[+] Test target running at: http://localhost:%d\n", *port)
	fmt.Println("[+] Custom query options: ?status=CODE&size=BYTES&words=N&lines=N&delay=MS&redirect=/path")

	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}

func writeCustomResponse(w http.ResponseWriter, q map[string][]string) {
	status := parseIntWithDefault(firstValue(q, "status"), http.StatusOK)
	if status < 100 || status > 599 {
		status = http.StatusOK
	}

	body := firstValue(q, "body")
	if body == "" {
		body = "custom response"
	}

	words := parseIntWithDefault(firstValue(q, "words"), 0)
	if words > 0 {
		body = strings.TrimSpace(strings.Repeat("word ", words))
	}

	lines := parseIntWithDefault(firstValue(q, "lines"), 0)
	if lines > 0 {
		line := body
		if line == "" {
			line = "line"
		}
		body = strings.TrimSuffix(strings.Repeat(line+"\n", lines), "\n")
	}

	size := parseIntWithDefault(firstValue(q, "size"), 0)
	if size > 0 {
		body = fitToSize(body, size)
	}

	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func fitToSize(body string, size int) string {
	if size <= 0 {
		return body
	}

	if len(body) >= size {
		return body[:size]
	}

	padding := strings.Repeat("x", size-len(body))
	return body + padding
}

func parseIntWithDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return v
}

func firstValue(q map[string][]string, key string) string {
	v := q[key]
	if len(v) == 0 {
		return ""
	}

	return v[0]
}

func serveHomePage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>FMirage Mock Site</title>
</head>
<body>
  <h1>FMirage Demo Website</h1>
  <p>This is a realistic mock target for fuzzing.</p>
  <ul>
    <li><a href="/healthz">/healthz</a></li>
    <li><a href="/readyz">/readyz</a></li>
    <li><a href="/api/v1/users">/api/v1/users (auth required)</a></li>
    <li><a href="/api/v1/products">/api/v1/products (auth required)</a></li>
  </ul>
  <script src="/assets/app.js"></script>
</body>
</html>`
	_, _ = w.Write([]byte(page))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writePrometheusLikeMetrics(w http.ResponseWriter, count uint64, uptime time.Duration) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(fmt.Sprintf("requests_total %d\n", count)))
	_, _ = w.Write([]byte(fmt.Sprintf("uptime_seconds %.0f\n", uptime.Seconds())))
}

func checkAuth(w http.ResponseWriter, r *http.Request) bool {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "Bearer "+demoToken {
		return true
	}

	writeJSON(w, http.StatusUnauthorized, map[string]any{
		"error":   "unauthorized",
		"message": "Use Authorization: Bearer dev-token",
	})
	return false
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	type loginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var in loginInput
	_ = json.NewDecoder(r.Body).Decode(&in)
	_ = r.Body.Close()

	if in.Username == "" || in.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "username and password are required"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": demoToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func handleListUsers(w http.ResponseWriter, r *http.Request, users []apiUser) {
	page := max(1, parseIntWithDefault(r.URL.Query().Get("page"), 1))
	limit := clamp(parseIntWithDefault(r.URL.Query().Get("limit"), 2), 1, 100)
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))

	filtered := users
	if search != "" {
		filtered = make([]apiUser, 0, len(users))
		for _, u := range users {
			if strings.Contains(strings.ToLower(u.Name), search) || strings.Contains(strings.ToLower(u.Email), search) {
				filtered = append(filtered, u)
			}
		}
	}

	items, total, totalPages := paginateUsers(filtered, page, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"meta": map[string]any{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func handleListProducts(w http.ResponseWriter, r *http.Request, products []apiProduct) {
	page := max(1, parseIntWithDefault(r.URL.Query().Get("page"), 1))
	limit := clamp(parseIntWithDefault(r.URL.Query().Get("limit"), 2), 1, 100)

	start := (page - 1) * limit
	if start >= len(products) {
		writeJSON(w, http.StatusOK, map[string]any{"items": []apiProduct{}, "meta": map[string]any{"page": page, "limit": limit, "total": len(products)}})
		return
	}

	end := start + limit
	if end > len(products) {
		end = len(products)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": products[start:end],
		"meta":  map[string]any{"page": page, "limit": limit, "total": len(products)},
	})
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_order_id"})
		return
	}

	if id == "404" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "order_not_found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":     id,
		"status": "processing",
		"items": []map[string]any{
			{"sku": "LAP-14-STD", "qty": 1, "price": 799.0},
			{"sku": "MOU-WL-ERG", "qty": 2, "price": 59.0},
		},
	})
}

func handleFlakyEndpoint(w http.ResponseWriter, r *http.Request) {
	failRate := clamp(parseIntWithDefault(r.URL.Query().Get("failRate"), 30), 0, 100)
	n := rand.Intn(100)
	if n < failRate {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "random_failure", "chance": failRate})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "random": n})
}

func handleSlowEndpoint(w http.ResponseWriter, r *http.Request) {
	baseDelay := clamp(parseIntWithDefault(r.URL.Query().Get("base"), 250), 0, 10000)
	jitter := clamp(parseIntWithDefault(r.URL.Query().Get("jitter"), 400), 0, 10000)
	if jitter > 0 {
		baseDelay += rand.Intn(jitter)
	}
	time.Sleep(time.Duration(baseDelay) * time.Millisecond)

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "delay_ms": baseDelay})
}

func paginateUsers(items []apiUser, page, limit int) ([]apiUser, int, int) {
	total := len(items)
	totalPages := 1
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	start := (page - 1) * limit
	if start >= total {
		return []apiUser{}, total, totalPages
	}

	end := start + limit
	if end > total {
		end = total
	}

	return items[start:end], total, totalPages
}

func (s *serverState) rateLimited(r *http.Request) bool {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.hitByIP[ip]++
	return s.hitByIP[ip] > 80
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func clamp(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
