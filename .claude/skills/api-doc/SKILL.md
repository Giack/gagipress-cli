---
name: api-doc
description: Generate API documentation from code for social media integrations
user-invocable: true
---

You are an API documentation generator for Gagipress CLI.

## Your Task

Generate comprehensive API documentation for Instagram and TikTok integrations.

## Process

1. **Scan Social Media Clients**
   - Read `internal/social/` directory
   - Identify API client implementations
   - Extract endpoint definitions, parameters, and responses

2. **Analyze API Patterns**
   For each API client, document:
   - **Authentication**: OAuth flow, token handling
   - **Endpoints**: URL, method, parameters
   - **Request**: Headers, body structure
   - **Response**: Success/error formats
   - **Rate Limits**: If mentioned in code
   - **Examples**: Code snippets from implementation

3. **Generate Documentation**
   Create structured documentation in `docs/api/`:
   - `instagram-api.md` - Instagram Graph API integration
   - `tiktok-api.md` - TikTok Creator API integration
   - `openapi.yaml` - OpenAPI 3.0 spec (if applicable)

4. **OpenAPI Spec Format** (if generating)
   ```yaml
   openapi: 3.0.0
   info:
     title: Gagipress Social Media API
     version: 1.0.0
     description: Instagram and TikTok integration APIs

   paths:
     /instagram/publish:
       post:
         summary: Publish Instagram Reel
         parameters: [...]
         responses: [...]
   ```

5. **Documentation Sections**
   Each API doc should include:
   - **Overview**: What the API does
   - **Authentication**: How to set up credentials
   - **Endpoints**: List of available operations
   - **Error Handling**: Common errors and solutions
   - **Examples**: Complete code examples
   - **Rate Limits**: Known limitations
   - **References**: Links to official docs

6. **Output Location**
   ```
   docs/api/
   ├── instagram-api.md
   ├── tiktok-api.md
   └── openapi.yaml (optional)
   ```

## Example Documentation Structure

```markdown
# Instagram Graph API Integration

## Overview
Gagipress uses the Instagram Graph API for publishing Reels...

## Authentication
OAuth 2.0 flow with Facebook Developer App...

## Endpoints

### POST /media
Publish a new Reel to Instagram

**Parameters:**
- `caption` (string): Post caption
- `media_url` (string): Video URL
- `hashtags` (array): List of hashtags

**Response:**
...

## Error Handling
...

## Examples
\`\`\`go
client.PublishReel(ctx, &PublishRequest{...})
\`\`\`
```

## Notes

- Focus on **practical usage** not just API specs
- Include **authentication setup** instructions
- Add **troubleshooting** section for common issues
- Link to **official documentation** for reference
- Keep **examples up-to-date** with actual code
