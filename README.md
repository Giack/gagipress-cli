# Gagipress CLI

🚀 Powerful social media automation tool for Amazon KDP publishers

## Overview

Gagipress CLI automates content generation, scheduling, and analytics for TikTok and Instagram Reels, specifically designed for self-publishers on Amazon KDP.

### Features

- ✅ **AI-Powered Content Generation**: Generate 7-10+ social media scripts per week using OpenAI and Gemini
- ✅ **Intelligent Scheduling**: Smart weekly planning with peak time optimization
- ✅ **Automated Publishing**: Cron-based publishing to Instagram and TikTok
- ✅ **Performance Analytics**: Track engagement and correlate with KDP sales
- ✅ **Self-Hosted**: Full control with minimal recurring costs

## Quick Start

### Prerequisites

- Go 1.24+ ([Download](https://go.dev/dl/))
- Supabase CLI ([Install](https://supabase.com/docs/guides/cli/getting-started))
- Supabase account ([Sign up](https://supabase.com))
- OpenAI API key ([Get one](https://platform.openai.com))

### Installation

```bash
# Clone repository
git clone https://github.com/gagipress/gagipress-cli
cd gagipress-cli

# Build
go build -o gagipress

# Install globally (optional)
sudo mv gagipress /usr/local/bin/
```

### Setup

```bash
# Initialize configuration
gagipress init

# Create database schema
gagipress db migrate

# Add your first book
gagipress books add

# Generate content ideas
gagipress generate ideas
```

## Configuration

Configuration is stored in `~/.gagipress/config.yaml`

Required:
- Supabase URL and keys
- OpenAI API key

Optional:
- Instagram access token
- TikTok access token
- Amazon KDP credentials

### Troubleshooting

**Commands fail with "Run 'gagipress init' first" even after running init:**

This was a bug in earlier versions where the config file was created with incorrect field names.

Fix:
```bash
gagipress fix-config
```

This will migrate your config file to the correct format. You only need to run this once.

## Commands

### Content Generation

```bash
# Generate content ideas (20 by default)
gagipress generate ideas
gagipress generate ideas --count 30
gagipress generate ideas --book <book-id>
gagipress generate ideas --gemini  # Force Gemini usage

# List generated ideas
gagipress ideas list
gagipress ideas list --status pending
gagipress ideas list --status approved --limit 10

# Approve/reject ideas
gagipress ideas approve <idea-id>
gagipress ideas reject <idea-id>

# Generate script from approved idea
gagipress generate script <idea-id>
gagipress generate script <idea-id> --platform instagram
gagipress generate script <idea-id> --gemini
```

### Scheduling

```bash
# Create intelligent weekly plan
gagipress calendar plan

# Approve/modify scheduled content
gagipress calendar approve

# View calendar
gagipress calendar show

# Force publish immediately
gagipress calendar publish <id>
```

### Analytics

```bash
# View performance dashboard
gagipress stats show
gagipress stats show --period 7d
gagipress stats show --period 30d
gagipress stats show --platform instagram

# Analyze social → sales correlation
gagipress stats correlate --book <book-id>
gagipress stats correlate --book <book-id> --days 60
```

### Book Management

```bash
# Add book to catalog
gagipress books add

# List all books
gagipress books list

# Edit/delete books
gagipress books edit <book-id>
gagipress books delete <book-id>

# Import sales data from Amazon KDP
gagipress books sales import <csv-file>
gagipress books sales show <book-id>
```

### Database Management

```bash
# Check database connection and schema version
gagipress db status

# Apply pending migrations
gagipress db migrate
```

### API Testing

```bash
# Test OpenAI API connection
gagipress auth openai

# Test Instagram API (requires OAuth setup)
gagipress auth instagram

# Test TikTok API (requires OAuth setup)
gagipress auth tiktok

# Test Gemini browser automation
gagipress test gemini "Write a short story"
gagipress test gemini --headless=false "Ciao!"
```

## Architecture

**Tech Stack:**
- **CLI**: Go with Cobra framework
- **Database**: Supabase (PostgreSQL)
- **AI**: OpenAI API + Gemini (browser automation)
- **Social APIs**: Instagram Graph API, TikTok Creator API
- **Automation**: Supabase Edge Functions (cron jobs)

**Deployment:**
- CLI runs locally or on VPS
- Database and cron jobs on Supabase (serverless)
- Zero downtime with automatic scaling

## Development

### Project Structure

```
gagipress-cli/
├── cmd/                    # CLI commands
│   ├── root.go
│   ├── init.go
│   ├── generate/
│   ├── calendar/
│   └── stats/
├── internal/               # Internal packages
│   ├── config/            # Configuration management
│   ├── supabase/          # Supabase client
│   ├── ai/                # OpenAI & Gemini
│   ├── social/            # Instagram & TikTok APIs
│   ├── models/            # Data models
│   ├── repository/        # Database operations
│   ├── generator/         # Content generation logic
│   ├── scheduler/         # Scheduling algorithms
│   └── analytics/         # Analytics & correlation
├── supabase/
│   └── functions/         # Edge Functions
├── migrations/            # Database migrations
├── templates/             # Prompt templates
└── docs/                  # Documentation
```

### Running Tests

```bash
# Run all tests
mise exec -- go test ./...

# Run with coverage
./scripts/test-coverage.sh

# Run only unit tests
mise exec -- go test ./internal/...

# Run only integration tests (requires credentials)
SUPABASE_URL=xxx SUPABASE_KEY=xxx mise exec -- go test ./test/integration/... -v

# Run specific package tests
mise exec -- go test ./internal/models/... -v
mise exec -- go test ./internal/parser/... -v
```

**Test Coverage:**
- Parser: ~97%
- Error Handling: ~83%
- Scheduler: ~52%
- Models: ~37%
- Overall: ~40%

### Building

```bash
# Build for current platform
go build -o gagipress

# Cross-compile for all platforms
make build-all

# Build and install
make install
```

## Documentation

- [Design Document](docs/plans/2026-02-08-gagipress-social-automation-design.md)
- [Implementation Plan](docs/plans/2026-02-08-implementation-plan.md)
- [User Guide](docs/USER_GUIDE.md) _(coming soon)_
- [API Setup Guide](docs/API_SETUP.md) _(coming soon)_

## Roadmap

### Phase 1: MVP (Current - Week 4 Complete!)
- [x] CLI foundation
- [x] Content generation
- [x] Scheduling & approval workflow
- [x] Analytics & correlation
- [x] Amazon KDP sales import
- [ ] Automated publishing (OAuth setup required)
- [ ] Automated metrics collection

### Phase 2: Video Automation
- [ ] Template-based video generation
- [ ] FFmpeg compositing
- [ ] Text-to-speech voiceover
- [ ] Auto-subtitles

### Phase 3: Interaction Automation
- [ ] Auto-reply to comments
- [ ] DM automation
- [ ] Community management

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- 📧 Email: support@gagipress.com
- 🐛 Issues: [GitHub Issues](https://github.com/gagipress/gagipress-cli/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/gagipress/gagipress-cli/discussions)

---

Built with ❤️ for independent publishers

**Current Version**: v0.1.0 (MVP Development)
