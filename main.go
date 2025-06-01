package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Yashh56/web-crawler/utils"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
)

// ANSI color codes for better output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

type Queue struct {
	totalQueued int
	number      int
	elements    []string
	mu          sync.Mutex
}

func (q *Queue) enqueue(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.elements = append(q.elements, url)
	q.totalQueued++
	q.number++
}

func (q *Queue) dequeue() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.elements) == 0 {
		return ""
	}
	url := q.elements[0]
	q.elements = q.elements[1:]
	q.number--
	return url
}

func (q *Queue) size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.number
}

type CrawledSet struct {
	data   map[uint64]bool
	number int
	mu     sync.Mutex
}

func (c *CrawledSet) add(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[hashUrl(url)] = true
	c.number++
}

func (c *CrawledSet) contains(url string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data[hashUrl(url)]
}

func (c *CrawledSet) size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.number
}

// Enhanced crawler stats with more metrics
type CrawlerStats struct {
	startTime       time.Time
	lastUpdateTime  time.Time
	totalPages      int
	successfulPages int
	failedPages     int
	totalBytes      int64
	pagesPerSecond  float64
	errorsPerMinute int
	averagePageSize int64
	lastMinutePages int
	mu              sync.RWMutex
}

func NewCrawlerStats() *CrawlerStats {
	now := time.Now()
	return &CrawlerStats{
		startTime:      now,
		lastUpdateTime: now,
	}
}

func (cs *CrawlerStats) updateSuccess(pageSize int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.successfulPages++
	cs.totalPages++
	cs.totalBytes += pageSize
	if cs.totalPages > 0 {
		cs.averagePageSize = cs.totalBytes / int64(cs.totalPages)
	}
}

func (cs *CrawlerStats) updateFailure() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.failedPages++
	cs.totalPages++
}

func (cs *CrawlerStats) calculateRate() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	elapsed := time.Since(cs.startTime).Seconds()
	if elapsed > 0 {
		cs.pagesPerSecond = float64(cs.successfulPages) / elapsed
	}
}

func (cs *CrawlerStats) getStats() (int, int, int, float64, int64) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.successfulPages, cs.failedPages, cs.totalPages, cs.pagesPerSecond, cs.averagePageSize
}

// Progress display functions
func clearLine() {
	fmt.Print("\r\033[K")
}

func moveCursorUp(lines int) {
	fmt.Printf("\033[%dA", lines)
}

func printHeader() {
	fmt.Printf("%s%s╔══════════════════════════════════════════════════════════════════════════════╗%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s║                            🕷️  WEB CRAWLER STATUS                            ║%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s%s╚══════════════════════════════════════════════════════════════════════════════╝%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Println()
}

func printProgress(crawled, queued, maxPages int, stats *CrawlerStats, currentUrl, currentTitle string, crawlLog []string) {
	// Move cursor up to overwrite previous output
	moveCursorUp(15) // Increased for more lines

	successful, failed, total, rate, avgSize := stats.getStats()

	// Progress bar
	progress := float64(crawled) / float64(maxPages)
	barWidth := 50
	filledWidth := int(progress * float64(barWidth))

	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)

	fmt.Printf("%s📊 Progress:%s [%s%s%s] %.1f%% (%d/%d)\n",
		ColorBold, ColorReset, ColorGreen, bar, ColorReset, progress*100, crawled, maxPages)

	// Current stats
	fmt.Printf("%s🎯 Queue Size:%s %s%-8d%s %s⚡ Rate:%s %s%.2f pages/sec%s\n",
		ColorBold, ColorReset, ColorYellow, queued, ColorReset,
		ColorBold, ColorReset, ColorGreen, rate, ColorReset)

	fmt.Printf("%s✅ Successful:%s %s%-8d%s %s❌ Failed:%s %s%-8d%s\n",
		ColorBold, ColorReset, ColorGreen, successful, ColorReset,
		ColorBold, ColorReset, ColorRed, failed, ColorReset)

	fmt.Printf("%s📈 Total Processed:%s %s%-6d%s %s📦 Avg Size:%s %s%s%s\n",
		ColorBold, ColorReset, ColorCyan, total, ColorReset,
		ColorBold, ColorReset, ColorPurple, formatBytes(avgSize), ColorReset)

	// Runtime info
	elapsed := time.Since(stats.startTime)
	fmt.Printf("%s⏱️  Runtime:%s %s%s%s\n",
		ColorBold, ColorReset, ColorYellow, formatDuration(elapsed), ColorReset)

	// Current crawling status
	fmt.Printf("%s🔄 Currently crawling:%s\n", ColorBold, ColorReset)
	if currentUrl != "" {
		displayUrl := currentUrl
		if len(displayUrl) > 70 {
			displayUrl = displayUrl[:67] + "..."
		}
		displayTitle := currentTitle
		if len(displayTitle) > 50 {
			displayTitle = displayTitle[:47] + "..."
		}
		fmt.Printf("   %s🌐 URL:%s %s%s%s\n", ColorBold, ColorReset, ColorCyan, displayUrl, ColorReset)
		fmt.Printf("   %s📄 Title:%s %s%s%s\n", ColorBold, ColorReset, ColorWhite, displayTitle, ColorReset)
	} else {
		fmt.Printf("   %s⏸️  Waiting for next URL...%s\n", ColorYellow, ColorReset)
	}

	// Recent crawling activity log
	fmt.Printf("\n%s📋 Recent Activity:%s\n", ColorBold, ColorReset)
	for i, logEntry := range crawlLog {
		if i < 5 { // Show last 5 entries
			fmt.Printf("   %s\n", logEntry)
		}
	}

	// Add padding lines
	fmt.Printf("\n")
}

