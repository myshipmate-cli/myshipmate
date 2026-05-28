package main

import (
	"fmt"
	"os"
)

// GenerateDockerfile creates a Dockerfile based on project type
func GenerateDockerfile(project *ProjectInfo) error {
	var dockerfile string

	switch project.Type {
	case ProjectNextJS:
		dockerfile = generateNextJSDockerfile()
	case ProjectReact, ProjectVue, ProjectAngular, ProjectSvelte, ProjectTanStack, ProjectAstro, ProjectGatsby, ProjectSolidJS, ProjectQwik:
		dockerfile = generateFrontendDockerfile(project)
	case ProjectNuxt, ProjectRemix, ProjectSvelteKit:
		dockerfile = generateSSRDockerfile(project)
	case ProjectGo:
		dockerfile = generateGoDockerfile(project)
	case ProjectNode:
		dockerfile = generateNodeDockerfile(project)
	case ProjectPython:
		dockerfile = generatePythonDockerfile(project)
	case ProjectDjango:
		dockerfile = generateDjangoDockerfile(project)
	case ProjectFlask:
		dockerfile = generateFlaskDockerfile(project)
	case ProjectFastAPI:
		dockerfile = generateFastAPIDockerfile(project)
	case ProjectRuby:
		dockerfile = generateRubyDockerfile(project)
	case ProjectRails:
		dockerfile = generateRailsDockerfile(project)
	case ProjectSinatra:
		dockerfile = generateSinatraDockerfile(project)
	case ProjectJava, ProjectKotlin:
		dockerfile = generateJavaDockerfile(project)
	case ProjectSpring:
		dockerfile = generateSpringDockerfile(project)
	case ProjectRust:
		dockerfile = generateRustDockerfile(project)
	case ProjectCSharp:
		dockerfile = generateCSharpDockerfile(project)
	case ProjectPHP:
		dockerfile = generatePHPDockerfile(project)
	case ProjectLaravel:
		dockerfile = generateLaravelDockerfile(project)
	case ProjectSymfony:
		dockerfile = generateSymfonyDockerfile(project)
	case ProjectElixir:
		dockerfile = generateElixirDockerfile(project)
	case ProjectDart:
		dockerfile = generateDartDockerfile(project)
	case ProjectStatic:
		dockerfile = generateStaticDockerfile(project)
	default:
		dockerfile = generateGenericDockerfile(project)
	}

	return os.WriteFile("Dockerfile", []byte(dockerfile), 0644)
}

// GenerateDockerIgnore creates a .dockerignore file
func GenerateDockerIgnore(project *ProjectInfo) error {
	var dockerignore string

	switch project.Type {
	case ProjectNextJS, ProjectReact, ProjectVue, ProjectAngular, ProjectSvelte, ProjectTanStack, ProjectAstro, ProjectNuxt, ProjectRemix, ProjectSvelteKit, ProjectGatsby, ProjectSolidJS, ProjectQwik, ProjectNode:
		dockerignore = `node_modules
.next
.nuxt
dist
build
.git
.gitignore
.env
.env.local
.env.production
README.md
.vscode
.idea
*.log
coverage
.nyc_output`
	case ProjectGo:
		dockerignore = `.git
.gitignore
.env
.env.local
README.md
vendor
*.exe
*.test
*.out
.vscode
.idea`
	case ProjectPython, ProjectDjango, ProjectFlask, ProjectFastAPI:
		dockerignore = `__pycache__
*.pyc
*.pyo
*.pyd
.Python
env/
venv/
.env
.env.local
.git
.gitignore
README.md
.vscode
.idea
.pytest_cache
.coverage
htmlcov`
	case ProjectRuby, ProjectRails, ProjectSinatra:
		dockerignore = `.git
.gitignore
.env
.env.local
README.md
log/*
tmp/*
vendor/bundle
.bundle
.vscode
.idea`
	case ProjectJava, ProjectKotlin, ProjectSpring:
		dockerignore = `.git
.gitignore
.env
README.md
target/
build/
.gradle/
*.class
*.jar
*.war
.vscode
.idea`
	case ProjectRust:
		dockerignore = `.git
.gitignore
.env
README.md
target/
Cargo.lock
.vscode
.idea`
	default:
		dockerignore = `.git
.gitignore
.env
.env.local
README.md
.vscode
.idea`
	}

	return os.WriteFile(".dockerignore", []byte(dockerignore), 0644)
}

