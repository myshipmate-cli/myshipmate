package main

// Platform represents a deployment platform
type Platform string

const (
	PlatformVercel  Platform = "vercel"
	PlatformRailway Platform = "railway"
	PlatformRender  Platform = "render"
	PlatformNetlify Platform = "netlify"
	PlatformFlyIO   Platform = "flyio"
	PlatformHeroku  Platform = "heroku"
)

// PlatformInfo holds platform-specific configuration
type PlatformInfo struct {
	Name        Platform
	DisplayName string
	URL         string
	AuthURL     string
}

// GetRecommendedPlatforms returns recommended platforms based on project type
func GetRecommendedPlatforms(projectType ProjectType) []PlatformInfo {
	switch projectType {
	case ProjectNextJS:
		return []PlatformInfo{
			{PlatformVercel, "Vercel", "https://vercel.com", "https://vercel.com/oauth/authorize"},
			{PlatformNetlify, "Netlify", "https://netlify.com", "https://app.netlify.com/authorize"},
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
		}
	case ProjectReact, ProjectVue, ProjectSvelte:
		return []PlatformInfo{
			{PlatformVercel, "Vercel", "https://vercel.com", "https://vercel.com/oauth/authorize"},
			{PlatformNetlify, "Netlify", "https://netlify.com", "https://app.netlify.com/authorize"},
		}
	case ProjectGo, ProjectNode, ProjectPython, ProjectRuby:
		return []PlatformInfo{
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
			{PlatformFlyIO, "Fly.io", "https://fly.io", "https://fly.io/oauth"},
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
		}
	case ProjectDjango, ProjectFlask, ProjectFastAPI:
		return []PlatformInfo{
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
			{PlatformFlyIO, "Fly.io", "https://fly.io", "https://fly.io/oauth"},
		}
	case ProjectRails, ProjectSinatra:
		return []PlatformInfo{
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
		}
	case ProjectJava, ProjectSpring, ProjectKotlin:
		return []PlatformInfo{
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
		}
	case ProjectRust:
		return []PlatformInfo{
			{PlatformFlyIO, "Fly.io", "https://fly.io", "https://fly.io/oauth"},
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
		}
	case ProjectPHP, ProjectLaravel, ProjectSymfony:
		return []PlatformInfo{
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
		}
	case ProjectStatic:
		return []PlatformInfo{
			{PlatformNetlify, "Netlify", "https://netlify.com", "https://app.netlify.com/authorize"},
			{PlatformVercel, "Vercel", "https://vercel.com", "https://vercel.com/oauth/authorize"},
		}
	case ProjectDocker:
		return []PlatformInfo{
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformFlyIO, "Fly.io", "https://fly.io", "https://fly.io/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
		}
	default:
		return []PlatformInfo{
			{PlatformRailway, "Railway", "https://railway.app", "https://railway.app/oauth"},
			{PlatformRender, "Render", "https://render.com", "https://render.com/oauth"},
			{PlatformHeroku, "Heroku", "https://heroku.com", "https://api.heroku.com/auth/heroku"},
		}
	}
}