func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-(d.Minutes()*60))
	} else {
		return fmt.Sprintf("%.0fh %.0fm", d.Hours(), d.Minutes()-(d.Hours()*60))
	}
}

// Global variables to track current crawling state
var (
	currentCrawlUrl   string
	currentCrawlTitle string
	crawlLog          []string
	crawlLogMutex     sync.Mutex
)

func addToCrawlLog(entry string) {
	crawlLogMutex.Lock()
	defer crawlLogMutex.Unlock()

	// Keep only last 10 entries
	crawlLog = append(crawlLog, entry)
	if len(crawlLog) > 10 {
		crawlLog = crawlLog[1:]
	}
}

func updateCurrentCrawl(url, title string) {
	crawlLogMutex.Lock()
	defer crawlLogMutex.Unlock()
	currentCrawlUrl = url
	currentCrawlTitle = title
}

func logCrawlResult(count int, url, title string, pageSize int64, isSuccess bool) {
	status := "✅"
	color := ColorGreen
	statusText := "SUCCESS"
	if !isSuccess {
		status = "❌"
		color = ColorRed
		statusText = "FAILED"
	}

	displayUrl := url
	displayTitle := title

	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("%s[%s] %s#%03d %s %s%s %s(%s) - %s",
		color, timestamp, status, count, statusText, displayUrl, ColorReset,
		ColorWhite, formatBytes(pageSize), displayTitle)

	addToCrawlLog(logEntry)
}

func hashUrl(url string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(url))
	return h.Sum64()
}

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			if len(a.Val) == 0 || !strings.HasPrefix(a.Val, "http") {
				ok = false
				href = a.Val
				return ok, href
			}
			href = a.Val
			ok = true
		}
	}
	return ok, href
}

func fetchPage(url string, c chan []byte) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Get(url)
	if err != nil {
		c <- []byte("")
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		c <- []byte("")
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		body = []byte("")
	}
	c <- body
}

func parseHTML(currUrl string, content []byte, q *Queue, crawled *CrawledSet, db utils.DatabaseConnection, stats *CrawlerStats) {
	updateCurrentCrawl(currUrl, "Parsing...")

	if len(content) == 0 {
		stats.updateFailure()
		logCrawlResult(crawled.size(), currUrl, "Failed to fetch", 0, false)
		updateCurrentCrawl("", "")
		return
	}

	z := html.NewTokenizer(bytes.NewReader(content))
	tokentCount := 0
	pageContentLength := 0
	body := false
	webPage := utils.WebPage{Url: currUrl, Title: "", Content: ""}

	for {
		if z.Next() == html.ErrorToken || tokentCount > 500 {
			if crawled.size() < 1000 {
				db.InsertWebPage(&webPage)
			}
			stats.updateSuccess(int64(len(content)))
			logCrawlResult(crawled.size(), currUrl, webPage.Title, int64(len(content)), true)
			updateCurrentCrawl("", "")
			return
		}

		t := z.Token()
		if t.Type == html.StartTagToken {
			if t.Data == "body" {
				body = true
			}
			if t.Data == "javascript" || t.Data == "script" || t.Data == "style" {
				z.Next()
				continue
			}
			if t.Data == "title" {
				z.Next()
				title := z.Token().Data
				webPage.Title = strings.TrimSpace(title)
				// Update current crawl status with the title
				updateCurrentCrawl(currUrl, webPage.Title)
			}
			if t.Data == "a" {
				ok, href := getHref(t)
				if !ok {
					continue
				}
				if crawled.contains(href) {
					continue
				} else {
					q.enqueue(href)
				}
			}
		}

		if body && t.Type == html.TextToken && pageContentLength < 500 {
			webPage.Content += strings.TrimSpace(t.Data)
			pageContentLength += len(t.Data)
		}
		tokentCount++
	}
}