// Dockerfile generators for each project type

func generateNextJSDockerfile() string {
	return `# Next.js Dockerfile
FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else npm install; \
  fi

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN \
  if [ -f yarn.lock ]; then yarn run build; \
  elif [ -f package-lock.json ]; then npm run build; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm run build; \
  else npm run build; \
  fi

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public

# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]
`
}

func generateFrontendDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# %s Dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else npm install; \
  fi

COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
`, project.Type)
}

func generateSSRDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# %s SSR Dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else npm install; \
  fi

COPY . .
RUN npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json ./package.json

ENV NODE_ENV=production
EXPOSE %s

CMD ["node", ".output/server/index.mjs"]
`, project.Type, project.Port)
}

func generateGoDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Go Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE %s

CMD ["./server"]
`, project.Port)
}

func generateNodeDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Node.js Dockerfile
FROM node:20-alpine

WORKDIR /app

COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then corepack enable pnpm && pnpm i --frozen-lockfile; \
  else npm install; \
  fi

COPY . .

EXPOSE %s

CMD ["node", "index.js"]
`, project.Port)
}

func generatePythonDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Python Dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE %s

CMD ["python", "app.py"]
`, project.Port)
}

func generateDjangoDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Django Dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt gunicorn

COPY . .

# Collect static files
RUN python manage.py collectstatic --noinput || true

EXPOSE %s

CMD ["gunicorn", "--bind", "0.0.0.0:%s", "config.wsgi:application"]
`, project.Port, project.Port)
}

func generateFlaskDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Flask Dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt gunicorn

COPY . .

EXPOSE %s

CMD ["gunicorn", "--bind", "0.0.0.0:%s", "app:app"]
`, project.Port, project.Port)
}

func generateFastAPIDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# FastAPI Dockerfile
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE %s

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "%s"]
`, project.Port, project.Port)
}

func generateRubyDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Ruby Dockerfile
FROM ruby:3.2-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .

EXPOSE %s

CMD ["ruby", "app.rb"]
`, project.Port)
}

func generateRailsDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Rails Dockerfile
FROM ruby:3.2-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    libpq-dev \
    nodejs \
    && rm -rf /var/lib/apt/lists/*

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .

# Precompile assets
RUN bundle exec rails assets:precompile || true

EXPOSE %s

CMD ["bundle", "exec", "rails", "server", "-b", "0.0.0.0"]
`, project.Port)
}

func generateSinatraDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Sinatra Dockerfile
FROM ruby:3.2-slim

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY . .

EXPOSE %s

CMD ["ruby", "app.rb", "-o", "0.0.0.0"]
`, project.Port)
}

func generateJavaDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Java Dockerfile
FROM maven:3.9-eclipse-temurin-17 AS builder

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline

COPY . .
RUN mvn clean package -DskipTests

# Production stage
FROM eclipse-temurin:17-jre-alpine

WORKDIR /app

COPY --from=builder /app/target/*.jar app.jar

EXPOSE %s

CMD ["java", "-jar", "app.jar"]
`, project.Port)
}

func generateSpringDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Spring Boot Dockerfile
FROM maven:3.9-eclipse-temurin-17 AS builder

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline

COPY . .
RUN mvn clean package -DskipTests

# Production stage
FROM eclipse-temurin:17-jre-alpine

WORKDIR /app

