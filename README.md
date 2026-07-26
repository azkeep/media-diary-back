# Media Diary Monorepo

This repository contains the full stack for the Media Diary application, organized as a monorepo for easier cross-service development.

## 📂 Project Structure

| Folder | Technology | Description | IDE |
| :--- | :--- | :--- | :--- |
| **`backend-java`** | Java / Spring Boot | Primary API and Business Logic | IntelliJ IDEA |
| **`frontend-react`** | React / JavaScript | Web User Interface | WebStorm |
| **`service-golang`** | Go (Golang) | High-performance Microservice | GoLand |

---

## 🚀 Quick Start (Development)

To work on a specific part of the project, open the corresponding subfolder in its dedicated IDE.

### 1. Backend (Java)
- **Path:** `/backend-java`
- **Setup:** Ensure JDK 17+ is installed. Import as a Maven/Gradle project.
- **Port:** `8080` (Standard)

### 2. Frontend (React)
- **Path:** `/frontend-react`
- **Setup:** Run `npm install` inside the folder.
- **Start:** `npm start`
- **Port:** `3000`

### 3. Microservice (Go)
- **Path:** `/service-golang`
- **Setup:** Run `go mod tidy` to install dependencies.
- **Port:** `8081` (Suggested)

---

## 🛠 Shared Commands (Root)

While each project has its own scripts, use these at the root for repository maintenance:

- **Check status of all apps:** `git status`
- **Update everything:** `git pull origin main`
- **Save all changes:** `git add . && git commit -m "Your message"`