func main() {
	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Clear screen and print header
	fmt.Print("\033[2J\033[H")
	printHeader()

	webArchiveAccess := true
	if godotenv.Load() != nil {
		fmt.Printf("%s⚠️  Warning:%s .env File Error - No Access to Web Archive\n", ColorYellow, ColorReset)
		webArchiveAccess = false
	}

	db := utils.DatabaseConnection{Access: webArchiveAccess, Uri: "", Client: nil, Collection: nil}
	db.Connect()
	defer db.Disconnect()

	crawled := CrawledSet{data: make(map[uint64]bool)}
	seed := "https://google.com"

	if len(os.Args) > 1 {
		seed = os.Args[1]
	}

	fmt.Printf("%s🌱 Starting crawl from:%s %s%s%s\n\n", ColorBold, ColorReset, ColorGreen, seed, ColorReset)

	queue := Queue{
		totalQueued: 0,
		number:      0,
		elements:    make([]string, 0),
	}

	stats := NewCrawlerStats()
	maxPages := 5000

	// Setup progress display area
	for i := 0; i < 15; i++ { // Increased for more display lines
		fmt.Println()
	}

	// Update display every second
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				stats.calculateRate()
				printProgress(crawled.size(), queue.size(), maxPages, stats, currentCrawlUrl, currentCrawlTitle, crawlLog)
			}
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-c
		fmt.Printf("\n%s🛑 Shutting down gracefully...%s\n", ColorRed, ColorReset)
		done <- true
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()

	queue.enqueue(seed)

	fetchChan := make(chan []byte, 10) // Buffered channel for better performance

	for queue.size() > 0 && crawled.size() < maxPages {
		url := queue.dequeue()
		if url == "" {
			continue
		}

		crawled.add(url)

		// Show that we're about to fetch this URL
		updateCurrentCrawl(url, "Fetching...")

		go fetchPage(url, fetchChan)
		content := <-fetchChan

		parseHTML(url, content, &queue, &crawled, db, stats)

		// Small delay to make the UI readable and prevent overwhelming
		time.Sleep(100 * time.Millisecond)
	}

	done <- true

	// Final summary
	fmt.Printf("\n%s%s╔══════════════════════════════════════════════════════════════════════════════╗%s\n", ColorBold, ColorGreen, ColorReset)
	fmt.Printf("%s%s║                              🎉 CRAWL COMPLETE                               ║%s\n", ColorBold, ColorGreen, ColorReset)
	fmt.Printf("%s%s╚══════════════════════════════════════════════════════════════════════════════╝%s\n", ColorBold, ColorGreen, ColorReset)

	successful, failed, total, rate, avgSize := stats.getStats()
	fmt.Printf("\n%s📊 Final Results:%s\n", ColorBold, ColorReset)
	fmt.Printf("   %s✅ Successfully crawled:%s %s%d pages%s\n", ColorBold, ColorReset, ColorGreen, successful, ColorReset)
	fmt.Printf("   %s❌ Failed attempts:%s %s%d pages%s\n", ColorBold, ColorReset, ColorRed, failed, ColorReset)
	fmt.Printf("   %s📈 Total processed:%s %s%d pages%s\n", ColorBold, ColorReset, ColorCyan, total, ColorReset)
	fmt.Printf("   %s⚡ Average rate:%s %s%.2f pages/sec%s\n", ColorBold, ColorReset, ColorYellow, rate, ColorReset)
	fmt.Printf("   %s📦 Average page size:%s %s%s%s\n", ColorBold, ColorReset, ColorPurple, formatBytes(avgSize), ColorReset)
	fmt.Printf("   %s📋 Remaining in queue:%s %s%d URLs%s\n", ColorBold, ColorReset, ColorWhite, queue.size(), ColorReset)
	fmt.Printf("   %s⏱️  Total runtime:%s %s%s%s\n", ColorBold, ColorReset, ColorCyan, formatDuration(time.Since(stats.startTime)), ColorReset)
}