COPY --from=builder /app/target/*.jar app.jar

EXPOSE %s

ENTRYPOINT ["java", "-jar", "app.jar"]
`, project.Port)
}

func generateRustDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Rust Dockerfile
FROM rust:1.75 AS builder

WORKDIR /app

COPY . .

RUN cargo build --release && \
    cp "$(find target/release -maxdepth 1 -type f -executable | head -n1)" /tmp/server

# Production stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /tmp/server ./server

EXPOSE %s

CMD ["./server"]
`, project.Port)
}

func generateCSharpDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# .NET Dockerfile
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS builder

WORKDIR /app

COPY *.csproj .
RUN dotnet restore

COPY . .
RUN dotnet publish -c Release -o out

# Production stage
FROM mcr.microsoft.com/dotnet/aspnet:8.0

WORKDIR /app

COPY --from=builder /app/out .

EXPOSE %s

ENTRYPOINT ["dotnet", "app.dll"]
`, project.Port)
}

func generatePHPDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# PHP Dockerfile
FROM php:8.2-apache

WORKDIR /var/www/html

COPY . .

EXPOSE %s

CMD ["apache2-foreground"]
`, project.Port)
}

func generateLaravelDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Laravel Dockerfile
FROM php:8.2-fpm

WORKDIR /var/www

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    libpng-dev \
    libonig-dev \
    libxml2-dev \
    zip \
    unzip \
    && docker-php-ext-install pdo_mysql mbstring exif pcntl bcmath gd

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

COPY . .

RUN composer install --no-dev --optimize-autoloader

EXPOSE %s

CMD ["php", "artisan", "serve", "--host=0.0.0.0", "--port=%s"]
`, project.Port, project.Port)
}

func generateSymfonyDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Symfony Dockerfile
FROM php:8.2-fpm

WORKDIR /var/www

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    libpng-dev \
    libonig-dev \
    libxml2-dev \
    zip \
    unzip \
    && docker-php-ext-install pdo_mysql mbstring exif pcntl bcmath gd

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

COPY . .

RUN composer install --no-dev --optimize-autoloader

EXPOSE %s

CMD ["php", "-S", "0.0.0.0:%s", "-t", "public"]
`, project.Port, project.Port)
}

func generateElixirDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Elixir Dockerfile
FROM elixir:1.16-alpine

WORKDIR /app

RUN apk add --no-cache build-base git && \
    mix local.hex --force && \
    mix local.rebar --force

ENV MIX_ENV=prod

COPY mix.exs mix.lock ./
RUN mix deps.get --only prod && mix deps.compile

COPY . .
RUN mix compile

EXPOSE %s

CMD ["mix", "phx.server"]
`, project.Port)
}

func generateDartDockerfile(project *ProjectInfo) string {
	return fmt.Sprintf(`# Dart/Flutter Dockerfile
FROM debian:latest AS builder

RUN apt-get update && apt-get install -y \
    curl \
    git \
    unzip \
    xz-utils \
    zip \
    libglu1-mesa

RUN git clone https://github.com/flutter/flutter.git /usr/local/flutter
ENV PATH="/usr/local/flutter/bin:/usr/local/flutter/bin/cache/dart-sdk/bin:${PATH}"

RUN flutter doctor

WORKDIR /app
COPY . .

RUN flutter pub get
RUN flutter build web

# Production stage
FROM nginx:alpine

COPY --from=builder /app/build/web /usr/share/nginx/html

EXPOSE %s

CMD ["nginx", "-g", "daemon off;"]
`, project.Port)
}

func generateStaticDockerfile(project *ProjectInfo) string {
	return `# Static Site Dockerfile
FROM nginx:alpine

COPY . /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
`
}

func generateGenericDockerfile(project *ProjectInfo) string {
	return `# Generic Dockerfile
FROM alpine:latest

WORKDIR /app

COPY . .

EXPOSE 8080

CMD ["./app"]
`
}
