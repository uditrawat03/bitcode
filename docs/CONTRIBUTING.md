# Contributing to Bitcode

Thank you for your interest in contributing to Bitcode! This document will help you get started as a contributor.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Getting Started](#getting-started)
- [Codebase Structure & Components](#codebase-structure--components)
- [Development Workflow](#development-workflow)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Community & Support](#community--support)

---

## Project Overview

Bitcode is a terminal-based code editor written in Go. It features a sidebar file tree, text editor, status bar, and dialog system, all rendered in the terminal using the `tcell` library.

---

## Getting Started

1. **Clone the repository:**
   ```sh
   git clone https://github.com/your-org/bitcode.git
   cd bitcode
   ```

2. **Install Go (version 1.20+ recommended):**
   - [Go installation guide](https://golang.org/doc/install)

3. **Install dependencies:**
   ```sh
   go mod tidy
   ```

4. **Run the application:**
   ```sh
   go run ./cmd/main.go
   ```

---

## Codebase Structure & Components

The codebase is organized into logical packages, each with a clear responsibility:

- **`cmd/main.go`**  
  The entrypoint of the application. Sets up and starts the main app loop.

- **`internal/app/`**  
  Application initialization, shutdown, and the main event loop.  
  - *Purpose:* Manages the tcell screen, initializes UI, and runs the event loop.

- **`internal/ui/`**  
  UI orchestration and screen management.  
  - *Purpose:* The `ScreenManager` here coordinates all UI components, handles focus, and delegates events.

- **`internal/sidebar/`**  
  Sidebar (file tree) logic and rendering.  
  - *Purpose:* Displays and manages navigation of the project’s file/folder structure.

- **`internal/editor/`**  
  Text editor logic and rendering.  
  - *Purpose:* Handles text buffers, cursor movement, editing, and clipboard operations.

- **`internal/statusbar/`**  
  Status bar logic and rendering.  
  - *Purpose:* Shows status messages and context at the bottom of the screen.

- **`internal/topbar/`**  
  Top bar logic and rendering.  
  - *Purpose:* Displays static or contextual information at the top of the screen.

- **`internal/dialog/`**  
  Dialogs and modal windows.  
  - *Purpose:* Handles pop-up dialogs for file creation, deletion, and other user prompts.

- **`internal/layout/`**  
  Layout management.  
  - *Purpose:* Calculates and manages the layout and sizing of UI components based on terminal dimensions.

- **`internal/buffer/`**  
  Buffer management.  
  - *Purpose:* Manages open file buffers for the editor, including loading and saving files.

- **`internal/treeview/`**  
  File tree logic.  
  - *Purpose:* Manages the data structure for the sidebar’s file/folder tree.

## Community & Support

- For questions or help, open a [GitHub Issue](https://github.com/your-org/bitcode/issues).

---

Thank you for helping make bitcode better