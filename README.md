# 🕷️ Advanced Web Crawler

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Cross--Platform-lightgrey?style=for-the-badge)

**A high-performance, feature-rich web crawler built in Go with real-time CLI visualization**

[Features](#features) • [Installation](#installation) • [Usage](#usage) • [Configuration](#configuration) • [Contributing](#contributing)

![Demo GIF Placeholder](./samples/web-crawler-demo.gif)

</div>

---

## ✨ Features

### 🚀 **Core Capabilities**

- **High-Performance Crawling** - Concurrent processing with goroutines
- **Smart URL Management** - Intelligent queuing and duplicate detection
- **Database Integration** - MongoDB support for data persistence
- **Real-time Statistics** - Live performance metrics and analytics
- **Graceful Shutdown** - Clean exit with Ctrl+C handling

### 🎨 **Beautiful CLI Interface**

- **Live Progress Bars** - Visual progress tracking
- **Real-time Activity Log** - See what's being crawled instantly
- **Color-coded Output** - Enhanced readability with ANSI colors
- **Performance Metrics** - Pages per second, success rates, and more
- **Current Status Display** - Know exactly what's happening

### 🛡️ **Robust & Reliable**

- **Error Handling** - Comprehensive error tracking and recovery
- **HTTP Timeouts** - Prevents hanging on slow requests
- **Memory Efficient** - Optimized data structures and algorithms
- **Thread Safe** - Concurrent-safe operations throughout

### 📊 **Advanced Analytics**

- **Success/Failure Tracking** - Detailed crawl statistics
- **Performance Monitoring** - Real-time rate calculations
- **Data Size Analysis** - Track bandwidth usage
- **Runtime Metrics** - Complete session analytics

---

## 🚀 Installation

### Prerequisites

- **Go 1.19+** installed on your system
- **MongoDB** (optional, for data persistence)
- **Git** for cloning the repository

### Quick Start

```bash
# Clone the repository
git clone https://github.com/Yashh56/web-crawler.git
cd web-crawler

# Install dependencies
go mod tidy

# Build the application
go build -o crawler main.go

# Run the crawler
./crawler https://example.com
```

### Using Go Install

```bash
go install github.com/Yashh56/web-crawler@latest
```

---

## 🎯 Usage

### Basic Usage

```bash
# Crawl starting from Google
./crawler https://google.com

# Crawl with default seed URL
./crawler
```

### Environment Setup

Create a `.env` file in the project root:

```env
# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
```

### Command Line Options

```bash
# Start crawling from a specific URL
./crawler https://example.com

# The crawler will automatically:
# ✅ Fetch and parse HTML content
# ✅ Extract links and queue them
# ✅ Store page data in MongoDB
# ✅ Display real-time progress
# ✅ Handle errors gracefully
```

---

## 📸 Screenshots

### Live Crawling Interface

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                            🕷️  WEB CRAWLER STATUS                            ║
╚══════════════════════════════════════════════════════════════════════════════╝

📊 Progress: [████████████████████████████░░░░░░░░░░░░░░░░░░░░░░] 65.4% (327/500)
🎯 Queue Size: 1,247     ⚡ Rate: 12.3 pages/sec
✅ Successful: 315       ❌ Failed: 12
📈 Total Processed: 327   📦 Avg Size: 45.2 KB
⏱️  Runtime: 2m 34s

🔄 Currently crawling:
   🌐 URL: https://github.com/trending
   📄 Title: Trending repositories on GitHub today

📋 Recent Activity:
   [14:23:45] ✅ #327 SUCCESS https://github.com/trending... (67.8 KB) - Trending
   [14:23:44] ✅ #326 SUCCESS https://stackoverflow.com... (123 KB) - Stack Overflow
   [14:23:43] ❌ #325 FAILED https://timeout-site.com... (0 B) - Request timeout
   [14:23:42] ✅ #324 SUCCESS https://reddit.com/r/golang... (89.4 KB) - r/golang
   [14:23:41] ✅ #323 SUCCESS https://news.ycombinator.com... (34.1 KB) - Hacker News
```

### Final Results Summary

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                              🎉 CRAWL COMPLETE                               ║
╚══════════════════════════════════════════════════════════════════════════════╝

📊 Final Results:
   ✅ Successfully crawled: 4,847 pages
   ❌ Failed attempts: 153 pages
   📈 Total processed: 5,000 pages
   ⚡ Average rate: 15.7 pages/sec
   📦 Average page size: 52.3 KB
   📋 Remaining in queue: 2,847 URLs
   ⏱️  Total runtime: 5m 18s
```

---

## Configuration ⚙️

### Database Configuration

The crawler supports MongoDB for storing crawled data:

```go
type WebPage struct {
    URL     string    `bson:"url"`
    Title   string    `bson:"title"`
    Content string    `bson:"content"`
}
```

### Crawler Settings

Modify these constants in `main.go` to customize behavior:

```go
const (
    MaxPages     = 5000              // Maximum pages to crawl
    CrawlDelay   = 100 * time.Millisecond // Delay between requests
    HTTPTimeout  = 10 * time.Second  // HTTP request timeout
    MaxTokens    = 500               // HTML tokens to process per page
    ContentLimit = 500               // Max content length to extract
)
```

---

## 🏗️ Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   URL Queue     │    │  Crawled Set    │    │  Database       │
│                 │    │                 │    │                 │
│ • Thread-safe   │    │ • Hash-based    │    │ • MongoDB       │
│ • FIFO ordering │    │ • Duplicate     │    │ • Async writes  │
│ • Auto-scaling  │    │   detection     │    │ • Error handling│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Crawler Engine │
                    │                 │
                    │ • Goroutine pool│
                    │ • Rate limiting │
                    │ • Error recovery│
                    │ • Progress track│
                    └─────────────────┘
```

### Data Flow

1. **URL Discovery** - Extract links from HTML
2. **Queue Management** - Add unique URLs to crawl queue
3. **Concurrent Fetching** - Download pages with HTTP client
4. **HTML Parsing** - Extract content and metadata
5. **Data Storage** - Persist to MongoDB
6. **Progress Tracking** - Update real-time statistics

---

## 🔧 Development

### Project Structure

```
web-crawler/
├── main.go                 # Main application entry
├── utils/
│   ├── database.go        # Database
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
├── .env.example           # Environment template
├── README.md              # This file
└── LICENSE                # MIT License
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/Yashh56/web-crawler.git
cd web-crawler

# Install dependencies
go mod download

# Build optimized binary
go build -ldflags="-s -w" -o crawler main.go
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Generate documentation
go doc -all
```

---

## 📈 Performance

### Benchmarks

| Metric            | Value                  |
| ----------------- | ---------------------- |
| **Average Speed** | 10-15 pages/second     |
| **Memory Usage**  | ~50MB for 10K pages    |
| **CPU Usage**     | 2-5% on modern systems |
| **Success Rate**  | 95%+ on healthy sites  |

### Optimization Tips

- **Adjust `CrawlDelay`** - Balance speed vs. server load
- **Monitor Queue Size** - Prevent memory overuse
- **Database Batching** - Group writes for better performance
- **Filter URLs** - Skip non-HTML content types

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### Ways to Contribute

- 🐛 **Bug Reports** - Found an issue? Let us know!
- 💡 **Feature Requests** - Have ideas? We'd love to hear them!
- 📖 **Documentation** - Help improve our docs
- 🔧 **Code Contributions** - Submit pull requests

### Development Setup

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Commit** your changes: `git commit -m 'Add amazing feature'`
4. **Push** to the branch: `git push origin feature/amazing-feature`
5. **Open** a Pull Request

### Code Style Guidelines

- Follow **Go conventions** and `gofmt` formatting
- Add **comments** for public functions
- Include **tests** for new features
- Update **documentation** as needed

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2025 Yash

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

---

## 🌟 Acknowledgments

- **Go Community** - For the amazing language and ecosystem
- **MongoDB** - For robust data storage capabilities
- **Contributors** - Everyone who helped improve this project
- **Open Source** - For inspiring collaborative development

---

<div align="center">

**⭐ Star this repository if you found it helpful!**

Made with ❤️ by [Yash](https://github.com/Yashh56)

[🔝 Back to Top](#-advanced-web-crawler)

</div>
