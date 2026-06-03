# Guna Men's Salon & Beauty Parlour | Full-Stack Application

An elegant, premium full-stack web application for **Guna Men's Salon & Beauty Parlour**, featuring a Go backend (REST API + static file embedding) and a luxury-themed Vanilla CSS & JS frontend. Data is stored and managed using MongoDB.

---

## Key Features

1. **Luxury Grooming Experience landing page**: Curated animations, premium typography, responsive layout, and beautiful imagery.
2. **Dynamic Appointment Booking**: Customers can choose services, select dates/times, add styling instructions, and book appointments instantly.
3. **Client Testimonials Carousel**: Customers can submit reviews and ratings directly from the UI.
4. **Secure Admin Dashboard**: Salon managers can view all bookings, modify booking statuses (Confirm, Complete, Cancel), view statistics, and approve/reject client reviews before they go live.
5. **No CGO dependencies**: The backend uses pure-Go libraries making it fully cross-platform and compile-ready for both Windows and Linux environments.

---

## Tech Stack

- **Backend**: Go (Standard library router, `net/http` router with path parameters, `embed` for serving static assets).
- **Database**: MongoDB (Official Go driver).
- **Frontend**: Vanilla HTML5, CSS3 (glassmorphic styling, charcoal/metallic gold color palette), and Javascript.

---

## Local Development Setup

### 1. Prerequisites
- **Go** (v1.22+ recommended for standard router parameters)
- **MongoDB** (running locally on `localhost:27017` or Atlas cluster)

### 2. Configuration
Copy `.env.example` to `.env` and fill in your details:
```bash
PORT=8080
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=saloon_db
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

### 3. Run the application
```bash
go run main.go
```
The server will start at:
- **Client App**: [http://localhost:8080](http://localhost:8080)
- **Admin Dashboard**: [http://localhost:8080/admin.html](http://localhost:8080/admin.html)

---

## Deploying to Render

This project is fully compatible with Render's native Go environments.

### 1. Create a Web Service
Create a new **Web Service** on Render and link your GitHub repository.

### 2. Configuration Settings
Set the following properties during creation:
- **Runtime**: `Go`
- **Build Command**: `go build -o bin/saloon main.go`
- **Start Command**: `./bin/saloon`

### 3. Environment Variables
Add the following key-value pairs in the **Environment** tab:
- `MONGODB_URI`: `<your_mongodb_connection_string>` (e.g. MongoDB Atlas cluster link)
- `MONGODB_DB`: `saloon_db`
- `ADMIN_USERNAME`: `admin`
- `ADMIN_PASSWORD`: `<your_secure_password>`

Render will automatically inject a `PORT` variable. The Go backend reads this variable and binds to it, making deployment fully automated!
