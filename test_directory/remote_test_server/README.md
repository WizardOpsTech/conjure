# Conjure Remote Test Server

This directory contains a self-contained Docker-based nginx web server for testing Conjure's remote source functionality locally. The server clones the `conjure-get-started` repository from GitHub during the build process, making it fully portable and ideal for demos.

## Purpose

Test the remote source feature of Conjure CLI using a local HTTP server. This server:

- Clones the official `conjure-get-started` repository from GitHub
- Serves templates and bundles via HTTP at `http://localhost:8080`
- Provides a local testing environment without external dependencies
- Is fully portable and can be distributed for demos

This allows you to test:

- Remote template/bundle discovery via `index.json`
- Remote file downloading
- SHA256 verification
- Caching behavior
- Multi-source resolution (local + remote)
- Remote-first vs local-first priority
